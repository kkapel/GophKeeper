// Package client реализует CLI-клиент менеджера паролей GophKeeper.
package client

import "github.com/spf13/cobra"

// rootCmd — корневая команда приложения.
var rootCmd = &cobra.Command{
	Use:   "gophkeeper",
	Short: "GophKeeper — клиент менеджера паролей",
	Long:  "Клиент-серверный менеджер паролей для безопасного хранения приватных данных.",
}

var serverAddress string

func init() {
	rootCmd.PersistentFlags().StringVarP(&serverAddress, "server", "address", "localhost:8080", "Адрес grpc-сервера в формате host:port")
}

// Execute запускает разбор и выполнение команд.
func Execute() error {
	return rootCmd.Execute()
}
