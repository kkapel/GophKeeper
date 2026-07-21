package client

import (
	"context"
	"fmt"
	"time"

	pb "github.com/kkapel/gophkeeper/internal/proto/gophkeeper/v1"
	"github.com/spf13/cobra"
)

// newLoginCmd реализует команду авторизации пользователя
func newLoginCmd() *cobra.Command {
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Войти в систему",
		Long:  `Войти в систему, используя ваши учетные данные.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			login, err := cmd.Flags().GetString("login")
			if err != nil {
				return err
			}
			if login == "" {
				return fmt.Errorf("укажите логин через --login")
			}

			password, err := readPassword("Введите пароль: ")
			if err != nil {
				return err
			}

			if password == "" {
				return fmt.Errorf("укажите пароль через --password")
			}

			auth, conn, err := newAuthClient(serverAddress, certPath)
			if err != nil {
				return fmt.Errorf("не удалось подключиться к серверу: %w", err)
			}

			defer func() { _ = conn.Close() }()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resp, err := auth.Login(ctx, pb.LoginRequest_builder{
				Login:    &login,
				Password: &password,
			}.Build())
			if err != nil {
				return fmt.Errorf("не удалось войти в систему: %w", err)
			}

			// Сохраняем токен
			if err := saveToken(resp.GetAccessToken()); err != nil {
				return fmt.Errorf("не удалось сохранить токен: %w", err)
			}

			fmt.Printf("Успешный вход в систему. Токен сохранен.\n")

			return nil
		}}

	loginCmd.Flags().String("login", "", "Логин пользователя")

	return loginCmd
}
