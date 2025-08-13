// internal/models/interfaces.go
package models

import (
	"context"
)

// UserRepository интерфейс репозитория пользователей
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id int) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int) error
	UpdateRating(ctx context.Context, userID int, newRating int) error
	IncrementWins(ctx context.Context, userID int) error
	IncrementLosses(ctx context.Context, userID int) error
}

// RoomRepository интерфейс репозитория комнат
type RoomRepository interface {
	Create(ctx context.Context, room *Room) error
	GetByID(ctx context.Context, id int) (*Room, error)
	GetAll(ctx context.Context) ([]Room, error)
	Update(ctx context.Context, room *Room) error
	Delete(ctx context.Context, id int) error
	AddParticipant(ctx context.Context, roomID, userID int) error
	RemoveParticipant(ctx context.Context, roomID, userID int) error
	GetParticipants(ctx context.Context, roomID int) ([]User, error)
	IsParticipant(ctx context.Context, roomID, userID int) (bool, error)
}

// TournamentRepository интерфейс репозитория турниров
type TournamentRepository interface {
	Create(ctx context.Context, tournament *Tournament) error
	GetByID(ctx context.Context, id int) (*Tournament, error)
	GetByRoomID(ctx context.Context, roomID int) (*Tournament, error)
	Update(ctx context.Context, tournament *Tournament) error
	Delete(ctx context.Context, id int) error
}

// MatchRepository интерфейс репозитория матчей
type MatchRepository interface {
	Create(ctx context.Context, match *Match) error
	GetByID(ctx context.Context, id int) (*Match, error)
	GetByTournamentID(ctx context.Context, tournamentID int) ([]Match, error)
	Update(ctx context.Context, match *Match) error
	Delete(ctx context.Context, id int) error
}

// MessageRepository интерфейс репозитория сообщений
type MessageRepository interface {
	Create(ctx context.Context, message *Message) error
	GetByRoomID(ctx context.Context, roomID int, limit, offset int) ([]Message, error)
	Delete(ctx context.Context, id int) error
	DeleteByRoomID(ctx context.Context, roomID int) error
}

// HeroRepository интерфейс репозитория героев
type HeroRepository interface {
	Create(ctx context.Context, hero *Hero) error
	GetByID(ctx context.Context, id int) (*Hero, error)
	GetAll(ctx context.Context) ([]Hero, error)
	GetByElement(ctx context.Context, element string) ([]Hero, error)
	GetByRarity(ctx context.Context, rarity string) ([]Hero, error)
	GetByRole(ctx context.Context, role string) ([]Hero, error)
	Update(ctx context.Context, hero *Hero) error
	Delete(ctx context.Context, id int) error
}

// AuthRepository интерфейс репозитория аутентификации
type AuthRepository interface {
	SaveRefreshToken(ctx context.Context, token *RefreshToken) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, userID int) error
	CleanupExpiredTokens(ctx context.Context) error

	SavePasswordResetToken(ctx context.Context, token *PasswordResetToken) error
	GetPasswordResetToken(ctx context.Context, token string) (*PasswordResetToken, error)
	DeletePasswordResetToken(ctx context.Context, userID int) error

	LogSecurityEvent(ctx context.Context, event *SecurityEvent) error
}
