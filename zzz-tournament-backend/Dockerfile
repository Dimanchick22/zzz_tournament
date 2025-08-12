# Dockerfile для ZZZ Tournament Backend

# Многоэтапная сборка
FROM golang:1.21-alpine AS builder

# Установка необходимых пакетов
RUN apk add --no-cache git ca-certificates tzdata

# Рабочая директория
WORKDIR /app

# Копируем go mod файлы
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# Финальная стадия
FROM alpine:latest AS production

# Установка ca-certificates для HTTPS запросов
RUN apk --no-cache add ca-certificates wget

# Создание пользователя для безопасности
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Рабочая директория
WORKDIR /app

# Копируем бинарный файл из builder стадии
COPY --from=builder /app/main .

# Копируем миграции
COPY --from=builder /app/internal/db/migrations ./migrations

# Создаем директории для логов
RUN mkdir -p logs && \
    chown -R appuser:appgroup /app

# Переключаемся на непривилегированного пользователя
USER appuser

# Порт приложения
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Запуск приложения
CMD ["./main"]

# Стадия для разработки
FROM golang:1.21-alpine AS development

# Установка необходимых пакетов для разработки
RUN apk add --no-cache git ca-certificates tzdata make curl

# Установка air для hot reload
RUN go install github.com/cosmtrek/air@latest

# Рабочая директория
WORKDIR /app

# Копируем все файлы
COPY . .

# Загружаем зависимости
RUN go mod download

# Порт приложения
EXPOSE 8080

# Запуск с hot reload
CMD ["air"]