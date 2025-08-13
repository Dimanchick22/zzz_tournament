// internal/handlers/users.go - исправленная версия
package handlers

import (
	"database/sql"
	"log/slog"
	"strconv"

	"zzz-tournament/internal/models"
	"zzz-tournament/internal/websocket"
	"zzz-tournament/pkg/utils"
	"zzz-tournament/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// UserHandlers обработчики пользователей
type UserHandlers struct {
	BaseHandlers
}

// NewUserHandlers создает новый экземпляр UserHandlers
func NewUserHandlers(db *sqlx.DB, hub *websocket.Hub, logger *slog.Logger) *UserHandlers {
	return &UserHandlers{
		BaseHandlers: newBaseHandlers(db, hub, logger),
	}
}

// UpdateProfileRequest структура запроса обновления профиля
type UpdateProfileRequest struct {
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
}

// GetProfile получение профиля текущего пользователя
func (h *UserHandlers) GetProfile(c *gin.Context) {
	userID := c.GetInt("user_id")

	var user models.User
	err := h.DB.Get(&user, `
		SELECT id, username, email, rating, wins, losses, created_at, updated_at
		FROM users WHERE id = $1
	`, userID)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "User not found")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	utils.SuccessResponse(c, user)
}

// UpdateProfile обновление профиля пользователя
func (h *UserHandlers) UpdateProfile(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Валидация
	var errors validator.ValidationErrors

	if req.Username != "" {
		if err := validator.ValidateUsername(req.Username); err != nil {
			errors = append(errors, *err)
		}
	}

	if req.Email != "" {
		if err := validator.ValidateEmail(req.Email); err != nil {
			errors = append(errors, *err)
		}
	}

	if errors.HasErrors() {
		utils.BadRequestResponse(c, errors.Error())
		return
	}

	// Начинаем транзакцию
	tx, err := h.DB.Beginx()
	if err != nil {
		utils.InternalErrorResponse(c, "Database transaction error")
		return
	}
	defer tx.Rollback()

	// Проверяем уникальность username и email
	if req.Username != "" {
		var exists bool
		err = tx.Get(&exists, `
			SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND id != $2)
		`, req.Username, userID)

		if err != nil {
			utils.InternalErrorResponse(c, "Database error")
			return
		}

		if exists {
			utils.ConflictResponse(c, "Username already taken")
			return
		}
	}

	if req.Email != "" {
		var exists bool
		err = tx.Get(&exists, `
			SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND id != $2)
		`, req.Email, userID)

		if err != nil {
			utils.InternalErrorResponse(c, "Database error")
			return
		}

		if exists {
			utils.ConflictResponse(c, "Email already taken")
			return
		}
	}

	// Обновляем профиль
	updateFields := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Username != "" {
		updateFields = append(updateFields, "username = $"+strconv.Itoa(argIndex))
		args = append(args, req.Username)
		argIndex++
	}

	if req.Email != "" {
		updateFields = append(updateFields, "email = $"+strconv.Itoa(argIndex))
		args = append(args, req.Email)
		argIndex++
	}

	if len(updateFields) == 0 {
		utils.BadRequestResponse(c, "No fields to update")
		return
	}

	updateFields = append(updateFields, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, userID)

	query := "UPDATE users SET " + joinStrings(updateFields, ", ") + " WHERE id = $" + strconv.Itoa(argIndex)

	_, err = tx.Exec(query, args...)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to update profile")
		return
	}

	// Получаем обновленные данные
	var user models.User
	err = tx.Get(&user, `
		SELECT id, username, email, rating, wins, losses, created_at, updated_at
		FROM users WHERE id = $1
	`, userID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to fetch updated profile")
		return
	}

	// Коммитим транзакцию
	if err = tx.Commit(); err != nil {
		utils.InternalErrorResponse(c, "Failed to commit transaction")
		return
	}

	utils.SuccessResponse(c, user, "Profile updated successfully")
}

// GetLeaderboard получение рейтинговой таблицы
func (h *UserHandlers) GetLeaderboard(c *gin.Context) {
	// Параметры пагинации
	page := getPageFromQuery(c, 1)
	perPage := getPerPageFromQuery(c, 50)

	// Валидация пагинации
	if err := validator.ValidatePage(page); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	if err := validator.ValidatePerPage(perPage); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	offset := (page - 1) * perPage

	// Получаем общее количество пользователей
	var total int
	err := h.DB.Get(&total, `SELECT COUNT(*) FROM users`)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to count users")
		return
	}

	// Получаем пользователей с пагинацией
	var users []models.User
	err = h.DB.Select(&users, `
		SELECT id, username, rating, wins, losses, created_at
		FROM users
		ORDER BY rating DESC, wins DESC, username ASC
		LIMIT $1 OFFSET $2
	`, perPage, offset)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to fetch leaderboard")
		return
	}

	// Создаем мета информацию для пагинации
	pagination := utils.NewPaginationMeta(page, perPage, total)

	utils.PaginatedSuccessResponse(c, users, pagination, "Leaderboard fetched successfully")
}

