// internal/models/websocket.go
package models

// WSMessage структура WebSocket сообщения
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// JoinRoomData данные для присоединения к комнате
type JoinRoomData struct {
	RoomID   int    `json:"room_id"`
	Password string `json:"password,omitempty"`
}

// LeaveRoomData данные для выхода из комнаты
type LeaveRoomData struct {
	RoomID int `json:"room_id"`
}

// ChatMessageData данные чат сообщения
type ChatMessageData struct {
	RoomID  int    `json:"room_id"`
	Content string `json:"content"`
}

// MatchResultData данные результата матча
type MatchResultData struct {
	MatchID  int `json:"match_id"`
	WinnerID int `json:"winner_id"`
}

// RoomUpdateData данные обновления комнаты
type RoomUpdateData struct {
	Room *Room `json:"room"`
}

// TournamentUpdateData данные обновления турнира
type TournamentUpdateData struct {
	Tournament *Tournament `json:"tournament"`
}

// UserJoinedData данные о присоединении пользователя
type UserJoinedData struct {
	RoomID int  `json:"room_id"`
	User   User `json:"user"`
}

// UserLeftData данные о выходе пользователя
type UserLeftData struct {
	RoomID int  `json:"room_id"`
	User   User `json:"user"`
}

// ErrorData данные ошибки
type ErrorData struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// WebSocket message types
const (
	WSTypeJoinRoom         = "join_room"
	WSTypeLeaveRoom        = "leave_room"
	WSTypeChatMessage      = "chat_message"
	WSTypeMatchResult      = "match_result"
	WSTypeRoomUpdate       = "room_update"
	WSTypeTournamentUpdate = "tournament_update"
	WSTypeUserJoined       = "user_joined"
	WSTypeUserLeft         = "user_left"
	WSTypeError            = "error"
	WSTypeNotification     = "notification"
)

// IsValidWSMessageType проверяет валидность типа WebSocket сообщения
func IsValidWSMessageType(msgType string) bool {
	switch msgType {
	case WSTypeJoinRoom, WSTypeLeaveRoom, WSTypeChatMessage, WSTypeMatchResult,
		WSTypeRoomUpdate, WSTypeTournamentUpdate, WSTypeUserJoined, WSTypeUserLeft,
		WSTypeError, WSTypeNotification:
		return true
	default:
		return false
	}
}
