package server

import (
	"context"
	"log/slog"
	"net"
	"time"

	"github.com/docker/docker/libnetwork/config"
	"github.com/kkapel/gophkeeper/internal/db/connections"
	"github.com/kkapel/gophkeeper/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func Run() error {
	// Логгер
	if err := logger.Initialize("INFO"); err != nil {
		return err
	}

	logger.Log.Info("Logger initialized")
	logger.Log.Info("Starting gRPC server...")

	// Инициализация БД
	var dbstr string = ""
	db, err := connections.InitDB(dbstr, "./migrations")
	if err != nil {
		return err
	}

	// Инициализация сервиса
	//gophKeeper := &GophKeeper{
	//	db: db,
	//	}
	// Закрываем БД-соединение
	if db != nil {
		defer func() { _ = db.Close() }()
	}

	// Запуск gRPC-сервера
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(LoggingInterceptor), // добавляем логирование
	)

	grpcServer.Serve(nil) // Здесь нужно передать net.Listener, но для примера оставим nil

	return nil
}

func LoggingInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()

	// вызываем сам обработчик (аналог h.ServeHTTP)
	resp, err := handler(ctx, req)

	duration := time.Since(start)
	code := status.Code(err) // gRPC-код из ошибки (OK, если err == nil)

	logger.Log.Info("incoming gRPC request",
		slog.String("method", info.FullMethod), // аналог r.URL.Path
		slog.Duration("duration", duration),
		slog.String("code", code.String()), // аналог статуса ответа
	)

	return resp, err
}

// Функция для запуска grpc-сервера
func startGRPCServer(svc *service.ShortenerService, cfg *config.Config) (*grpc.Server, error) {
	// 1. Открываем listener на нужном порту
	listener, err := net.Listen("tcp", cfg.Grpc)
	if err != nil {
		return nil, err
	}

	// Загружаем TLS-сертификаты
	// to do : сгенерировать сертификаты
	// пути сделать настраиваемыми через конфиг
	creds, err := credentials.NewServerTLSFromFile("cert.pem", "key.pem")
	if err != nil {
		return nil, err
	}

	// 2. Создаём экземпляр gRPC-сервера
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(LoggingInterceptor), // добавляем логирование
		grpc.Creds(creds), // добавляем TLS
	)

	// 3. Создаём свой сервер с бизнес-логикой
	srv := serverGRPC.NewShortenerGRPCServer(svc, cfg)

	// 4. Регистрируем его в gRPC-сервере
	pb.RegisterShortenerServiceServer(grpcServer, srv)

	// 5. Запускаем в горутине, чтобы не блокировать main
	go func() {
		logger.Log.Info("Grpc server starts")
		if err := grpcServer.Serve(listener); err != nil {
			logger.Log.Error("ошибка работы gRPC-сервера", slog.Any("error", err))
		}
	}()

	return grpcServer, nil
}
