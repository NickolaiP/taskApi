
# Инструкция по использованию сервиса

## Запуск сервиса

1. Клонируем [репозиторий](https://github.com/NickolaiP/taskApi) с github
```
git clone https://github.com/NickolaiP/taskApi.git
```

2. Переходим в директорию приложения
```
cd taskApi
```

3. Запускаем docker-compose
```
docker-compose up
```

4. Ждем вывод в консоль с информацией о том, что приложение успешно запущено
```
Server started on :8000
```

## Выполнение комманд

**Вместо {id} укажите айди интересующей вас задачи**

1. Создание задачи:
```
curl -X POST http://localhost:8000/tasks \
-H "Content-Type: application/json" \
-d '{
  "title": "Заголовок задачи",
  "description": "Описание задачи",
  "due_date": "2024-12-31T23:59:59Z"
}'
```

2. Получение списка задач:
```
curl -X GET http://localhost:8000/tasks
```

3. Получение задачи по ID:
```
curl -X GET http://localhost:8000/tasks/{id}
```

4. Обновление задачи:
```
curl -X PUT http://localhost:8000/tasks/{id} \
-H "Content-Type: application/json" \
-d '{
  "title": "Обновленный заголовок",
  "description": "Обновленное описание",
  "due_date": "2025-01-15T23:59:59Z"
}'
```

5. Удаление задачи:
```
curl -X DELETE http://localhost:8000/tasks/{id}
```