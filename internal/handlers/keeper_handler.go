package handlers

import (
	"context"

	"github.com/google/uuid"
	"github.com/kkapel/gophkeeper/internal/auth"
	"github.com/kkapel/gophkeeper/internal/db/sqlc"
	pb "github.com/kkapel/gophkeeper/internal/proto/gophkeeper/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type KeepService interface {
	// CreateItem создает новый элемент в хранилище.
	CreateItem(ctx context.Context, userID uuid.UUID, itemType int16, encryptedPayload []byte, metadata string) (sqlc.Item, error)
}

type KeeperHandler struct {
	pb.UnimplementedKeeperServiceServer
	service KeepService
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

func NewKeeperHandler(svc KeepService) *KeeperHandler {
	return &KeeperHandler{service: svc}
}
