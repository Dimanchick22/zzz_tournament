// internal/middleware/cors.go
package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CORSConfig конфигурация для CORS
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// DefaultCORSConfig возвращает стандартную конфигурацию CORS
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{
			"http://localhost:3000",  // React dev server
			"http://localhost:3001",  // Alternative React port
			"http://127.0.0.1:3000",
			"http://127.0.0.1:3001",
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodHead,
		},
		AllowHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Requested-With",
			"X-API-Key",
			"X-Client-Version",
			"Cache-Control",
		},
		ExposeHeaders: []string{
			"X-New-Token",
			"X-Rate-Limit-Remaining",
			"X-Rate-Limit-Reset",
			"X-Total-Count",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
}

// ProductionCORSConfig возвращает конфигурацию CORS для продакшена
func ProductionCORSConfig(allowedOrigins []string) CORSConfig {
	config := DefaultCORSConfig()
	config.AllowOrigins = allowedOrigins
	return config
}

// CORSMiddleware создает middleware для обработки CORS
func CORSMiddleware() gin.HandlerFunc {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig создает middleware с кастомной конфигурацией
func CORSWithConfig(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Проверяем, разрешен ли origin
		if origin != "" && isOriginAllowed(origin, config.AllowOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		// Устанавливаем другие CORS заголовки
		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if len(config.AllowMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
		}

		if len(config.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
		}

		if len(config.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
		}

		if config.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", formatDuration(config.MaxAge))
		}

		// Обрабатываем preflight запросы
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// StrictCORSMiddleware строгий CORS только для определенных origins
func StrictCORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		if origin == "" {
			// Разрешаем запросы без Origin (например, от Postman)
			c.Next()
			return
		}

		if !isOriginAllowed(origin, allowedOrigins) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Origin not allowed",
			})
			return
		}

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// DevCORSMiddleware разрешительный CORS для разработки
func DevCORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// В режиме разработки разрешаем все localhost origins
		if origin != "" && (strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1")) {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if origin == "" {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "*")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Access-Control-Expose-Headers", "*")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// NoSniffMiddleware добавляет заголовки безопасности
func NoSniffMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// SecureHeadersMiddleware добавляет заголовки безопасности для продакшена
func SecureHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Next()
	}
}

// isOriginAllowed проверяет, разрешен ли origin
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" {
			return true
		}
		if origin == allowedOrigin {
			return true
		}
		// Поддержка wildcard субдоменов (например: *.example.com)
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := allowedOrigin[2:]
			if strings.HasSuffix(origin, "."+domain) || origin == domain {
				return true
			}
		}
	}
	return false
}

// formatDuration форматирует duration в секунды для заголовка Max-Age
func formatDuration(d time.Duration) string {
	return string(int(d.Seconds()))
}

// WebSocketCORSMiddleware специальный CORS для WebSocket соединений
func WebSocketCORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Разрешаем WebSocket соединения от разрешенных origins
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:3001",
		}

		if origin != "" && isOriginAllowed(origin, allowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		c.Next()
	}
}

// APICORSMiddleware CORS для API endpoints
func APICORSMiddleware() gin.HandlerFunc {
	config := CORSConfig{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"https://zzz-tournament.example.com", // Замените на ваш домен
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-API-Key",
		},
		ExposeHeaders: []string{
			"X-Rate-Limit-Remaining",
			"X-Rate-Limit-Reset",
		},
		AllowCredentials: false, // API обычно не требует credentials
		MaxAge:           1 * time.Hour,
	}

	return CORSWithConfig(config)
}