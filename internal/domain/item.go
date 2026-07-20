package domain

import (
	"time"

	"github.com/google/uuid"
)

// Item — доменная модель записи пользователя.
type Item struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	Type             int16
	EncryptedPayload []byte
	Metadata         string
	Version          int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
