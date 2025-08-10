// pkg/utils/response.go
package utils

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Response стандартная структура ответа API
type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// PaginationMeta метаданные для пагинации
type PaginationMeta struct {
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// PaginatedResponse ответ с пагинацией
type PaginatedResponse struct {
	Success    bool           `json:"success"`
	Message    string         `json:"message,omitempty"`
	Data       interface{}    `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
	Timestamp  time.Time      `json:"timestamp"`
	RequestID  string         `json:"request_id,omitempty"`
}

// ErrorDetail детальная информация об ошибке
type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse ответ с детальными ошибками
type ErrorResponse struct {
	Success   bool          `json:"success"`
	Error     string        `json:"error"`
	Details   []ErrorDetail `json:"details,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
	RequestID string        `json:"request_id,omitempty"`
}

// SuccessResponse отправляет успешный ответ
func SuccessResponse(c *gin.Context, data interface{}, message ...string) {
	resp := Response{
		Success:   true,
		Data:      data,
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}

	if len(message) > 0 {
		resp.Message = message[0]
	}

	c.JSON(http.StatusOK, resp)
}

// CreatedResponse отправляет ответ о создании ресурса
func CreatedResponse(c *gin.Context, data interface{}, message ...string) {
	resp := Response{
		Success:   true,
		Data:      data,
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}

	if len(message) > 0 {
		resp.Message = message[0]
	} else {
		resp.Message = "Resource created successfully"
	}

	c.JSON(http.StatusCreated, resp)
}

// NoContentResponse отправляет ответ без контента
func NoContentResponse(c *gin.Context, message ...string) {
	resp := Response{
		Success:   true,
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}

	if len(message) > 0 {
		resp.Message = message[0]
	}

	c.JSON(http.StatusOK, resp)
}

// PaginatedSuccessResponse отправляет успешный ответ с пагинацией
func PaginatedSuccessResponse(c *gin.Context, data interface{}, pagination PaginationMeta, message ...string) {
	resp := PaginatedResponse{
		Success:    true,
		Data:       data,
		Pagination: pagination,
		Timestamp:  time.Now(),
		RequestID:  getRequestID(c),
	}

	if len(message) > 0 {
		resp.Message = message[0]
	}

	c.JSON(http.StatusOK, resp)
}

// ErrorResponse отправляет ответ с ошибкой
func ErrorResponseWithDetails(c *gin.Context, statusCode int, err string, details ...ErrorDetail) {
	resp := ErrorResponse{
		Success:   false,
		Error:     err,
		Details:   details,
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}

	c.JSON(statusCode, resp)
}

// BadRequestResponse отправляет ответ с ошибкой 400
func BadRequestResponse(c *gin.Context, err string, details ...ErrorDetail) {
	ErrorResponseWithDetails(c, http.StatusBadRequest, err, details...)
}

// UnauthorizedResponse отправляет ответ с ошибкой 401
func UnauthorizedResponse(c *gin.Context, err string) {
	ErrorResponseWithDetails(c, http.StatusUnauthorized, err)
}

// ForbiddenResponse отправляет ответ с ошибкой 403
func ForbiddenResponse(c *gin.Context, err string) {
	ErrorResponseWithDetails(c, http.StatusForbidden, err)
}

// NotFoundResponse отправляет ответ с ошибкой 404
func NotFoundResponse(c *gin.Context, err string) {
	ErrorResponseWithDetails(c, http.StatusNotFound, err)
}

// ConflictResponse отправляет ответ с ошибкой 409
func ConflictResponse(c *gin.Context, err string) {
	ErrorResponseWithDetails(c, http.StatusConflict, err)
}

// UnprocessableEntityResponse отправляет ответ с ошибкой 422
func UnprocessableEntityResponse(c *gin.Context, err string, details ...ErrorDetail) {
	ErrorResponseWithDetails(c, http.StatusUnprocessableEntity, err, details...)
}

// TooManyRequestsResponse отправляет ответ с ошибкой 429
func TooManyRequestsResponse(c *gin.Context, err string) {
	ErrorResponseWithDetails(c, http.StatusTooManyRequests, err)
}

// InternalErrorResponse отправляет ответ с ошибкой 500
func InternalErrorResponse(c *gin.Context, err string) {
	ErrorResponseWithDetails(c, http.StatusInternalServerError, err)
}

// ServiceUnavailableResponse отправляет ответ с ошибкой 503
func ServiceUnavailableResponse(c *gin.Context, err string) {
	ErrorResponseWithDetails(c, http.StatusServiceUnavailable, err)
}

// ValidationErrorResponse отправляет ответ с ошибками валидации
func ValidationErrorResponse(c *gin.Context, errors []ErrorDetail) {
	ErrorResponseWithDetails(c, http.StatusUnprocessableEntity, "Validation failed", errors...)
}

// getRequestID извлекает ID запроса из контекста
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// NewPaginationMeta создает новую структуру пагинации
func NewPaginationMeta(page, perPage, total int) PaginationMeta {
	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}

	return PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// JSONResponse отправляет произвольный JSON ответ
func JSONResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

// HealthCheckResponse отправляет ответ для проверки здоровья сервиса
func HealthCheckResponse(c *gin.Context, status string, checks map[string]interface{}) {
	resp := map[string]interface{}{
		"status":    status,
		"timestamp": time.Now(),
		"checks":    checks,
	}

	statusCode := http.StatusOK
	if status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, resp)
}

