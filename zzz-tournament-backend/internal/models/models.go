// internal/models/models.go
package models

// Этот файл служит основной точкой входа для пакета models
// Все модели определены в отдельных файлах:
//
// user.go - модель пользователя и связанные методы
// auth.go - модели аутентификации (токены, события безопасности)
// hero.go - модель героя и константы
// room.go - модель комнаты и участников
// tournament.go - модели турнира и матчей
// message.go - модель сообщений
// websocket.go - модели WebSocket сообщений
// constants.go - общие константы и ограничения
// interfaces.go - интерфейсы репозиториев
// validation.go - валидационные функции

// Вспомогательные функции для работы с моделями

// NewUser создает нового пользователя с значениями по умолчанию
func NewUser(username, email, passwordHash string) *User {
	return &User{
		Username:      username,
		Email:         email,
		Password:      passwordHash,
		Rating:        DefaultUserRating,
		Wins:          0,
		Losses:        0,
		IsActive:      true,
		IsVerified:    false,
		LoginAttempts: 0,
	}
}

// NewRoom создает новую комнату с значениями по умолчанию
func NewRoom(name, description string, hostID int, isPrivate bool, password string) *Room {
	return &Room{
		Name:         name,
		Description:  description,
		HostID:       hostID,
		MaxPlayers:   DefaultMaxPlayers,
		CurrentCount: 0,
		Status:       RoomStatusWaiting,
		IsPrivate:    isPrivate,
		Password:     password,
	}
}

// NewTournament создает новый турнир
func NewTournament(roomID int, name string) *Tournament {
	return &Tournament{
		RoomID:  roomID,
		Name:    name,
		Status:  TournamentStatusCreated,
		Bracket: make(map[string]interface{}),
	}
}

// NewMatch создает новый матч
func NewMatch(tournamentID, round, player1ID, player2ID int) *Match {
	return &Match{
		TournamentID: tournamentID,
		Round:        round,
		Player1ID:    player1ID,
		Player2ID:    player2ID,
		Status:       MatchStatusPending,
	}
}

// NewMessage создает новое сообщение
func NewMessage(roomID, userID int, content, msgType string) *Message {
	return &Message{
		RoomID:  roomID,
		UserID:  userID,
		Content: content,
		Type:    msgType,
	}
}

// NewSystemMessage создает системное сообщение
func NewSystemMessage(roomID int, content string) *Message {
	return &Message{
		RoomID:  roomID,
		UserID:  0, // Системные сообщения не имеют пользователя
		Content: content,
		Type:    MessageTypeSystem,
	}
}

// NewWSMessage создает новое WebSocket сообщение
func NewWSMessage(msgType string, data interface{}) *WSMessage {
	return &WSMessage{
		Type: msgType,
		Data: data,
	}
}

// NewErrorWSMessage создает WebSocket сообщение об ошибке
func NewErrorWSMessage(message, code string) *WSMessage {
	return &WSMessage{
		Type: WSTypeError,
		Data: ErrorData{
			Message: message,
			Code:    code,
		},
	}
}
