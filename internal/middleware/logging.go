// internal/middleware/logging.go
package middleware

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LogConfig конфигурация для логирования
type LogConfig struct {
	TimeFormat    string
	UTC           bool
	SkipPaths     []string
	SkipPathRegex []string
	Output        io.Writer
	LogLevel      LogLevel
}

// LogLevel уровни логирования
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// LogEntry структура записи лога
type LogEntry struct {
	RequestID    string        `json:"request_id"`
	Timestamp    time.Time     `json:"timestamp"`
	Method       string        `json:"method"`
	Path         string        `json:"path"`
	Query        string        `json:"query,omitempty"`
	StatusCode   int           `json:"status_code"`
	ResponseTime time.Duration `json:"response_time"`
	ClientIP     string        `json:"client_ip"`
	UserAgent    string        `json:"user_agent"`
	UserID       interface{}   `json:"user_id,omitempty"`
	RequestSize  int64         `json:"request_size"`
	ResponseSize int64         `json:"response_size"`
	Referer      string        `json:"referer,omitempty"`
	Error        string        `json:"error,omitempty"`
	Level        string        `json:"level"`
}

// responseBodyWriter захватывает response body для логирования
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// DefaultLogConfig возвращает стандартную конфигурацию логирования
func DefaultLogConfig() LogConfig {
	return LogConfig{
		TimeFormat: time.RFC3339,
		UTC:        true,
		SkipPaths:  []string{"/health", "/metrics", "/favicon.ico"},
		Output:     os.Stdout,
		LogLevel:   LogLevelInfo,
	}
}

// LoggingMiddleware создает middleware для логирования запросов
func LoggingMiddleware() gin.HandlerFunc {
	return LoggingWithConfig(DefaultLogConfig())
}

// LoggingWithConfig создает middleware с кастомной конфигурацией
func LoggingWithConfig(config LogConfig) gin.HandlerFunc {
	formatter := getLogFormatter(config)

	return func(c *gin.Context) {
		// Пропускаем указанные пути
		path := c.Request.URL.Path
		for _, skipPath := range config.SkipPaths {
			if path == skipPath {
				c.Next()
				return
			}
		}

		// Генерируем уникальный ID запроса
		requestID := generateRequestID()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		// Засекаем время начала
		start := time.Now()

		// Захватываем размер запроса
		var requestSize int64
		if c.Request.Body != nil {
			requestSize = c.Request.ContentLength
		}

		// Захватываем response body для подсчета размера
		blw := &responseBodyWriter{
			body:           bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
		}
		c.Writer = blw

		// Выполняем запрос
		c.Next()

		// Рассчитываем время выполнения
		duration := time.Since(start)

		// Получаем размер ответа
		responseSize := int64(blw.body.Len())

		// Создаем запись лога
		entry := LogEntry{
			RequestID:    requestID,
			Timestamp:    getTimestamp(config),
			Method:       c.Request.Method,
			Path:         path,
			Query:        c.Request.URL.RawQuery,
			StatusCode:   c.Writer.Status(),
			ResponseTime: duration,
			ClientIP:     c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			RequestSize:  requestSize,
			ResponseSize: responseSize,
			Referer:      c.Request.Referer(),
			Level:        getLogLevel(c.Writer.Status()),
		}

		// Добавляем user ID если доступен
		if userID, exists := c.Get("user_id"); exists {
			entry.UserID = userID
		}

		// Добавляем ошибку если есть
		if len(c.Errors) > 0 {
			entry.Error = c.Errors.String()
		}

		// Записываем лог
		formatter(config.Output, entry)
	}
}

// StructuredLoggingMiddleware создает структурированный JSON лог
func StructuredLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := generateRequestID()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		start := time.Now()

		c.Next()

		duration := time.Since(start)

		logData := map[string]interface{}{
			"request_id":    requestID,
			"timestamp":     time.Now().UTC().Format(time.RFC3339),
			"method":        c.Request.Method,
			"path":          c.Request.URL.Path,
			"query":         c.Request.URL.RawQuery,
			"status_code":   c.Writer.Status(),
			"response_time": duration.Milliseconds(),
			"client_ip":     c.ClientIP(),
			"user_agent":    c.Request.UserAgent(),
			"referer":       c.Request.Referer(),
		}

		if userID, exists := c.Get("user_id"); exists {
			logData["user_id"] = userID
		}

		if len(c.Errors) > 0 {
			logData["errors"] = c.Errors.String()
		}

		// Логируем в зависимости от статус кода
		if c.Writer.Status() >= 500 {
			log.Printf("ERROR: %+v", logData)
		} else if c.Writer.Status() >= 400 {
			log.Printf("WARN: %+v", logData)
		} else {
			log.Printf("INFO: %+v", logData)
		}
	}
}

// ColoredLoggingMiddleware создает цветной лог для разработки
func ColoredLoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		var statusColor, methodColor, resetColor string

		if param.IsOutputColor() {
			statusColor = getStatusColor(param.StatusCode)
			methodColor = getMethodColor(param.Method)
			resetColor = "\033[0m"
		}

		return fmt.Sprintf("%s[%s]%s %s%3d%s %13v | %15s | %s%-7s%s %s\n%s",
			"\033[90m", param.TimeStamp.Format("2006/01/02 - 15:04:05"), resetColor,
			statusColor, param.StatusCode, resetColor,
			param.Latency,
			param.ClientIP,
			methodColor, param.Method, resetColor,
			param.Path,
			param.ErrorMessage,
		)
	})
}

