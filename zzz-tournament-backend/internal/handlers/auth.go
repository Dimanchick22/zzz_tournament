// internal/handlers/auth.go
package handlers

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"zzz-tournament/internal/models"
	"zzz-tournament/internal/websocket"
	"zzz-tournament/pkg/auth"
	"zzz-tournament/pkg/config"
	"zzz-tournament/pkg/utils"
	"zzz-tournament/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandlers обработчики аутентификации
type AuthHandlers struct {
	BaseHandlers
	logger       *slog.Logger
	config       *config.AuthConfig
	userMutexes  map[int]*sync.Mutex
	mutexMapLock sync.RWMutex
}

// NewAuthHandlers создает новый экземпляр AuthHandlers
func NewAuthHandlers(db *sqlx.DB, hub *websocket.Hub, logger *slog.Logger, cfg *config.AuthConfig) *AuthHandlers {
	return &AuthHandlers{
		BaseHandlers: newBaseHandlers(db, hub, logger),
		logger:       logger,
		config:       cfg,
		userMutexes:  make(map[int]*sync.Mutex),
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

// getUserMutex получает мьютекс для конкретного пользователя
func (h *AuthHandlers) getUserMutex(userID int) *sync.Mutex {
	h.mutexMapLock.RLock()
	mutex, exists := h.userMutexes[userID]
	h.mutexMapLock.RUnlock()

	if exists {
		return mutex
	}

	h.mutexMapLock.Lock()
	defer h.mutexMapLock.Unlock()

	// Проверяем еще раз после получения write lock
	if mutex, exists := h.userMutexes[userID]; exists {
		return mutex
	}

	mutex = &sync.Mutex{}
	h.userMutexes[userID] = mutex
	return mutex
}

// hashRefreshToken хеширует refresh token для безопасного хранения
func (h *AuthHandlers) hashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// logSecurityEvent логирует события безопасности
func (h *AuthHandlers) logSecurityEvent(ctx context.Context, event string, userID int, clientIP string, details map[string]interface{}) {
	h.logger.InfoContext(ctx, "Security event",
		slog.String("event", event),
		slog.Int("user_id", userID),
		slog.String("client_ip", clientIP),
		slog.Any("details", details),
	)
}

// logSecurityError логирует ошибки безопасности
func (h *AuthHandlers) logSecurityError(ctx context.Context, event string, clientIP string, err error, details map[string]interface{}) {
	h.logger.ErrorContext(ctx, "Security error",
		slog.String("event", event),
		slog.String("client_ip", clientIP),
		slog.String("error", err.Error()),
		slog.Any("details", details),
	)
}

// Register регистрация нового пользователя
func (h *AuthHandlers) Register(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.config.DatabaseTimeout)
	defer cancel()

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logSecurityError(ctx, "invalid_register_request", c.ClientIP(), err, map[string]interface{}{
			"error_type": "validation",
		})
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Валидация
	errors := validator.ValidateUserRegistration(req.Username, req.Email, req.Password)
	if errors.HasErrors() {
		h.logSecurityError(ctx, "register_validation_failed", c.ClientIP(), fmt.Errorf("validation failed"), map[string]interface{}{
			"username": req.Username,
			"email":    req.Email,
		})
		utils.BadRequestResponse(c, errors.Error())
		return
	}

	// Проверяем, не существует ли пользователь
	var exists bool
	err := h.DB.GetContext(ctx, &exists, `
		SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 OR email = $2)
	`, req.Username, req.Email)

	if err != nil {
		h.logger.ErrorContext(ctx, "Database error during user existence check",
			slog.String("error", err.Error()),
			slog.String("client_ip", c.ClientIP()),
		)
		utils.InternalErrorResponse(c, "Database error")
		return
	}

	if exists {
		h.logSecurityEvent(ctx, "register_user_exists", 0, c.ClientIP(), map[string]interface{}{
			"username": req.Username,
			"email":    req.Email,
		})
		utils.ConflictResponse(c, "Username or email already exists")
		return
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to hash password",
			slog.String("error", err.Error()),
			slog.String("client_ip", c.ClientIP()),
		)
		utils.InternalErrorResponse(c, "Failed to process registration")
		return
	}

	// Создаем пользователя
	var userID int
	err = h.DB.QueryRowContext(ctx, `
		INSERT INTO users (username, email, password_hash, is_active, is_verified, created_at)
		VALUES ($1, $2, $3, true, false, CURRENT_TIMESTAMP)
		RETURNING id
	`, req.Username, req.Email, string(hashedPassword)).Scan(&userID)

	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to create user",
			slog.String("error", err.Error()),
			slog.String("client_ip", c.ClientIP()),
		)
		utils.InternalErrorResponse(c, "Failed to create user")
		return
	}

	// Генерируем токены
	accessToken, err := auth.GenerateToken(userID, req.Username)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to generate access token",
			slog.String("error", err.Error()),
			slog.Int("user_id", userID),
		)
		utils.InternalErrorResponse(c, "Failed to generate token")
		return
	}

	refreshToken, err := auth.GenerateRefreshToken(userID)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to generate refresh token",
			slog.String("error", err.Error()),
			slog.Int("user_id", userID),
		)
		utils.InternalErrorResponse(c, "Failed to generate refresh token")
		return
	}

	// Сохраняем refresh token в базе (хешированный)
	hashedRefreshToken := h.hashRefreshToken(refreshToken)
	_, err = h.DB.ExecContext(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, hashedRefreshToken, time.Now().Add(h.config.RefreshTokenTTL))

	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to save refresh token",
			slog.String("error", err.Error()),
			slog.Int("user_id", userID),
		)
		utils.InternalErrorResponse(c, "Failed to save refresh token")
		return
	}

	h.logSecurityEvent(ctx, "user_registered", userID, c.ClientIP(), map[string]interface{}{
		"username": req.Username,
		"email":    req.Email,
	})

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
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.config.DatabaseTimeout)
	defer cancel()

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logSecurityError(ctx, "invalid_login_request", c.ClientIP(), err, map[string]interface{}{
			"error_type": "validation",
		})
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Валидация
	errors := validator.ValidateUserLogin(req.Username, req.Password)
	if errors.HasErrors() {
		h.logSecurityError(ctx, "login_validation_failed", c.ClientIP(), fmt.Errorf("validation failed"), map[string]interface{}{
			"username": req.Username,
		})
		utils.BadRequestResponse(c, errors.Error())
		return
	}

	// Находим пользователя
	var user models.User
	err := h.DB.GetContext(ctx, &user, `
		SELECT id, username, email, password_hash, rating, wins, losses, is_active, is_verified
		FROM users WHERE username = $1
	`, req.Username)

	if err != nil {
		if err == sql.ErrNoRows {
			h.logSecurityEvent(ctx, "login_user_not_found", 0, c.ClientIP(), map[string]interface{}{
				"username": req.Username,
			})
			utils.UnauthorizedResponse(c, "Invalid credentials")
		} else {
			h.logger.ErrorContext(ctx, "Database error during login",
				slog.String("error", err.Error()),
				slog.String("client_ip", c.ClientIP()),
			)
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	// Проверяем статус пользователя
	if !user.IsActive {
		h.logSecurityEvent(ctx, "login_inactive_user", user.ID, c.ClientIP(), map[string]interface{}{
			"username": req.Username,
		})
		utils.UnauthorizedResponse(c, "Account is inactive")
		return
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		h.logSecurityEvent(ctx, "login_invalid_password", user.ID, c.ClientIP(), map[string]interface{}{
			"username": req.Username,
		})
		utils.UnauthorizedResponse(c, "Invalid credentials")
		return
	}

	// Блокируем операции для данного пользователя
	userMutex := h.getUserMutex(user.ID)
	userMutex.Lock()
	defer userMutex.Unlock()

	// Генерируем токены
	accessToken, err := auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to generate access token",
			slog.String("error", err.Error()),
			slog.Int("user_id", user.ID),
		)
		utils.InternalErrorResponse(c, "Failed to generate token")
		return
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to generate refresh token",
			slog.String("error", err.Error()),
			slog.Int("user_id", user.ID),
		)
		utils.InternalErrorResponse(c, "Failed to generate refresh token")
		return
	}

	// Начинаем транзакцию
	tx, err := h.DB.BeginTxx(ctx, nil)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to begin transaction",
			slog.String("error", err.Error()),
			slog.Int("user_id", user.ID),
		)
		utils.InternalErrorResponse(c, "Database transaction error")
		return
	}
	defer tx.Rollback()

	// Обновляем последний вход
	_, err = tx.ExecContext(ctx, `
		UPDATE users SET 
			last_login = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1
	`, user.ID)

	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to update last login",
			slog.String("error", err.Error()),
			slog.Int("user_id", user.ID),
		)
		utils.InternalErrorResponse(c, "Failed to update login time")
		return
	}

	// Сохраняем новый refresh token (хешированный)
	hashedRefreshToken := h.hashRefreshToken(refreshToken)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE SET
			token_hash = EXCLUDED.token_hash,
			expires_at = EXCLUDED.expires_at,
			updated_at = CURRENT_TIMESTAMP
	`, user.ID, hashedRefreshToken, time.Now().Add(h.config.RefreshTokenTTL))

	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to save refresh token",
			slog.String("error", err.Error()),
			slog.Int("user_id", user.ID),
		)
		utils.InternalErrorResponse(c, "Failed to save refresh token")
		return
	}

	// Коммитим транзакцию
	if err = tx.Commit(); err != nil {
		h.logger.ErrorContext(ctx, "Failed to commit login transaction",
			slog.String("error", err.Error()),
			slog.Int("user_id", user.ID),
		)
		utils.InternalErrorResponse(c, "Failed to complete login")
		return
	}

	// Скрываем пароль
	user.Password = ""

	h.logSecurityEvent(ctx, "user_logged_in", user.ID, c.ClientIP(), map[string]interface{}{
		"username": req.Username,
	})

	utils.SuccessResponse(c, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          user,
	}, "Login successful")
}

// RefreshToken обновление токена доступа
func (h *AuthHandlers) RefreshToken(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.config.DatabaseTimeout)
	defer cancel()

	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logSecurityError(ctx, "invalid_refresh_request", c.ClientIP(), err, map[string]interface{}{
			"error_type": "validation",
		})
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Хешируем переданный токен для поиска в БД
	hashedToken := h.hashRefreshToken(req.RefreshToken)

	// Получаем информацию о токене и пользователе одним запросом
	var result struct {
		UserID     int       `db:"user_id"`
		ExpiresAt  time.Time `db:"expires_at"`
		Username   string    `db:"username"`
		Email      string    `db:"email"`
		Rating     int       `db:"rating"`
		Wins       int       `db:"wins"`
		Losses     int       `db:"losses"`
		IsActive   bool      `db:"is_active"`
		IsVerified bool      `db:"is_verified"`
	}

	err := h.DB.GetContext(ctx, &result, `
		SELECT rt.user_id, rt.expires_at, u.username, u.email, u.rating, u.wins, u.losses, u.is_active, u.is_verified
		FROM refresh_tokens rt
		JOIN users u ON rt.user_id = u.id
		WHERE rt.token_hash = $1 AND rt.expires_at > CURRENT_TIMESTAMP
	`, hashedToken)

	if err != nil {
		if err == sql.ErrNoRows {
			h.logSecurityEvent(ctx, "refresh_token_invalid", 0, c.ClientIP(), nil)
			utils.UnauthorizedResponse(c, "Invalid or expired refresh token")
		} else {
			h.logger.ErrorContext(ctx, "Database error during token refresh",
				slog.String("error", err.Error()),
				slog.String("client_ip", c.ClientIP()),
			)
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	// Проверяем статус пользователя
	if !result.IsActive {
		h.logSecurityEvent(ctx, "refresh_inactive_user", result.UserID, c.ClientIP(), nil)
		utils.UnauthorizedResponse(c, "Account is inactive")
		return
	}

	// Блокируем операции для данного пользователя
	userMutex := h.getUserMutex(result.UserID)
	userMutex.Lock()
	defer userMutex.Unlock()

	// Генерируем новый access token
	accessToken, err := auth.GenerateToken(result.UserID, result.Username)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to generate access token during refresh",
			slog.String("error", err.Error()),
			slog.Int("user_id", result.UserID),
		)
		utils.InternalErrorResponse(c, "Failed to generate token")
		return
	}

	// Опционально генерируем новый refresh token если старый скоро истекает
	var newRefreshToken string
	if time.Until(result.ExpiresAt) < h.config.RefreshTokenRotationThreshold {
		newRefreshToken, err = auth.GenerateRefreshToken(result.UserID)
		if err != nil {
			h.logger.ErrorContext(ctx, "Failed to generate new refresh token",
				slog.String("error", err.Error()),
				slog.Int("user_id", result.UserID),
			)
			utils.InternalErrorResponse(c, "Failed to generate refresh token")
			return
		}

		// Обновляем refresh token в базе
		newHashedToken := h.hashRefreshToken(newRefreshToken)
		_, err = h.DB.ExecContext(ctx, `
			UPDATE refresh_tokens 
			SET token_hash = $1, expires_at = $2, updated_at = CURRENT_TIMESTAMP
			WHERE user_id = $3
		`, newHashedToken, time.Now().Add(h.config.RefreshTokenTTL), result.UserID)

		if err != nil {
			h.logger.ErrorContext(ctx, "Failed to update refresh token",
				slog.String("error", err.Error()),
				slog.Int("user_id", result.UserID),
			)
			utils.InternalErrorResponse(c, "Failed to update refresh token")
			return
		}

		h.logSecurityEvent(ctx, "refresh_token_rotated", result.UserID, c.ClientIP(), nil)
	} else {
		newRefreshToken = req.RefreshToken
	}

	h.logSecurityEvent(ctx, "token_refreshed", result.UserID, c.ClientIP(), nil)

	utils.SuccessResponse(c, gin.H{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
	}, "Token refreshed successfully")
}

// Logout выход пользователя
func (h *AuthHandlers) Logout(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.config.DatabaseTimeout)
	defer cancel()

	userID := c.GetInt("user_id")

	// Удаляем refresh token из базы
	_, err := h.DB.ExecContext(ctx, `
		DELETE FROM refresh_tokens WHERE user_id = $1
	`, userID)

	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to logout user",
			slog.String("error", err.Error()),
			slog.Int("user_id", userID),
		)
		utils.InternalErrorResponse(c, "Failed to logout")
		return
	}

	h.logSecurityEvent(ctx, "user_logged_out", userID, c.ClientIP(), nil)

	utils.NoContentResponse(c, "Logout successful")
}

// ChangePassword смена пароля
func (h *AuthHandlers) ChangePassword(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.config.DatabaseTimeout)
	defer cancel()

	userID := c.GetInt("user_id")

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logSecurityError(ctx, "invalid_change_password_request", c.ClientIP(), err, map[string]interface{}{
			"user_id": userID,
		})
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Валидация нового пароля
	if err := validator.ValidatePassword(req.NewPassword); err != nil {
		h.logSecurityEvent(ctx, "change_password_validation_failed", userID, c.ClientIP(), nil)
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Блокируем операции для данного пользователя
	userMutex := h.getUserMutex(userID)
	userMutex.Lock()
	defer userMutex.Unlock()

	// Получаем текущий хеш пароля
	var currentHash string
	err := h.DB.GetContext(ctx, &currentHash, `
		SELECT password_hash FROM users WHERE id = $1 AND is_active = true
	`, userID)

	if err != nil {
		if err == sql.ErrNoRows {
			h.logSecurityEvent(ctx, "change_password_user_not_found", userID, c.ClientIP(), nil)
			utils.UnauthorizedResponse(c, "User not found")
		} else {
			h.logger.ErrorContext(ctx, "Database error during password change",
				slog.String("error", err.Error()),
				slog.Int("user_id", userID),
			)
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	// Проверяем текущий пароль
	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(req.CurrentPassword)); err != nil {
		h.logSecurityEvent(ctx, "change_password_invalid_current", userID, c.ClientIP(), nil)
		utils.UnauthorizedResponse(c, "Current password is incorrect")
		return
	}

	// Хешируем новый пароль
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to hash new password",
			slog.String("error", err.Error()),
			slog.Int("user_id", userID),
		)
		utils.InternalErrorResponse(c, "Failed to process password")
		return
	}

	// Начинаем транзакцию
	tx, err := h.DB.BeginTxx(ctx, nil)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to begin password change transaction",
			slog.String("error", err.Error()),
			slog.Int("user_id", userID),
		)
		utils.InternalErrorResponse(c, "Database transaction error")
		return
	}
	defer tx.Rollback()

	// Обновляем пароль
	_, err = tx.ExecContext(ctx, `
		UPDATE users 
		SET password_hash = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, string(newHash), userID)

	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to update password",
			slog.String("error", err.Error()),
			slog.Int("user_id", userID),
		)
		utils.InternalErrorResponse(c, "Failed to update password")
		return
	}

	// Удаляем все refresh токены для принудительного переавторизации
	_, err = tx.ExecContext(ctx, `
		DELETE FROM refresh_tokens WHERE user_id = $1
	`, userID)

	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to revoke refresh tokens during password change",
			slog.String("error", err.Error()),
			slog.Int("user_id", userID),
		)
		utils.InternalErrorResponse(c, "Failed to revoke tokens")
		return
	}

	// Коммитим транзакцию
	if err = tx.Commit(); err != nil {
		h.logger.ErrorContext(ctx, "Failed to commit password change transaction",
			slog.String("error", err.Error()),
			slog.Int("user_id", userID),
		)
		utils.InternalErrorResponse(c, "Failed to complete password change")
		return
	}

	h.logSecurityEvent(ctx, "password_changed", userID, c.ClientIP(), nil)

	utils.NoContentResponse(c, "Password changed successfully")
}

