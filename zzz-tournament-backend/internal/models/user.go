// internal/models/user.go
package models

import (
	"time"
)

// User модель пользователя
type User struct {
	ID            int        `json:"id" db:"id"`
	Username      string     `json:"username" db:"username"`
	Email         string     `json:"email" db:"email"`
	Password      string     `json:"-" db:"password_hash"` // Скрыто в JSON
	Rating        int        `json:"rating" db:"rating"`
	Wins          int        `json:"wins" db:"wins"`
	Losses        int        `json:"losses" db:"losses"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	IsVerified    bool       `json:"is_verified" db:"is_verified"`
	LastLogin     *time.Time `json:"last_login,omitempty" db:"last_login"`
	LoginAttempts int        `json:"-" db:"login_attempts"` // Скрыто в JSON
	LockedUntil   *time.Time `json:"-" db:"locked_until"`   // Скрыто в JSON
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// IsLocked проверяет, заблокирован ли аккаунт
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.LockedUntil)
}

// CanAttemptLogin проверяет, может ли пользователь попытаться войти
func (u *User) CanAttemptLogin(maxAttempts int) bool {
	return u.IsActive && !u.IsLocked() && u.LoginAttempts < maxAttempts
}

// IncrementLoginAttempts увеличивает счетчик неудачных попыток входа
func (u *User) IncrementLoginAttempts(maxAttempts int, lockoutDuration time.Duration) {
	u.LoginAttempts++
	if u.LoginAttempts >= maxAttempts {
		lockUntil := time.Now().Add(lockoutDuration)
		u.LockedUntil = &lockUntil
	}
}

// ResetLoginAttempts сбрасывает счетчик попыток входа
func (u *User) ResetLoginAttempts() {
	u.LoginAttempts = 0
	u.LockedUntil = nil
}
