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

// deleteCmd удаляет запись по id.
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Удалить запись по id",
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

		_, err = client.DeleteItem(ctx, pb.DeleteItemRequest_builder{
			Id: &id,
		}.Build())
		if err != nil {
			if status.Code(err) == codes.Unauthenticated {
				return errors.New("не авторизованы — выполните login заново")
			}
			return fmt.Errorf("не удалось удалить запись: %w", err)
		}

		fmt.Println("Запись удалена.")
		return nil
	},
}

func init() {
	deleteCmd.Flags().String("id", "", "id записи")
	rootCmd.AddCommand(deleteCmd)
}
