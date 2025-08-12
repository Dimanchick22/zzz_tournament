// internal/handlers/handlers.go
package handlers

import (
	"zzz-tournament/internal/websocket"

	"github.com/jmoiron/sqlx"
)

// Handlers структура содержащая все обработчики
type Handlers struct {
	DB  *sqlx.DB
	Hub *websocket.Hub

	// Отдельные группы хендлеров
	Auth        *AuthHandlers
	Users       *UserHandlers
	Heroes      *HeroHandlers
	Rooms       *RoomHandlers
	Tournaments *TournamentHandlers
	Chat        *ChatHandlers
}

// New создает новый экземпляр всех хендлеров
func New(db *sqlx.DB, hub *websocket.Hub) *Handlers {
	h := &Handlers{
		DB:  db,
		Hub: hub,
	}

	// Инициализируем отдельные группы хендлеров
	h.Auth = NewAuthHandlers(db, hub)
	h.Users = NewUserHandlers(db, hub)
	h.Heroes = NewHeroHandlers(db, hub)
	h.Rooms = NewRoomHandlers(db, hub)
	h.Tournaments = NewTournamentHandlers(db, hub)
	h.Chat = NewChatHandlers(db, hub)

	return h
}

// BaseHandlers базовая структура для всех групп хендлеров
type BaseHandlers struct {
	DB  *sqlx.DB
	Hub *websocket.Hub
}

// newBaseHandlers создает базовый хендлер
func newBaseHandlers(db *sqlx.DB, hub *websocket.Hub) BaseHandlers {
	return BaseHandlers{
		DB:  db,
		Hub: hub,
	}
}
