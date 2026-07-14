package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kkapel/gophkeeper/internal/auth"
	"github.com/kkapel/gophkeeper/internal/db/sqlc"
	pb "github.com/kkapel/gophkeeper/internal/proto/gophkeeper/v1"
	"github.com/kkapel/gophkeeper/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockKeeperService — заглушка сервиса для тестов KeeperHandler.
type mockKeeperService struct {
	createResult sqlc.Item
	createErr    error
	getResult    sqlc.Item
	getErr       error
	listResult   []sqlc.Item
	listErr      error
	updateResult sqlc.Item
	updateErr    error
	deleteErr    error
}

func (m *mockKeeperService) CreateItem(ctx context.Context, userID uuid.UUID, itemType int16, payload []byte, metadata string) (sqlc.Item, error) {
	return m.createResult, m.createErr
}
func (m *mockKeeperService) GetItem(ctx context.Context, userID, itemID uuid.UUID) (sqlc.Item, error) {
	return m.getResult, m.getErr
}
func (m *mockKeeperService) ListItems(ctx context.Context, userID uuid.UUID) ([]sqlc.Item, error) {
	return m.listResult, m.listErr
}
func (m *mockKeeperService) UpdateItem(ctx context.Context, userID, itemID uuid.UUID, payload []byte, metadata string, version int64) (sqlc.Item, error) {
	return m.updateResult, m.updateErr
}
func (m *mockKeeperService) DeleteItem(ctx context.Context, userID, itemID uuid.UUID) error {
	return m.deleteErr
}

// ctxWithUser — контекст с user_id, как его кладёт интерцептор.
func ctxWithUser() context.Context {
	return auth.WithUserID(context.Background(), uuid.New())
}

// --- CreateItem ---

func TestKeeperHandler_CreateItem_NoUserID(t *testing.T) {
	h := NewKeeperHandler(&mockKeeperService{})
	req := pb.CreateItemRequest_builder{}.Build()

	_, err := h.CreateItem(context.Background(), req) // без user_id
	if status.Code(err) != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", status.Code(err))
	}
}

func TestKeeperHandler_CreateItem_Success(t *testing.T) {
	h := NewKeeperHandler(&mockKeeperService{
		createResult: sqlc.Item{ID: uuid.New(), Version: 1},
	})
	req := pb.CreateItemRequest_builder{}.Build()

	resp, err := h.CreateItem(ctxWithUser(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response")
	}
}

func TestKeeperHandler_CreateItem_ServiceError(t *testing.T) {
	h := NewKeeperHandler(&mockKeeperService{createErr: errors.New("boom")})
	req := pb.CreateItemRequest_builder{}.Build()

	_, err := h.CreateItem(ctxWithUser(), req)
	if status.Code(err) != codes.Internal {
		t.Errorf("expected Internal, got %v", status.Code(err))
	}
}

// --- GetItem ---

func TestKeeperHandler_GetItem_NoUserID(t *testing.T) {
	h := NewKeeperHandler(&mockKeeperService{})
	id := uuid.New().String()
	req := pb.GetItemRequest_builder{Id: &id}.Build()

	_, err := h.GetItem(context.Background(), req)
	if status.Code(err) != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", status.Code(err))
	}
}

func TestKeeperHandler_GetItem_InvalidID(t *testing.T) {
	h := NewKeeperHandler(&mockKeeperService{})
	bad := "not-a-uuid"
	req := pb.GetItemRequest_builder{Id: &bad}.Build()

	_, err := h.GetItem(ctxWithUser(), req)
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", status.Code(err))
	}
}

func TestKeeperHandler_GetItem_NotFound(t *testing.T) {
	h := NewKeeperHandler(&mockKeeperService{getErr: service.ErrItemNotFound})
	id := uuid.New().String()
	req := pb.GetItemRequest_builder{Id: &id}.Build()

	_, err := h.GetItem(ctxWithUser(), req)
	if status.Code(err) != codes.NotFound {
		t.Errorf("expected NotFound, got %v", status.Code(err))
	}
}

func TestKeeperHandler_GetItem_Success(t *testing.T) {
	h := NewKeeperHandler(&mockKeeperService{
		getResult: sqlc.Item{ID: uuid.New(), Version: 1},
	})
	id := uuid.New().String()
	req := pb.GetItemRequest_builder{Id: &id}.Build()

	resp, err := h.GetItem(ctxWithUser(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response")
	}
}

// --- ListItems ---

func TestKeeperHandler_ListItems_NoUserID(t *testing.T) {
	h := NewKeeperHandler(&mockKeeperService{})
	req := pb.ListItemsRequest_builder{}.Build()

	_, err := h.ListItems(context.Background(), req)
	if status.Code(err) != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", status.Code(err))
	}
}

func TestKeeperHandler_ListItems_Success(t *testing.T) {
	h := NewKeeperHandler(&mockKeeperService{
		listResult: []sqlc.Item{{ID: uuid.New()}, {ID: uuid.New()}},
	})
	req := pb.ListItemsRequest_builder{}.Build()

	resp, err := h.ListItems(ctxWithUser(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response")
	}
}

// --- UpdateItem ---

func TestKeeperHandler_UpdateItem_NoUserID(t *testing.T) {
	h := NewKeeperHandler(&mockKeeperService{})
	id := uuid.New().String()
	req := pb.UpdateItemRequest_builder{Id: &id}.Build()

	_, err := h.UpdateItem(context.Background(), req)
	if status.Code(err) != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", status.Code(err))
	}
}

func TestKeeperHandler_UpdateItem_VersionConflict(t *testing.T) {
	h := NewKeeperHandler(&mockKeeperService{updateErr: service.ErrVersionConflict})
	id := uuid.New().String()
	req := pb.UpdateItemRequest_builder{Id: &id}.Build()

	_, err := h.UpdateItem(ctxWithUser(), req)
	if status.Code(err) != codes.Aborted {
		t.Errorf("expected Aborted, got %v", status.Code(err))
	}
}

func TestKeeperHandler_UpdateItem_Success(t *testing.T) {
	h := NewKeeperHandler(&mockKeeperService{
		updateResult: sqlc.Item{ID: uuid.New(), Version: 2},
	})
	id := uuid.New().String()
	req := pb.UpdateItemRequest_builder{Id: &id}.Build()

	resp, err := h.UpdateItem(ctxWithUser(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response")
	}
}

// --- DeleteItem ---

func TestKeeperHandler_DeleteItem_NoUserID(t *testing.T) {
	h := NewKeeperHandler(&mockKeeperService{})
	id := uuid.New().String()
	req := pb.DeleteItemRequest_builder{Id: &id}.Build()

	_, err := h.DeleteItem(context.Background(), req)
	if status.Code(err) != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", status.Code(err))
	}
}

func TestKeeperHandler_DeleteItem_Success(t *testing.T) {
	h := NewKeeperHandler(&mockKeeperService{})
	id := uuid.New().String()
	req := pb.DeleteItemRequest_builder{Id: &id}.Build()

	_, err := h.DeleteItem(ctxWithUser(), req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
