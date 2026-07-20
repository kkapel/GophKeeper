package handlers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/kkapel/gophkeeper/internal/auth"
	"github.com/kkapel/gophkeeper/internal/domain"
	pb "github.com/kkapel/gophkeeper/internal/proto/gophkeeper/v1"
	"github.com/kkapel/gophkeeper/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type KeepService interface {
	// CreateItem создает новый элемент в хранилище.
	CreateItem(ctx context.Context, userID uuid.UUID, itemType int16, encryptedPayload []byte, metadata string) (domain.Item, error)
	GetItem(ctx context.Context, userID, itemID uuid.UUID) (domain.Item, error)
	ListItems(ctx context.Context, userID uuid.UUID) ([]domain.Item, error)
	UpdateItem(ctx context.Context, userID, itemID uuid.UUID, payload []byte, metadata string, version int64) (domain.Item, error)
	DeleteItem(ctx context.Context, userID, itemID uuid.UUID) error
}

type KeeperHandler struct {
	pb.UnimplementedKeeperServiceServer
	service KeepService
}

// GetItem возвращает запись пользователя по id.
func (h *KeeperHandler) GetItem(ctx context.Context, req *pb.GetItemRequest) (*pb.GetItemResponse, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no user in context")
	}

	itemID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid item id")
	}

	item, err := h.service.GetItem(ctx, userID, itemID)
	if err != nil {
		if errors.Is(err, service.ErrItemNotFound) {
			return nil, status.Error(codes.NotFound, "item not found")
		}
		return nil, status.Error(codes.Internal, "failed to get item")
	}

	return pb.GetItemResponse_builder{
		Item: itemToProto(item),
	}.Build(), nil
}

// CreateItem обрабатывает gRPC-запрос на создание нового элемента в хранилище.
func (h *KeeperHandler) CreateItem(ctx context.Context, req *pb.CreateItemRequest) (*pb.CreateItemResponse, error) {

	userID, ok := auth.UserIDFromContext(ctx)

	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no user in context")
	}

	item, err := h.service.CreateItem(
		ctx,
		userID,
		int16(req.GetType()),
		req.GetEncryptedPayload(),
		req.GetMetadata(),
	)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create item")
	}

	return pb.CreateItemResponse_builder{
		Id:      proto.String(item.ID.String()),
		Version: &item.Version,
	}.Build(), nil
}

// ListItems возвращает все записи пользователя.
func (h *KeeperHandler) ListItems(ctx context.Context, req *pb.ListItemsRequest) (*pb.ListItemsResponse, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no user in context")
	}

	items, err := h.service.ListItems(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list items")
	}

	// Формируем ответ
	protoItems := make([]*pb.Item, 0, len(items))
	for _, item := range items {
		protoItems = append(protoItems, itemToProto(item))
	}

	return pb.ListItemsResponse_builder{
		Items: protoItems,
	}.Build(), nil

}

// UpdateItem обновляет запись.
func (h *KeeperHandler) UpdateItem(ctx context.Context, req *pb.UpdateItemRequest) (*pb.UpdateItemResponse, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no user in context")
	}

	itemID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid item id")
	}

	item, err := h.service.UpdateItem(ctx, userID, itemID, req.GetEncryptedPayload(), req.GetMetadata(), req.GetVersion())
	if err != nil {
		if errors.Is(err, service.ErrItemNotFound) {
			return nil, status.Error(codes.NotFound, "item not found")
		}
		if errors.Is(err, service.ErrVersionConflict) {
			return nil, status.Error(codes.Aborted, "version conflict, reload and retry")
		}
		return nil, status.Error(codes.Internal, "failed to update item")
	}

	version := item.Version
	return pb.UpdateItemResponse_builder{
		Version: &version,
	}.Build(), nil
}

// DeleteItem удаляет запись (мягко).
func (h *KeeperHandler) DeleteItem(ctx context.Context, req *pb.DeleteItemRequest) (*pb.DeleteItemResponse, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no user in context")
	}

	itemID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid item id")
	}

	if err := h.service.DeleteItem(ctx, userID, itemID); err != nil {
		return nil, status.Error(codes.Internal, "failed to delete item")
	}

	return pb.DeleteItemResponse_builder{}.Build(), nil
}

// Создание KeepHandler
func NewKeeperHandler(svc KeepService) *KeeperHandler {
	return &KeeperHandler{service: svc}
}

// Конвертирует sqlc.Item в pb.Item
func itemToProto(item domain.Item) *pb.Item {
	id := item.ID.String()
	itemType := pb.ItemType(item.Type)

	return pb.Item_builder{
		Id:               &id,
		Type:             &itemType,
		EncryptedPayload: item.EncryptedPayload,
		Metadata:         &item.Metadata,
		Version:          &item.Version,
	}.Build()
}
