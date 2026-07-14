package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kkapel/gophkeeper/internal/auth"
	"github.com/kkapel/gophkeeper/internal/db/sqlc"
	pb "github.com/kkapel/gophkeeper/internal/proto/gophkeeper/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockKeeperService struct {
	result sqlc.Item
	err    error
}

func (m *mockKeeperService) CreateItem(ctx context.Context, userID uuid.UUID, itemType int16, payload []byte, metadata string) (sqlc.Item, error) {
	return m.result, m.err
}

func TestKeeperHandler_CreateItem_Success(t *testing.T) {
	mock := &mockKeeperService{result: sqlc.Item{ID: uuid.New(), Version: 1}}
	h := NewKeeperHandler(mock)

	// имитируем интерцептор — кладём user_id в контекст
	ctx := auth.WithUserID(context.Background(), uuid.New())

	itemType := pb.ItemType_ITEM_TYPE_CARD
	payload := []byte("encrypted")
	meta := "Visa"
	req := pb.CreateItemRequest_builder{
		Type:             &itemType,
		EncryptedPayload: payload,
		Metadata:         &meta,
	}.Build()

	resp, err := h.CreateItem(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response")
	}
}

func TestKeeperHandler_CreateItem_NoUserID(t *testing.T) {
	mock := &mockKeeperService{}
	h := NewKeeperHandler(mock)

	req := pb.CreateItemRequest_builder{}.Build()

	// БЕЗ user_id в контексте → Unauthenticated
	_, err := h.CreateItem(context.Background(), req)
	if status.Code(err) != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", status.Code(err))
	}
}

func TestKeeperHandler_CreateItem_ServiceError(t *testing.T) {
	mock := &mockKeeperService{err: errors.New("boom")}
	h := NewKeeperHandler(mock)
	ctx := auth.WithUserID(context.Background(), uuid.New())

	req := pb.CreateItemRequest_builder{}.Build()
	_, err := h.CreateItem(ctx, req)
	if status.Code(err) != codes.Internal {
		t.Errorf("expected Internal, got %v", status.Code(err))
	}
}
