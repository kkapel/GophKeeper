package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kkapel/gophkeeper/internal/db/sqlc"
)

// mockItemStorage — заглушка хранилища для тестов KeeperService.
// Запоминает переданные параметры и возвращает заранее заданные результаты.
type mockItemStorage struct {
	createdParams sqlc.CreateItemParams
	createResult  sqlc.Item
	createErr     error

	getResult sqlc.Item
	getErr    error

	listResult []sqlc.Item
	listErr    error

	updateResult sqlc.Item
	updateErr    error

	deleteErr error
}

func (m *mockItemStorage) CreateItem(ctx context.Context, arg sqlc.CreateItemParams) (sqlc.Item, error) {
	m.createdParams = arg
	return m.createResult, m.createErr
}

func (m *mockItemStorage) GetItem(ctx context.Context, arg sqlc.GetItemParams) (sqlc.Item, error) {
	return m.getResult, m.getErr
}

func (m *mockItemStorage) ListItems(ctx context.Context, userID uuid.UUID) ([]sqlc.Item, error) {
	return m.listResult, m.listErr
}

func (m *mockItemStorage) UpdateItem(ctx context.Context, arg sqlc.UpdateItemParams) (sqlc.Item, error) {
	return m.updateResult, m.updateErr
}

func (m *mockItemStorage) DeleteItem(ctx context.Context, arg sqlc.DeleteItemParams) error {
	return m.deleteErr
}

// --- CreateItem ---

func TestKeeperService_CreateItem_Success(t *testing.T) {
	userID := uuid.New()
	mock := &mockItemStorage{
		createResult: sqlc.Item{ID: uuid.New(), Version: 1},
	}
	svc := NewKeeperService(mock)

	item, err := svc.CreateItem(context.Background(), userID, 4, []byte("encrypted-data"), "Visa")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// проверяем, что в storage ушли правильные данные
	if mock.createdParams.UserID != userID {
		t.Errorf("wrong userID: got %v, want %v", mock.createdParams.UserID, userID)
	}
	if mock.createdParams.Type != 4 {
		t.Errorf("wrong type: got %d, want 4", mock.createdParams.Type)
	}
	if string(mock.createdParams.EncryptedPayload) != "encrypted-data" {
		t.Error("payload mismatch")
	}
	if mock.createdParams.Metadata != "Visa" {
		t.Errorf("wrong metadata: got %q", mock.createdParams.Metadata)
	}
	if item.Version != 1 {
		t.Errorf("wrong version: got %d", item.Version)
	}
}

func TestKeeperService_CreateItem_StorageError(t *testing.T) {
	mock := &mockItemStorage{createErr: errors.New("db error")}
	svc := NewKeeperService(mock)

	_, err := svc.CreateItem(context.Background(), uuid.New(), 1, []byte("x"), "")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// --- GetItem ---

func TestKeeperService_GetItem_Success(t *testing.T) {
	itemID := uuid.New()
	mock := &mockItemStorage{
		getResult: sqlc.Item{ID: itemID, Version: 2},
	}
	svc := NewKeeperService(mock)

	item, err := svc.GetItem(context.Background(), uuid.New(), itemID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != itemID {
		t.Errorf("wrong item id: got %v, want %v", item.ID, itemID)
	}
}

func TestKeeperService_GetItem_NotFound(t *testing.T) {
	mock := &mockItemStorage{getErr: sql.ErrNoRows}
	svc := NewKeeperService(mock)

	_, err := svc.GetItem(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, ErrItemNotFound) {
		t.Errorf("expected ErrItemNotFound, got %v", err)
	}
}

func TestKeeperService_GetItem_StorageError(t *testing.T) {
	mock := &mockItemStorage{getErr: errors.New("connection lost")}
	svc := NewKeeperService(mock)

	_, err := svc.GetItem(context.Background(), uuid.New(), uuid.New())
	// реальный сбой БД НЕ должен маскироваться под ErrItemNotFound
	if errors.Is(err, ErrItemNotFound) {
		t.Error("db error masked as ErrItemNotFound")
	}
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// --- ListItems ---

func TestKeeperService_ListItems_Success(t *testing.T) {
	mock := &mockItemStorage{
		listResult: []sqlc.Item{
			{ID: uuid.New()},
			{ID: uuid.New()},
		},
	}
	svc := NewKeeperService(mock)

	items, err := svc.ListItems(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("wrong count: got %d, want 2", len(items))
	}
}

func TestKeeperService_ListItems_StorageError(t *testing.T) {
	mock := &mockItemStorage{listErr: errors.New("db error")}
	svc := NewKeeperService(mock)

	_, err := svc.ListItems(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// --- UpdateItem ---

func TestKeeperService_UpdateItem_NotFound(t *testing.T) {
	mock := &mockItemStorage{getErr: sql.ErrNoRows}
	svc := NewKeeperService(mock)
	_, err := svc.UpdateItem(context.Background(), uuid.New(), uuid.New(), []byte("x"), "", 1)
	if !errors.Is(err, ErrItemNotFound) {
		t.Errorf("expected ErrItemNotFound, got %v", err)
	}
}

func TestKeeperService_UpdateItem_Success(t *testing.T) {
	mock := &mockItemStorage{
		getResult:    sqlc.Item{ID: uuid.New()},
		updateResult: sqlc.Item{ID: uuid.New(), Version: 3},
	}
	svc := NewKeeperService(mock)
	item, err := svc.UpdateItem(context.Background(), uuid.New(), uuid.New(), []byte("new"), "meta", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Version != 3 {
		t.Errorf("wrong version: got %d", item.Version)
	}
}

func TestKeeperService_UpdateItem_VersionConflict(t *testing.T) {
	mock := &mockItemStorage{
		getResult: sqlc.Item{ID: uuid.New()}, // SELECT нашёл запись
		getErr:    nil,
		updateErr: sql.ErrNoRows, // но UPDATE не сработал → версия
	}
	svc := NewKeeperService(mock)
	_, err := svc.UpdateItem(context.Background(), uuid.New(), uuid.New(), []byte("x"), "", 1)
	if !errors.Is(err, ErrVersionConflict) {
		t.Errorf("expected ErrVersionConflict, got %v", err)
	}
}

func TestKeeperService_UpdateItem_StorageError(t *testing.T) {
	mock := &mockItemStorage{
		getResult: sqlc.Item{ID: uuid.New()}, // SELECT нашёл запись (явно)
		updateErr: errors.New("db error"),    // UPDATE упал реальной ошибкой
	}
	svc := NewKeeperService(mock)

	_, err := svc.UpdateItem(context.Background(), uuid.New(), uuid.New(), []byte("x"), "", 1)
	if errors.Is(err, ErrVersionConflict) {
		t.Error("db error masked as ErrVersionConflict")
	}
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// --- DeleteItem ---

func TestKeeperService_DeleteItem_Success(t *testing.T) {
	mock := &mockItemStorage{}
	svc := NewKeeperService(mock)

	err := svc.DeleteItem(context.Background(), uuid.New(), uuid.New())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestKeeperService_DeleteItem_StorageError(t *testing.T) {
	mock := &mockItemStorage{deleteErr: errors.New("db error")}
	svc := NewKeeperService(mock)

	err := svc.DeleteItem(context.Background(), uuid.New(), uuid.New())
	if err == nil {
		t.Error("expected error, got nil")
	}
}
