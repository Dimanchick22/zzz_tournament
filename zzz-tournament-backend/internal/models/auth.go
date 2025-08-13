// internal/models/auth.go
package models

import (
	"time"
)

// RefreshToken модель refresh токена
type RefreshToken struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	TokenHash string    `json:"-" db:"token_hash"` // Хешированный токен
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// PasswordResetToken модель токена сброса пароля
type PasswordResetToken struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Token     string    `json:"-" db:"token"` // В отличие от refresh, reset токены можно не хешировать
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// SecurityEvent модель события безопасности
type SecurityEvent struct {
	ID        int                    `json:"id" db:"id"`
	UserID    *int                   `json:"user_id,omitempty" db:"user_id"`
	Event     string                 `json:"event" db:"event"`
	ClientIP  string                 `json:"client_ip" db:"client_ip"`
	UserAgent string                 `json:"user_agent" db:"user_agent"`
	Details   map[string]interface{} `json:"details" db:"details"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}
