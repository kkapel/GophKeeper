// Package handlers содержит gRPC-хендлеры сервера GophKeeper.
package handlers

import (
	"context"
	"errors"

	pb "github.com/kkapel/gophkeeper/internal/proto/gophkeeper/v1"
	"github.com/kkapel/gophkeeper/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthServiceAPI описывает то, что хендлеру нужно от слоя сервиса.
type AuthServiceAPI interface {
	Register(ctx context.Context, login, password string) (string, error)
	Login(ctx context.Context, login, password string) (string, error)
}

// AuthHandler реализует gRPC-сервис аутентификации.
type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	service AuthServiceAPI
}

// NewAuthHandler создаёт хендлер аутентификации.
func NewAuthHandler(svc AuthServiceAPI) *AuthHandler {
	return &AuthHandler{service: svc}
}

// Register регистрирует нового пользователя и возвращает токен доступа.
func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
	login := req.GetLogin()
	password := req.GetPassword()

	if login == "" || password == "" {
		return nil, status.Error(codes.InvalidArgument, "login and password are required")
	}

	token, err := h.service.Register(ctx, login, password)
	if err != nil {
		if errors.Is(err, service.ErrLoginTaken) {
			return nil, status.Error(codes.AlreadyExists, "login already taken")
		}
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return pb.AuthResponse_builder{
		AccessToken: &token,
	}.Build(), nil
}

// Login аутентифицирует пользователя и возвращает токен доступа.
func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	login := req.GetLogin()
	password := req.GetPassword()

	// Проверка на пустые поля
	if login == "" || password == "" {
		return nil, status.Error(codes.InvalidArgument, "login and password are required")
	}

	// Вызов сервиса для аутентификации
	token, err := h.service.Login(ctx, login, password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid login or password")
		}
		return nil, status.Error(codes.Internal, "failed to authenticate user")
	}

	return pb.AuthResponse_builder{
		AccessToken: &token,
	}.Build(), nil
}
