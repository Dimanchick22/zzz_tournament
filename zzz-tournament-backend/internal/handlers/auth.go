// internal/handlers/auth.go
package handlers

import (
	"database/sql"
	"fmt"
	"time"

	"zzz-tournament/internal/models"
	"zzz-tournament/internal/websocket"
	"zzz-tournament/pkg/auth"
	"zzz-tournament/pkg/utils"
	"zzz-tournament/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandlers обработчики аутентификации
type AuthHandlers struct {
	BaseHandlers
}

// NewAuthHandlers создает новый экземпляр AuthHandlers
func NewAuthHandlers(db *sqlx.DB, hub *websocket.Hub) *AuthHandlers {
	return &AuthHandlers{
		BaseHandlers: newBaseHandlers(db, hub),
	}
}

// RegisterRequest структура запроса регистрации
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginRequest структура запроса входа
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest структура запроса обновления токена
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePasswordRequest структура запроса смены пароля
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

// ForgotPasswordRequest структура запроса восстановления пароля
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required"`
}

// ResetPasswordRequest структура запроса сброса пароля
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// Register регистрация нового пользователя
func (h *AuthHandlers) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Валидация
	errors := validator.ValidateUserRegistration(req.Username, req.Email, req.Password)
	if errors.HasErrors() {
		utils.BadRequestResponse(c, errors.Error())
		return
	}

	// Проверяем, не существует ли пользователь
	var exists bool
	err := h.DB.Get(&exists, `
		SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 OR email = $2)
	`, req.Username, req.Email)

	if err != nil {
		utils.InternalErrorResponse(c, "Database error")
		return
	}

	if exists {
		utils.ConflictResponse(c, "Username or email already exists")
		return
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to hash password")
		return
	}

	// Создаем пользователя
	var userID int
	err = h.DB.QueryRow(`
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id
	`, req.Username, req.Email, string(hashedPassword)).Scan(&userID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to create user")
		return
	}

	// Генерируем токены
	accessToken, err := auth.GenerateToken(userID, req.Username)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to generate token")
		return
	}

	refreshToken, err := auth.GenerateRefreshToken(userID)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to generate refresh token")
		return
	}

	// Сохраняем refresh token в базе
	_, err = h.DB.Exec(`
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`, userID, refreshToken, time.Now().Add(7*24*time.Hour))

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to save refresh token")
		return
	}

	utils.CreatedResponse(c, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": gin.H{
			"id":       userID,
			"username": req.Username,
			"email":    req.Email,
		},
	}, "User registered successfully")
}

// Login вход пользователя
func (h *AuthHandlers) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Валидация
	errors := validator.ValidateUserLogin(req.Username, req.Password)
	if errors.HasErrors() {
		utils.BadRequestResponse(c, errors.Error())
		return
	}

	// Находим пользователя
	var user models.User
	err := h.DB.Get(&user, `
		SELECT id, username, email, password_hash, rating, wins, losses
		FROM users WHERE username = $1
	`, req.Username)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.UnauthorizedResponse(c, "Invalid credentials")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.UnauthorizedResponse(c, "Invalid credentials")
		return
	}

	// Генерируем токены
	accessToken, err := auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to generate token")
		return
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to generate refresh token")
		return
	}

	// Обновляем последний вход
	_, err = h.DB.Exec(`
		UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = $1
	`, user.ID)

	if err != nil {
		// Не критично, логируем и продолжаем
		fmt.Printf("Failed to update last login for user %d: %v\n", user.ID, err)
	}

	// Сохраняем новый refresh token
	_, err = h.DB.Exec(`
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE SET
			token = EXCLUDED.token,
			expires_at = EXCLUDED.expires_at,
			updated_at = CURRENT_TIMESTAMP
	`, user.ID, refreshToken, time.Now().Add(7*24*time.Hour))

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to save refresh token")
		return
	}

	// Скрываем пароль
	user.Password = ""

	utils.SuccessResponse(c, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          user,
	}, "Login successful")
}

// RefreshToken обновление токена доступа
func (h *AuthHandlers) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Проверяем refresh token
	var userID int
	var expiresAt time.Time
	err := h.DB.QueryRow(`
		SELECT user_id, expires_at FROM refresh_tokens 
		WHERE token = $1 AND expires_at > CURRENT_TIMESTAMP
	`, req.RefreshToken).Scan(&userID, &expiresAt)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.UnauthorizedResponse(c, "Invalid or expired refresh token")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	// Получаем информацию о пользователе
	var user models.User
	err = h.DB.Get(&user, `
		SELECT id, username, email, rating, wins, losses
		FROM users WHERE id = $1
	`, userID)

	if err != nil {
		utils.UnauthorizedResponse(c, "User not found")
		return
	}

	// Генерируем новый access token
	accessToken, err := auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to generate token")
		return
	}

	// Опционально генерируем новый refresh token если старый скоро истекает
	var newRefreshToken string
	if time.Until(expiresAt) < 24*time.Hour {
		newRefreshToken, err = auth.GenerateRefreshToken(user.ID)
		if err != nil {
			utils.InternalErrorResponse(c, "Failed to generate refresh token")
			return
		}

		// Обновляем refresh token в базе
		_, err = h.DB.Exec(`
			UPDATE refresh_tokens 
			SET token = $1, expires_at = $2, updated_at = CURRENT_TIMESTAMP
			WHERE user_id = $3
		`, newRefreshToken, time.Now().Add(7*24*time.Hour), user.ID)

		if err != nil {
			utils.InternalErrorResponse(c, "Failed to update refresh token")
			return
		}
	} else {
		newRefreshToken = req.RefreshToken
	}

	utils.SuccessResponse(c, gin.H{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
	}, "Token refreshed successfully")
}