// RedirectResponse отправляет редирект
func RedirectResponse(c *gin.Context, url string, permanent bool) {
	statusCode := http.StatusFound
	if permanent {
		statusCode = http.StatusMovedPermanently
	}

	c.Redirect(statusCode, url)
}

// FileResponse отправляет файл
func FileResponse(c *gin.Context, filepath string, filename ...string) {
	if len(filename) > 0 {
		c.Header("Content-Disposition", "attachment; filename="+filename[0])
	}
	c.File(filepath)
}

// StreamResponse отправляет потоковый ответ
func StreamResponse(c *gin.Context, contentType string, data []byte) {
	c.Header("Content-Type", contentType)
	c.Data(http.StatusOK, contentType, data)
}

// CacheResponse устанавливает заголовки кэширования
func CacheResponse(c *gin.Context, maxAge int) {
	c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
}

// NoCacheResponse отключает кэширование
func NoCacheResponse(c *gin.Context) {
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
}

// CSVResponse отправляет CSV файл
func CSVResponse(c *gin.Context, data [][]string, filename string) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename="+filename)

	// Простая реализация CSV
	var csvData string
	for _, row := range data {
		for i, cell := range row {
			if i > 0 {
				csvData += ","
			}
			csvData += `"` + strings.ReplaceAll(cell, `"`, `""`) + `"`
		}
		csvData += "\n"
	}

	c.Data(http.StatusOK, "text/csv", []byte(csvData))
}

// WebSocketUpgradeResponse отправляет ответ для обновления до WebSocket
func WebSocketUpgradeResponse(c *gin.Context, err string) {
	c.JSON(http.StatusBadRequest, map[string]interface{}{
		"error":     "WebSocket upgrade failed",
		"details":   err,
		"timestamp": time.Now(),
	})
}

// MaintenanceResponse отправляет ответ о техническом обслуживании
func MaintenanceResponse(c *gin.Context, message string, retryAfter int) {
	c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
	c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
		"error":       "Service temporarily unavailable",
		"message":     message,
		"retry_after": retryAfter,
		"timestamp":   time.Now(),
	})
}

// APIVersionResponse отправляет информацию о версии API
func APIVersionResponse(c *gin.Context, version, buildTime, commit string) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"version":    version,
		"build_time": buildTime,
		"commit":     commit,
		"timestamp":  time.Now(),
	})
}
