# Makefile для ZZZ Tournament Backend

# Переменные
BINARY_NAME=zzz-tournament-backend
MAIN_PATH=cmd/server/main.go
BUILD_DIR=build
GO_VERSION=1.21
DOCKER_IMAGE=zzz-tournament
DOCKER_TAG=latest

# Go параметры
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Цвета для вывода
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

.PHONY: help build run test clean deps tidy fmt vet lint security docker-build docker-run docker-stop migrate-up migrate-down dev prod install-tools

# Помощь
help: ## Показать это сообщение помощи
	@echo "$(GREEN)ZZZ Tournament Backend Makefile$(NC)"
	@echo ""
	@echo "$(YELLOW)Основные команды:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Установка зависимостей
deps: ## Установить зависимости
	@echo "$(YELLOW)Установка зависимостей...$(NC)"
	$(GOGET) -d ./...
	@echo "$(GREEN)Зависимости установлены$(NC)"

# Обновление go.mod
tidy: ## Очистить и обновить зависимости
	@echo "$(YELLOW)Обновление go.mod...$(NC)"
	$(GOMOD) tidy
	@echo "$(GREEN)go.mod обновлен$(NC)"

# Форматирование кода
fmt: ## Форматировать код
	@echo "$(YELLOW)Форматирование кода...$(NC)"
	$(GOFMT) ./...
	@echo "$(GREEN)Код отформатирован$(NC)"

# Проверка кода
vet: ## Проверить код с go vet
	@echo "$(YELLOW)Проверка кода...$(NC)"
	$(GOCMD) vet ./...
	@echo "$(GREEN)Проверка завершена$(NC)"

# Сборка приложения
build: fmt vet ## Собрать приложение
	@echo "$(YELLOW)Сборка приложения...$(NC)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -a -installsuffix cgo -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)Приложение собрано: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

# Сборка для разработки
build-dev: ## Собрать для разработки
	@echo "$(YELLOW)Сборка для разработки...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)Приложение собрано: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

# Запуск в режиме разработки
run: ## Запустить сервер в режиме разработки
	@echo "$(YELLOW)Запуск сервера...$(NC)"
	$(GOCMD) run $(MAIN_PATH)

# Запуск с hot reload (требует air)
dev: ## Запустить с hot reload (требует air)
	@echo "$(YELLOW)Запуск с hot reload...$(NC)"
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "$(RED)air не установлен. Установите: go install github.com/cosmtrek/air@latest$(NC)"; \
		echo "$(YELLOW)Запуск без hot reload...$(NC)"; \
		$(GOCMD) run $(MAIN_PATH); \
	fi

# Тестирование
test: ## Запустить тесты
	@echo "$(YELLOW)Запуск тестов...$(NC)"
	$(GOTEST) -v ./...

# Тесты с покрытием
test-coverage: ## Запустить тесты с покрытием
	@echo "$(YELLOW)Запуск тестов с покрытием...$(NC)"
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Отчет о покрытии: coverage.html$(NC)"

# Бенчмарки
bench: ## Запустить бенчмарки
	@echo "$(YELLOW)Запуск бенчмарков...$(NC)"
	$(GOTEST) -bench=. -benchmem ./...

# Очистка
clean: ## Очистить сборочные файлы
	@echo "$(YELLOW)Очистка...$(NC)"
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "$(GREEN)Очистка завершена$(NC)"

# Линтинг (требует golangci-lint)
lint: ## Запустить линтер
	@echo "$(YELLOW)Запуск линтера...$(NC)"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "$(RED)golangci-lint не установлен. Установите: https://golangci-lint.run/usage/install/$(NC)"; \
	fi

# Проверка безопасности (требует gosec)
security: ## Проверка безопасности
	@echo "$(YELLOW)Проверка безопасности...$(NC)"
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "$(RED)gosec не установлен. Установите: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest$(NC)"; \
	fi

