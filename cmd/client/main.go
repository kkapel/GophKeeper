// Package main реализует клиент менеджера паролей GophKeeper.
package main

import (
	"log"

	"github.com/kkapel/gophkeeper/internal/client"
)

// Дефолтные значения, присваивыемые переменным уровня пакета при их объявлении, могут быть перезаписаны на этапе линковки
// с помощью флагов -ldflags
//
//	go build -ldflags "-X main.buildVersion=v1.0.0 -X main.buildDate=2026-07-02"
var (
	buildVersion = "N/A"
	buildDate    = "N/A"
)

func main() {
	client.SetBuildInfo(buildVersion, buildDate) // прокидываем инфу о сборке
	if err := client.Execute(); err != nil {
		log.Fatal(err)
	}
}
