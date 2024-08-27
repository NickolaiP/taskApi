package hand

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"taskApi/internal/database"
	"taskApi/internal/logger"
	"taskApi/internal/models"

	"github.com/gorilla/mux"
)

// taskHandler представляет собой структуру обработчика для управления задачами.
// Включает в себя подключение к базе данных и логгер.
type taskHandler struct {
	db     database.Database
	logger *logger.Logger
}

// NewTaskHandler создает новый экземпляр taskHandler с заданными базой данных и логгером.
func NewTaskHandler(db database.Database, logger *logger.Logger) *taskHandler {
	return &taskHandler{
		db:     db,
		logger: logger,
	}
}

// CreateTask обрабатывает запрос на создание новой задачи.
// Декодирует тело запроса в структуру задачи, сохраняет задачу в базе данных
// и возвращает созданную задачу в формате JSON.
func (h *taskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	// Декодируем JSON-запрос в структуру task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		// Возвращаем ошибку при некорректном запросе
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Создаем контекст с таймаутом для операции с базой данных
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Устанавливаем время создания и обновления задачи
	task.CreatedAt = time.Now().Format(time.RFC3339)
	task.UpdatedAt = task.CreatedAt

	// Выполняем запрос на вставку новой задачи в базу данных и получаем её ID
	err := h.db.QueryRow(ctx, "INSERT INTO tasks (title, description, due_date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		task.Title, task.Description, task.DueDate, task.CreatedAt, task.UpdatedAt).Scan(&task.ID)
	if err != nil {
		// Возвращаем ошибку сервера, если вставка не удалась
		http.Error(w, "Error creating task", http.StatusInternalServerError)
		return
	}

	// Устанавливаем статус ответа как Created и возвращаем созданную задачу
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// GetTasks обрабатывает запрос на получение списка всех задач.
// Выполняет запрос к базе данных и возвращает задачи в формате JSON.
func (h *taskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	// Создаем контекст с таймаутом для операции с базой данных
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Выполняем запрос на выборку всех задач из базы данных
	rows, err := h.db.Query(ctx, "SELECT id, title, description, due_date, created_at, updated_at FROM tasks")
	if err != nil {
		// Возвращаем ошибку сервера при сбое запроса
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []models.Task
	// Итерируем по результатам выборки и заполняем срез задач
	for rows.Next() {
		var task models.Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.CreatedAt, &task.UpdatedAt); err != nil {
			// Возвращаем ошибку сервера при сбое сканирования
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}

	// Возвращаем задачи в формате JSON
	json.NewEncoder(w).Encode(tasks)
}

// GetTaskByID обрабатывает запрос на получение задачи по её ID.
// Выполняет запрос к базе данных и возвращает задачу в формате JSON.
func (h *taskHandler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	// Создаем контекст с таймаутом для операции с базой данных
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Извлекаем ID задачи из параметров запроса
	vars := mux.Vars(r)
	taskID, err := strconv.Atoi(vars["id"])
	if err != nil {
		// Возвращаем ошибку при некорректном ID
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var task models.Task
	// Выполняем запрос на выборку задачи по ID
	err = h.db.QueryRow(ctx, "SELECT id, title, description, due_date, created_at, updated_at FROM tasks WHERE id=$1", taskID).
		Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		// Возвращаем ошибку, если задача не найдена
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Возвращаем найденную задачу в формате JSON
	json.NewEncoder(w).Encode(task)
}

// UpdateTask обрабатывает запрос на обновление задачи по её ID.
// Декодирует тело запроса, обновляет соответствующую запись в базе данных
// и возвращает обновленную задачу в формате JSON.
func (h *taskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	// Декодируем JSON-запрос в структуру task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		// Возвращаем ошибку при некорректном запросе
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Создаем контекст с таймаутом для операции с базой данных
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Извлекаем ID задачи из параметров запроса
	vars := mux.Vars(r)
	taskID, err := strconv.Atoi(vars["id"])
	if err != nil {
		// Возвращаем ошибку при некорректном ID
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// Получаем существующую задачу для сохранения её поля CreatedAt
	var existingTask models.Task
	err = h.db.QueryRow(ctx, "SELECT created_at FROM tasks WHERE id=$1", taskID).Scan(&existingTask.CreatedAt)
	if err != nil {
		// Возвращаем ошибку, если задача не найдена
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Обновляем время изменения задачи
	task.UpdatedAt = time.Now().Format(time.RFC3339)

	// Обновляем запись задачи в базе данных
	_, err = h.db.Exec(ctx, "UPDATE tasks SET title=$1, description=$2, due_date=$3, updated_at=$4 WHERE id=$5",
		task.Title, task.Description, task.DueDate, task.UpdatedAt, taskID)
	if err != nil {
		// Возвращаем ошибку сервера при сбое обновления
		http.Error(w, "Error updating task", http.StatusInternalServerError)
		return
	}

	// Возвращаем обновленную задачу с сохранением оригинального поля CreatedAt
	task.CreatedAt = existingTask.CreatedAt
	task.ID = taskID
	json.NewEncoder(w).Encode(task)
}

// DeleteTask обрабатывает запрос на удаление задачи по её ID.
// Выполняет запрос к базе данных для удаления задачи.
func (h *taskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	// Создаем контекст с таймаутом для операции с базой данных
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Извлекаем ID задачи из параметров запроса
	vars := mux.Vars(r)
	taskID, err := strconv.Atoi(vars["id"])
	if err != nil {
		// Возвращаем ошибку при некорректном ID
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// Выполняем запрос на удаление задачи по ID
	_, err = h.db.Exec(ctx, "DELETE FROM tasks WHERE id=$1", taskID)
	if err != nil {
		// Возвращаем ошибку сервера при сбое удаления
		http.Error(w, "Error deleting task", http.StatusInternalServerError)
		return
	}

	// Устанавливаем статус ответа как No Content (204) при успешном удалении
	w.WriteHeader(http.StatusNoContent)
}
