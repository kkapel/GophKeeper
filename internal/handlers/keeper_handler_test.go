package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kkapel/gophkeeper/internal/auth"
	"github.com/kkapel/gophkeeper/internal/domain"
	"github.com/kkapel/gophkeeper/internal/handlers/mocks"
	pb "github.com/kkapel/gophkeeper/internal/proto/gophkeeper/v1"
	"github.com/kkapel/gophkeeper/internal/service"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ctxWithUser — контекст с user_id, как его кладёт интерцептор.
func ctxWithUser() context.Context {
	return auth.WithUserID(context.Background(), uuid.New())
}

// --- CreateItem ---

func TestKeeperHandler_CreateItem_NoUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockKeepService(ctrl)
	// сервис не должен вызваться — EXPECT не задаём

	h := NewKeeperHandler(mockSvc)
	req := pb.CreateItemRequest_builder{}.Build()

	_, err := h.CreateItem(context.Background(), req)
	if status.Code(err) != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", status.Code(err))
	}
}

func TestKeeperHandler_CreateItem_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockKeepService(ctrl)

	mockSvc.EXPECT().
		CreateItem(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.Item{ID: uuid.New(), Version: 1}, nil)

	h := NewKeeperHandler(mockSvc)
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
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockKeepService(ctrl)

	mockSvc.EXPECT().
		CreateItem(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.Item{}, errors.New("boom"))

	h := NewKeeperHandler(mockSvc)
	req := pb.CreateItemRequest_builder{}.Build()

	_, err := h.CreateItem(ctxWithUser(), req)
	if status.Code(err) != codes.Internal {
		t.Errorf("expected Internal, got %v", status.Code(err))
	}
}

// --- GetItem ---

func TestKeeperHandler_GetItem_NoUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockKeepService(ctrl)

	h := NewKeeperHandler(mockSvc)
	id := uuid.New().String()
	req := pb.GetItemRequest_builder{Id: &id}.Build()

	_, err := h.GetItem(context.Background(), req)
	if status.Code(err) != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", status.Code(err))
	}
}

func TestKeeperHandler_GetItem_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockKeepService(ctrl)

	h := NewKeeperHandler(mockSvc)
	bad := "not-a-uuid"
	req := pb.GetItemRequest_builder{Id: &bad}.Build()

	_, err := h.GetItem(ctxWithUser(), req)
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", status.Code(err))
	}
}

func TestKeeperHandler_GetItem_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockKeepService(ctrl)

	mockSvc.EXPECT().
		GetItem(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.Item{}, service.ErrItemNotFound)

	h := NewKeeperHandler(mockSvc)
	id := uuid.New().String()
	req := pb.GetItemRequest_builder{Id: &id}.Build()

	_, err := h.GetItem(ctxWithUser(), req)
	if status.Code(err) != codes.NotFound {
		t.Errorf("expected NotFound, got %v", status.Code(err))
	}
}

func TestKeeperHandler_GetItem_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockKeepService(ctrl)

	mockSvc.EXPECT().
		GetItem(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.Item{ID: uuid.New(), Version: 1}, nil)

	h := NewKeeperHandler(mockSvc)
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
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockKeepService(ctrl)

	h := NewKeeperHandler(mockSvc)
	req := pb.ListItemsRequest_builder{}.Build()

	_, err := h.ListItems(context.Background(), req)
	if status.Code(err) != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", status.Code(err))
	}
}

func TestKeeperHandler_ListItems_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockKeepService(ctrl)

	mockSvc.EXPECT().
		ListItems(gomock.Any(), gomock.Any()).
		Return([]domain.Item{{ID: uuid.New()}, {ID: uuid.New()}}, nil)

	h := NewKeeperHandler(mockSvc)
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
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockKeepService(ctrl)

	h := NewKeeperHandler(mockSvc)
	id := uuid.New().String()
	req := pb.UpdateItemRequest_builder{Id: &id}.Build()

	_, err := h.UpdateItem(context.Background(), req)
	if status.Code(err) != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", status.Code(err))
	}
}

func TestKeeperHandler_UpdateItem_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockKeepService(ctrl)

	h := NewKeeperHandler(mockSvc)
	bad := "not-a-uuid"
	req := pb.UpdateItemRequest_builder{Id: &bad}.Build()

	_, err := h.UpdateItem(ctxWithUser(), req)
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", status.Code(err))
	}
}

func TestKeeperHandler_UpdateItem_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockKeepService(ctrl)

	mockSvc.EXPECT().
		UpdateItem(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.Item{}, service.ErrItemNotFound)

	h := NewKeeperHandler(mockSvc)
	id := uuid.New().String()
	req := pb.UpdateItemRequest_builder{Id: &id}.Build()

	_, err := h.UpdateItem(ctxWithUser(), req)
	if status.Code(err) != codes.NotFound {
		t.Errorf("expected NotFound, got %v", status.Code(err))
	}
}

func TestKeeperHandler_UpdateItem_VersionConflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockKeepService(ctrl)

	mockSvc.EXPECT().
		UpdateItem(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.Item{}, service.ErrVersionConflict)

	h := NewKeeperHandler(mockSvc)
	id := uuid.New().String()
	req := pb.UpdateItemRequest_builder{Id: &id}.Build()

	_, err := h.UpdateItem(ctxWithUser(), req)
	if status.Code(err) != codes.Aborted {
		t.Errorf("expected Aborted, got %v", status.Code(err))
	}
}

func TestKeeperHandler_UpdateItem_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockKeepService(ctrl)

	mockSvc.EXPECT().
		UpdateItem(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.Item{ID: uuid.New(), Version: 2}, nil)

	h := NewKeeperHandler(mockSvc)
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
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockKeepService(ctrl)

	h := NewKeeperHandler(mockSvc)
	id := uuid.New().String()
	req := pb.DeleteItemRequest_builder{Id: &id}.Build()

	_, err := h.DeleteItem(context.Background(), req)
	if status.Code(err) != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", status.Code(err))
	}
}

func TestKeeperHandler_DeleteItem_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockKeepService(ctrl)

	mockSvc.EXPECT().
		DeleteItem(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	h := NewKeeperHandler(mockSvc)
	id := uuid.New().String()
	req := pb.DeleteItemRequest_builder{Id: &id}.Build()

	_, err := h.DeleteItem(ctxWithUser(), req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestKeeperHandler_DeleteItem_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockKeepService(ctrl)

	mockSvc.EXPECT().
		DeleteItem(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(service.ErrItemNotFound)

	h := NewKeeperHandler(mockSvc)
	id := uuid.New().String()
	req := pb.DeleteItemRequest_builder{Id: &id}.Build()

	_, err := h.DeleteItem(ctxWithUser(), req)
	if status.Code(err) != codes.NotFound {
		t.Errorf("expected NotFound, got %v", status.Code(err))
		return
	}
}
