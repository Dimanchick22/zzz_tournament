// internal/models/tournament.go
package models

import (
	"time"
)

// Tournament модель турнира
type Tournament struct {
	ID        int                    `json:"id" db:"id"`
	RoomID    int                    `json:"room_id" db:"room_id"`
	Name      string                 `json:"name" db:"name"`
	Status    string                 `json:"status" db:"status"` // created, started, finished
	Bracket   map[string]interface{} `json:"bracket" db:"bracket"`
	WinnerID  *int                   `json:"winner_id" db:"winner_id"`
	Winner    *User                  `json:"winner,omitempty"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
	Matches   []Match                `json:"matches,omitempty"`
}

// Match модель матча
type Match struct {
	ID           int       `json:"id" db:"id"`
	TournamentID int       `json:"tournament_id" db:"tournament_id"`
	Round        int       `json:"round" db:"round"`
	Player1ID    int       `json:"player1_id" db:"player1_id"`
	Player2ID    int       `json:"player2_id" db:"player2_id"`
	Player1      *User     `json:"player1,omitempty"`
	Player2      *User     `json:"player2,omitempty"`
	WinnerID     *int      `json:"winner_id" db:"winner_id"`
	Winner       *User     `json:"winner,omitempty"`
	Status       string    `json:"status" db:"status"` // pending, in_progress, finished
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// TournamentStatus константы статусов турниров
const (
	TournamentStatusCreated  = "created"
	TournamentStatusStarted  = "started"
	TournamentStatusFinished = "finished"
)

// MatchStatus константы статусов матчей
const (
	MatchStatusPending    = "pending"
	MatchStatusInProgress = "in_progress"
	MatchStatusFinished   = "finished"
)

// IsValidTournamentStatus проверяет валидность статуса турнира
func IsValidTournamentStatus(status string) bool {
	switch status {
	case TournamentStatusCreated, TournamentStatusStarted, TournamentStatusFinished:
		return true
	default:
		return false
	}
}

// IsValidMatchStatus проверяет валидность статуса матча
func IsValidMatchStatus(status string) bool {
	switch status {
	case MatchStatusPending, MatchStatusInProgress, MatchStatusFinished:
		return true
	default:
		return false
	}
}

// IsFinished проверяет, завершен ли турнир
func (t *Tournament) IsFinished() bool {
	return t.Status == TournamentStatusFinished
}

// HasWinner проверяет, есть ли победитель в турнире
func (t *Tournament) HasWinner() bool {
	return t.WinnerID != nil
}

// CanStart проверяет, можно ли начать турнир
func (t *Tournament) CanStart() bool {
	return t.Status == TournamentStatusCreated
}

// IsParticipant проверяет, является ли пользователь участником матча
func (m *Match) IsParticipant(userID int) bool {
	return m.Player1ID == userID || m.Player2ID == userID
}

// GetOpponentID возвращает ID оппонента для данного пользователя
func (m *Match) GetOpponentID(userID int) *int {
	if m.Player1ID == userID {
		return &m.Player2ID
	}
	if m.Player2ID == userID {
		return &m.Player1ID
	}
	return nil
}

// IsFinished проверяет, завершен ли матч
func (m *Match) IsFinished() bool {
	return m.Status == MatchStatusFinished
}

// HasWinner проверяет, есть ли победитель в матче
func (m *Match) HasWinner() bool {
	return m.WinnerID != nil
}
