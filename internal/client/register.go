package client

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	pb "github.com/kkapel/gophkeeper/internal/proto/gophkeeper/v1"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// registerCmd реализует команду регистрации нового пользователя.
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Зарегистрировать нового пользователя",
	Long:  "Команда для регистрации нового пользователя в системе GophKeeper.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Читаем логин и пароль из флагов
		login, err := cmd.Flags().GetString("login")
		if err != nil {
			return err
		}

		if login == "" {
			return errors.New("укажите логин через --login")
		}

		// Читаем пароль скрытым вводом
		password, err := readPassword("Введите пароль: ")
		if err != nil {
			return err
		}
		if password == "" {
			return errors.New("пароль не может быть пустым")
		}

		// Подключаемся к gRPC-серверу
		authClient, conn, err := newAuthClient(serverAddress, certPath)
		if err != nil {
			return fmt.Errorf("не удалось подключиться к серверу: %v", err)
		}
		defer func() { _ = conn.Close() }()

		// Вызываем метод регистрации
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := authClient.Register(ctx, pb.RegisterRequest_builder{
			Login:    &login,
			Password: &password,
		}.Build())

		if err != nil {
			return fmt.Errorf("ошибка регистрации: %v", err)
		}

		// Сохраняем токен
		if err := saveToken(resp.GetAccessToken()); err != nil {
			return fmt.Errorf("не удалось сохранить токен: %v", err)
		}

		fmt.Println("Регистрация успешна! Токен сохранён.")
		return nil

	},
}

// readPassword читает пароль из терминала без отображения введённых символов.
func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println() // перенос строки после ввода
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// init подключает команду register к корневой команде.
func init() {
	registerCmd.Flags().String("login", "", "Логин для регистрации")
	rootCmd.AddCommand(registerCmd)
}