// ForgotPassword запрос на восстановление пароля
func (h *AuthHandlers) ForgotPassword(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.config.DatabaseTimeout)
	defer cancel()

	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logSecurityError(ctx, "invalid_forgot_password_request", c.ClientIP(), err, nil)
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Валидация email
	if err := validator.ValidateEmail(req.Email); err != nil {
		h.logSecurityEvent(ctx, "forgot_password_invalid_email", 0, c.ClientIP(), map[string]interface{}{
			"email": req.Email,
		})
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Проверяем существование активного пользователя
	var userID int
	err := h.DB.GetContext(ctx, &userID, `
		SELECT id FROM users WHERE email = $1 AND is_active = true
	`, req.Email)

	if err != nil {
		// Не раскрываем информацию о существовании пользователя
		if err != sql.ErrNoRows {
			h.logger.ErrorContext(ctx, "Database error during forgot password",
				slog.String("error", err.Error()),
				slog.String("client_ip", c.ClientIP()),
			)
		}
		utils.SuccessResponse(c, nil, "If the email exists, a reset link has been sent")
		return
	}

	h.logSecurityEvent(ctx, "password_reset_requested", userID, c.ClientIP(), map[string]interface{}{
		"email": req.Email,
	})

	// Генерируем токен сброса
	resetToken, err := auth.GenerateResetToken(userID)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to generate reset token",
			slog.String("error", err.Error()),
			slog.Int("user_id", userID),
		)
		utils.InternalErrorResponse(c, "Failed to generate reset token")
		return
	}

	// Сохраняем токен в базе (действителен согласно конфигурации)
	_, err = h.DB.ExecContext(ctx, `
		INSERT INTO password_reset_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE SET
			token = EXCLUDED.token,
			expires_at = EXCLUDED.expires_at,
			created_at = CURRENT_TIMESTAMP
	`, userID, resetToken, time.Now().Add(h.config.ResetTokenTTL))

	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to save reset token",
			slog.String("error", err.Error()),
			slog.Int("user_id", userID),
		)
		utils.InternalErrorResponse(c, "Failed to save reset token")
		return
	}

	// TODO: Отправить email с ссылкой для сброса
	// В реальном приложении здесь должна быть отправка email

	utils.SuccessResponse(c, nil, "If the email exists, a reset link has been sent")
}

