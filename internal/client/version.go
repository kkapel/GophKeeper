package client

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Переменные сборки, пробрасываются сюда из main через SetBuildInfo.
var (
	buildVersion = "N/A"
	buildDate    = "N/A"
)

// SetBuildInfo передаёт информацию о сборке из main в пакет команд.
func SetBuildInfo(version, date string) {
	buildVersion = version
	buildDate = date
}

// newVersionCmd выводит версию и дату сборки клиента.
func newVersionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Показать версию и дату сборки",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Build version:", buildVersion)
			fmt.Println("Build date:", buildDate)
		},
	}

	return versionCmd
}
