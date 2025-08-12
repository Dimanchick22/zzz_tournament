// internal/handlers/heroes.go - исправленная версия
package handlers

import (
	"database/sql"
	"strconv"

	"zzz-tournament/internal/models"
	"zzz-tournament/internal/websocket"
	"zzz-tournament/pkg/utils"
	"zzz-tournament/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// HeroHandlers обработчики героев
type HeroHandlers struct {
	BaseHandlers
}

// NewHeroHandlers создает новый экземпляр HeroHandlers
func NewHeroHandlers(db *sqlx.DB, hub *websocket.Hub) *HeroHandlers {
	return &HeroHandlers{
		BaseHandlers: newBaseHandlers(db, hub),
	}
}

// CreateHeroRequest структура запроса создания героя
type CreateHeroRequest struct {
	Name        string `json:"name" binding:"required"`
	Element     string `json:"element" binding:"required"`
	Rarity      string `json:"rarity" binding:"required"`
	Role        string `json:"role" binding:"required"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
}

// UpdateHeroRequest структура запроса обновления героя
type UpdateHeroRequest struct {
	Name        string `json:"name,omitempty"`
	Element     string `json:"element,omitempty"`
	Rarity      string `json:"rarity,omitempty"`
	Role        string `json:"role,omitempty"`
	Description string `json:"description,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
	IsActive    *bool  `json:"is_active,omitempty"`
}

// GetHeroesQuery параметры фильтрации героев
type GetHeroesQuery struct {
	Element  string `form:"element"`
	Rarity   string `form:"rarity"`
	Role     string `form:"role"`
	Active   *bool  `form:"active"`
	Page     int    `form:"page"`
	PerPage  int    `form:"per_page"`
	SortBy   string `form:"sort_by"`
	SortDesc bool   `form:"sort_desc"`
}

// GetHeroes получение списка героев с фильтрацией
func (h *HeroHandlers) GetHeroes(c *gin.Context) {
	var query GetHeroesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Устанавливаем значения по умолчанию
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PerPage <= 0 {
		query.PerPage = 50
	}
	if query.PerPage > 100 {
		query.PerPage = 100
	}
	if query.SortBy == "" {
		query.SortBy = "name"
	}

	// Валидация
	if err := validator.ValidatePage(query.Page); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	if err := validator.ValidatePerPage(query.PerPage); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Строим WHERE условие
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	// Фильтр по активности (по умолчанию только активные)
	if query.Active == nil {
		active := true
		query.Active = &active
	}
	whereConditions = append(whereConditions, "is_active = $"+strconv.Itoa(argIndex))
	args = append(args, *query.Active)
	argIndex++

	// Фильтр по элементу
	if query.Element != "" {
		whereConditions = append(whereConditions, "element = $"+strconv.Itoa(argIndex))
		args = append(args, query.Element)
		argIndex++
	}

	// Фильтр по редкости
	if query.Rarity != "" {
		whereConditions = append(whereConditions, "rarity = $"+strconv.Itoa(argIndex))
		args = append(args, query.Rarity)
		argIndex++
	}

	// Фильтр по роли
	if query.Role != "" {
		whereConditions = append(whereConditions, "role = $"+strconv.Itoa(argIndex))
		args = append(args, query.Role)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + joinStrings(whereConditions, " AND ")
	}

	// Получаем общее количество
	countQuery := "SELECT COUNT(*) FROM heroes " + whereClause
	var total int
	err := h.DB.Get(&total, countQuery, args...)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to count heroes")
		return
	}

	// Строим ORDER BY
	validSortFields := map[string]string{
		"name":    "name",
		"element": "element",
		"rarity":  "rarity",
		"role":    "role",
		"created": "id", // сортировка по времени создания через ID
	}

	sortField, exists := validSortFields[query.SortBy]
	if !exists {
		sortField = "name"
	}

	sortDirection := "ASC"
	if query.SortDesc {
		sortDirection = "DESC"
	}

	// Особая логика для сортировки по редкости
	orderClause := ""
	if query.SortBy == "rarity" {
		// S ранги выше A рангов
		orderClause = "ORDER BY CASE WHEN rarity = 'S' THEN 1 ELSE 2 END " + sortDirection + ", name ASC"
	} else {
		orderClause = "ORDER BY " + sortField + " " + sortDirection
	}

	// Добавляем LIMIT и OFFSET
	offset := (query.Page - 1) * query.PerPage
	args = append(args, query.PerPage, offset)
	limitClause := "LIMIT $" + strconv.Itoa(argIndex) + " OFFSET $" + strconv.Itoa(argIndex+1)

	// Основной запрос
	mainQuery := "SELECT id, name, element, rarity, role, description, image_url, is_active FROM heroes " +
		whereClause + " " + orderClause + " " + limitClause

	var heroes []models.Hero
	err = h.DB.Select(&heroes, mainQuery, args...)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to fetch heroes")
		return
	}

	// Добавляем статистику использования героев (закомментировано для будущей реализации)
	/*
		for heroIndex := range heroes {
			var usageCount int
			err = h.DB.Get(&usageCount, `
				SELECT COUNT(*) FROM matches
				WHERE (player1_id IN (SELECT id FROM users) OR player2_id IN (SELECT id FROM users))
				-- TODO: Добавить связь с героями когда будет таблица выбранных героев в матчах
			`)
			if err == nil {
				// heroes[heroIndex].UsageCount = usageCount
				_ = usageCount // Используем переменную чтобы избежать ошибки компиляции
			}
		}
	*/

	pagination := utils.NewPaginationMeta(query.Page, query.PerPage, total)
	utils.PaginatedSuccessResponse(c, heroes, pagination, "Heroes fetched successfully")
}

