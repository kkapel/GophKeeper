// Package client реализует CLI-клиент менеджера паролей GophKeeper.
package client

import "github.com/spf13/cobra"

// clientConfig содержит общие настройки подключения к серверу.
type clientConfig struct {
	serverAddress string
	certPath      string
}

// NewRootCmd — корневая команда приложения.
func NewRootCmd() *cobra.Command {
	cfg := &clientConfig{}

	root := &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper — клиент менеджера паролей",
	}

	// persistent-флаги корня
	root.PersistentFlags().StringVar(&cfg.serverAddress, "address", "127.0.0.1:8080", "адрес gRPC-сервера")
	root.PersistentFlags().StringVar(&cfg.certPath, "cert", "", "путь к TLS-сертификату сервера")

	// явно добавляем все подкоманды
	root.AddCommand(
		newRegisterCmd(cfg),
		newLoginCmd(cfg),
		newAddCmd(cfg),
		newGetCmd(cfg),
		newListCmd(cfg),
		newDeleteCmd(cfg),
		newVersionCmd(),
	)

	return root
}
