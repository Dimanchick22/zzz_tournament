// internal/models/validation.go
package models

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// ValidationError структура ошибки валидации
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors слайс ошибок валидации
type ValidationErrors []ValidationError

// Error реализует интерфейс error
func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, "; ")
}

// HasErrors проверяет наличие ошибок
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// Add добавляет новую ошибку валидации
func (ve *ValidationErrors) Add(field, message string) {
	*ve = append(*ve, ValidationError{Field: field, Message: message})
}

// ValidateUser валидирует пользователя
func (u *User) Validate() ValidationErrors {
	var errors ValidationErrors

	// Валидация username
	if len(u.Username) < MinUsernameLength {
		errors.Add("username", fmt.Sprintf("Username must be at least %d characters", MinUsernameLength))
	}
	if len(u.Username) > MaxUsernameLength {
		errors.Add("username", fmt.Sprintf("Username must be no more than %d characters", MaxUsernameLength))
	}
	if !usernameRegex.MatchString(u.Username) {
		errors.Add("username", "Username can only contain letters, numbers, underscores and hyphens")
	}

	// Валидация email
	if !emailRegex.MatchString(u.Email) {
		errors.Add("email", "Invalid email format")
	}

	// Валидация rating
	if u.Rating < 0 {
		errors.Add("rating", "Rating cannot be negative")
	}

	return errors
}

// ValidatePassword валидирует пароль
func ValidatePassword(password string) ValidationErrors {
	var errors ValidationErrors

	if len(password) < MinPasswordLength {
		errors.Add("password", fmt.Sprintf("Password must be at least %d characters", MinPasswordLength))
	}
	if len(password) > MaxPasswordLength {
		errors.Add("password", fmt.Sprintf("Password must be no more than %d characters", MaxPasswordLength))
	}

	// Проверка на наличие различных типов символов
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		errors.Add("password", "Password must contain at least one uppercase letter")
	}
	if !hasLower {
		errors.Add("password", "Password must contain at least one lowercase letter")
	}
	if !hasDigit {
		errors.Add("password", "Password must contain at least one digit")
	}
	if !hasSpecial {
		errors.Add("password", "Password must contain at least one special character")
	}

	return errors
}

// ValidateRoom валидирует комнату
func (r *Room) Validate() ValidationErrors {
	var errors ValidationErrors

	// Валидация названия
	if len(strings.TrimSpace(r.Name)) < MinRoomNameLength {
		errors.Add("name", fmt.Sprintf("Room name must be at least %d characters", MinRoomNameLength))
	}
	if len(r.Name) > MaxRoomNameLength {
		errors.Add("name", fmt.Sprintf("Room name must be no more than %d characters", MaxRoomNameLength))
	}

	// Валидация описания
	if len(r.Description) > MaxRoomDescription {
		errors.Add("description", fmt.Sprintf("Room description must be no more than %d characters", MaxRoomDescription))
	}

	// Валидация максимального количества игроков
	if r.MaxPlayers < MinPlayersInRoom {
		errors.Add("max_players", fmt.Sprintf("Room must allow at least %d players", MinPlayersInRoom))
	}
	if r.MaxPlayers > MaxPlayersInRoom {
		errors.Add("max_players", fmt.Sprintf("Room can have no more than %d players", MaxPlayersInRoom))
	}

	// Валидация статуса
	if !IsValidRoomStatus(r.Status) {
		errors.Add("status", "Invalid room status")
	}

	// Валидация пароля для приватных комнат
	if r.IsPrivate {
		if len(r.Password) == 0 {
			errors.Add("password", "Private room must have a password")
		}
		if len(r.Password) > MaxRoomPasswordLength {
			errors.Add("password", fmt.Sprintf("Room password must be no more than %d characters", MaxRoomPasswordLength))
		}
	}

	return errors
}

// ValidateHero валидирует героя
func (h *Hero) Validate() ValidationErrors {
	var errors ValidationErrors

	// Валидация имени
	if len(strings.TrimSpace(h.Name)) == 0 {
		errors.Add("name", "Hero name is required")
	}
	if len(h.Name) > MaxHeroNameLength {
		errors.Add("name", fmt.Sprintf("Hero name must be no more than %d characters", MaxHeroNameLength))
	}

	// Валидация элемента
	if !IsValidElement(h.Element) {
		errors.Add("element", "Invalid hero element")
	}

	// Валидация редкости
	if !IsValidRarity(h.Rarity) {
		errors.Add("rarity", "Invalid hero rarity")
	}

	// Валидация роли
	if !IsValidRole(h.Role) {
		errors.Add("role", "Invalid hero role")
	}

	// Валидация описания
	if len(h.Description) > MaxHeroDescription {
		errors.Add("description", fmt.Sprintf("Hero description must be no more than %d characters", MaxHeroDescription))
	}

	return errors
}

