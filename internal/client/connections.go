package client

import (
	pb "github.com/kkapel/gophkeeper/internal/proto/gophkeeper/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// newAuthClient создаёт gRPC-клиент AuthService по адресу сервера.
func newAuthClient(address string) (pb.AuthServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, err
	}
	return pb.NewAuthServiceClient(conn), conn, nil
}

// newKeeperClient создаёт gRPC-клиент KeeperService.
func newKeeperClient(address string) (pb.KeeperServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, nil, err
	}

	return pb.NewKeeperServiceClient(conn), conn, nil
}
