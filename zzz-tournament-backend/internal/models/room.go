// internal/models/room.go
package models

import (
	"time"
)

// Room модель комнаты
type Room struct {
	ID           int       `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Description  string    `json:"description" db:"description"`
	HostID       int       `json:"host_id" db:"host_id"`
	Host         *User     `json:"host,omitempty"`
	MaxPlayers   int       `json:"max_players" db:"max_players"`
	CurrentCount int       `json:"current_count" db:"current_count"`
	Status       string    `json:"status" db:"status"` // waiting, in_progress, finished
	IsPrivate    bool      `json:"is_private" db:"is_private"`
	Password     string    `json:"-" db:"password"` // Скрыто в JSON
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	Participants []User    `json:"participants,omitempty"`
}

// RoomParticipant связь участника с комнатой
type RoomParticipant struct {
	RoomID   int       `json:"room_id" db:"room_id"`
	UserID   int       `json:"user_id" db:"user_id"`
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`
}

// RoomStatus константы статусов комнат
const (
	RoomStatusWaiting    = "waiting"
	RoomStatusInProgress = "in_progress"
	RoomStatusFinished   = "finished"
)

// IsValidRoomStatus проверяет валидность статуса комнаты
func IsValidRoomStatus(status string) bool {
	switch status {
	case RoomStatusWaiting, RoomStatusInProgress, RoomStatusFinished:
		return true
	default:
		return false
	}
}

// CanJoin проверяет, можно ли присоединиться к комнате
func (r *Room) CanJoin() bool {
	return r.Status == RoomStatusWaiting && r.CurrentCount < r.MaxPlayers
}

// IsFull проверяет, заполнена ли комната
func (r *Room) IsFull() bool {
	return r.CurrentCount >= r.MaxPlayers
}

// IsHost проверяет, является ли пользователь хостом комнаты
func (r *Room) IsHost(userID int) bool {
	return r.HostID == userID
}