// ResetPassword сброс пароля по токену
func (h *AuthHandlers) ResetPassword(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.config.DatabaseTimeout)
	defer cancel()

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logSecurityError(ctx, "invalid_reset_password_request", c.ClientIP(), err, nil)
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Валидация нового пароля
	if err := validator.ValidatePassword(req.NewPassword); err != nil {
		h.logSecurityEvent(ctx, "reset_password_validation_failed", 0, c.ClientIP(), nil)
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Проверяем токен сброса
	var userID int
	err := h.DB.QueryRowContext(ctx, `
		SELECT user_id FROM password_reset_tokens 
		WHERE token = $1 AND expires_at > CURRENT_TIMESTAMP
	`, req.Token).Scan(&userID)

	if err != nil {
		if err == sql.ErrNoRows {
			h.logSecurityEvent(ctx, "reset_password_invalid_token", 0, c.ClientIP(), nil)
			utils.UnauthorizedResponse(c, "Invalid or expired reset token")
		} else {
			h.logger.ErrorContext(ctx, "Database error during password reset",
				slog.String("error", err.Error()),
				slog.String("client_ip", c.ClientIP()),
			)
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	// Проверяем, что пользователь активен
	var isActive bool
	err = h.DB.GetContext(ctx, &isActive, `
		SELECT is_active FROM users WHERE id = $1
	`, userID)

	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to check user status during password reset",
			slog.String("error", err.Error()),
			slog.Int("user_id", userID),
		)
		utils.InternalErrorResponse(c, "Database error")
		return
	}

	if !isActive {
		h.logSecurityEvent(ctx, "reset_password_inactive_user", userID, c.ClientIP(), nil)
		utils.UnauthorizedResponse(c, "Account is inactive")
		return
	}

	// Блокируем операции для данного пользователя
	userMutex := h.getUserMutex(userID)
	userMutex.Lock()
	defer userMutex.Unlock()

	// Хешируем новый пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to hash password during reset",
			slog.String("error", err.Error()),
			slog.Int("user_id", userID),
		)
		utils.InternalErrorResponse(c, "Failed to process password")
		return
	}

	// Начинаем транзакцию
	tx, err := h.DB.BeginTxx(ctx, nil)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to begin password reset transaction",
			slog.String("error", err.Error()),
			slog.Int("user_id", userID),
		)
		utils.InternalErrorResponse(c, "Database transaction error")
		return
	}
	defer tx.Rollback()

	// Обновляем пароль
	_, err = tx.ExecContext(ctx, `
		UPDATE users 
		SET password_hash = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, string(hashedPassword), userID)

	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to update password during reset",
			slog.String("error", err.Error()),
			slog.Int("user_id", userID),
		)
		utils.InternalErrorResponse(c, "Failed to update password")
		return
	}

	// Удаляем использованный токен сброса
	_, err = tx.ExecContext(ctx, `
		DELETE FROM password_reset_tokens WHERE user_id = $1
	`, userID)

	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to cleanup reset token",
			slog.String("error", err.Error()),
			slog.Int("user_id", userID),
		)
		utils.InternalErrorResponse(c, "Failed to cleanup reset token")
		return
	}

	// Удаляем все refresh токены
	_, err = tx.ExecContext(ctx, `
		DELETE FROM refresh_tokens WHERE user_id = $1
	`, userID)

	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to revoke tokens during password reset",
			slog.String("error", err.Error()),
			slog.Int("user_id", userID),
		)
		utils.InternalErrorResponse(c, "Failed to revoke tokens")
		return
	}

	// Коммитим транзакцию
	if err = tx.Commit(); err != nil {
		h.logger.ErrorContext(ctx, "Failed to commit password reset transaction",
			slog.String("error", err.Error()),
			slog.Int("user_id", userID),
		)
		utils.InternalErrorResponse(c, "Failed to commit transaction")
		return
	}

	h.logSecurityEvent(ctx, "password_reset_completed", userID, c.ClientIP(), nil)

	utils.NoContentResponse(c, "Password reset successfully")
}

// CleanupExpiredTokens очищает просроченные токены (должно вызываться периодически)
func (h *AuthHandlers) CleanupExpiredTokens() error {
	ctx, cancel := context.WithTimeout(context.Background(), h.config.DatabaseTimeout)
	defer cancel()

	// Удаляем просроченные refresh токены
	result1, err := h.DB.ExecContext(ctx, `
		DELETE FROM refresh_tokens WHERE expires_at <= CURRENT_TIMESTAMP
	`)
	if err != nil {
		h.logger.Error("Failed to cleanup expired refresh tokens", slog.String("error", err.Error()))
		return err
	}

	refreshCount, _ := result1.RowsAffected()

	// Удаляем просроченные reset токены
	result2, err := h.DB.ExecContext(ctx, `
		DELETE FROM password_reset_tokens WHERE expires_at <= CURRENT_TIMESTAMP
	`)
	if err != nil {
		h.logger.Error("Failed to cleanup expired reset tokens", slog.String("error", err.Error()))
		return err
	}

	resetCount, _ := result2.RowsAffected()

	if refreshCount > 0 || resetCount > 0 {
		h.logger.Info("Cleaned up expired tokens",
			slog.Int64("refresh_tokens", refreshCount),
			slog.Int64("reset_tokens", resetCount),
		)
	}

	return nil
}
