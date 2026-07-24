package handlers

import (
	"context"
	"errors"
	"testing"

	pb "github.com/kkapel/gophkeeper/internal/proto/gophkeeper/v1"
	"github.com/kkapel/gophkeeper/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockAuthService — заглушка сервиса. Возвращает заранее заданные token и err.
type mockAuthService struct {
	token string
	err   error
}

func (m *mockAuthService) Register(ctx context.Context, login, password string) (string, error) {
	return m.token, m.err
}

func (m *mockAuthService) Login(ctx context.Context, login, password string) (string, error) {
	return m.token, m.err
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name      string
		login     string
		password  string
		mockToken string
		mockErr   error
		wantCode  codes.Code
		wantToken string
	}{
		{
			name:      "success",
			login:     "user1",
			password:  "secret123",
			mockToken: "some.jwt.token",
			mockErr:   nil,
			wantCode:  codes.OK,
			wantToken: "some.jwt.token",
		},
		{
			name:     "empty login",
			login:    "",
			password: "secret123",
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "empty password",
			login:    "user1",
			password: "",
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "login already taken",
			login:    "user1",
			password: "secret123",
			mockErr:  service.ErrLoginTaken,
			wantCode: codes.AlreadyExists,
		},
		{
			name:     "internal error",
			login:    "user1",
			password: "secret123",
			mockErr:  errors.New("db exploded"),
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// подставляем мок вместо реального сервиса
			h := NewAuthHandler(&mockAuthService{token: tt.mockToken, err: tt.mockErr})

			req := pb.RegisterRequest_builder{
				Login:    &tt.login,
				Password: &tt.password,
			}.Build()

			resp, err := h.Register(context.Background(), req)

			// проверяем gRPC-код ошибки
			if status.Code(err) != tt.wantCode {
				t.Errorf("got code %v, want %v (err: %v)", status.Code(err), tt.wantCode, err)
				return
			}

			// при успехе проверяем токен в ответе
			if tt.wantCode == codes.OK {
				if resp == nil {
					t.Fatal("expected response, got nil")
				}
				if resp.GetAccessToken() != tt.wantToken {
					t.Errorf("got token %q, want %q", resp.GetAccessToken(), tt.wantToken)
					return
				}
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name      string
		login     string
		password  string
		mockToken string
		mockErr   error
		wantCode  codes.Code
		wantToken string
	}{
		{
			name:      "success",
			login:     "user1",
			password:  "secret123",
			mockToken: "some.jwt.token",
			wantCode:  codes.OK,
			wantToken: "some.jwt.token",
		},
		{
			name:     "empty login",
			login:    "",
			password: "secret123",
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "empty password",
			login:    "user1",
			password: "",
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "invalid credentials",
			login:    "user1",
			password: "wrong",
			mockErr:  service.ErrInvalidCredentials,
			wantCode: codes.Unauthenticated,
		},
		{
			name:     "internal error",
			login:    "user1",
			password: "secret123",
			mockErr:  errors.New("db exploded"),
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAuthHandler(&mockAuthService{token: tt.mockToken, err: tt.mockErr})

			req := pb.LoginRequest_builder{
				Login:    &tt.login,
				Password: &tt.password,
			}.Build()

			resp, err := h.Login(context.Background(), req)

			if status.Code(err) != tt.wantCode {
				t.Errorf("got code %v, want %v (err: %v)", status.Code(err), tt.wantCode, err)
				return
			}

			if tt.wantCode == codes.OK {
				if resp == nil {
					t.Fatal("expected response, got nil")
				}
				if resp.GetAccessToken() != tt.wantToken {
					t.Errorf("got token %q, want %q", resp.GetAccessToken(), tt.wantToken)
					return
				}
			}
		})
	}
}
