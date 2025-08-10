// internal/middleware/recovery.go
package middleware

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"zzz-tournament/pkg/utils"

	"github.com/gin-gonic/gin"
)

// RecoveryConfig конфигурация для recovery middleware
type RecoveryConfig struct {
	Skipper          func(*gin.Context) bool
	BeforeRecover    func(*gin.Context, interface{})
	AfterRecover     func(*gin.Context, interface{})
	LogStack         bool
	PrintStack       bool
	LogLevel         RecoveryLogLevel
	Output           io.Writer
	DisableStackAll  bool
	DisableColorWhen func() bool
}

// RecoveryLogLevel уровни логирования для recovery
type RecoveryLogLevel int

const (
	RecoveryLogError RecoveryLogLevel = iota
	RecoveryLogFatal
	RecoveryLogPanic
)

// PanicInfo информация о панике
type PanicInfo struct {
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
	Method    string      `json:"method"`
	Path      string      `json:"path"`
	ClientIP  string      `json:"client_ip"`
	UserAgent string      `json:"user_agent"`
	UserID    interface{} `json:"user_id,omitempty"`
	Recovery  interface{} `json:"recovery"`
	Stack     string      `json:"stack"`
	Request   string      `json:"request,omitempty"`
}

// DefaultRecoveryConfig возвращает стандартную конфигурацию recovery
func DefaultRecoveryConfig() RecoveryConfig {
	return RecoveryConfig{
		Skipper:         nil,
		BeforeRecover:   nil,
		AfterRecover:    nil,
		LogStack:        true,
		PrintStack:      true,
		LogLevel:        RecoveryLogError,
		Output:          os.Stderr,
		DisableStackAll: false,
		DisableColorWhen: func() bool {
			return gin.Mode() == gin.ReleaseMode
		},
	}
}

// RecoveryMiddleware создает middleware для обработки паник
func RecoveryMiddleware() gin.HandlerFunc {
	return RecoveryWithConfig(DefaultRecoveryConfig())
}

// RecoveryWithConfig создает middleware с кастомной конфигурацией
func RecoveryWithConfig(config RecoveryConfig) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// Проверяем skipper
		if config.Skipper != nil && config.Skipper(c) {
			c.Abort()
			return
		}

		// Вызываем BeforeRecover hook
		if config.BeforeRecover != nil {
			config.BeforeRecover(c, recovered)
		}

		// Собираем информацию о панике
		panicInfo := collectPanicInfo(c, recovered)

		// Логируем панику
		logPanic(config, panicInfo)

		// Определяем, является ли это проблемой соединения
		if isBrokenPipe(recovered) {
			handleBrokenPipe(c, recovered)
			return
		}

		// Отправляем ответ об ошибке
		sendErrorResponse(c, recovered)

		// Вызываем AfterRecover hook
		if config.AfterRecover != nil {
			config.AfterRecover(c, recovered)
		}
	})
}

// DetailedRecoveryMiddleware расширенная обработка паник с детальным логированием
func DetailedRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				panicInfo := collectPanicInfo(c, recovered)

				// Детальное логирование
				logDetailedPanic(panicInfo)

				// Уведомление администраторов (в продакшене)
				if gin.Mode() == gin.ReleaseMode {
					go notifyAdmins(panicInfo)
				}

				// Сохранение в базу данных для анализа
				go savePanicToDB(panicInfo)

				sendErrorResponse(c, recovered)
			}
		}()

		c.Next()
	}
}

func sendErrorResponse(c *gin.Context, recovered interface{}) {
	if gin.Mode() == gin.ReleaseMode {
		// В продакшене не показываем детали ошибки
		utils.InternalErrorResponse(c, "Internal server error occurred")
	} else {
		// В режиме разработки показываем детали
		utils.ErrorResponseWithDetails(c, http.StatusInternalServerError,
			fmt.Sprintf("Panic recovered: %v", recovered))
	}
}

func collectPanicInfo(c *gin.Context, recovered interface{}) PanicInfo {
	stack := string(debug.Stack())

	// Получаем dump запроса
	var requestDump string
	if gin.Mode() != gin.ReleaseMode {
		if dump, err := httputil.DumpRequest(c.Request, false); err == nil {
			requestDump = string(dump)
		}
	}

	info := PanicInfo{
		Timestamp: time.Now().UTC(),
		Method:    c.Request.Method,
		Path:      c.Request.URL.Path,
		ClientIP:  c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Recovery:  recovered,
		Stack:     stack,
		Request:   requestDump,
	}

	// Добавляем request ID если есть
	if requestID, exists := c.Get("request_id"); exists {
		info.RequestID = requestID.(string)
	}

	// Добавляем user ID если есть
	if userID, exists := c.Get("user_id"); exists {
		info.UserID = userID
	}

	return info
}

func logPanic(config RecoveryConfig, info PanicInfo) {
	if config.LogStack {
		fmt.Fprintf(config.Output, "[PANIC RECOVERED] %s\n", time.Now().Format("2006/01/02 - 15:04:05"))
		fmt.Fprintf(config.Output, "Request: %s %s\n", info.Method, info.Path)
		fmt.Fprintf(config.Output, "Client IP: %s\n", info.ClientIP)
		if info.UserID != nil {
			fmt.Fprintf(config.Output, "User ID: %v\n", info.UserID)
		}
		fmt.Fprintf(config.Output, "Recovery: %v\n", info.Recovery)

		if config.PrintStack {
			fmt.Fprintf(config.Output, "Stack trace:\n%s\n", info.Stack)
		}

		fmt.Fprintf(config.Output, "Request dump:\n%s\n", info.Request)
		fmt.Fprintf(config.Output, "--- END PANIC REPORT ---\n\n")
	}
}

func logDetailedPanic(info PanicInfo) {
	log.Printf(`
=== DETAILED PANIC REPORT ===
Timestamp: %s
Request ID: %s
Method: %s
Path: %s
Client IP: %s
User Agent: %s
User ID: %v
Recovery: %v
Stack Trace:
%s
Request Details:
%s
=== END PANIC REPORT ===
`,
		info.Timestamp.Format(time.RFC3339),
		info.RequestID,
		info.Method,
		info.Path,
		info.ClientIP,
		info.UserAgent,
		info.UserID,
		info.Recovery,
		info.Stack,
		info.Request,
	)
}

func isBrokenPipe(recovered interface{}) bool {
	if ne, ok := recovered.(*net.OpError); ok {
		if se, ok := ne.Err.(*os.SyscallError); ok {
			errStr := strings.ToLower(se.Error())
			return strings.Contains(errStr, "broken pipe") ||
				strings.Contains(errStr, "connection reset by peer")
		}
	}
	return false
}

func handleBrokenPipe(c *gin.Context, recovered interface{}) {
	log.Printf("BROKEN_PIPE: Connection broken - %s %s from %s",
		c.Request.Method, c.Request.URL.Path, c.ClientIP())
	c.Error(fmt.Errorf("broken pipe: %v", recovered))
	c.Abort()
}

func notifyAdmins(info PanicInfo) {
	// Здесь должна быть логика отправки уведомлений администраторам
	log.Printf("ADMIN_NOTIFICATION: Panic occurred - %s %s", info.Method, info.Path)
}

func savePanicToDB(info PanicInfo) {
	// Здесь должна быть логика сохранения информации о панике в базу данных
	log.Printf("DB_SAVE: Saving panic info to database - Request ID: %s", info.RequestID)
}
