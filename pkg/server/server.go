package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/fr11c/yandex-go-final-project/pkg/api"
)

const (
	defaultPort = 7540
	webDir      = "./web"
)

// Start запускает веб-сервер на указанном порту или порту по умолчанию
func Start() error {
	port := getPort()

	api.Init()

	http.Handle("/", http.FileServer(http.Dir(webDir)))

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Запуск сервера на http://localhost%s", addr)

	return http.ListenAndServe(addr, nil)
}

// getPort возвращает порт из переменной окружения TODO_PORT или порт по умолчанию
func getPort() int {
	envPort := os.Getenv("TODO_PORT")
	if envPort != "" {
		if port, err := strconv.Atoi(envPort); err == nil && port > 0 && port <= 65535 {
			return port
		}
		log.Printf("Предупреждение: некорректное значение TODO_PORT=%s, используется порт по умолчанию %d", envPort, defaultPort)
	}
	return defaultPort
}