// Logout выход пользователя
func (h *AuthHandlers) Logout(c *gin.Context) {
	userID := c.GetInt("user_id")

	// Удаляем refresh token из базы
	_, err := h.DB.Exec(`
		DELETE FROM refresh_tokens WHERE user_id = $1
	`, userID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to logout")
		return
	}

	utils.NoContentResponse(c, "Logout successful")
}

// ChangePassword смена пароля
func (h *AuthHandlers) ChangePassword(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Валидация нового пароля
	if err := validator.ValidatePassword(req.NewPassword); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Получаем текущий хеш пароля
	var currentHash string
	err := h.DB.Get(&currentHash, `
		SELECT password_hash FROM users WHERE id = $1
	`, userID)

	if err != nil {
		utils.InternalErrorResponse(c, "Database error")
		return
	}

	// Проверяем текущий пароль
	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(req.CurrentPassword)); err != nil {
		utils.UnauthorizedResponse(c, "Current password is incorrect")
		return
	}

	// Хешируем новый пароль
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to hash password")
		return
	}

	// Обновляем пароль
	_, err = h.DB.Exec(`
		UPDATE users 
		SET password_hash = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, string(newHash), userID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to update password")
		return
	}

	// Удаляем все refresh токены для принудительного переавторизации
	_, err = h.DB.Exec(`
		DELETE FROM refresh_tokens WHERE user_id = $1
	`, userID)

	if err != nil {
		// Не критично, продолжаем
		fmt.Printf("Failed to revoke refresh tokens for user %d: %v\n", userID, err)
	}

	utils.NoContentResponse(c, "Password changed successfully")
}

// ForgotPassword запрос на восстановление пароля
func (h *AuthHandlers) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Валидация email
	if err := validator.ValidateEmail(req.Email); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Проверяем существование пользователя
	var userID int
	err := h.DB.Get(&userID, `
		SELECT id FROM users WHERE email = $1
	`, req.Email)

	if err != nil {
		// Не раскрываем информацию о существовании пользователя
		utils.SuccessResponse(c, nil, "If the email exists, a reset link has been sent")
		return
	}

	// Генерируем токен сброса
	resetToken, err := auth.GenerateResetToken(userID)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to generate reset token")
		return
	}

	// Сохраняем токен в базе (действителен 1 час)
	_, err = h.DB.Exec(`
		INSERT INTO password_reset_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE SET
			token = EXCLUDED.token,
			expires_at = EXCLUDED.expires_at,
			created_at = CURRENT_TIMESTAMP
	`, userID, resetToken, time.Now().Add(time.Hour))

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to save reset token")
		return
	}

	// TODO: Отправить email с ссылкой для сброса
	// В реальном приложении здесь должна быть отправка email

	utils.SuccessResponse(c, nil, "If the email exists, a reset link has been sent")
}

// ResetPassword сброс пароля по токену
func (h *AuthHandlers) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Валидация нового пароля
	if err := validator.ValidatePassword(req.NewPassword); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Проверяем токен сброса
	var userID int
	err := h.DB.QueryRow(`
		SELECT user_id FROM password_reset_tokens 
		WHERE token = $1 AND expires_at > CURRENT_TIMESTAMP
	`, req.Token).Scan(&userID)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.UnauthorizedResponse(c, "Invalid or expired reset token")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	// Хешируем новый пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to hash password")
		return
	}

	// Начинаем транзакцию
	tx, err := h.DB.Beginx()
	if err != nil {
		utils.InternalErrorResponse(c, "Database transaction error")
		return
	}
	defer tx.Rollback()

	// Обновляем пароль
	_, err = tx.Exec(`
		UPDATE users 
		SET password_hash = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, string(hashedPassword), userID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to update password")
		return
	}

	// Удаляем использованный токен сброса
	_, err = tx.Exec(`
		DELETE FROM password_reset_tokens WHERE user_id = $1
	`, userID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to cleanup reset token")
		return
	}

	// Удаляем все refresh токены
	_, err = tx.Exec(`
		DELETE FROM refresh_tokens WHERE user_id = $1
	`, userID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to revoke tokens")
		return
	}

	// Коммитим транзакцию
	if err = tx.Commit(); err != nil {
		utils.InternalErrorResponse(c, "Failed to commit transaction")
		return
	}

	utils.NoContentResponse(c, "Password reset successfully")
}
