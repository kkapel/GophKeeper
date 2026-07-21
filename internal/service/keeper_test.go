package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kkapel/gophkeeper/internal/db/sqlc"
	"github.com/kkapel/gophkeeper/internal/service/mocks"
	"go.uber.org/mock/gomock"
)

// --- CreateItem ---

func TestKeeperService_CreateItem_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockStorage := mocks.NewMockItemStorage(ctrl)

	userID := uuid.New()
	mockStorage.EXPECT().
		CreateItem(gomock.Any(), gomock.Cond(func(arg sqlc.CreateItemParams) bool {
			return arg.UserID == userID &&
				arg.Type == 4 &&
				string(arg.EncryptedPayload) == "encrypted-data" &&
				arg.Metadata == "Visa"
		})).
		Return(sqlc.Item{ID: uuid.New(), Version: 1}, nil)

	svc := NewKeeperService(mockStorage)

	item, err := svc.CreateItem(context.Background(), userID, 4, []byte("encrypted-data"), "Visa")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Version != 1 {
		t.Errorf("wrong version: got %d", item.Version)
	}
}

func TestKeeperService_CreateItem_StorageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockStorage := mocks.NewMockItemStorage(ctrl)

	mockStorage.EXPECT().
		CreateItem(gomock.Any(), gomock.Any()).
		Return(sqlc.Item{}, errors.New("db error"))

	svc := NewKeeperService(mockStorage)

	_, err := svc.CreateItem(context.Background(), uuid.New(), 1, []byte("x"), "")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// --- GetItem ---

func TestKeeperService_GetItem_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockStorage := mocks.NewMockItemStorage(ctrl)

	itemID := uuid.New()
	mockStorage.EXPECT().
		GetItem(gomock.Any(), gomock.Any()).
		Return(sqlc.Item{ID: itemID, Version: 2}, nil)

	svc := NewKeeperService(mockStorage)

	item, err := svc.GetItem(context.Background(), uuid.New(), itemID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != itemID {
		t.Errorf("wrong item id: got %v, want %v", item.ID, itemID)
	}
}

func TestKeeperService_GetItem_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockStorage := mocks.NewMockItemStorage(ctrl)

	mockStorage.EXPECT().
		GetItem(gomock.Any(), gomock.Any()).
		Return(sqlc.Item{}, sql.ErrNoRows)

	svc := NewKeeperService(mockStorage)

	_, err := svc.GetItem(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, ErrItemNotFound) {
		t.Errorf("expected ErrItemNotFound, got %v", err)
	}
}

func TestKeeperService_GetItem_StorageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockStorage := mocks.NewMockItemStorage(ctrl)

	mockStorage.EXPECT().
		GetItem(gomock.Any(), gomock.Any()).
		Return(sqlc.Item{}, errors.New("connection lost"))

	svc := NewKeeperService(mockStorage)

	_, err := svc.GetItem(context.Background(), uuid.New(), uuid.New())
	if errors.Is(err, ErrItemNotFound) {
		t.Error("db error masked as ErrItemNotFound")
	}
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// --- ListItems ---

func TestKeeperService_ListItems_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockStorage := mocks.NewMockItemStorage(ctrl)

	mockStorage.EXPECT().
		ListItems(gomock.Any(), gomock.Any()).
		Return([]sqlc.Item{{ID: uuid.New()}, {ID: uuid.New()}}, nil)

	svc := NewKeeperService(mockStorage)

	items, err := svc.ListItems(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("wrong count: got %d, want 2", len(items))
	}
}

func TestKeeperService_ListItems_StorageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockStorage := mocks.NewMockItemStorage(ctrl)

	mockStorage.EXPECT().
		ListItems(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("db error"))

	svc := NewKeeperService(mockStorage)

	_, err := svc.ListItems(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// --- UpdateItem ---

func TestKeeperService_UpdateItem_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockStorage := mocks.NewMockItemStorage(ctrl)

	// GetItem не находит запись
	mockStorage.EXPECT().
		GetItem(gomock.Any(), gomock.Any()).
		Return(sqlc.Item{}, sql.ErrNoRows)

	svc := NewKeeperService(mockStorage)

	_, err := svc.UpdateItem(context.Background(), uuid.New(), uuid.New(), []byte("x"), "", 1)
	if !errors.Is(err, ErrItemNotFound) {
		t.Errorf("expected ErrItemNotFound, got %v", err)
	}
}

func TestKeeperService_UpdateItem_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockStorage := mocks.NewMockItemStorage(ctrl)

	// GetItem находит запись
	mockStorage.EXPECT().
		GetItem(gomock.Any(), gomock.Any()).
		Return(sqlc.Item{ID: uuid.New()}, nil)
	// UpdateItem успешно
	mockStorage.EXPECT().
		UpdateItem(gomock.Any(), gomock.Any()).
		Return(sqlc.Item{ID: uuid.New(), Version: 3}, nil)

	svc := NewKeeperService(mockStorage)

	item, err := svc.UpdateItem(context.Background(), uuid.New(), uuid.New(), []byte("new"), "meta", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Version != 3 {
		t.Errorf("wrong version: got %d", item.Version)
	}
}

func TestKeeperService_UpdateItem_VersionConflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockStorage := mocks.NewMockItemStorage(ctrl)

	// GetItem находит запись
	mockStorage.EXPECT().
		GetItem(gomock.Any(), gomock.Any()).
		Return(sqlc.Item{ID: uuid.New()}, nil)
	// UpdateItem не срабатывает — версия разошлась
	mockStorage.EXPECT().
		UpdateItem(gomock.Any(), gomock.Any()).
		Return(sqlc.Item{}, sql.ErrNoRows)

	svc := NewKeeperService(mockStorage)

	_, err := svc.UpdateItem(context.Background(), uuid.New(), uuid.New(), []byte("x"), "", 1)
	if !errors.Is(err, ErrVersionConflict) {
		t.Errorf("expected ErrVersionConflict, got %v", err)
	}
}

func TestKeeperService_UpdateItem_StorageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockStorage := mocks.NewMockItemStorage(ctrl)

	// GetItem находит запись
	mockStorage.EXPECT().
		GetItem(gomock.Any(), gomock.Any()).
		Return(sqlc.Item{ID: uuid.New()}, nil)
	// UpdateItem падает реальной ошибкой
	mockStorage.EXPECT().
		UpdateItem(gomock.Any(), gomock.Any()).
		Return(sqlc.Item{}, errors.New("db error"))

	svc := NewKeeperService(mockStorage)

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
	ctrl := gomock.NewController(t)
	mockStorage := mocks.NewMockItemStorage(ctrl)

	mockStorage.EXPECT().
		DeleteItem(gomock.Any(), gomock.Any()).
		Return(nil)

	svc := NewKeeperService(mockStorage)

	err := svc.DeleteItem(context.Background(), uuid.New(), uuid.New())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestKeeperService_DeleteItem_StorageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockStorage := mocks.NewMockItemStorage(ctrl)

	mockStorage.EXPECT().
		DeleteItem(gomock.Any(), gomock.Any()).
		Return(errors.New("db error"))

	svc := NewKeeperService(mockStorage)

	err := svc.DeleteItem(context.Background(), uuid.New(), uuid.New())
	if err == nil {
		t.Error("expected error, got nil")
	}
}
