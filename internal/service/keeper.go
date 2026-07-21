package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kkapel/gophkeeper/internal/db/sqlc"
	"github.com/kkapel/gophkeeper/internal/domain"
)

var (
	// ErrVersionConflict возвращается при несовпадении версии записи.
	ErrVersionConflict = errors.New("version conflict")
	// ErrItemNotFound возвращается, когда запись не найдена
	// или принадлежит другому пользователю.
	ErrItemNotFound = errors.New("item not found")
)

//go:generate mockgen -source=keeper.go -destination=mocks/mock_item_storage.go -package=mocks ItemStorage

// ItemStorage описывает операции с приватными данными,
// необходимые сервису для работы с хранилищем.
type ItemStorage interface {
	// CreateItem создает новый элемент в хранилище.
	CreateItem(ctx context.Context, arg sqlc.CreateItemParams) (sqlc.Item, error)
	GetItem(ctx context.Context, arg sqlc.GetItemParams) (sqlc.Item, error)
	ListItems(ctx context.Context, userID uuid.UUID) ([]sqlc.Item, error)
	UpdateItem(ctx context.Context, arg sqlc.UpdateItemParams) (sqlc.Item, error)
	DeleteItem(ctx context.Context, arg sqlc.DeleteItemParams) error
}

// KeeperService реализует бизнес-логику работы с приватными данными пользователя:
// создание, чтение, обновление и удаление записей.
type KeeperService struct {
	storage ItemStorage
}

// CreateItem создает новый элемент в хранилище и возвращает его.
func (s *KeeperService) CreateItem(
	ctx context.Context,
	userID uuid.UUID,
	itemType int16,
	encryptedPayload []byte,
	metadata string) (domain.Item, error) {

	item, err := s.storage.CreateItem(ctx, sqlc.CreateItemParams{
		UserID:           userID,
		Type:             itemType,
		EncryptedPayload: encryptedPayload,
		Metadata:         metadata,
	})

	if err != nil {
		return domain.Item{}, fmt.Errorf("failed to create item: %w", err)
	}

	return toDomainItem(item), nil
}

// GetItem возвращает запись пользователя по id.
func (s *KeeperService) GetItem(
	ctx context.Context,
	userID uuid.UUID,
	id uuid.UUID,
) (domain.Item, error) {

	item, err := s.storage.GetItem(ctx, sqlc.GetItemParams{
		ID:     id,
		UserID: userID,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Item{}, ErrItemNotFound
		}
		return domain.Item{}, fmt.Errorf("GetItem: %w", err)
	}

	return toDomainItem(item), nil
}

// ListItems возвращает все неудалённые записи пользователя.
func (s *KeeperService) ListItems(ctx context.Context, userID uuid.UUID) ([]domain.Item, error) {
	items, err := s.storage.ListItems(ctx, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to get list items: %w", err)
	}

	domainItems := make([]domain.Item, 0, len(items))
	for _, value := range items {
		domainItems = append(domainItems, toDomainItem(value))
	}
	return domainItems, nil
}

// UpdateItem обновляет запись с проверкой версии.
func (s *KeeperService) UpdateItem(ctx context.Context, userID, itemID uuid.UUID, payload []byte, metadata string, version int64) (domain.Item, error) {

	// Смотрим, есть ли запись
	_, err := s.storage.GetItem(ctx, sqlc.GetItemParams{
		ID:     itemID,
		UserID: userID,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Item{}, ErrItemNotFound // записи нет
		}
		return domain.Item{}, fmt.Errorf("UpdateItem: %w", err)
	}

	// Запись есть — обновляем с проверкой версии
	item, err := s.storage.UpdateItem(ctx, sqlc.UpdateItemParams{
		ID:               itemID,
		UserID:           userID,
		EncryptedPayload: payload,
		Metadata:         metadata,
		Version:          version,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// версия разошлась
			return domain.Item{}, ErrVersionConflict
		}
		return domain.Item{}, fmt.Errorf("UpdateItem: %w", err)
	}

	return toDomainItem(item), nil

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

// NewKeeperService создаёт сервис работы с приватными данными.
func NewKeeperService(storage ItemStorage) *KeeperService {
	return &KeeperService{
		storage: storage,
	}
}

// toDomainItem конвертирует запись хранилища в доменную модель.
func toDomainItem(i sqlc.Item) domain.Item {
	return domain.Item{
		ID:               i.ID,
		UserID:           i.UserID,
		Type:             i.Type,
		EncryptedPayload: i.EncryptedPayload,
		Metadata:         i.Metadata,
		Version:          i.Version,
		CreatedAt:        i.CreatedAt,
		UpdatedAt:        i.UpdatedAt,
	}
}
