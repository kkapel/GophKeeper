package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	pb "github.com/kkapel/gophkeeper/internal/proto/gophkeeper/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// newAuthClient создаёт gRPC-клиент AuthService по адресу сервера.
func newAuthClient(address, certPath string) (pb.AuthServiceClient, *grpc.ClientConn, error) {
	creds, err := loadClientTLS(certPath)
	if err != nil {
		return nil, nil, err
	}
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		return nil, nil, err
	}
	return pb.NewAuthServiceClient(conn), conn, nil
}

// newKeeperClient создаёт gRPC-клиент KeeperService.
func newKeeperClient(address, certPath string) (pb.KeeperServiceClient, *grpc.ClientConn, error) {
	creds, err := loadClientTLS(certPath)
	if err != nil {
		return nil, nil, err
	}
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(creds),
	)

	if err != nil {
		return nil, nil, err
	}

	return pb.NewKeeperServiceClient(conn), conn, nil
}

// loadClientTLS создаёт TLS-креды клиента, доверяя сертификату сервера.
func loadClientTLS(certPath string) (credentials.TransportCredentials, error) {
	pemServerCA, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("read server cert: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server cert to pool")
	}

	return credentials.NewTLS(&tls.Config{
		RootCAs: certPool,
	}), nil
}
