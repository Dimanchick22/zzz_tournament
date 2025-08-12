# ZZZ Tournament Backend

Бэкенд для турнирной системы игры Zenless Zone Zero, построенный на Go с использованием Gin, PostgreSQL и WebSocket.

## 🚀 Основные возможности

- **JWT аутентификация** с refresh токенами
- **Система комнат** для организации турниров
- **Турнирная сетка** с автоматической генерацией bracket'ов
- **Рейтинговая система ELO** для ранжирования игроков
- **Real-time чат** через WebSocket
- **База героев ZZZ** с фильтрацией и поиском
- **Rate limiting** и защита от спама
- **Логирование** и мониторинг

## 📋 Требования

- Go 1.21+
- PostgreSQL 15+
- Redis (опционально)

## 🛠 Установка

1. Клонируйте репозиторий:
```bash
git clone <repository-url>
cd zzz-tournament-backend
```

2. Создайте `.env` файл:
```bash
make init
```

3. Настройте переменные окружения в `.env`

4. Запустите базу данных:
```bash
docker-compose up postgres redis -d
```

5. Запустите приложение:
```bash
make dev  # для разработки с hot reload
# или
make run  # обычный запуск
```

## 📚 API Документация

В режиме разработки доступна по адресу: `http://localhost:8080/docs`

### Основные эндпоинты:

#### Аутентификация
- `POST /api/v1/auth/register` - Регистрация
- `POST /api/v1/auth/login` - Вход
- `POST /api/v1/auth/refresh` - Обновление токена

#### Пользователи
- `GET /api/v1/users/profile` - Профиль пользователя
- `GET /api/v1/users/leaderboard` - Рейтинговая таблица

#### Комнаты
- `GET /api/v1/rooms` - Список комнат
- `POST /api/v1/rooms` - Создать комнату
- `POST /api/v1/rooms/:id/join` - Присоединиться

#### Турниры
- `POST /api/v1/rooms/:id/tournament/start` - Запустить турнир
- `GET /api/v1/tournaments/:id` - Информация о турнире

#### WebSocket
- `WS /ws` - WebSocket соединение для real-time обновлений

## 🎮 Использование

### Создание турнира

1. Создайте комнату
2. Игроки присоединяются к комнате
3. Хост запускает турнир
4. Система автоматически генерирует bracket
5. Игроки сражаются и отправляют результаты

### WebSocket события

```javascript
// Присоединиться к комнате
ws.send(JSON.stringify({
  type: "join_room",
  data: { room_id: 123 }
}));

// Отправить сообщение в чат
ws.send(JSON.stringify({
  type: "chat_message", 
  data: { room_id: 123, content: "Hello!" }
}));
```

## 🏗 Архитектура

```
cmd/server/          # Точка входа приложения
internal/
├── config/          # Конфигурация
├── db/             # Подключение к БД и миграции
├── handlers/       # HTTP обработчики
├── middleware/     # Middleware (auth, cors, logging)
├── models/         # Структуры данных
└── websocket/      # WebSocket hub
pkg/
├── auth/           # JWT аутентификация
├── rating/         # ELO рейтинговая система
├── tournament/     # Генерация турнирных сеток
├── utils/          # Утилиты и ответы API
└── validator/      # Валидация данных
```

## 🔧 Команды Make

```bash
make help           # Показать все команды
make init           # Инициализировать проект
make dev            # Запуск с hot reload
make run            # Обычный запуск
make build          # Сборка приложения
make test           # Запуск тестов
make docker-build   # Сборка Docker образа
make migrate-up     # Применить миграции
```

## 🐳 Docker

```bash
# Запуск полного стека
docker-compose up

# Только база данных
docker-compose up postgres redis -d

# Продакшен режим
docker-compose --profile production up
```

## 🔐 Безопасность

- JWT токены с автоматическим обновлением
- Bcrypt хеширование паролей
- Rate limiting по IP и пользователям
- CORS защита
- Валидация всех входных данных
- SQL инъекции защита через sqlx

## 📊 Мониторинг

- Structured logging
- Prometheus метрики (планируется)
- Health check endpoints
- Performance monitoring

## 🧪 Тестирование

```bash
make test                # Все тесты
make test-coverage      # Тесты с покрытием
make bench             # Бенчмарки
```

## 🤝 Разработка

1. Форкните репозиторий
2. Создайте feature ветку
3. Внесите изменения
4. Добавьте тесты
5. Отправьте Pull Request

## 📄 Лицензия

MIT License - см. файл LICENSE

## 🆘 Поддержка

Создайте issue в GitHub репозитории для:
- Сообщения об ошибках
- Запросы новых функций  
- Вопросы по использованию

---

**Автор:** ZZZ Tournament Team  
**Версия:** 1.0.0