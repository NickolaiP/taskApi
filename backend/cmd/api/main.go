package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"taskApi/internal/config"
	"taskApi/internal/database"
	"taskApi/internal/hand"
	"taskApi/internal/logger"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	// Загружаем конфигурацию приложения
	cfg := config.LoadConfig()

	// Инициализируем логгер для записи логов в стандартный вывод (stdout)
	logger := logger.InitLogger(os.Stdout)

	// Подключаемся к базе данных PostgreSQL с использованием настроек из конфигурации
	db, err := database.NewPostgresDB(cfg.DB)
	if err != nil {
		// Логируем ошибку при подключении к базе данных и выходим из программы
		logger.Error("Failed to connect to database", "error", err)
		return
	}
	// Закрываем соединение с базой данных при завершении программы
	defer db.Close()

	// Выполняем миграции базы данных для обновления её структуры
	database.RunMigrations(db)

	// Создаём новый маршрутизатор для обработки HTTP-запросов
	r := mux.NewRouter()

	// Инициализируем обработчик задач с подключением к базе данных и логгером
	taskHandler := hand.NewTaskHandler(db, logger)

	// Настраиваем маршруты для работы с задачами
	// Создание новой задачи
	r.HandleFunc("/tasks", taskHandler.CreateTask).Methods("POST")
	// Получение всех задач
	r.HandleFunc("/tasks", taskHandler.GetTasks).Methods("GET")
	// Получение задачи по ID
	r.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.GetTaskByID).Methods("GET")
	// Обновление задачи по ID
	r.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.UpdateTask).Methods("PUT")
	// Удаление задачи по ID
	r.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.DeleteTask).Methods("DELETE")

	// Создаём HTTP-сервер с конфигурацией CORS и маршрутизатором
	server := &http.Server{
		Addr: ":8000", // Адрес, на котором будет запущен сервер
		Handler: handlers.CORS(
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}), // Разрешённые методы HTTP
			handlers.AllowedHeaders([]string{"Authorization", "Content-Type"}),           // Разрешённые заголовки
		)(r), // Передача маршрутизатора в качестве обработчика запросов
	}

	// Запуск сервера в отдельной горутине, чтобы не блокировать основной поток
	go func() {
		logger.Info("Server started on :8000")
		// Запуск HTTP-сервера и логирование ошибок, если сервер не может быть запущен
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Could not listen on :8000", "error", err)
		}
	}()

	// Создание канала для получения сигналов прерывания (например, Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit // Ожидание сигнала прерывания

	// Создаём контекст с таймаутом для корректного завершения работы сервера
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Завершаем работу сервера с использованием созданного контекста
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	// Логируем сообщение о завершении работы сервера
	logger.Info("Server exiting")
}
