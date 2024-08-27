package database

import (
	"context"
	"log"
	"time"
)

// RunMigrations выполняет миграции базы данных, создавая необходимые таблицы,
// если они еще не существуют. Это необходимо для обеспечения структуры
// базы данных перед запуском приложения.
func RunMigrations(db Database) {
	// SQL-запрос для создания таблицы задач.
	taskTable := `CREATE TABLE IF NOT EXISTS tasks (
        id SERIAL PRIMARY KEY,
        title VARCHAR(255) NOT NULL,
        description TEXT NOT NULL,
        due_date TIMESTAMP,
        created_at TIMESTAMP NOT NULL,
        updated_at TIMESTAMP NOT NULL
    );`

	// Создание контекста с таймаутом для выполнения SQL-запросов.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Выполнение SQL-запроса для создания таблицы задач.
	_, err := db.Exec(ctx, taskTable)
	if err != nil {
		log.Fatal(err)
	}
}
