// Package main реализует сервер менеджера паролей GophKeeper.
package main

import (
	"fmt"
	"log"

	"github.com/kkapel/gophkeeper/internal/server"
)

// Дефолтные значения, присваиваемые переменным уровня пакета при их объявлении, могут быть перезаписаны на этапе линковки
// с помощью флагов -ldflags
//
//	go build -ldflags "-X main.buildVersion=v1.0.0 -X main.buildDate=2026-07-02"
var (
	buildVersion = "N/A"
	buildDate    = "N/A"
)

func main() {
	// stdout для отображения информации о сборке при запуске приложения
	fmt.Println("Build version:", buildVersion)
	fmt.Println("Build date:", buildDate)

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
