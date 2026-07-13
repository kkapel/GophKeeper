package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kkapel/gophkeeper/internal/db/sqlc"
)

type ItemStorage interface {
	// CreateItem создает новый элемент в хранилище.
	CreateItem(ctx context.Context, arg sqlc.CreateItemParams) (sqlc.Item, error)
}

type KeeperService struct {
	storage ItemStorage
}

// CreateItem создает новый элемент в хранилище и возвращает его.
func (s *KeeperService) CreateItem(
	ctx context.Context,
	arg sqlc.CreateItemParams,
	userID uuid.UUID,
	itemType int16,
	encryptedPayload []byte,
	metadata string) (sqlc.Item, error) {

	item, err := s.storage.CreateItem(ctx, sqlc.CreateItemParams{
		UserID:           userID,
		Type:             itemType,
		EncryptedPayload: encryptedPayload,
		Metadata:         metadata,
	})

	if err != nil {
		return sqlc.Item{}, fmt.Errorf("failed to create item: %w", err)
	}

	return item, nil
}
