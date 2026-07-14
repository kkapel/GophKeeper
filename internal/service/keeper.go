package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kkapel/gophkeeper/internal/db/sqlc"
)

// ErrVersionConflict возвращается при несовпадении версии.
var ErrVersionConflict = errors.New("version conflict")
var ErrItemNotFound = errors.New("item not found")

type ItemStorage interface {
	// CreateItem создает новый элемент в хранилище.
	CreateItem(ctx context.Context, arg sqlc.CreateItemParams) (sqlc.Item, error)
	GetItem(ctx context.Context, arg sqlc.GetItemParams) (sqlc.Item, error)
	ListItems(ctx context.Context, userID uuid.UUID) ([]sqlc.Item, error)
	UpdateItem(ctx context.Context, arg sqlc.UpdateItemParams) (sqlc.Item, error)
	DeleteItem(ctx context.Context, arg sqlc.DeleteItemParams) error
}

type KeeperService struct {
	storage ItemStorage
}

// CreateItem создает новый элемент в хранилище и возвращает его.
func (s *KeeperService) CreateItem(
	ctx context.Context,
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

// GetItem возвращает запись пользователя по id.
func (s *KeeperService) GetItem(
	ctx context.Context,
	userID uuid.UUID,
	id uuid.UUID,
) (sqlc.Item, error) {

	item, err := s.storage.GetItem(ctx, sqlc.GetItemParams{
		ID:     id,
		UserID: userID,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlc.Item{}, ErrItemNotFound
		}
		return sqlc.Item{}, fmt.Errorf("GetItem: %w", err)
	}

	return item, nil
}

// ListItems возвращает все неудалённые записи пользователя.
func (s *KeeperService) ListItems(ctx context.Context, userID uuid.UUID) ([]sqlc.Item, error) {
	items, err := s.storage.ListItems(ctx, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to get list items: %w", err)
	}

	return items, nil
}

// UpdateItem обновляет запись с проверкой версии.
func (s *KeeperService) UpdateItem(ctx context.Context, userID, itemID uuid.UUID, payload []byte, metadata string, version int64) (sqlc.Item, error) {
	item, err := s.storage.UpdateItem(ctx, sqlc.UpdateItemParams{
		ID:               itemID,
		UserID:           userID,
		EncryptedPayload: payload,
		Metadata:         metadata,
		Version:          version,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// либо версия разошлась, либо записи нет
			return sqlc.Item{}, ErrVersionConflict
		}
		return sqlc.Item{}, fmt.Errorf("UpdateItem: %w", err)
	}

	return item, nil

}

// DeleteItem удаляет запись (проставляет флаг).
func (s *KeeperService) DeleteItem(ctx context.Context, userID, id uuid.UUID) error {
	err := s.storage.DeleteItem(ctx, sqlc.DeleteItemParams{
		ID:     id,
		UserID: userID,
	})

	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	return nil
}

// Создание KeeperService
func NewKeeperService(storage ItemStorage) *KeeperService {
	return &KeeperService{
		storage: storage,
	}
}
