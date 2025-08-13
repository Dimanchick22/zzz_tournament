// pkg/config/auth.go
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// AuthConfig содержит настройки аутентификации
type AuthConfig struct {
	// Время жизни токенов
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl" env:"ACCESS_TOKEN_TTL" default:"15m"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl" env:"REFRESH_TOKEN_TTL" default:"168h"` // 7 дней
	ResetTokenTTL   time.Duration `yaml:"reset_token_ttl" env:"RESET_TOKEN_TTL" default:"1h"`

	// Пороговое значение для ротации refresh токена
	RefreshTokenRotationThreshold time.Duration `yaml:"refresh_token_rotation_threshold" env:"REFRESH_TOKEN_ROTATION_THRESHOLD" default:"24h"`

	// Таймауты для операций с БД
	DatabaseTimeout time.Duration `yaml:"database_timeout" env:"DATABASE_TIMEOUT" default:"5s"`

	// Настройки rate limiting
	RateLimitEnabled bool          `yaml:"rate_limit_enabled" env:"RATE_LIMIT_ENABLED" default:"true"`
	LoginRateLimit   int           `yaml:"login_rate_limit" env:"LOGIN_RATE_LIMIT" default:"5"`     // попыток в окне
	LoginRateWindow  time.Duration `yaml:"login_rate_window" env:"LOGIN_RATE_WINDOW" default:"15m"` // размер окна

	RegisterRateLimit  int           `yaml:"register_rate_limit" env:"REGISTER_RATE_LIMIT" default:"3"`
	RegisterRateWindow time.Duration `yaml:"register_rate_window" env:"REGISTER_RATE_WINDOW" default:"1h"`

	ForgotPasswordRateLimit  int           `yaml:"forgot_password_rate_limit" env:"FORGOT_PASSWORD_RATE_LIMIT" default:"3"`
	ForgotPasswordRateWindow time.Duration `yaml:"forgot_password_rate_window" env:"FORGOT_PASSWORD_RATE_WINDOW" default:"1h"`

	// Настройки безопасности
	RequireEmailVerification bool          `yaml:"require_email_verification" env:"REQUIRE_EMAIL_VERIFICATION" default:"false"`
	MaxLoginAttempts         int           `yaml:"max_login_attempts" env:"MAX_LOGIN_ATTEMPTS" default:"5"`
	AccountLockoutDuration   time.Duration `yaml:"account_lockout_duration" env:"ACCOUNT_LOCKOUT_DURATION" default:"30m"`

	// Настройки логирования
	LogSecurityEvents bool   `yaml:"log_security_events" env:"LOG_SECURITY_EVENTS" default:"true"`
	LogLevel          string `yaml:"log_level" env:"LOG_LEVEL" default:"INFO"`

	// JWT настройки
	JWTSecret   string `yaml:"jwt_secret" env:"JWT_SECRET" required:"true"`
	JWTIssuer   string `yaml:"jwt_issuer" env:"JWT_ISSUER" default:"zzz-tournament"`
	JWTAudience string `yaml:"jwt_audience" env:"JWT_AUDIENCE" default:"zzz-tournament-users"`
}

// getEnv получает значение переменной окружения с fallback
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool получает boolean значение из переменной окружения
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvInt получает int значение из переменной окружения
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvDuration получает duration значение из переменной окружения
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// LoadAuthConfig загружает конфигурацию аутентификации
func LoadAuthConfig() (*AuthConfig, error) {
	config := &AuthConfig{
		// Время жизни токенов
		AccessTokenTTL:  getEnvDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL: getEnvDuration("REFRESH_TOKEN_TTL", 7*24*time.Hour),
		ResetTokenTTL:   getEnvDuration("RESET_TOKEN_TTL", time.Hour),

		// Ротация токенов
		RefreshTokenRotationThreshold: getEnvDuration("REFRESH_TOKEN_ROTATION_THRESHOLD", 24*time.Hour),

		// Таймауты
		DatabaseTimeout: getEnvDuration("DATABASE_TIMEOUT", 5*time.Second),

		// Rate limiting
		RateLimitEnabled:         getEnvBool("RATE_LIMIT_ENABLED", true),
		LoginRateLimit:           getEnvInt("LOGIN_RATE_LIMIT", 5),
		LoginRateWindow:          getEnvDuration("LOGIN_RATE_WINDOW", 15*time.Minute),
		RegisterRateLimit:        getEnvInt("REGISTER_RATE_LIMIT", 3),
		RegisterRateWindow:       getEnvDuration("REGISTER_RATE_WINDOW", time.Hour),
		ForgotPasswordRateLimit:  getEnvInt("FORGOT_PASSWORD_RATE_LIMIT", 3),
		ForgotPasswordRateWindow: getEnvDuration("FORGOT_PASSWORD_RATE_WINDOW", time.Hour),

		// Безопасность
		RequireEmailVerification: getEnvBool("REQUIRE_EMAIL_VERIFICATION", false),
		MaxLoginAttempts:         getEnvInt("MAX_LOGIN_ATTEMPTS", 5),
		AccountLockoutDuration:   getEnvDuration("ACCOUNT_LOCKOUT_DURATION", 30*time.Minute),

		// Логирование
		LogSecurityEvents: getEnvBool("LOG_SECURITY_EVENTS", true),
		LogLevel:          getEnv("LOG_LEVEL", "INFO"),

		// JWT
		JWTSecret:   getEnv("JWT_SECRET", ""),
		JWTIssuer:   getEnv("JWT_ISSUER", "zzz-tournament"),
		JWTAudience: getEnv("JWT_AUDIENCE", "zzz-tournament-users"),
	}

	return config, nil
}

// Validate проверяет корректность конфигурации
func (c *AuthConfig) Validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters long")
	}

	if c.AccessTokenTTL <= 0 {
		return fmt.Errorf("access token TTL must be positive")
	}

	if c.RefreshTokenTTL <= 0 {
		return fmt.Errorf("refresh token TTL must be positive")
	}

	if c.ResetTokenTTL <= 0 {
		return fmt.Errorf("reset token TTL must be positive")
	}

	if c.DatabaseTimeout <= 0 {
		return fmt.Errorf("database timeout must be positive")
	}

	if c.MaxLoginAttempts <= 0 {
		return fmt.Errorf("max login attempts must be positive")
	}

	return nil
}