// SecurityLoggingMiddleware логирует события безопасности
func SecurityLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Логируем подозрительные запросы
		suspiciousPatterns := []string{
			"../", "..\\", "/etc/passwd", "/proc/", "cmd=", "eval(",
			"<script", "javascript:", "vbscript:", "onload=", "onerror=",
		}

		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		userAgent := c.Request.UserAgent()

		for _, pattern := range suspiciousPatterns {
			if contains(path, pattern) || contains(query, pattern) || contains(userAgent, pattern) {
				log.Printf("SECURITY: Suspicious request detected - IP: %s, Path: %s, Query: %s, UA: %s",
					c.ClientIP(), path, query, userAgent)
				break
			}
		}

		c.Next()

		// Логируем неудачные попытки аутентификации
		if c.Writer.Status() == 401 || c.Writer.Status() == 403 {
			log.Printf("SECURITY: Authentication/Authorization failure - IP: %s, Path: %s, Status: %d",
				c.ClientIP(), path, c.Writer.Status())
		}
	}
}

// AuditLoggingMiddleware логирует действия пользователей для аудита
func AuditLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Логируем только определенные действия
		auditPaths := map[string]bool{
			"/api/v1/auth/login":    true,
			"/api/v1/auth/register": true,
			"/api/v1/rooms":         c.Request.Method == "POST" || c.Request.Method == "DELETE",
			"/api/v1/tournaments":   c.Request.Method == "POST",
		}

		path := c.Request.URL.Path
		shouldAudit := false

		for auditPath, condition := range auditPaths {
			if matchPath(path, auditPath) && condition {
				shouldAudit = true
				break
			}
		}

		if !shouldAudit {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")

		auditEntry := map[string]interface{}{
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
			"action":      fmt.Sprintf("%s %s", c.Request.Method, path),
			"user_id":     userID,
			"username":    username,
			"client_ip":   c.ClientIP(),
			"user_agent":  c.Request.UserAgent(),
			"status_code": c.Writer.Status(),
			"duration":    time.Since(start).Milliseconds(),
		}

		log.Printf("AUDIT: %+v", auditEntry)
	}
}

// PerformanceLoggingMiddleware логирует медленные запросы
func PerformanceLoggingMiddleware(threshold time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		duration := time.Since(start)

		if duration > threshold {
			log.Printf("PERFORMANCE: Slow request detected - Duration: %v, Path: %s %s, Status: %d, IP: %s",
				duration, c.Request.Method, c.Request.URL.Path, c.Writer.Status(), c.ClientIP())
		}
	}
}

// DatabaseLoggingMiddleware логирует запросы в базу данных
func DatabaseLoggingMiddleware(logDB func(LogEntry)) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := generateRequestID()
		c.Set("request_id", requestID)

		start := time.Now()
		c.Next()

		entry := LogEntry{
			RequestID:    requestID,
			Timestamp:    time.Now().UTC(),
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			Query:        c.Request.URL.RawQuery,
			StatusCode:   c.Writer.Status(),
			ResponseTime: time.Since(start),
			ClientIP:     c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			Level:        getLogLevel(c.Writer.Status()),
		}

		if userID, exists := c.Get("user_id"); exists {
			entry.UserID = userID
		}

		if len(c.Errors) > 0 {
			entry.Error = c.Errors.String()
		}

		// Асинхронно сохраняем в базу данных
		go logDB(entry)
	}
}

// Вспомогательные функции

func generateRequestID() string {
	return uuid.New().String()
}

func getTimestamp(config LogConfig) time.Time {
	if config.UTC {
		return time.Now().UTC()
	}
	return time.Now()
}

func getLogLevel(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "ERROR"
	case statusCode >= 400:
		return "WARN"
	default:
		return "INFO"
	}
}

func getLogFormatter(config LogConfig) func(io.Writer, LogEntry) {
	return func(writer io.Writer, entry LogEntry) {
		timestamp := entry.Timestamp.Format(config.TimeFormat)

		fmt.Fprintf(writer, "[%s] %s %s %s %d %v %s \"%s\" \"%s\"",
			timestamp,
			entry.Level,
			entry.RequestID,
			entry.Method,
			entry.StatusCode,
			entry.ResponseTime,
			entry.Path,
			entry.UserAgent,
			entry.ClientIP,
		)

		if entry.UserID != nil {
			fmt.Fprintf(writer, " user=%v", entry.UserID)
		}

		if entry.Error != "" {
			fmt.Fprintf(writer, " error=\"%s\"", entry.Error)
		}

		fmt.Fprintln(writer)
	}
}

func getStatusColor(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "\033[97;42m" // Green
	case code >= 300 && code < 400:
		return "\033[90;47m" // White
	case code >= 400 && code < 500:
		return "\033[90;43m" // Yellow
	default:
		return "\033[97;41m" // Red
	}
}

func getMethodColor(method string) string {
	switch method {
	case "GET":
		return "\033[94m" // Blue
	case "POST":
		return "\033[92m" // Green
	case "PUT":
		return "\033[93m" // Yellow
	case "DELETE":
		return "\033[91m" // Red
	case "PATCH":
		return "\033[95m" // Magenta
	case "HEAD":
		return "\033[96m" // Cyan
	case "OPTIONS":
		return "\033[90m" // Gray
	default:
		return "\033[0m" // Reset
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			contains(s[1:], substr))))
}

func matchPath(path, pattern string) bool {
	return path == pattern || (len(path) > len(pattern) && path[:len(pattern)] == pattern)
}
