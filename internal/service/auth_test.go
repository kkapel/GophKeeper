package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kkapel/gophkeeper/internal/db/sqlc"
	"golang.org/x/crypto/bcrypt"
)

// mockUserStorage — заглушка хранилища. Запоминает, что ему передали,
// и возвращает заранее заданные результаты.
type mockUserStorage struct {
	createdParams sqlc.CreateUserParams // что реально ушло в CreateUser
	createResult  sqlc.CreateUserRow
	createErr     error
}

func (m *mockUserStorage) CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.CreateUserRow, error) {
	m.createdParams = arg // запоминаем аргумент для последующей проверки
	return m.createResult, m.createErr
}

func (m *mockUserStorage) GetUserByLogin(ctx context.Context, login string) (sqlc.User, error) {
	return sqlc.User{}, nil // для регистрации не нужен
}

func TestAuthService_Register_Success(t *testing.T) {
	mock := &mockUserStorage{
		createResult: sqlc.CreateUserRow{ID: uuid.New()},
	}
	svc := NewAuthService(mock, "test-secret")

	token, err := svc.Register(context.Background(), "user1", "secret123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestAuthService_Register_HashesPassword(t *testing.T) {
	mock := &mockUserStorage{
		createResult: sqlc.CreateUserRow{ID: uuid.New()},
	}
	svc := NewAuthService(mock, "test-secret")

	const password = "secret123"
	_, err := svc.Register(context.Background(), "user1", password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// в БД не должен уйти открытый пароль
	if mock.createdParams.PasswordHash == password {
		t.Fatal("password stored in plaintext!")
	}
	// а то, что ушло, должно быть валидным bcrypt-хешем этого пароля
	err = bcrypt.CompareHashAndPassword([]byte(mock.createdParams.PasswordHash), []byte(password))
	if err != nil {
		t.Errorf("stored hash does not match password: %v", err)
	}
}

func TestAuthService_Register_LoginTaken(t *testing.T) {
	mock := &mockUserStorage{
		createErr: &pgconn.PgError{Code: "23505"}, // unique_violation
	}
	svc := NewAuthService(mock, "test-secret")

	_, err := svc.Register(context.Background(), "user1", "secret123")
	if !errors.Is(err, ErrLoginTaken) {
		t.Errorf("expected ErrLoginTaken, got %v", err)
	}
}
