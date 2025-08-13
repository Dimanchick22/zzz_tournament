// cmd/server/main.go
package main

import (
	"context"
	"log"
	"log/slog"
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
	authConfig "zzz-tournament/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные окружения
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()

	// Инициализация структурированного логгера
	var logger *slog.Logger
	if cfg.Environment == "production" {
		// JSON логирование для продакшена
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	} else {
		// Текстовое логирование для разработки
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}

	// Загружаем конфигурацию аутентификации
	authCfg, err := authConfig.LoadAuthConfig()
	if err != nil {
		logger.Error("Failed to load auth config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Валидируем конфигурацию аутентификации
	if err := authCfg.Validate(); err != nil {
		logger.Error("Invalid auth config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Устанавливаем JWT секрет из конфигурации
	if authCfg.JWTSecret == "" {
		authCfg.JWTSecret = cfg.JWTSecret // Fallback на старую конфигурацию
	}
	auth.SetSecret(authCfg.JWTSecret)

	logger.Info("Application starting",
		slog.String("environment", cfg.Environment),
		slog.String("port", cfg.Port),
		slog.Bool("rate_limiting", authCfg.RateLimitEnabled),
	)

	// Подключение к БД
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Error("Failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer database.Close()

	logger.Info("Database connected successfully")

	// Миграции
	if err := db.Migrate(database); err != nil {
		logger.Error("Failed to run migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("Database migrations completed")

	// WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	logger.Info("WebSocket hub started")

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

	// Global rate limiting (только если включено в конфиге)
	if authCfg.RateLimitEnabled {
		r.Use(middleware.GlobalRateLimiter())
		logger.Info("Global rate limiting enabled")
	}

	// === HANDLERS ===
	h := handlers.New(database, hub, logger, authCfg)

	logger.Info("Handlers initialized successfully")

	// === HEALTH CHECK ===
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
			"features": gin.H{
				"database":      "connected",
				"websocket":     "active",
				"rate_limiting": authCfg.RateLimitEnabled,
				"environment":   cfg.Environment,
			},
		})
	})

	// === API ROUTES ===
	api := r.Group("/api/v1")

	// === AUTH ROUTES ===
	authGroup := api.Group("/auth")
	{
		// Rate limiting для аутентификации (если включено)
		if authCfg.RateLimitEnabled {
			authGroup.Use(middleware.AuthRateLimiter())
		}

		// Аудит логирование для критичных действий
		authGroup.Use(middleware.AuditLoggingMiddleware())

		authGroup.POST("/register", h.Auth.Register)
		authGroup.POST("/login", h.Auth.Login)
		authGroup.POST("/refresh", h.Auth.RefreshToken)
		authGroup.POST("/logout", middleware.AuthMiddleware(), h.Auth.Logout)
		authGroup.POST("/change-password", middleware.AuthMiddleware(), h.Auth.ChangePassword)
		authGroup.POST("/forgot-password", h.Auth.ForgotPassword)
		authGroup.POST("/reset-password", h.Auth.ResetPassword)
	}

	// === PROTECTED ROUTES ===
	protected := api.Group("/")

	// Middleware для защищенных роутов
	protected.Use(middleware.AuthMiddleware()) // JWT аутентификация

	// Автообновление токенов и пользовательские лимиты (если включено)
	if authCfg.RateLimitEnabled {
		protected.Use(middleware.RefreshTokenMiddleware())          // Автообновление токенов
		protected.Use(createUserBasedRateLimiter(time.Second, 120)) // Лимиты по пользователям
	}

	{
		// === USER ROUTES ===
		users := protected.Group("/users")
		{
			users.GET("/profile", h.Users.GetProfile)
			users.PUT("/profile", h.Users.UpdateProfile)
			users.GET("/leaderboard", h.Users.GetLeaderboard)
			users.GET("/search", h.Users.SearchUsers)
			users.GET("/:id", h.Users.GetUserByID)
			users.GET("/:id/stats", h.Users.GetUserStats)
		}

		// === HERO ROUTES ===
		heroes := protected.Group("/heroes")
		{
			heroes.GET("", h.Heroes.GetHeroes)
			heroes.GET("/:id", h.Heroes.GetHero)
			heroes.GET("/:id/stats", h.Heroes.GetHeroStats)

			// Только администраторы могут управлять героями
			admin := heroes.Group("")
			admin.Use(middleware.AdminOnlyMiddleware())
			{
				admin.POST("", h.Heroes.CreateHero)
				admin.PUT("/:id", h.Heroes.UpdateHero)
				admin.DELETE("/:id", h.Heroes.DeleteHero)
				admin.POST("/:id/restore", h.Heroes.RestoreHero)
			}
		}

		// === ROOM ROUTES ===
		rooms := protected.Group("/rooms")
		{
			rooms.GET("", h.Rooms.GetRooms)
			rooms.POST("", h.Rooms.CreateRoom)
			rooms.GET("/:id", h.Rooms.GetRoom)
			rooms.PUT("/:id", h.Rooms.UpdateRoom)
			rooms.DELETE("/:id", h.Rooms.DeleteRoom)

			// Действия для участников
			rooms.POST("/:id/join", h.Rooms.JoinRoom)
			rooms.POST("/:id/leave", h.Rooms.LeaveRoom)
			rooms.GET("/:id/participants", h.Rooms.GetRoomParticipants)

			// Действия для хоста
			rooms.POST("/:id/kick", h.Rooms.KickPlayer)
			rooms.PUT("/:id/password", h.Rooms.SetRoomPassword)

			// Чат
			rooms.GET("/:id/messages", h.Chat.GetRoomMessages)
			rooms.POST("/:id/messages", h.Chat.SendMessage)
			rooms.PUT("/:id/messages/:message_id", h.Chat.EditMessage)
			rooms.DELETE("/:id/messages/:message_id", h.Chat.DeleteMessage)
			rooms.GET("/:id/chat/stats", h.Chat.GetChatStats)
			rooms.DELETE("/:id/chat/clear", h.Chat.ClearChatHistory)
			rooms.POST("/:id/chat/mute/:user_id", h.Chat.MuteUser)
			rooms.DELETE("/:id/chat/mute/:user_id", h.Chat.UnmuteUser)
		}

		// === TOURNAMENT ROUTES ===
		tournaments := protected.Group("/tournaments")
		{
			tournaments.GET("", h.Tournaments.GetTournaments)
			tournaments.GET("/:id", h.Tournaments.GetTournament)
			tournaments.GET("/:id/stats", h.Tournaments.GetTournamentStats)
			tournaments.POST("/:id/cancel", h.Tournaments.CancelTournament)
			tournaments.GET("/:id/matches/:match_id", h.Tournaments.GetMatch)
			tournaments.POST("/:id/matches/:match_id/result", h.Tournaments.SubmitMatchResult)

			// Запуск турнира
			protected.POST("/rooms/:id/tournament/start", h.Tournaments.StartTournament)
		}
	}

	// === WEBSOCKET ENDPOINT ===
	ws := r.Group("/ws")
	{
		// WebSocket specific middleware
		ws.Use(middleware.WebSocketCORSMiddleware())

		// WebSocket rate limiting (если включено)
		if authCfg.RateLimitEnabled {
			ws.Use(middleware.WebSocketRateLimiter())
		}

		ws.GET("", func(c *gin.Context) {
			websocket.HandleWebSocket(hub, c.Writer, c.Request)
		})
	}

	// === ADMIN ROUTES (только для разработки и специальных эндпоинтов) ===
	if gin.Mode() != gin.ReleaseMode {
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware())
		admin.Use(middleware.AdminOnlyMiddleware())
		{
			// Эндпоинт для очистки просроченных токенов
			admin.POST("/cleanup-tokens", func(c *gin.Context) {
				if err := h.Auth.CleanupExpiredTokens(); err != nil {
					logger.Error("Failed to cleanup tokens", slog.String("error", err.Error()))
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cleanup tokens"})
					return
				}
				c.JSON(http.StatusOK, gin.H{"message": "Tokens cleaned up successfully"})
			})

			// Статистика системы
			admin.GET("/stats", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"auth_config": gin.H{
						"access_token_ttl":   authCfg.AccessTokenTTL.String(),
						"refresh_token_ttl":  authCfg.RefreshTokenTTL.String(),
						"rate_limiting":      authCfg.RateLimitEnabled,
						"max_login_attempts": authCfg.MaxLoginAttempts,
					},
					"server_info": gin.H{
						"environment": cfg.Environment,
						"uptime":      time.Now().UTC(),
					},
				})
			})
		}
	}

	// === API DOCUMENTATION (в режиме разработки) ===
	if gin.Mode() != gin.ReleaseMode {
		r.GET("/docs", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "ZZZ Tournament API Documentation",
				"version": "1.0.0",
				"security": gin.H{
					"rate_limiting":      authCfg.RateLimitEnabled,
					"max_login_attempts": authCfg.MaxLoginAttempts,
					"token_rotation":     authCfg.RefreshTokenRotationThreshold.String(),
				},
				"endpoints": map[string]interface{}{
					"auth": map[string]string{
						"POST /api/v1/auth/register":        "Регистрация пользователя",
						"POST /api/v1/auth/login":           "Авторизация",
						"POST /api/v1/auth/refresh":         "Обновление токена",
						"POST /api/v1/auth/logout":          "Выход из системы",
						"POST /api/v1/auth/change-password": "Смена пароля",
						"POST /api/v1/auth/forgot-password": "Восстановление пароля",
						"POST /api/v1/auth/reset-password":  "Сброс пароля",
					},
					"users": map[string]string{
						"GET /api/v1/users/profile":     "Получить профиль",
						"PUT /api/v1/users/profile":     "Обновить профиль",
						"GET /api/v1/users/leaderboard": "Рейтинговая таблица",
						"GET /api/v1/users/search":      "Поиск пользователей",
						"GET /api/v1/users/:id":         "Информация о пользователе",
						"GET /api/v1/users/:id/stats":   "Статистика пользователя",
					},
					"heroes": map[string]string{
						"GET /api/v1/heroes":           "Список героев",
						"GET /api/v1/heroes/:id":       "Информация о герое",
						"GET /api/v1/heroes/:id/stats": "Статистика героя",
						"POST /api/v1/heroes":          "Создать героя (админ)",
						"PUT /api/v1/heroes/:id":       "Обновить героя (админ)",
						"DELETE /api/v1/heroes/:id":    "Удалить героя (админ)",
					},
					"rooms": map[string]string{
						"GET /api/v1/rooms":               "Список комнат",
						"POST /api/v1/rooms":              "Создать комнату",
						"GET /api/v1/rooms/:id":           "Информация о комнате",
						"PUT /api/v1/rooms/:id":           "Обновить комнату",
						"DELETE /api/v1/rooms/:id":        "Удалить комнату",
						"POST /api/v1/rooms/:id/join":     "Присоединиться к комнате",
						"POST /api/v1/rooms/:id/leave":    "Покинуть комнату",
						"POST /api/v1/rooms/:id/kick":     "Исключить игрока",
						"GET /api/v1/rooms/:id/messages":  "Сообщения чата",
						"POST /api/v1/rooms/:id/messages": "Отправить сообщение",
					},
					"tournaments": map[string]string{
						"GET /api/v1/tournaments":                               "Список турниров",
						"POST /api/v1/rooms/:id/tournament/start":               "Запустить турнир",
						"GET /api/v1/tournaments/:id":                           "Информация о турнире",
						"POST /api/v1/tournaments/:id/matches/:match_id/result": "Результат матча",
						"POST /api/v1/tournaments/:id/cancel":                   "Отменить турнир",
					},
					"admin": map[string]string{
						"POST /api/v1/admin/cleanup-tokens": "Очистка просроченных токенов",
						"GET /api/v1/admin/stats":           "Статистика системы",
					},
					"websocket": "/ws - WebSocket соединение",
				},
				"features": []string{
					"JWT Authentication with enhanced security",
					"Real-time WebSocket chat",
					"Tournament bracket generation",
					"ELO rating system",
					"Room management",
					"Hero database",
					"User statistics",
					"Configurable rate limiting",
					"CORS protection",
					"Structured logging with slog",
					"Account lockout protection",
					"Token rotation",
					"Password reset functionality",
					"Security event logging",
					"Automatic token cleanup",
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

	// Запуск фоновой задачи очистки токенов (каждые 6 часов)
	if authCfg.RateLimitEnabled {
		go func() {
			ticker := time.NewTicker(6 * time.Hour)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					if err := h.Auth.CleanupExpiredTokens(); err != nil {
						logger.Error("Failed to cleanup expired tokens", slog.String("error", err.Error()))
					} else {
						logger.Info("Expired tokens cleaned up successfully")
					}
				}
			}
		}()

		logger.Info("Token cleanup task started (runs every 6 hours)")
	}

	// === GRACEFUL SHUTDOWN ===
	go func() {
		logger.Info("Server starting",
			slog.String("port", cfg.Port),
			slog.String("environment", cfg.Environment),
		)

		if gin.Mode() != gin.ReleaseMode {
			logger.Info("Development endpoints available",
				slog.String("docs", "http://localhost:"+cfg.Port+"/docs"),
				slog.String("health", "http://localhost:"+cfg.Port+"/health"),
				slog.String("websocket", "ws://localhost:"+cfg.Port+"/ws"),
			)
		}

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start server", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Ждем сигнал для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("Server exited successfully")
}

// === HELPER FUNCTIONS ===

// createUserBasedRateLimiter создает rate limiter для пользователей
func createUserBasedRateLimiter(rate time.Duration, burst int) gin.HandlerFunc {
	return middleware.NewUserBasedRateLimiter(rate, burst).Middleware()
}
