// Package auth предоставляет примитивы для передачи идентификатора
// пользователя через контекст запроса.
package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	// ErrNoUserID возвращается, когда в контексте отсутствует идентификатор пользователя.
	ErrNoUserID = errors.New("user id not found in context")
	// ErrInvalidUserID возвращается, когда значение в контексте имеет неожиданный тип.
	ErrInvalidUserID = errors.New("user id in context has invalid type")
)

type contextKey string

const userIDKey contextKey = "user_id"

// WithUserID возвращает контекст с записанным идентификатором пользователя.
func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext извлекает идентификатор пользователя из контекста.
// Возвращает ErrNoUserID, если значение отсутствует, и ErrInvalidUserID,
// если значение имеет неожиданный тип.
func UserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	value := ctx.Value(userIDKey)
	if value == nil {
		return uuid.Nil, ErrNoUserID
	}

	userID, ok := value.(uuid.UUID)
	if !ok {
		return uuid.Nil, ErrInvalidUserID
	}
	return userID, nil
}