// GetHero получение героя по ID
func (h *HeroHandlers) GetHero(c *gin.Context) {
	heroID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid hero ID")
		return
	}

	var hero models.Hero
	err = h.DB.Get(&hero, `
		SELECT id, name, element, rarity, role, description, image_url, is_active
		FROM heroes WHERE id = $1
	`, heroID)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "Hero not found")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	utils.SuccessResponse(c, hero)
}

// CreateHero создание нового героя (только для администраторов)
func (h *HeroHandlers) CreateHero(c *gin.Context) {
	var req CreateHeroRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Валидация
	errors := validator.ValidateHeroCreation(req.Name, req.Element, req.Rarity, req.Role)
	if errors.HasErrors() {
		utils.BadRequestResponse(c, errors.Error())
		return
	}

	// Валидация URL изображения
	if req.ImageURL != "" {
		if err := validator.ValidateURL(req.ImageURL); err != nil {
			utils.BadRequestResponse(c, err.Error())
			return
		}
	}

	// Проверяем уникальность имени героя
	var exists bool
	err := h.DB.Get(&exists, `
		SELECT EXISTS(SELECT 1 FROM heroes WHERE name = $1)
	`, req.Name)

	if err != nil {
		utils.InternalErrorResponse(c, "Database error")
		return
	}

	if exists {
		utils.ConflictResponse(c, "Hero with this name already exists")
		return
	}

	// Создаем героя
	var heroID int
	err = h.DB.QueryRow(`
		INSERT INTO heroes (name, element, rarity, role, description, image_url)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, req.Name, req.Element, req.Rarity, req.Role, req.Description, req.ImageURL).Scan(&heroID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to create hero")
		return
	}

	// Получаем созданного героя
	var hero models.Hero
	err = h.DB.Get(&hero, `
		SELECT id, name, element, rarity, role, description, image_url, is_active
		FROM heroes WHERE id = $1
	`, heroID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to fetch created hero")
		return
	}

	utils.CreatedResponse(c, hero, "Hero created successfully")
}

// UpdateHero обновление героя (только для администраторов)
func (h *HeroHandlers) UpdateHero(c *gin.Context) {
	heroID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid hero ID")
		return
	}

	var req UpdateHeroRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Проверяем существование героя
	var exists bool
	err = h.DB.Get(&exists, `SELECT EXISTS(SELECT 1 FROM heroes WHERE id = $1)`, heroID)
	if err != nil {
		utils.InternalErrorResponse(c, "Database error")
		return
	}
	if !exists {
		utils.NotFoundResponse(c, "Hero not found")
		return
	}

	// Валидация обновляемых полей
	var errors validator.ValidationErrors

	if req.Name != "" {
		if err := validator.ValidateHeroName(req.Name); err != nil {
			errors = append(errors, *err)
		}

		// Проверяем уникальность имени
		var nameExists bool
		err = h.DB.Get(&nameExists, `
			SELECT EXISTS(SELECT 1 FROM heroes WHERE name = $1 AND id != $2)
		`, req.Name, heroID)
		if err == nil && nameExists {
			errors = append(errors, validator.ValidationError{
				Field:   "name",
				Message: "Hero with this name already exists",
				Code:    "DUPLICATE",
				Value:   req.Name,
			})
		}
	}

	if req.Element != "" {
		if err := validator.ValidateHeroElement(req.Element); err != nil {
			errors = append(errors, *err)
		}
	}

	if req.Rarity != "" {
		if err := validator.ValidateHeroRarity(req.Rarity); err != nil {
			errors = append(errors, *err)
		}
	}

	if req.Role != "" {
		if err := validator.ValidateHeroRole(req.Role); err != nil {
			errors = append(errors, *err)
		}
	}

	if req.ImageURL != "" {
		if err := validator.ValidateURL(req.ImageURL); err != nil {
			errors = append(errors, *err)
		}
	}

	if errors.HasErrors() {
		utils.BadRequestResponse(c, errors.Error())
		return
	}

	// Строим UPDATE запрос
	updateFields := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Name != "" {
		updateFields = append(updateFields, "name = $"+strconv.Itoa(argIndex))
		args = append(args, req.Name)
		argIndex++
	}

	if req.Element != "" {
		updateFields = append(updateFields, "element = $"+strconv.Itoa(argIndex))
		args = append(args, req.Element)
		argIndex++
	}

	if req.Rarity != "" {
		updateFields = append(updateFields, "rarity = $"+strconv.Itoa(argIndex))
		args = append(args, req.Rarity)
		argIndex++
	}

	if req.Role != "" {
		updateFields = append(updateFields, "role = $"+strconv.Itoa(argIndex))
		args = append(args, req.Role)
		argIndex++
	}

	if req.Description != "" {
		updateFields = append(updateFields, "description = $"+strconv.Itoa(argIndex))
		args = append(args, req.Description)
		argIndex++
	}

	if req.ImageURL != "" {
		updateFields = append(updateFields, "image_url = $"+strconv.Itoa(argIndex))
		args = append(args, req.ImageURL)
		argIndex++
	}

	if req.IsActive != nil {
		updateFields = append(updateFields, "is_active = $"+strconv.Itoa(argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}

	if len(updateFields) == 0 {
		utils.BadRequestResponse(c, "No fields to update")
		return
	}

	args = append(args, heroID)
	query := "UPDATE heroes SET " + joinStrings(updateFields, ", ") + " WHERE id = $" + strconv.Itoa(argIndex)

	_, err = h.DB.Exec(query, args...)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to update hero")
		return
	}

	// Получаем обновленного героя
	var hero models.Hero
	err = h.DB.Get(&hero, `
		SELECT id, name, element, rarity, role, description, image_url, is_active
		FROM heroes WHERE id = $1
	`, heroID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to fetch updated hero")
		return
	}

	utils.SuccessResponse(c, hero, "Hero updated successfully")
}

// DeleteHero мягкое удаление героя (только для администраторов)
func (h *HeroHandlers) DeleteHero(c *gin.Context) {
	heroID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid hero ID")
		return
	}

	// Проверяем существование героя
	var exists bool
	err = h.DB.Get(&exists, `SELECT EXISTS(SELECT 1 FROM heroes WHERE id = $1)`, heroID)
	if err != nil {
		utils.InternalErrorResponse(c, "Database error")
		return
	}
	if !exists {
		utils.NotFoundResponse(c, "Hero not found")
		return
	}

	// Мягкое удаление (деактивация)
	_, err = h.DB.Exec(`UPDATE heroes SET is_active = false WHERE id = $1`, heroID)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to delete hero")
		return
	}

	utils.NoContentResponse(c, "Hero deleted successfully")
}

// RestoreHero восстановление удаленного героя (только для администраторов)
func (h *HeroHandlers) RestoreHero(c *gin.Context) {
	heroID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid hero ID")
		return
	}

	// Проверяем существование героя
	var exists bool
	err = h.DB.Get(&exists, `SELECT EXISTS(SELECT 1 FROM heroes WHERE id = $1)`, heroID)
	if err != nil {
		utils.InternalErrorResponse(c, "Database error")
		return
	}
	if !exists {
		utils.NotFoundResponse(c, "Hero not found")
		return
	}

	// Восстанавливаем героя
	_, err = h.DB.Exec(`UPDATE heroes SET is_active = true WHERE id = $1`, heroID)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to restore hero")
		return
	}

	// Получаем восстановленного героя
	var hero models.Hero
	err = h.DB.Get(&hero, `
		SELECT id, name, element, rarity, role, description, image_url, is_active
		FROM heroes WHERE id = $1
	`, heroID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to fetch restored hero")
		return
	}

	utils.SuccessResponse(c, hero, "Hero restored successfully")
}

// GetHeroStats получение статистики героя
func (h *HeroHandlers) GetHeroStats(c *gin.Context) {
	heroID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid hero ID")
		return
	}

	// Проверяем существование героя
	var hero models.Hero
	err = h.DB.Get(&hero, `
		SELECT id, name, element, rarity, role, description, image_url, is_active
		FROM heroes WHERE id = $1
	`, heroID)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "Hero not found")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	// TODO: Добавить статистику когда будет таблица выбранных героев в матчах
	stats := gin.H{
		"hero":        hero,
		"total_picks": 0,
		"win_rate":    0.0,
		"popularity":  0.0,
		"ban_rate":    0.0,
		"tier":        "Unknown",
	}

	utils.SuccessResponse(c, stats, "Hero statistics fetched successfully")
}

// joinStrings определена в helpers.go
