package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kkapel/gophkeeper/internal/db/sqlc"
	"golang.org/x/crypto/bcrypt"
)

var ErrLoginTaken = errors.New("login already taken")

// UserStorage описывает операции с пользователями, нужные сервису аутентификации.
type UserStorage interface {
	CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.CreateUserRow, error)
	GetUserByLogin(ctx context.Context, login string) (sqlc.User, error)
}

// AuthService реализует бизнес-логику регистрации и аутентификации.
type AuthService struct {
	storage   UserStorage
	jwtSecret string
}

// NewAuthService создаёт сервис аутентификации.
func NewAuthService(storage UserStorage, jwtSecret string) *AuthService {
	return &AuthService{storage: storage, jwtSecret: jwtSecret}
}

// Register регистрирует нового пользователя: хеширует пароль, сохраняет
// пользователя и возвращает подписанный JWT-токен.
func (s *AuthService) Register(ctx context.Context, login, password string) (string, error) {
	// 1. Хешируем пароль — открытый пароль никуда дальше не идёт
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("Register: hash password: %w", err)
	}

	// 2. Сохраняем пользователя
	user, err := s.storage.CreateUser(ctx, sqlc.CreateUserParams{
		Login:        login,
		PasswordHash: string(hash),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return "", ErrLoginTaken
		}
		return "", fmt.Errorf("Register: create user: %w", err)
	}

	// 3. Генерируем токен для нового пользователя
	token, err := s.generateToken(user.ID)
	if err != nil {
		return "", fmt.Errorf("Register: generate token: %w", err)
	}

	return token, nil
}

func (s *AuthService) generateToken(userID uuid.UUID) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
