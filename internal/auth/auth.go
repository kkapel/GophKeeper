// Package auth предоставляет примитивы для передачи идентификатора
// пользователя через контекст запроса.
package auth

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "user_id"

// WithUserID возвращает контекст с записанным идентификатором пользователя.
func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext извлекает идентификатор пользователя из контекста.
// Второе значение — false, если идентификатор отсутствует.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}
