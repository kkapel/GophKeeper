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

// itemTypeFromString переводит человекочитаемый тип в enum ItemType.
func itemTypeFromString(s string) (pb.ItemType, error) {
	switch s {
	case "credentials":
		return pb.ItemType_ITEM_TYPE_CREDENTIALS, nil
	case "text":
		return pb.ItemType_ITEM_TYPE_TEXT, nil
	case "binary":
		return pb.ItemType_ITEM_TYPE_BINARY, nil
	case "card":
		return pb.ItemType_ITEM_TYPE_CARD, nil
	default:
		return pb.ItemType_ITEM_TYPE_UNSPECIFIED, fmt.Errorf("неизвестный тип: %s (допустимо: credentials, text, binary, card)", s)
	}
}

// addCmd добавляет новую запись приватных данных.
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Добавить новую запись",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Тип записи
		typeStr, err := cmd.Flags().GetString("type")
		if err != nil {
			return err
		}
		itemType, err := itemTypeFromString(typeStr)
		if err != nil {
			return err
		}

		// Данные и метаинформация
		data, err := cmd.Flags().GetString("data")
		if err != nil {
			return err
		}
		if data == "" {
			return errors.New("укажите данные через --data")
		}
		meta, err := cmd.Flags().GetString("meta")
		if err != nil {
			return err
		}

		// payload
		payload := []byte(data)

		// Токен из файла
		token, err := loadToken()
		if err != nil {
			return fmt.Errorf("не удалось прочитать токен (выполните login): %w", err)
		}

		// Подключение к серверу
		client, conn, err := newKeeperClient(serverAddress)
		if err != nil {
			return fmt.Errorf("подключение к серверу: %w", err)
		}
		defer func() { _ = conn.Close() }()

		// Вызов CreateItem с токеном в metadata
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ctx = withToken(ctx, token)

		resp, err := client.CreateItem(ctx, pb.CreateItemRequest_builder{
			Type:             &itemType,
			EncryptedPayload: payload,
			Metadata:         &meta,
		}.Build())

		if err != nil {
			// разбираем gRPC-код
			if status.Code(err) == codes.Unauthenticated {
				return errors.New("не авторизованы — выполните login заново")
			}
			return fmt.Errorf("не удалось создать запись: %w", err)
		}

		fmt.Printf("Запись создана: id=%s\n", resp.GetId())
		return nil

	},
}

func init() {
	addCmd.Flags().String("type", "text", "тип записи: credentials, text, binary, card")
	addCmd.Flags().String("data", "", "данные для сохранения")
	addCmd.Flags().String("meta", "", "метаинформация (сайт, банк и т.п.)")
	rootCmd.AddCommand(addCmd)
}