// ValidateMessage валидирует сообщение
func (m *Message) Validate() ValidationErrors {
	var errors ValidationErrors

	// Валидация содержимого
	if len(strings.TrimSpace(m.Content)) == 0 {
		errors.Add("content", "Message content is required")
	}
	if len(m.Content) > MaxMessageLength {
		errors.Add("content", fmt.Sprintf("Message must be no more than %d characters", MaxMessageLength))
	}

	// Валидация типа
	if !IsValidMessageType(m.Type) {
		errors.Add("type", "Invalid message type")
	}

	// Валидация идентификаторов
	if m.RoomID <= 0 {
		errors.Add("room_id", "Invalid room ID")
	}
	if m.UserID <= 0 && m.Type == MessageTypeMessage {
		errors.Add("user_id", "User ID is required for user messages")
	}

	return errors
}

// ValidateTournament валидирует турнир
func (t *Tournament) Validate() ValidationErrors {
	var errors ValidationErrors

	// Валидация названия
	if len(strings.TrimSpace(t.Name)) < MinTournamentNameLength {
		errors.Add("name", fmt.Sprintf("Tournament name must be at least %d characters", MinTournamentNameLength))
	}
	if len(t.Name) > MaxTournamentNameLength {
		errors.Add("name", fmt.Sprintf("Tournament name must be no more than %d characters", MaxTournamentNameLength))
	}

	// Валидация статуса
	if !IsValidTournamentStatus(t.Status) {
		errors.Add("status", "Invalid tournament status")
	}

	// Валидация комнаты
	if t.RoomID <= 0 {
		errors.Add("room_id", "Invalid room ID")
	}

	return errors
}

// ValidateMatch валидирует матч
func (m *Match) Validate() ValidationErrors {
	var errors ValidationErrors

	// Валидация турнира
	if m.TournamentID <= 0 {
		errors.Add("tournament_id", "Invalid tournament ID")
	}

	// Валидация игроков
	if m.Player1ID <= 0 {
		errors.Add("player1_id", "Invalid player 1 ID")
	}
	if m.Player2ID <= 0 {
		errors.Add("player2_id", "Invalid player 2 ID")
	}
	if m.Player1ID == m.Player2ID {
		errors.Add("players", "Player 1 and Player 2 cannot be the same")
	}

	// Валидация раунда
	if m.Round <= 0 {
		errors.Add("round", "Round must be positive")
	}

	// Валидация статуса
	if !IsValidMatchStatus(m.Status) {
		errors.Add("status", "Invalid match status")
	}

	// Валидация победителя
	if m.WinnerID != nil {
		if *m.WinnerID != m.Player1ID && *m.WinnerID != m.Player2ID {
			errors.Add("winner_id", "Winner must be one of the match participants")
		}
		if m.Status != MatchStatusFinished {
			errors.Add("winner_id", "Winner can only be set for finished matches")
		}
	}

	return errors
}

// ValidateWSMessage валидирует WebSocket сообщение
func (ws *WSMessage) Validate() ValidationErrors {
	var errors ValidationErrors

	// Валидация типа
	if !IsValidWSMessageType(ws.Type) {
		errors.Add("type", "Invalid WebSocket message type")
	}

	// Валидация данных в зависимости от типа
	switch ws.Type {
	case WSTypeJoinRoom:
		if data, ok := ws.Data.(JoinRoomData); ok {
			if data.RoomID <= 0 {
				errors.Add("data.room_id", "Invalid room ID")
			}
		} else {
			errors.Add("data", "Invalid join room data format")
		}

	case WSTypeLeaveRoom:
		if data, ok := ws.Data.(LeaveRoomData); ok {
			if data.RoomID <= 0 {
				errors.Add("data.room_id", "Invalid room ID")
			}
		} else {
			errors.Add("data", "Invalid leave room data format")
		}

	case WSTypeChatMessage:
		if data, ok := ws.Data.(ChatMessageData); ok {
			if data.RoomID <= 0 {
				errors.Add("data.room_id", "Invalid room ID")
			}
			if len(strings.TrimSpace(data.Content)) == 0 {
				errors.Add("data.content", "Message content is required")
			}
			if len(data.Content) > MaxMessageLength {
				errors.Add("data.content", fmt.Sprintf("Message must be no more than %d characters", MaxMessageLength))
			}
		} else {
			errors.Add("data", "Invalid chat message data format")
		}

	case WSTypeMatchResult:
		if data, ok := ws.Data.(MatchResultData); ok {
			if data.MatchID <= 0 {
				errors.Add("data.match_id", "Invalid match ID")
			}
			if data.WinnerID <= 0 {
				errors.Add("data.winner_id", "Invalid winner ID")
			}
		} else {
			errors.Add("data", "Invalid match result data format")
		}
	}

	return errors
}
