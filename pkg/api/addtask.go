package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/fr11c/yandex-go-final-project/pkg/db"
)

// getTaskHandler обрабатывает GET-запросы для получения задачи по ID
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, map[string]string{"error": "Не указан идентификатор"}, http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJSON(w, map[string]string{"error": err.Error()}, http.StatusNotFound)
		return
	}

	writeJSON(w, task, http.StatusOK)
}

// addTaskHandler обрабатывает POST-запросы для добавления новой задачи
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Ошибка чтения тела запроса: %v", err)
		writeJSON(w, map[string]string{"error": "Ошибка чтения запроса"}, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &task); err != nil {
		log.Printf("Ошибка десериализации JSON: %v", err)
		writeJSON(w, map[string]string{"error": "Ошибка десериализации JSON"}, http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		writeJSON(w, map[string]string{"error": "Не указан заголовок задачи"}, http.StatusBadRequest)
		return
	}

	if err := checkDate(&task); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		writeJSON(w, map[string]string{"error": fmt.Sprintf("Ошибка добавления задачи: %v", err)}, http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"id": fmt.Sprintf("%d", id)}, http.StatusOK)
}

// updateTaskHandler обрабатывает PUT-запросы для обновления существующей задачи
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Ошибка чтения тела запроса: %v", err)
		writeJSON(w, map[string]string{"error": "Ошибка чтения запроса"}, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &task); err != nil {
		log.Printf("Ошибка десериализации JSON: %v", err)
		writeJSON(w, map[string]string{"error": "Ошибка десериализации JSON"}, http.StatusBadRequest)
		return
	}

	if task.ID == "" {
		writeJSON(w, map[string]string{"error": "Не указан идентификатор"}, http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		writeJSON(w, map[string]string{"error": "Не указан заголовок задачи"}, http.StatusBadRequest)
		return
	}

	if err := checkDate(&task); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	if err := db.UpdateTask(&task); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()}, http.StatusNotFound)
		return
	}

	writeJSON(w, map[string]string{}, http.StatusOK)
}

// beforeToday проверяет, что дата строго меньше сегодняшней (не включая сегодня)
func beforeToday(date, now time.Time) bool {
	y1, m1, d1 := date.Date()
	y2, m2, d2 := now.Date()

	if y1 < y2 {
		return true
	}
	if y1 == y2 && m1 < m2 {
		return true
	}
	if y1 == y2 && m1 == m2 && d1 < d2 {
		return true
	}
	return false
}

// checkDate проверяет и корректирует дату задачи
func checkDate(task *db.Task) error {
	now := time.Now()

	if task.Date == "" {
		task.Date = now.Format(dateFormat)
	}

	t, err := time.Parse(dateFormat, task.Date)
	if err != nil {
		return errors.New("дата представлена в неверном формате")
	}

	var next string
	if task.Repeat != "" {
		next, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return fmt.Errorf("некорректное правило повторения: %v", err)
		}
	}

	if beforeToday(t, now) {
		if task.Repeat == "" {
			task.Date = now.Format(dateFormat)
		} else {
			task.Date = next
		}
	}

	return nil
}

// doneTaskHandler обрабатывает POST-запросы для отметки задачи как выполненной
func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, map[string]string{"error": "Не указан идентификатор"}, http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJSON(w, map[string]string{"error": err.Error()}, http.StatusNotFound)
		return
	}

	if task.Repeat == "" {
		if err := db.DeleteTask(id); err != nil {
			writeJSON(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]string{}, http.StatusOK)
		return
	}

	now := time.Now()
	next, err := NextDate(now, task.Date, task.Repeat)
	if err != nil {
		writeJSON(w, map[string]string{"error": fmt.Sprintf("Ошибка вычисления следующей даты: %v", err)}, http.StatusBadRequest)
		return
	}

	if err := db.UpdateDate(next, id); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{}, http.StatusOK)
}

// deleteTaskHandler обрабатывает DELETE-запросы для удаления задачи
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, map[string]string{"error": "Не указан идентификатор"}, http.StatusBadRequest)
		return
	}

	if err := db.DeleteTask(id); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()}, http.StatusNotFound)
		return
	}

	writeJSON(w, map[string]string{}, http.StatusOK)
}

// writeJSON записывает данные в формате JSON в HTTP-ответ
func writeJSON(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Ошибка сериализации JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Ошибка сериализации JSON"}`))
		return
	}

	w.WriteHeader(statusCode)
	w.Write(jsonData)
}
