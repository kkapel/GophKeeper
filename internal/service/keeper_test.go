package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kkapel/gophkeeper/internal/db/sqlc"
)

// mockItemStorage — заглушка хранилища для тестов KeeperService.
type mockItemStorage struct {
	createdParams sqlc.CreateItemParams
	createResult  sqlc.Item
	createErr     error
}

func (m *mockItemStorage) CreateItem(ctx context.Context, arg sqlc.CreateItemParams) (sqlc.Item, error) {
	m.createdParams = arg
	return m.createResult, m.createErr
}

func TestKeeperService_CreateItem_Success(t *testing.T) {
	userID := uuid.New()
	mock := &mockItemStorage{
		createResult: sqlc.Item{ID: uuid.New(), Version: 1},
	}
	svc := NewKeepService(mock)

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
	svc := NewKeepService(mock)

	_, err := svc.CreateItem(context.Background(), uuid.New(), 1, []byte("x"), "")
	if err == nil {
		t.Error("expected error, got nil")
	}
}
