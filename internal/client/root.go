// Package client реализует CLI-клиент менеджера паролей GophKeeper.
package client

import "github.com/spf13/cobra"

var (
	serverAddress string
	certPath      string
)

// NewRootCmd — корневая команда приложения.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper — клиент менеджера паролей",
	}

	// persistent-флаги корня
	root.PersistentFlags().StringVar(&serverAddress, "address", "127.0.0.1:8080", "адрес gRPC-сервера")
	root.PersistentFlags().StringVar(&certPath, "cert", "", "путь к TLS-сертификату сервера")

	// явно добавляем все подкоманды
	root.AddCommand(
		newRegisterCmd(),
		newLoginCmd(),
		newAddCmd(),
		newGetCmd(),
		newListCmd(),
		newDeleteCmd(),
		newVersionCmd(),
	)

	return root
}
