// internal/middleware/ratelimit.go
package middleware

import (
	"fmt"
	"sync"
	"time"

	"zzz-tournament/pkg/utils"

	"github.com/gin-gonic/gin"
)

// RateLimiter структура для rate limiting
type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.RWMutex
	rate     time.Duration
	burst    int
	cleanup  time.Duration
}

// Visitor представляет посетителя с его лимитами
type Visitor struct {
	limiter   *TokenBucket
	lastSeen  time.Time
	requests  int
	resetTime time.Time
}

// TokenBucket реализация алгоритма token bucket
type TokenBucket struct {
	tokens     int
	capacity   int
	refillRate time.Duration
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter создает новый rate limiter
func NewRateLimiter(rate time.Duration, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     rate,
		burst:    burst,
		cleanup:  5 * time.Minute,
	}

	go rl.cleanupVisitors()
	return rl
}

// NewTokenBucket создает новый token bucket
func NewTokenBucket(capacity int, refillRate time.Duration) *TokenBucket {
	return &TokenBucket{
		tokens:     capacity,
		capacity:   capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow проверяет, можно ли выполнить запрос
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)

	// Добавляем токены основываясь на времени
	tokensToAdd := int(elapsed / tb.refillRate)
	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// GetRemainingTokens возвращает количество оставшихся токенов
func (tb *TokenBucket) GetRemainingTokens() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.tokens
}

// Middleware возвращает gin middleware для rate limiting
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := rl.getClientIP(c)

		rl.mu.RLock()
		visitor, exists := rl.visitors[ip]
		rl.mu.RUnlock()

		if !exists {
			rl.mu.Lock()
			rl.visitors[ip] = &Visitor{
				limiter:   NewTokenBucket(rl.burst, rl.rate),
				lastSeen:  time.Now(),
				requests:  0,
				resetTime: time.Now().Add(time.Hour),
			}
			visitor = rl.visitors[ip]
			rl.mu.Unlock()
		}

		visitor.lastSeen = time.Now()
		visitor.requests++

		// Сбрасываем счетчик каждый час
		if time.Now().After(visitor.resetTime) {
			visitor.requests = 1
			visitor.resetTime = time.Now().Add(time.Hour)
		}

		if !visitor.limiter.Allow() {
			// Устанавливаем заголовки rate limit
			c.Header("X-Rate-Limit-Limit", fmt.Sprintf("%d", rl.burst))
			c.Header("X-Rate-Limit-Remaining", "0")
			c.Header("X-Rate-Limit-Reset", fmt.Sprintf("%d", visitor.resetTime.Unix()))
			c.Header("Retry-After", fmt.Sprintf("%.0f", rl.rate.Seconds()))

			utils.TooManyRequestsResponse(c, "Rate limit exceeded. Please try again later.")
			c.Abort()
			return
		}

		// Устанавливаем заголовки rate limit
		remaining := visitor.limiter.GetRemainingTokens()
		c.Header("X-Rate-Limit-Limit", fmt.Sprintf("%d", rl.burst))
		c.Header("X-Rate-Limit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-Rate-Limit-Reset", fmt.Sprintf("%d", visitor.resetTime.Unix()))

		c.Next()
	}
}

// cleanupVisitors удаляет старых посетителей
func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(rl.cleanup)

		rl.mu.Lock()
		for ip, visitor := range rl.visitors {
			if time.Since(visitor.lastSeen) > rl.cleanup {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// getClientIP получает IP адрес клиента
func (rl *RateLimiter) getClientIP(c *gin.Context) string {
	// Проверяем заголовки прокси
	ip := c.GetHeader("X-Forwarded-For")
	if ip != "" {
		return ip
	}

	ip = c.GetHeader("X-Real-IP")
	if ip != "" {
		return ip
	}

	return c.ClientIP()
}

// Предустановленные rate limiters

// GlobalRateLimiter общий лимит для всех endpoints
func GlobalRateLimiter() gin.HandlerFunc {
	limiter := NewRateLimiter(100*time.Millisecond, 100) // 100 requests per 10 seconds
	return limiter.Middleware()
}

// AuthRateLimiter лимит для аутентификации
func AuthRateLimiter() gin.HandlerFunc {
	limiter := NewRateLimiter(time.Minute, 5) // 5 requests per minute
	return limiter.Middleware()
}

// APIRateLimiter лимит для API endpoints
func APIRateLimiter() gin.HandlerFunc {
	limiter := NewRateLimiter(time.Second, 60) // 60 requests per minute
	return limiter.Middleware()
}

// WebSocketRateLimiter лимит для WebSocket соединений
func WebSocketRateLimiter() gin.HandlerFunc {
	limiter := NewRateLimiter(time.Second, 10) // 10 connections per second
	return limiter.Middleware()
}

// UploadRateLimiter лимит для загрузки файлов
func UploadRateLimiter() gin.HandlerFunc {
	limiter := NewRateLimiter(time.Minute, 3) // 3 uploads per minute
	return limiter.Middleware()
}

// UserBasedRateLimiter лимит по пользователям
type UserBasedRateLimiter struct {
	userLimiters map[int]*RateLimiter
	mu           sync.RWMutex
	rate         time.Duration
	burst        int
}

// NewUserBasedRateLimiter создает rate limiter по пользователям
func NewUserBasedRateLimiter(rate time.Duration, burst int) *UserBasedRateLimiter {
	return &UserBasedRateLimiter{
		userLimiters: make(map[int]*RateLimiter),
		rate:         rate,
		burst:        burst,
	}
}

// Middleware возвращает middleware для rate limiting по пользователям
func (url *UserBasedRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			// Если пользователь не аутентифицирован, используем IP
			GlobalRateLimiter()(c)
			return
		}

		id := userID.(int)

		url.mu.RLock()
		limiter, exists := url.userLimiters[id]
		url.mu.RUnlock()

		if !exists {
			url.mu.Lock()
			url.userLimiters[id] = NewRateLimiter(url.rate, url.burst)
			limiter = url.userLimiters[id]
			url.mu.Unlock()
		}

		limiter.Middleware()(c)
	}
}

// BurstProtectionMiddleware защита от burst атак
func BurstProtectionMiddleware() gin.HandlerFunc {
	type BurstTracker struct {
		requests []time.Time
		mu       sync.Mutex
	}

	trackers := make(map[string]*BurstTracker)
	mu := sync.RWMutex{}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mu.RLock()
		tracker, exists := trackers[ip]
		mu.RUnlock()

		if !exists {
			mu.Lock()
			trackers[ip] = &BurstTracker{
				requests: []time.Time{now},
			}
			mu.Unlock()
			c.Next()
			return
		}

		tracker.mu.Lock()
		// Удаляем старые запросы (старше 1 минуты)
		cutoff := now.Add(-time.Minute)
		validRequests := make([]time.Time, 0)
		for _, reqTime := range tracker.requests {
			if reqTime.After(cutoff) {
				validRequests = append(validRequests, reqTime)
			}
		}
		validRequests = append(validRequests, now)
		tracker.requests = validRequests

		// Проверяем на burst (более 20 запросов в минуту)
		if len(tracker.requests) > 20 {
			tracker.mu.Unlock()
			utils.TooManyRequestsResponse(c, "Burst protection triggered")
			c.Abort()
			return
		}

		tracker.mu.Unlock()
		c.Next()
	}
}
