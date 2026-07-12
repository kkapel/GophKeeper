package service

import (
	"context"
	"database/sql"
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

	getResult sqlc.User
	getErr    error
}

func (m *mockUserStorage) CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.CreateUserRow, error) {
	m.createdParams = arg // запоминаем аргумент для последующей проверки
	return m.createResult, m.createErr
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

func (m *mockUserStorage) GetUserByLogin(ctx context.Context, login string) (sqlc.User, error) {
	return m.getResult, m.getErr
}

func TestAuthService_Login_Success(t *testing.T) {
	const password = "secret123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	mock := &mockUserStorage{
		getResult: sqlc.User{
			ID:           uuid.New(),
			Login:        "user1",
			PasswordHash: string(hash),
		},
	}
	svc := NewAuthService(mock, "test-secret")

	token, err := svc.Login(context.Background(), "user1", password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)

	mock := &mockUserStorage{
		getResult: sqlc.User{
			ID:           uuid.New(),
			PasswordHash: string(hash),
		},
	}
	svc := NewAuthService(mock, "test-secret")

	_, err := svc.Login(context.Background(), "user1", "wrong-password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	mock := &mockUserStorage{
		getErr: sql.ErrNoRows,
	}
	svc := NewAuthService(mock, "test-secret")

	_, err := svc.Login(context.Background(), "nobody", "secret123")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}
