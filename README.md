# ZZZ Tournament Backend

Бэкенд для турнирной системы по игре Zenless Zone Zero.

## Возможности

- 🔐 Аутентификация пользователей (регистрация/логин)
- 👥 Система рейтингов и лидерборд
- 🦸 Управление героями ZZZ
- 🏠 Создание и управление комнатами
- 🏆 Турнирная система с генерацией сетки
- 💬 Чат в реальном времени (WebSocket)
- 📊 Статистика побед/поражений

## Технологии

- **Backend**: Go + Gin
- **Database**: PostgreSQL
- **WebSocket**: Gorilla WebSocket
- **Auth**: JWT tokens
- **Migrations**: golang-migrate

## Установка и запуск

1. Клонируем репозиторий
2. Устанавливаем зависимости:
```bash
go mod download
```

3. Настраиваем базу данных PostgreSQL и создаем файл .env:
```env
DATABASE_URL=postgres://user:password@localhost/zzz_tournament?sslmode=disable
JWT_SECRET=your-super-secret-jwt-key-here
PORT=8080
```

4. Запускаем сервер:
```bash
go run cmd/server/main.go
```

## API Endpoints

### Auth
- `POST /api/v1/auth/register` - Регистрация
- `POST /api/v1/auth/login` - Авторизация
- `POST /api/v1/auth/refresh` - Обновление токена

### User
- `GET /api/v1/profile` - Получить профиль
- `PUT /api/v1/profile` - Обновить профиль
- `GET /api/v1/leaderboard` - Лидерборд

### Heroes
- `GET /api/v1/heroes` - Список героев
- `POST /api/v1/heroes` - Создать героя (админ)
- `PUT /api/v1/heroes/:id` - Обновить героя (админ)
- `DELETE /api/v1/heroes/:id` - Удалить героя (админ)

### Rooms
- `GET /api/v1/rooms` - Список комнат
- `POST /api/v1/rooms` - Создать комнату
- `GET /api/v1/rooms/:id` - Получить комнату
- `PUT /api/v1/rooms/:id` - Обновить комнату
- `DELETE /api/v1/rooms/:id` - Удалить комнату
- `POST /api/v1/rooms/:id/join` - Присоединиться к комнате
- `POST /api/v1/rooms/:id/leave` - Покинуть комнату

### Tournaments
- `POST /api/v1/rooms/:id/tournament/start` - Запустить турнир
- `GET /api/v1/tournaments/:id` - Получить турнир
- `POST /api/v1/tournaments/:id/matches/:match_id/result` - Отправить результат матча

### Chat
- `GET /api/v1/rooms/:id/messages` - Получить сообщения комнаты

### WebSocket
- `GET /ws` - WebSocket соединение

## WebSocket Events

### Отправляемые клиентом:
- `join_room` - Присоединиться к комнате
- `leave_room` - Покинуть комнату
- `chat_message` - Отправить сообщение в чат

### Получаемые от сервера:
- `room_updated` - Обновление комнаты
- `tournament_started` - Турнир начался
- `chat_message` - Новое сообщение в чате
- `match_assigned` - Назначен матч

## Структура базы данных

### Основные таблицы:
- `users` - Пользователи с рейтингом и статистикой
- `heroes` - Герои ZZZ с характеристиками
- `rooms` - Игровые комнаты
- `room_participants` - Участники комнат
- `tournaments` - Турниры
- `matches` - Матчи турниров
- `messages` - Сообщения чата

## Игровая логика

### Рейтинговая система:
- Начальный рейтинг: 1000
- За победу: +25 рейтинга
- За поражение: -15 рейтинга (минимум 0)

### Турнирная система:
- Поддержка от 2 до 16 участников
- Система на выбывание (single elimination)
- Автоматическая генерация сетки
- Отслеживание результатов матчей

### Комнаты:
- Публичные и приватные комнаты
- Максимум участников настраивается
- Хост может управлять комнатой
- Автоматическое удаление пустых комнат

## Дополнительные файлы конфигурации

### docker-compose.yml
```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: zzz_tournament
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  app:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - postgres
    environment:
      DATABASE_URL: postgres://user:password@postgres/zzz_tournament?sslmode=disable
      JWT_SECRET: your-super-secret-jwt-key-here
    volumes:
      - ./.env:/app/.env

volumes:
  postgres_data:
```

### Dockerfile
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/internal/db/migrations ./internal/db/migrations

CMD ["./main"]
```

### Makefile
```makefile
.PHONY: build run test migrate-up migrate-down docker-up docker-down

build:
	go build -o bin/server cmd/server/main.go

run:
	go run cmd/server/main.go

test:
	go test ./...

migrate-up:
	migrate -path internal/db/migrations -database "postgres://user:password@localhost/zzz_tournament?sslmode=disable" up

migrate-down:
	migrate -path internal/db/migrations -database "postgres://user:password@localhost/zzz_tournament?sslmode=disable" down

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

install-deps:
	go mod download
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Примеры использования API

### Регистрация пользователя:
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "player1",
    "email": "player1@example.com",
    "password": "password123"
  }'
```

### Создание комнаты:
```bash
curl -X POST http://localhost:8080/api/v1/rooms \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "name": "Epic Tournament Room",
    "description": "Only pros allowed",
    "max_players": 8,
    "is_private": false
  }'
```

### Запуск турнира:
```bash
curl -X POST http://localhost:8080/api/v1/rooms/1/tournament/start \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Расширения и улучшения

Для production-ready версии рекомендуется добавить:

1. **Кэширование**: Redis для сессий и часто запрашиваемых данных
2. **Логирование**: Структурированные логи с уровнями
3. **Мониторинг**: Prometheus метрики + Grafana
4. **Rate Limiting**: Ограничение запросов на пользователя
5. **Валидация**: Более строгая валидация входных данных
6. **Тесты**: Unit и integration тесты
7. **CI/CD**: Автоматические деплойменты
8. **Backup**: Регулярные бэкапы базы данных
9. **Admin Panel**: Веб-интерфейс для администрирования
10. **Push уведомления**: Уведомления о начале матчей

## Лицензия

MIT License