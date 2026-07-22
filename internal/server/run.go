package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kkapel/gophkeeper/internal/config"
	"github.com/kkapel/gophkeeper/internal/db/connections"
	"github.com/kkapel/gophkeeper/internal/db/sqlc"
	"github.com/kkapel/gophkeeper/internal/handlers"
	"github.com/kkapel/gophkeeper/internal/logger"
	"github.com/kkapel/gophkeeper/internal/service"
	"github.com/kkapel/gophkeeper/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	pb "github.com/kkapel/gophkeeper/internal/proto/gophkeeper/v1"
)

// Run запускает grpc-сервер и инциализирует конфиги и подключения
func Run() error {
	// Инициализация конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	// Логгер
	if err := logger.Initialize(cfg.LoggerLevel); err != nil {
		return err
	}

	logger.Log.Info("Logger initialized")
	logger.Log.Info("Starting gRPC server...")

	// Инициализация БД
	db, err := connections.InitDB(cfg.DataBaseURL, "migrations")
	if err != nil {
		return err
	}
	// Закрываем БД-соединение
	if db != nil {
		defer func() { _ = db.Close() }()
	}

	// Инициализация сервисов
	queries := sqlc.New(db.GetSqlDb())
	itemStorage := storage.NewItemStorage(queries)
	svc := service.NewAuthService(queries, cfg.JWTSecret)
	keepSvc := service.NewKeeperService(itemStorage)

	// Инициализация хендлеров gRPC-сервиса
	authHandler := handlers.NewAuthHandler(svc)
	keepHandler := handlers.NewKeeperHandler(keepSvc)

	// Запуск gRPC-сервера
	srv, err := startGRPCServer(cfg, authHandler, keepHandler)
	if err != nil {
		return err
	}

	// ждём сигнал остановки (Ctrl+C / SIGTERM)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	logger.Log.Info("Shutting down gRPC server")

	srv.GracefulStop() // Останавливаем сервер при завершении работы

	return nil
}

// LoggingInterceptor добавляет логирование в поступающие вызовы
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
func startGRPCServer(cfg *config.Config, authHandler *handlers.AuthHandler, keepHandler *handlers.KeeperHandler) (*grpc.Server, error) {
	// 1. Открываем listener на нужном порту
	listener, err := net.Listen("tcp", cfg.GRPCAddress)
	if err != nil {
		return nil, err
	}

	// Загружаем TLS-сертификаты
	creds, err := credentials.NewServerTLSFromFile(cfg.TLSCertPath, cfg.TLSKeyPath)
	if err != nil {
		return nil, fmt.Errorf("load TLS credentials: %w", err)
	}

	publicMethods := map[string]bool{
		"/gophkeeper.v1.AuthService/Register": true,
		"/gophkeeper.v1.AuthService/Login":    true,
	}

	authInterceptor := NewAuthInterceptor(cfg.JWTSecret, publicMethods)

	// 2. Создаём экземпляр gRPC-сервера
	grpcServer := grpc.NewServer(
		grpc.Creds(creds), // добавляем TLS
		grpc.ChainUnaryInterceptor(
			LoggingInterceptor, authInterceptor.Unary), // добавляем логирование и аутентификацию
	)

	// 3. Регистрируем его в gRPC-сервере
	pb.RegisterAuthServiceServer(grpcServer, authHandler)
	pb.RegisterKeeperServiceServer(grpcServer, keepHandler)

	// 4. Запускаем в горутине, чтобы не блокировать main
	go func() {
		logger.Log.Info("Grpc server starts")
		if err := grpcServer.Serve(listener); err != nil {
			logger.Log.Error("ошибка работы gRPC-сервера", slog.Any("error", err))
		}
	}()

	return grpcServer, nil
}
