package server

import (
	"context"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/kkapel/gophkeeper/internal/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor — структура для перехвата gRPC-запросов и проверки JWT-токена.
type AuthInterceptor struct {
	jwtSecret     string
	publicMethods map[string]bool
}

// NewAuthInterceptor создаёт новый экземпляр AuthInterceptor с заданным секретом JWT.
func NewAuthInterceptor(jwtSecret string, publicMethods map[string]bool) *AuthInterceptor {
	return &AuthInterceptor{
		jwtSecret:     jwtSecret,
		publicMethods: publicMethods,
	}
}

// Unary возвращает gRPC хэнделер для перехвата unary-запросов и проверки JWT-токена.
func (a *AuthInterceptor) Unary(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {

	if a.publicMethods[info.FullMethod] {
		// Если метод публичный, пропускаем проверку токена
		return handler(ctx, req)
	}

	userID, err := a.authorize(ctx)
	if err != nil {
		return nil, err
	}

	// Добавляем UUID пользователя в контекст
	ctx = auth.WithUserID(ctx, userID)

	return handler(ctx, req)

}

// authorize проверяет JWT-токен из контекста и возвращает UUID пользователя.
func (a *AuthInterceptor) authorize(ctx context.Context) (uuid.UUID, error) {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuid.Nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return uuid.Nil, status.Error(codes.Unauthenticated, "missing authorization token")
	}

	// Проверяем токен
	// Формат токена: "Bearer <token>"
	tokenStr := strings.TrimPrefix(values[0], "Bearer ")

	// Парсим токен и получаем UUID пользователя
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, status.Error(codes.Unauthenticated, "unexpected signing method")
		}
		return []byte(a.jwtSecret), nil

	})

	if err != nil || !token.Valid {
		return uuid.Nil, status.Error(codes.Unauthenticated, "invalid authorization token")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, status.Error(codes.Unauthenticated, "invalid authorization token")
	}
	return userID, nil
}
