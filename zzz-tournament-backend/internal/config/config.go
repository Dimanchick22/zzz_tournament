// internal/config/config.go
package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL    string
	JWTSecret      string
	Port           string
	Environment    string
	MetricsAPIKey  string
	RedisURL       string
	SMTPHost       string
	SMTPPort       string
	SMTPUsername   string
	SMTPPassword   string
	AWSAccessKey   string
	AWSSecretKey   string
	AWSRegion      string
	AWSBucket      string
	LogLevel       string
	MaxUploadSize  int64
	SessionTimeout int
	RateLimitRedis bool
}

func Load() *Config {
	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://user:password@localhost/zzz_tournament?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key"),
		Port:           getEnv("PORT", "8080"),
		Environment:    getEnv("ENVIRONMENT", "development"),
		MetricsAPIKey:  getEnv("METRICS_API_KEY", "metrics-secret-key"),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379"),
		SMTPHost:       getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:       getEnv("SMTP_PORT", "587"),
		SMTPUsername:   getEnv("SMTP_USERNAME", ""),
		SMTPPassword:   getEnv("SMTP_PASSWORD", ""),
		AWSAccessKey:   getEnv("AWS_ACCESS_KEY", ""),
		AWSSecretKey:   getEnv("AWS_SECRET_KEY", ""),
		AWSRegion:      getEnv("AWS_REGION", "us-east-1"),
		AWSBucket:      getEnv("AWS_BUCKET", "zzz-tournament-uploads"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		MaxUploadSize:  getEnvInt64("MAX_UPLOAD_SIZE", 10*1024*1024), // 10MB
		SessionTimeout: getEnvInt("SESSION_TIMEOUT", 24),             // 24 hours
		RateLimitRedis: getEnvBool("RATE_LIMIT_REDIS", false),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// IsDevelopment проверяет, находимся ли мы в режиме разработки
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development" || c.Environment == "dev"
}

// IsProduction проверяет, находимся ли мы в продакшене
func (c *Config) IsProduction() bool {
	return c.Environment == "production" || c.Environment == "prod"
}

// IsStaging проверяет, находимся ли мы в staging
func (c *Config) IsStaging() bool {
	return c.Environment == "staging"
}
