package main

import (
	"log"

	"github.com/fr11c/yandex-go-final-project/pkg/db"
	"github.com/fr11c/yandex-go-final-project/pkg/server"
)

func main() {
	if err := db.Init(""); err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}
	defer db.Close()

	if err := server.Start(); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
