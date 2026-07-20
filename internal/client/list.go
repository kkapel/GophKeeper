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

// listCmd выводит все записи пользователя.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Показать все записи",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := loadToken()
		if err != nil {
			return fmt.Errorf("не удалось прочитать токен (выполните login): %w", err)
		}

		client, conn, err := newKeeperClient(serverAddress, certPath)
		if err != nil {
			return fmt.Errorf("подключение к серверу: %w", err)
		}
		defer func() { _ = conn.Close() }()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ctx = withToken(ctx, token)

		resp, err := client.ListItems(ctx, pb.ListItemsRequest_builder{}.Build())
		if err != nil {
			if status.Code(err) == codes.Unauthenticated {
				return errors.New("не авторизованы — выполните login заново")
			}
			return fmt.Errorf("не удалось получить список: %w", err)
		}

		items := resp.GetItems()
		if len(items) == 0 {
			fmt.Println("Записей нет.")
			return nil
		}

		fmt.Printf("Найдено записей: %d\n\n", len(items))
		for _, item := range items {
			fmt.Printf("id: %s | type: %s | meta: %s | v%d\n",
				item.GetId(), item.GetType(), item.GetMetadata(), item.GetVersion())
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
