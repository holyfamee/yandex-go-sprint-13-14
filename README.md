# ToDo List

Веб приложение для планирования задач

# локальный запуск
git clone https://github.com/holyfamee/yandex-go-sprint-13-14

```bash
go mod download
```

```bash
go run .
```

http://localhost:7540

# Docker

```bash
docker build -t todo-scheduler .
```

Простой запуск:
```bash
docker run -p 7540:7540 todo-scheduler
```

С подключением внешней базы данных:
```bash
docker run -p 7540:7540 \
  -v /path/to/host/data:/app/data \
  todo-scheduler
```

С аутентификацией:
```bash
docker run -p 7540:7540 \
  -e TODO_PASSWORD=mypassword \
  -v /path/to/host/data:/app/data \
  todo-scheduler
```