# Установка инструментов для разработки
install-tools: ## Установить инструменты для разработки
	@echo "$(YELLOW)Установка инструментов...$(NC)"
	$(GOCMD) install github.com/cosmtrek/air@latest
	$(GOCMD) install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.54.2
	@echo "$(GREEN)Инструменты установлены$(NC)"

# Docker команды
docker-build: ## Собрать Docker образ
	@echo "$(YELLOW)Сборка Docker образа...$(NC)"
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "$(GREEN)Docker образ собран: $(DOCKER_IMAGE):$(DOCKER_TAG)$(NC)"

docker-run: ## Запустить в Docker
	@echo "$(YELLOW)Запуск в Docker...$(NC)"
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-stop: ## Остановить Docker контейнеры
	@echo "$(YELLOW)Остановка Docker контейнеров...$(NC)"
	docker stop $$(docker ps -q --filter ancestor=$(DOCKER_IMAGE):$(DOCKER_TAG))

# Docker Compose
docker-up: ## Запустить с Docker Compose
	@echo "$(YELLOW)Запуск с Docker Compose...$(NC)"
	docker-compose up -d

docker-down: ## Остановить Docker Compose
	@echo "$(YELLOW)Остановка Docker Compose...$(NC)"
	docker-compose down

docker-logs: ## Показать логи Docker Compose
	docker-compose logs -f

# Миграции базы данных
migrate-up: ## Применить миграции
	@echo "$(YELLOW)Применение миграций...$(NC)"
	@if command -v migrate > /dev/null; then \
		migrate -path internal/db/migrations -database "$(shell grep DATABASE_URL .env | cut -d '=' -f2)" up; \
	else \
		echo "$(RED)migrate не установлен. Установите: https://github.com/golang-migrate/migrate$(NC)"; \
	fi

migrate-down: ## Откатить миграции
	@echo "$(YELLOW)Откат миграций...$(NC)"
	@if command -v migrate > /dev/null; then \
		migrate -path internal/db/migrations -database "$(shell grep DATABASE_URL .env | cut -d '=' -f2)" down; \
	else \
		echo "$(RED)migrate не установлен. Установите: https://github.com/golang-migrate/migrate$(NC)"; \
	fi

migrate-create: ## Создать новую миграцию (make migrate-create name=имя_миграции)
	@echo "$(YELLOW)Создание миграции: $(name)$(NC)"
	@if command -v migrate > /dev/null; then \
		migrate create -ext sql -dir internal/db/migrations -seq $(name); \
	else \
		echo "$(RED)migrate не установлен. Установите: https://github.com/golang-migrate/migrate$(NC)"; \
	fi

# Развертывание
prod: build ## Подготовить к продакшену
	@echo "$(YELLOW)Подготовка к продакшену...$(NC)"
	@echo "$(GREEN)Готово к развертыванию$(NC)"

# Проверка всего
check: deps tidy fmt vet lint test ## Полная проверка проекта
	@echo "$(GREEN)Все проверки пройдены успешно!$(NC)"

# Инициализация проекта
init: ## Инициализировать проект
	@echo "$(YELLOW)Инициализация проекта...$(NC)"
	@if [ ! -f .env ]; then \
		cp .env.exemple .env; \
		echo "$(GREEN)Файл .env создан из .env.exemple$(NC)"; \
		echo "$(YELLOW)Не забудьте настроить переменные окружения в .env$(NC)"; \
	else \
		echo "$(YELLOW)Файл .env уже существует$(NC)"; \
	fi
	$(MAKE) deps
	$(MAKE) install-tools
	@echo "$(GREEN)Проект инициализирован!$(NC)"

# Информация о проекте
info: ## Показать информацию о проекте
	@echo "$(GREEN)ZZZ Tournament Backend$(NC)"
	@echo "Go version: $(shell go version)"
	@echo "Project path: $(shell pwd)"
	@echo "Binary name: $(BINARY_NAME)"
	@echo "Build directory: $(BUILD_DIR)"

# По умолчанию
.DEFAULT_GOAL := help