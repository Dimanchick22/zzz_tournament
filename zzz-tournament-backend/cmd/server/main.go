// cmd/server/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"zzz-tournament/internal/config"
	"zzz-tournament/internal/db"
	"zzz-tournament/internal/handlers"
	"zzz-tournament/internal/middleware"
	"zzz-tournament/internal/websocket"
	"zzz-tournament/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные окружения
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()

	// Устанавливаем JWT секрет
	auth.SetSecret(cfg.JWTSecret)

	// Подключение к БД
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Миграции
	if err := db.Migrate(database); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	// Настройка Gin в зависимости от окружения
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Gin router
	r := gin.New()

	// === GLOBAL MIDDLEWARE ===

	// Recovery middleware (должен быть первым)
	r.Use(middleware.RecoveryMiddleware())

	// Logging middleware
	if gin.Mode() == gin.ReleaseMode {
		r.Use(middleware.StructuredLoggingMiddleware())
	} else {
		r.Use(middleware.ColoredLoggingMiddleware())
	}

	// Security logging
	r.Use(middleware.SecurityLoggingMiddleware())

	// Performance monitoring (логируем запросы дольше 2 секунд)
	r.Use(middleware.PerformanceLoggingMiddleware(2 * time.Second))

	// CORS
	if gin.Mode() == gin.ReleaseMode {
		// Строгий CORS для продакшена
		allowedOrigins := []string{
			"https://zzz-tournament.example.com", // Замените на ваш домен
			"https://www.zzz-tournament.example.com",
		}
		r.Use(middleware.StrictCORSMiddleware(allowedOrigins))
		r.Use(middleware.SecureHeadersMiddleware())
	} else {
		// Разрешительный CORS для разработки
		r.Use(middleware.DevCORSMiddleware())
		r.Use(middleware.NoSniffMiddleware())
	}

	// Global rate limiting (100 requests per 10 seconds)
	r.Use(middleware.GlobalRateLimiter())

	// === HANDLERS ===
	h := handlers.New(database, hub)

	// === HEALTH CHECK ===
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		})
	})

	// === API ROUTES ===
	api := r.Group("/api/v1")

	// === AUTH ROUTES ===
	authGroup := api.Group("/auth")
	{
		// Строгий rate limiting для аутентификации
		authGroup.Use(middleware.AuthRateLimiter())

		// Аудит логирование для критичных действий
		authGroup.Use(middleware.AuditLoggingMiddleware())

		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)
		authGroup.POST("/refresh", h.RefreshToken)
	}

	// === PROTECTED ROUTES ===
	protected := api.Group("/")

	// Middleware для защищенных роутов
	protected.Use(middleware.AuthMiddleware())                  // JWT аутентификация
	protected.Use(middleware.RefreshTokenMiddleware())          // Автообновление токенов
	protected.Use(createUserBasedRateLimiter(time.Second, 120)) // Лимиты по пользователям

	{
		// === USER ROUTES ===
		users := protected.Group("/users")
		{
			users.GET("/profile", h.GetProfile)
			users.PUT("/profile", h.UpdateProfile)
			users.GET("/leaderboard", h.GetLeaderboard)
		}

		// === HERO ROUTES ===
		heroes := protected.Group("/heroes")
		{
			heroes.GET("", h.GetHeroes)

			// Только администраторы могут управлять героями
			admin := heroes.Group("")
			admin.Use(middleware.AdminOnlyMiddleware())
			{
				admin.POST("", h.CreateHero)
				admin.PUT("/:id", h.UpdateHero)
				admin.DELETE("/:id", h.DeleteHero)
			}
		}

		// === ROOM ROUTES ===
		rooms := protected.Group("/rooms")
		{
			rooms.GET("", h.GetRooms)
			rooms.POST("", h.CreateRoom)
			rooms.GET("/:id", h.GetRoom)
			rooms.PUT("/:id", h.UpdateRoom)
			rooms.DELETE("/:id", h.DeleteRoom)

			// Действия для участников
			rooms.POST("/:id/join", h.JoinRoom)
			rooms.POST("/:id/leave", h.LeaveRoom)
		}

		// === TOURNAMENT ROUTES ===
		tournaments := protected.Group("/tournaments")
		{
			// Запуск турнира
			protected.POST("/rooms/:id/tournament/start", h.StartTournament)

			tournaments.GET("/:id", h.GetTournament)
			tournaments.POST("/:id/matches/:match_id/result", h.SubmitMatchResult)
		}

		// === CHAT ROUTES ===
		chat := protected.Group("/chat")
		{
			chat.GET("/rooms/:id/messages", h.GetRoomMessages)
		}
	}

	// === WEBSOCKET ENDPOINT ===
	ws := r.Group("/ws")
	{
		// WebSocket specific middleware
		ws.Use(middleware.WebSocketCORSMiddleware())
		ws.Use(middleware.WebSocketRateLimiter())

		ws.GET("", func(c *gin.Context) {
			websocket.HandleWebSocket(hub, c.Writer, c.Request)
		})
	}

	// === API DOCUMENTATION (в режиме разработки) ===
	if gin.Mode() != gin.ReleaseMode {
		r.GET("/docs", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "API Documentation",
				"version": "1.0.0",
				"endpoints": map[string]interface{}{
					"auth":        "/api/v1/auth/*",
					"users":       "/api/v1/users/*",
					"heroes":      "/api/v1/heroes",
					"rooms":       "/api/v1/rooms",
					"tournaments": "/api/v1/tournaments",
					"websocket":   "/ws",
				},
			})
		})
	}

	// === НАСТРОЙКА СЕРВЕРА ===
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// === GRACEFUL SHUTDOWN ===
	go func() {
		log.Printf("🚀 Server starting on port %s", cfg.Port)
		log.Printf("🌍 Environment: %s", cfg.Environment)
		log.Printf("📊 Database: Connected")

		if gin.Mode() != gin.ReleaseMode {
			log.Printf("📚 API Documentation: http://localhost:%s/docs", cfg.Port)
			log.Printf("💊 Health Check: http://localhost:%s/health", cfg.Port)
		}

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Ждем сигнал для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("✅ Server exited")
}

// === HELPER FUNCTIONS ===

// createUserBasedRateLimiter создает rate limiter для пользователей
func createUserBasedRateLimiter(rate time.Duration, burst int) gin.HandlerFunc {
	return middleware.NewUserBasedRateLimiter(rate, burst).Middleware()
}
