package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "github.com/kkapel/gophkeeper/internal/proto/gophkeeper/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// getCmd получает одну запись по id.
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Получить запись по id",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := cmd.Flags().GetString("id")
		if err != nil {
			return err
		}
		if id == "" {
			return errors.New("укажите id через --id")
		}

		token, err := loadToken()
		if err != nil {
			return fmt.Errorf("не удалось прочитать токен (выполните login): %w", err)
		}

		client, conn, err := newKeeperClient(serverAddress)
		if err != nil {
			return fmt.Errorf("подключение к серверу: %w", err)
		}
		defer func() { _ = conn.Close() }()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ctx = withToken(ctx, token)

		resp, err := client.GetItem(ctx, pb.GetItemRequest_builder{
			Id: &id,
		}.Build())
		if err != nil {
			switch status.Code(err) {
			case codes.Unauthenticated:
				return errors.New("не авторизованы — выполните login заново")
			case codes.NotFound:
				return errors.New("запись не найдена")
			default:
				return fmt.Errorf("не удалось получить запись: %w", err)
			}
		}

		item := resp.GetItem()
		fmt.Printf("id:       %s\n", item.GetId())
		fmt.Printf("type:     %s\n", item.GetType())
		fmt.Printf("metadata: %s\n", item.GetMetadata())
		fmt.Printf("version:  %d\n", item.GetVersion())
		fmt.Printf("data:     %s\n", string(item.GetEncryptedPayload()))
		return nil
	},
}

func init() {
	getCmd.Flags().String("id", "", "id записи")
	rootCmd.AddCommand(getCmd)
}
