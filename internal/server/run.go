package server

import (
	"context"
	"log/slog"
	"time"

	"github.com/kkapel/gophkeeper/internal/db/connections"
	"github.com/kkapel/gophkeeper/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func Run() error {
	// Логгер
	if err := logger.Initialize("INFO"); err != nil {
		return err
	}

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
