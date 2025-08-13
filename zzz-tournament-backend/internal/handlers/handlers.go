// internal/handlers/handlers.go
package handlers

import (
	"log/slog"
	"zzz-tournament/internal/websocket"
	"zzz-tournament/pkg/config"

	"github.com/jmoiron/sqlx"
)

// Handlers структура содержащая все обработчики
type Handlers struct {
	DB     *sqlx.DB
	Hub    *websocket.Hub
	Logger *slog.Logger
	Config *config.AuthConfig

	// Отдельные группы хендлеров
	Auth        *AuthHandlers
	Users       *UserHandlers
	Heroes      *HeroHandlers
	Rooms       *RoomHandlers
	Tournaments *TournamentHandlers
	Chat        *ChatHandlers
}

// New создает новый экземпляр всех хендлеров
func New(db *sqlx.DB, hub *websocket.Hub, logger *slog.Logger, authConfig *config.AuthConfig) *Handlers {
	h := &Handlers{
		DB:     db,
		Hub:    hub,
		Logger: logger,
		Config: authConfig,
	}

	// Инициализируем отдельные группы хендлеров
	h.Auth = NewAuthHandlers(db, hub, logger, authConfig)
	h.Users = NewUserHandlers(db, hub, logger)
	h.Heroes = NewHeroHandlers(db, hub, logger)
	h.Rooms = NewRoomHandlers(db, hub, logger)
	h.Tournaments = NewTournamentHandlers(db, hub, logger)
	h.Chat = NewChatHandlers(db, hub, logger)

	return h
}

// BaseHandlers базовая структура для всех групп хендлеров
type BaseHandlers struct {
	DB     *sqlx.DB
	Hub    *websocket.Hub
	Logger *slog.Logger
}

// newBaseHandlers создает базовый хендлер
func newBaseHandlers(db *sqlx.DB, hub *websocket.Hub, logger *slog.Logger) BaseHandlers {
	return BaseHandlers{
		DB:     db,
		Hub:    hub,
		Logger: logger,
	}
}
