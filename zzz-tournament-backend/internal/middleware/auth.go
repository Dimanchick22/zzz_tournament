// internal/middleware/auth.go
package middleware

import (
	"fmt"
	"strings"

	"zzz-tournament/pkg/auth"
	"zzz-tournament/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware проверяет JWT токен и устанавливает данные пользователя в контекст
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.UnauthorizedResponse(c, "Authorization header required")
			c.Abort()
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			utils.UnauthorizedResponse(c, "Invalid authorization header format. Use: Bearer <token>")
			c.Abort()
			return
		}

		tokenString := bearerToken[1]
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			utils.UnauthorizedResponse(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Устанавливаем данные пользователя в контекст
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("token_claims", claims)

		c.Next()
	}
}

// OptionalAuthMiddleware похож на AuthMiddleware, но не прерывает выполнение при отсутствии токена
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := bearerToken[1]
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		// Устанавливаем данные пользователя в контекст если токен валидный
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("token_claims", claims)

		c.Next()
	}
}

// AdminOnlyMiddleware проверяет, что пользователь является администратором
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.UnauthorizedResponse(c, "Authentication required")
			c.Abort()
			return
		}

		// TODO: Проверить в базе данных, является ли пользователь администратором
		// Пока что считаем, что admin это пользователь с ID = 1
		if userID.(int) != 1 {
			utils.ForbiddenResponse(c, "Administrator access required")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RoomHostOnlyMiddleware проверяет, что пользователь является хостом комнаты
func RoomHostOnlyMiddleware(getRoomHostID func(roomID int) (int, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.UnauthorizedResponse(c, "Authentication required")
			c.Abort()
			return
		}

		roomIDStr := c.Param("id")
		if roomIDStr == "" {
			utils.BadRequestResponse(c, "Room ID is required")
			c.Abort()
			return
		}

		// Конвертируем room ID в int
		var roomID int
		if _, err := fmt.Sscanf(roomIDStr, "%d", &roomID); err != nil {
			utils.BadRequestResponse(c, "Invalid room ID format")
			c.Abort()
			return
		}

		// Получаем ID хоста комнаты
		hostID, err := getRoomHostID(roomID)
		if err != nil {
			utils.NotFoundResponse(c, "Room not found")
			c.Abort()
			return
		}

		if userID.(int) != hostID {
			utils.ForbiddenResponse(c, "Only room host can perform this action")
			c.Abort()
			return
		}

		c.Set("room_id", roomID)
		c.Set("host_id", hostID)
		c.Next()
	}
}

// TournamentParticipantMiddleware проверяет, что пользователь участвует в турнире
func TournamentParticipantMiddleware(isParticipant func(userID, tournamentID int) bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.UnauthorizedResponse(c, "Authentication required")
			c.Abort()
			return
		}

		tournamentIDStr := c.Param("id")
		if tournamentIDStr == "" {
			utils.BadRequestResponse(c, "Tournament ID is required")
			c.Abort()
			return
		}

		var tournamentID int
		if _, err := fmt.Sscanf(tournamentIDStr, "%d", &tournamentID); err != nil {
			utils.BadRequestResponse(c, "Invalid tournament ID format")
			c.Abort()
			return
		}

		if !isParticipant(userID.(int), tournamentID) {
			utils.ForbiddenResponse(c, "Only tournament participants can access this resource")
			c.Abort()
			return
		}

		c.Set("tournament_id", tournamentID)
		c.Next()
	}
}

// RefreshTokenMiddleware обновляет токен если он скоро истекает
func RefreshTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := bearerToken[1]
		newToken, err := auth.RefreshToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		// Если токен был обновлен, добавляем новый токен в заголовок ответа
		if newToken != tokenString {
			c.Header("X-New-Token", newToken)
		}

		c.Next()
	}
}

// ValidateUserExists проверяет существование пользователя в базе данных
func ValidateUserExists(getUserByID func(int) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		if err := getUserByID(userID.(int)); err != nil {
			utils.UnauthorizedResponse(c, "User account no longer exists")
			c.Abort()
			return
		}

		c.Next()
	}
}

// APIKeyMiddleware проверяет API ключ для внешних интеграций
func APIKeyMiddleware(validAPIKeys []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			utils.UnauthorizedResponse(c, "API key required")
			c.Abort()
			return
		}

		isValid := false
		for _, validKey := range validAPIKeys {
			if apiKey == validKey {
				isValid = true
				break
			}
		}

		if !isValid {
			utils.UnauthorizedResponse(c, "Invalid API key")
			c.Abort()
			return
		}

		c.Set("api_key", apiKey)
		c.Next()
	}
}

// GetUserID получает ID пользователя из контекста
func GetUserID(c *gin.Context) (int, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	return userID.(int), true
}

// GetUsername получает имя пользователя из контекста
func GetUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}
	return username.(string), true
}

// GetTokenClaims получает claims токена из контекста
func GetTokenClaims(c *gin.Context) (*auth.Claims, bool) {
	claims, exists := c.Get("token_claims")
	if !exists {
		return nil, false
	}
	return claims.(*auth.Claims), true
}

// RequireUserID middleware, который гарантирует наличие user_id в контексте
func RequireUserID() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := GetUserID(c); !exists {
			utils.UnauthorizedResponse(c, "User authentication required")
			c.Abort()
			return
		}
		c.Next()
	}
}
