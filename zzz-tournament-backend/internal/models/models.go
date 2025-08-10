// internal/models/models.go
package models

import (
	"time"
)

type User struct {
	ID        int       `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password_hash"`
	Rating    int       `json:"rating" db:"rating"`
	Wins      int       `json:"wins" db:"wins"`
	Losses    int       `json:"losses" db:"losses"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Hero struct {
	ID          int    `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Element     string `json:"element" db:"element"`
	Rarity      string `json:"rarity" db:"rarity"`
	Role        string `json:"role" db:"role"`
	Description string `json:"description" db:"description"`
	ImageURL    string `json:"image_url" db:"image_url"`
	IsActive    bool   `json:"is_active" db:"is_active"`
}

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
	Password     string    `json:"-" db:"password"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	Participants []User    `json:"participants,omitempty"`
}

type RoomParticipant struct {
	RoomID   int       `json:"room_id" db:"room_id"`
	UserID   int       `json:"user_id" db:"user_id"`
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`
}

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

type Message struct {
	ID        int       `json:"id" db:"id"`
	RoomID    int       `json:"room_id" db:"room_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	User      *User     `json:"user,omitempty"`
	Content   string    `json:"content" db:"content"`
	Type      string    `json:"type" db:"type"` // message, system, join, leave
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// WebSocket messages
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type JoinRoomData struct {
	RoomID int `json:"room_id"`
}

type LeaveRoomData struct {
	RoomID int `json:"room_id"`
}

type ChatMessageData struct {
	RoomID  int    `json:"room_id"`
	Content string `json:"content"`
}

type MatchResultData struct {
	MatchID  int `json:"match_id"`
	WinnerID int `json:"winner_id"`
}
