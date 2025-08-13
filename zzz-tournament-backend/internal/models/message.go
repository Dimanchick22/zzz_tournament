// internal/models/message.go
package models

import (
	"time"
)

// Message модель сообщения
type Message struct {
	ID        int       `json:"id" db:"id"`
	RoomID    int       `json:"room_id" db:"room_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	User      *User     `json:"user,omitempty"`
	Content   string    `json:"content" db:"content"`
	Type      string    `json:"type" db:"type"` // message, system, join, leave
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// MessageType константы типов сообщений
const (
	MessageTypeMessage = "message"
	MessageTypeSystem  = "system"
	MessageTypeJoin    = "join"
	MessageTypeLeave   = "leave"
)

// IsValidMessageType проверяет валидность типа сообщения
func IsValidMessageType(msgType string) bool {
	switch msgType {
	case MessageTypeMessage, MessageTypeSystem, MessageTypeJoin, MessageTypeLeave:
		return true
	default:
		return false
	}
}

// IsUserMessage проверяет, является ли сообщение пользовательским
func (m *Message) IsUserMessage() bool {
	return m.Type == MessageTypeMessage
}

// IsSystemMessage проверяет, является ли сообщение системным
func (m *Message) IsSystemMessage() bool {
	return m.Type == MessageTypeSystem || m.Type == MessageTypeJoin || m.Type == MessageTypeLeave
}
