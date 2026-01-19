package api

import (
	"net/http"

	"github.com/fr11c/yandex-go-final-project/pkg/db"
)

// TasksResp представляет ответ со списком задач
type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

// tasksHandler обрабатывает GET-запросы для получения списка задач
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := db.Tasks(50)
	if err != nil {
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, TasksResp{
		Tasks: tasks,
	})
}