// GetUserStats получение статистики пользователя
func (h *UserHandlers) GetUserStats(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID")
		return
	}

	// Получаем основную информацию о пользователе
	var user models.User
	err = h.DB.Get(&user, `
		SELECT id, username, rating, wins, losses, created_at
		FROM users WHERE id = $1
	`, userID)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "User not found")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	// Получаем дополнительную статистику
	type UserStats struct {
		models.User
		TotalGames       int     `json:"total_games"`
		WinRate          float64 `json:"win_rate"`
		TournamentsWon   int     `json:"tournaments_won"`
		TournamentsTotal int     `json:"tournaments_total"`
		CurrentStreak    int     `json:"current_streak"`
		BestStreak       int     `json:"best_streak"`
		RatingTier       string  `json:"rating_tier"`
		RatingColor      string  `json:"rating_color"`
		Rank             int     `json:"rank"`
	}

	stats := UserStats{
		User:       user,
		TotalGames: user.Wins + user.Losses,
	}

	// Рассчитываем процент побед
	if stats.TotalGames > 0 {
		stats.WinRate = float64(user.Wins) / float64(stats.TotalGames) * 100
	}

	// Получаем количество выигранных турниров
	err = h.DB.Get(&stats.TournamentsWon, `
		SELECT COUNT(*) FROM tournaments WHERE winner_id = $1 AND status = 'finished'
	`, userID)
	if err != nil {
		stats.TournamentsWon = 0
	}

	// Получаем общее количество турниров
	err = h.DB.Get(&stats.TournamentsTotal, `
		SELECT COUNT(DISTINCT t.id) 
		FROM tournaments t
		JOIN room_participants rp ON t.room_id = rp.room_id
		WHERE rp.user_id = $1 AND t.status = 'finished'
	`, userID)
	if err != nil {
		stats.TournamentsTotal = 0
	}

	// TODO: Рассчитать текущую и лучшую серии побед
	// Это требует более сложных запросов к истории матчей

	// Получаем ранг в рейтинге
	err = h.DB.Get(&stats.Rank, `
		SELECT COUNT(*) + 1 FROM users WHERE rating > $1
	`, user.Rating)
	if err != nil {
		stats.Rank = 0
	}

	// Устанавливаем тир и цвет рейтинга
	stats.RatingTier = getRatingTier(user.Rating)
	stats.RatingColor = getRatingColor(user.Rating)

	utils.SuccessResponse(c, stats, "User statistics fetched successfully")
}

// SearchUsers поиск пользователей
func (h *UserHandlers) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequestResponse(c, "Search query is required")
		return
	}

	if len(query) < 2 {
		utils.BadRequestResponse(c, "Search query must be at least 2 characters")
		return
	}

	// Параметры пагинации
	page := getPageFromQuery(c, 1)
	perPage := getPerPageFromQuery(c, 20)

	if perPage > 50 {
		perPage = 50 // Ограничиваем для поиска
	}

	offset := (page - 1) * perPage

	// Поиск пользователей
	searchPattern := "%" + query + "%"
	var users []models.User
	err := h.DB.Select(&users, `
		SELECT id, username, rating, wins, losses, created_at
		FROM users
		WHERE username ILIKE $1
		ORDER BY 
			CASE WHEN username ILIKE $2 THEN 1 ELSE 2 END,
			rating DESC,
			username ASC
		LIMIT $3 OFFSET $4
	`, searchPattern, query+"%", perPage, offset)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to search users")
		return
	}

	// Получаем общее количество результатов
	var total int
	err = h.DB.Get(&total, `
		SELECT COUNT(*) FROM users WHERE username ILIKE $1
	`, searchPattern)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to count search results")
		return
	}

	pagination := utils.NewPaginationMeta(page, perPage, total)

	utils.PaginatedSuccessResponse(c, users, pagination, "Users found")
}

// GetUserByID получение пользователя по ID (публичная информация)
func (h *UserHandlers) GetUserByID(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID")
		return
	}

	var user models.User
	err = h.DB.Get(&user, `
		SELECT id, username, rating, wins, losses, created_at
		FROM users WHERE id = $1
	`, userID)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "User not found")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	utils.SuccessResponse(c, user)
}

// Вспомогательные функции

func getPageFromQuery(c *gin.Context, defaultPage int) int {
	pageStr := c.DefaultQuery("page", strconv.Itoa(defaultPage))
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		return defaultPage
	}
	return page
}

func getPerPageFromQuery(c *gin.Context, defaultPerPage int) int {
	perPageStr := c.DefaultQuery("per_page", strconv.Itoa(defaultPerPage))
	perPage, err := strconv.Atoi(perPageStr)
	if err != nil || perPage < 1 {
		return defaultPerPage
	}
	return perPage
}

func getRatingTier(rating int) string {
	switch {
	case rating < 800:
		return "Bronze"
	case rating < 1200:
		return "Silver"
	case rating < 1600:
		return "Gold"
	case rating < 2000:
		return "Platinum"
	case rating < 2400:
		return "Diamond"
	case rating < 2800:
		return "Master"
	default:
		return "Grandmaster"
	}
}

func getRatingColor(rating int) string {
	switch {
	case rating < 800:
		return "#CD7F32" // Bronze
	case rating < 1200:
		return "#C0C0C0" // Silver
	case rating < 1600:
		return "#FFD700" // Gold
	case rating < 2000:
		return "#E5E4E2" // Platinum
	case rating < 2400:
		return "#B9F2FF" // Diamond
	case rating < 2800:
		return "#FF6347" // Master
	default:
		return "#9400D3" // Grandmaster
	}
}
