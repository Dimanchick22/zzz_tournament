// pkg/auth/jwt.go
package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var jwtSecret = []byte("your-secret-key") // TODO: move to config

func GenerateToken(userID int, username string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return claims, nil
}

func SetSecret(secret string) {
	jwtSecret = []byte(secret)
}

// RefreshToken обновляет токен, если он скоро истекает
func RefreshToken(tokenString string) (string, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Обновляем токен, если до истечения осталось меньше 6 часов
	if time.Until(claims.ExpiresAt.Time) < 6*time.Hour {
		return GenerateToken(claims.UserID, claims.Username)
	}

	return tokenString, nil
}

// GetUserFromToken извлекает данные пользователя из токена
func GetUserFromToken(tokenString string) (int, string, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return 0, "", err
	}

	return claims.UserID, claims.Username, nil
}