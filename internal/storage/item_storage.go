// Package storage содержит адаптеры доступа к хранилищу данных.
package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kkapel/gophkeeper/internal/db/sqlc"
	"github.com/kkapel/gophkeeper/internal/domain"
)

// Queries описывает сгенерированные sqlc-запросы, нужные адаптеру.
type Queries interface {
	CreateItem(ctx context.Context, arg sqlc.CreateItemParams) (sqlc.Item, error)
	GetItem(ctx context.Context, arg sqlc.GetItemParams) (sqlc.Item, error)
	ListItems(ctx context.Context, userID uuid.UUID) ([]sqlc.Item, error)
	UpdateItem(ctx context.Context, arg sqlc.UpdateItemParams) (sqlc.Item, error)
	DeleteItem(ctx context.Context, arg sqlc.DeleteItemParams) (int64, error)
}

// ItemStorage реализует хранение приватных данных поверх PostgreSQL.
type ItemStorage struct {
	queries Queries
}

// NewItemStorage создаёт адаптер хранилища приватных данных.
func NewItemStorage(queries Queries) *ItemStorage {
	return &ItemStorage{queries: queries}
}

// CreateItem сохраняет новую запись и возвращает её.
func (s *ItemStorage) CreateItem(ctx context.Context, item domain.Item) (domain.Item, error) {
	row, err := s.queries.CreateItem(ctx, sqlc.CreateItemParams{
		UserID:           item.UserID,
		Type:             item.Type,
		EncryptedPayload: item.EncryptedPayload,
		Metadata:         item.Metadata,
	})
	if err != nil {
		return domain.Item{}, fmt.Errorf("create item: %w", err)
	}
	return toDomainItem(row), nil
}

// GetItem возвращает запись пользователя по идентификатору.
func (s *ItemStorage) GetItem(ctx context.Context, userID, itemID uuid.UUID) (domain.Item, error) {
	row, err := s.queries.GetItem(ctx, sqlc.GetItemParams{
		ID:     itemID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Item{}, domain.ErrNotFound
		}
		return domain.Item{}, fmt.Errorf("get item: %w", err)
	}
	return toDomainItem(row), nil
}

// ListItems возвращает все неудалённые записи пользователя.
func (s *ItemStorage) ListItems(ctx context.Context, userID uuid.UUID) ([]domain.Item, error) {
	rows, err := s.queries.ListItems(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	items := make([]domain.Item, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomainItem(row))
	}
	return items, nil
}

// UpdateItem обновляет запись с проверкой версии.
func (s *ItemStorage) UpdateItem(ctx context.Context, item domain.Item) (domain.Item, error) {
	row, err := s.queries.UpdateItem(ctx, sqlc.UpdateItemParams{
		ID:               item.ID,
		UserID:           item.UserID,
		EncryptedPayload: item.EncryptedPayload,
		Metadata:         item.Metadata,
		Version:          item.Version,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Item{}, domain.ErrNotFound
		}
		return domain.Item{}, fmt.Errorf("update item: %w", err)
	}
	return toDomainItem(row), nil
}

// DeleteItem выполняет мягкое удаление записи.
// Возвращает domain.ErrNotFound, если запись отсутствует или принадлежит другому пользователю.
func (s *ItemStorage) DeleteItem(ctx context.Context, userID, itemID uuid.UUID) error {
	rows, err := s.queries.DeleteItem(ctx, sqlc.DeleteItemParams{
		ID:     itemID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
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
