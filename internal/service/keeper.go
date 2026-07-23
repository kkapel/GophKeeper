package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
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
	CreateItem(ctx context.Context, item domain.Item) (domain.Item, error)
	GetItem(ctx context.Context, userID, itemID uuid.UUID) (domain.Item, error)
	ListItems(ctx context.Context, userID uuid.UUID) ([]domain.Item, error)
	UpdateItem(ctx context.Context, item domain.Item) (domain.Item, error)
	DeleteItem(ctx context.Context, userID, itemID uuid.UUID) error
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

	item, err := s.storage.CreateItem(ctx, domain.Item{
		UserID:           userID,
		Type:             itemType,
		EncryptedPayload: encryptedPayload,
		Metadata:         metadata,
	})

	if err != nil {
		return domain.Item{}, fmt.Errorf("failed to create item: %w", err)
	}

	return item, nil
}

// GetItem возвращает запись пользователя по id.
func (s *KeeperService) GetItem(
	ctx context.Context,
	userID uuid.UUID,
	id uuid.UUID,
) (domain.Item, error) {

	item, err := s.storage.GetItem(ctx, userID, id)

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.Item{}, ErrItemNotFound
		}
		return domain.Item{}, fmt.Errorf("GetItem: %w", err)
	}

	return item, nil
}

// ListItems возвращает все неудалённые записи пользователя.
func (s *KeeperService) ListItems(ctx context.Context, userID uuid.UUID) ([]domain.Item, error) {
	items, err := s.storage.ListItems(ctx, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to get list items: %w", err)
	}

	return items, nil
}

// UpdateItem обновляет запись с проверкой версии.
func (s *KeeperService) UpdateItem(ctx context.Context, userID, itemID uuid.UUID, payload []byte, metadata string, version int64) (domain.Item, error) {

	// делаем update
	item, err := s.storage.UpdateItem(ctx, domain.Item{
		ID:               itemID,
		UserID:           userID,
		EncryptedPayload: payload,
		Metadata:         metadata,
		Version:          version,
	})

	// Если ошибки нет, возвращаем item
	if err == nil {
		return item, nil
	}

	if !errors.Is(err, domain.ErrNotFound) {
		return domain.Item{}, fmt.Errorf("UpdateItem: %w", err)
	}

	// либо записи нет, либо версия разошлась.
	// делаем дополнительный get-запрос
	if _, getErr := s.storage.GetItem(ctx, userID, itemID); getErr != nil {
		if errors.Is(getErr, domain.ErrNotFound) {
			return domain.Item{}, ErrItemNotFound
		}
		return domain.Item{}, fmt.Errorf("UpdateItem: %w", getErr)
	}

	return domain.Item{}, ErrVersionConflict

}

// DeleteItem выполняет мягкое удаление записи пользователя.
func (s *KeeperService) DeleteItem(ctx context.Context, userID, id uuid.UUID) error {
	err := s.storage.DeleteItem(ctx, userID, id)

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return ErrItemNotFound
		}
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
