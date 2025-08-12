// internal/handlers/rooms.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"zzz-tournament/internal/models"
	"zzz-tournament/internal/websocket"
	"zzz-tournament/pkg/utils"
	"zzz-tournament/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// RoomHandlers обработчики комнат
type RoomHandlers struct {
	BaseHandlers
}

// NewRoomHandlers создает новый экземпляр RoomHandlers
func NewRoomHandlers(db *sqlx.DB, hub *websocket.Hub) *RoomHandlers {
	return &RoomHandlers{
		BaseHandlers: newBaseHandlers(db, hub),
	}
}

// CreateRoomRequest структура запроса создания комнаты
type CreateRoomRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=255"`
	Description string `json:"description"`
	MaxPlayers  int    `json:"max_players" binding:"required,min=2,max=64"`
	IsPrivate   bool   `json:"is_private"`
	Password    string `json:"password"`
}

// UpdateRoomRequest структура запроса обновления комнаты
type UpdateRoomRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	MaxPlayers  int    `json:"max_players,omitempty"`
	IsPrivate   *bool  `json:"is_private,omitempty"`
	Password    string `json:"password,omitempty"`
}

// JoinRoomRequest структура запроса присоединения к комнате
type JoinRoomRequest struct {
	Password string `json:"password"`
}

// KickPlayerRequest структура запроса исключения игрока
type KickPlayerRequest struct {
	UserID int `json:"user_id" binding:"required"`
}

// GetRoomsQuery параметры фильтрации комнат
type GetRoomsQuery struct {
	Status     string `form:"status"`
	IsPrivate  *bool  `form:"is_private"`
	HasSpace   *bool  `form:"has_space"`
	MinPlayers int    `form:"min_players"`
	MaxPlayers int    `form:"max_players"`
	Page       int    `form:"page"`
	PerPage    int    `form:"per_page"`
	SortBy     string `form:"sort_by"`
	SortDesc   bool   `form:"sort_desc"`
}

// GetRooms получение списка комнат с фильтрацией
func (h *RoomHandlers) GetRooms(c *gin.Context) {
	var query GetRoomsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Устанавливаем значения по умолчанию
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PerPage <= 0 {
		query.PerPage = 20
	}
	if query.PerPage > 50 {
		query.PerPage = 50
	}
	if query.SortBy == "" {
		query.SortBy = "created_at"
	}

	// Строим WHERE условие
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	// По умолчанию показываем только комнаты в ожидании
	if query.Status == "" {
		query.Status = "waiting"
	}
	whereConditions = append(whereConditions, "r.status = $"+strconv.Itoa(argIndex))
	args = append(args, query.Status)
	argIndex++

	// Фильтр по приватности
	if query.IsPrivate != nil {
		whereConditions = append(whereConditions, "r.is_private = $"+strconv.Itoa(argIndex))
		args = append(args, *query.IsPrivate)
		argIndex++
	}

	// Фильтр по наличию свободных мест
	if query.HasSpace != nil && *query.HasSpace {
		whereConditions = append(whereConditions, "r.current_count < r.max_players")
	}

	// Фильтр по минимальному количеству игроков
	if query.MinPlayers > 0 {
		whereConditions = append(whereConditions, "r.current_count >= $"+strconv.Itoa(argIndex))
		args = append(args, query.MinPlayers)
		argIndex++
	}

	// Фильтр по максимальному количеству игроков
	if query.MaxPlayers > 0 {
		whereConditions = append(whereConditions, "r.max_players <= $"+strconv.Itoa(argIndex))
		args = append(args, query.MaxPlayers)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + joinStrings(whereConditions, " AND ")
	}

	// Получаем общее количество
	countQuery := `
		SELECT COUNT(*) 
		FROM rooms r 
		JOIN users u ON r.host_id = u.id ` + whereClause
	var total int
	err := h.DB.Get(&total, countQuery, args...)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to count rooms")
		return
	}

	// Строим ORDER BY
	validSortFields := map[string]string{
		"created_at":  "r.created_at",
		"updated_at":  "r.updated_at",
		"name":        "r.name",
		"players":     "r.current_count",
		"max_players": "r.max_players",
		"host":        "u.username",
	}

	sortField, exists := validSortFields[query.SortBy]
	if !exists {
		sortField = "r.created_at"
	}

	sortDirection := "DESC"
	if !query.SortDesc && query.SortBy != "created_at" {
		sortDirection = "ASC"
	}

	orderClause := "ORDER BY " + sortField + " " + sortDirection

	// Добавляем LIMIT и OFFSET
	offset := (query.Page - 1) * query.PerPage
	args = append(args, query.PerPage, offset)
	limitClause := "LIMIT $" + strconv.Itoa(argIndex) + " OFFSET $" + strconv.Itoa(argIndex+1)

	// Основной запрос
	mainQuery := `
		SELECT r.id, r.name, r.description, r.host_id, r.max_players, r.current_count, 
		       r.status, r.is_private, r.created_at, r.updated_at,
		       u.username as host_username, u.rating as host_rating
		FROM rooms r
		JOIN users u ON r.host_id = u.id ` +
		whereClause + " " + orderClause + " " + limitClause

	type RoomWithHost struct {
		models.Room
		HostUsername string `db:"host_username" json:"host_username"`
		HostRating   int    `db:"host_rating" json:"host_rating"`
	}

	var rooms []RoomWithHost
	err = h.DB.Select(&rooms, mainQuery, args...)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to fetch rooms")
		return
	}

	// Получаем участников для каждой комнаты
	for i := range rooms {
		var participants []models.User
		err := h.DB.Select(&participants, `
			SELECT u.id, u.username, u.rating
			FROM users u
			JOIN room_participants rp ON u.id = rp.user_id
			WHERE rp.room_id = $1
			ORDER BY rp.joined_at
		`, rooms[i].ID)

		if err == nil {
			rooms[i].Participants = participants
		}
	}

	pagination := utils.NewPaginationMeta(query.Page, query.PerPage, total)
	utils.PaginatedSuccessResponse(c, rooms, pagination, "Rooms fetched successfully")
}

// GetRoom получение комнаты по ID
func (h *RoomHandlers) GetRoom(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid room ID")
		return
	}

	var room models.Room
	err = h.DB.Get(&room, `
		SELECT id, name, description, host_id, max_players, current_count, 
		       status, is_private, created_at, updated_at
		FROM rooms WHERE id = $1
	`, roomID)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "Room not found")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	// Получаем информацию о хосте
	var host models.User
	err = h.DB.Get(&host, `
		SELECT id, username, rating, wins, losses
		FROM users WHERE id = $1
	`, room.HostID)

	if err == nil {
		room.Host = &host
	}

	// Получаем участников
	var participants []models.User
	err = h.DB.Select(&participants, `
		SELECT u.id, u.username, u.rating, u.wins, u.losses
		FROM users u
		JOIN room_participants rp ON u.id = rp.user_id
		WHERE rp.room_id = $1
		ORDER BY rp.joined_at
	`, roomID)

	if err == nil {
		room.Participants = participants
	}

	utils.SuccessResponse(c, room)
}

// CreateRoom создание новой комнаты
func (h *RoomHandlers) CreateRoom(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Валидация
	errors := validator.ValidateRoomCreation(req.Name, req.MaxPlayers)
	if errors.HasErrors() {
		utils.BadRequestResponse(c, errors.Error())
		return
	}

	// Проверяем, что пользователь не находится в другой активной комнате
	var activeRoomCount int
	err := h.DB.Get(&activeRoomCount, `
		SELECT COUNT(*) 
		FROM room_participants rp
		JOIN rooms r ON rp.room_id = r.id
		WHERE rp.user_id = $1 AND r.status IN ('waiting', 'in_progress')
	`, userID)

	if err != nil {
		utils.InternalErrorResponse(c, "Database error")
		return
	}

	if activeRoomCount > 0 {
		utils.ConflictResponse(c, "You are already in an active room")
		return
	}

	// Валидация пароля для приватных комнат
	if req.IsPrivate && req.Password == "" {
		utils.BadRequestResponse(c, "Private rooms must have a password")
		return
	}

	// Начинаем транзакцию
	tx, err := h.DB.Beginx()
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Создаем комнату
	var roomID int
	err = tx.QueryRow(`
		INSERT INTO rooms (name, description, host_id, max_players, is_private, password)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, req.Name, req.Description, userID, req.MaxPlayers, req.IsPrivate, req.Password).Scan(&roomID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to create room")
		return
	}

	// Добавляем хоста как участника
	_, err = tx.Exec(`
		INSERT INTO room_participants (room_id, user_id)
		VALUES ($1, $2)
	`, roomID, userID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to add host to room")
		return
	}

	// Коммитим транзакцию
	if err = tx.Commit(); err != nil {
		utils.InternalErrorResponse(c, "Failed to commit transaction")
		return
	}

	// Получаем созданную комнату
	var room models.Room
	err = h.DB.Get(&room, `
		SELECT id, name, description, host_id, max_players, current_count, 
		       status, is_private, created_at, updated_at
		FROM rooms WHERE id = $1
	`, roomID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to fetch created room")
		return
	}

	utils.CreatedResponse(c, gin.H{
		"room":    room,
		"message": "Room created successfully",
	})
}

// UpdateRoom обновление комнаты (только хост)
func (h *RoomHandlers) UpdateRoom(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid room ID")
		return
	}

	userID := c.GetInt("user_id")

	// Проверяем, что пользователь является хостом
	var hostID int
	var status string
	err = h.DB.QueryRow(`
		SELECT host_id, status FROM rooms WHERE id = $1
	`, roomID).Scan(&hostID, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "Room not found")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	if hostID != userID {
		utils.ForbiddenResponse(c, "Only room host can update room")
		return
	}

	if status != "waiting" {
		utils.BadRequestResponse(c, "Cannot update room that is not in waiting status")
		return
	}

	var req UpdateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Валидация
	var errors validator.ValidationErrors

	if req.Name != "" {
		if err := validator.ValidateRoomName(req.Name); err != nil {
			errors = append(errors, *err)
		}
	}

	if req.MaxPlayers > 0 {
		if err := validator.ValidateMaxPlayers(req.MaxPlayers); err != nil {
			errors = append(errors, *err)
		}

		// Проверяем, что новый лимит не меньше текущего количества игроков
		var currentCount int
		err = h.DB.Get(&currentCount, `SELECT current_count FROM rooms WHERE id = $1`, roomID)
		if err == nil && req.MaxPlayers < currentCount {
			errors = append(errors, validator.ValidationError{
				Field:   "max_players",
				Message: "Cannot set max players less than current player count",
				Code:    "TOO_SMALL",
				Value:   req.MaxPlayers,
			})
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

	if req.Description != "" {
		updateFields = append(updateFields, "description = $"+strconv.Itoa(argIndex))
		args = append(args, req.Description)
		argIndex++
	}

	if req.MaxPlayers > 0 {
		updateFields = append(updateFields, "max_players = $"+strconv.Itoa(argIndex))
		args = append(args, req.MaxPlayers)
		argIndex++
	}

	if req.IsPrivate != nil {
		updateFields = append(updateFields, "is_private = $"+strconv.Itoa(argIndex))
		args = append(args, *req.IsPrivate)
		argIndex++

		// Если комната становится приватной, требуем пароль
		if *req.IsPrivate && req.Password == "" {
			utils.BadRequestResponse(c, "Private rooms must have a password")
			return
		}
	}

	if req.Password != "" {
		updateFields = append(updateFields, "password = $"+strconv.Itoa(argIndex))
		args = append(args, req.Password)
		argIndex++
	}

	if len(updateFields) == 0 {
		utils.BadRequestResponse(c, "No fields to update")
		return
	}

	updateFields = append(updateFields, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, roomID)

	query := "UPDATE rooms SET " + joinStrings(updateFields, ", ") + " WHERE id = $" + strconv.Itoa(argIndex)

	_, err = h.DB.Exec(query, args...)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to update room")
		return
	}

	// Получаем обновленную комнату
	var room models.Room
	err = h.DB.Get(&room, `
		SELECT id, name, description, host_id, max_players, current_count, 
		       status, is_private, created_at, updated_at
		FROM rooms WHERE id = $1
	`, roomID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to fetch updated room")
		return
	}

	// Отправляем обновление через WebSocket
	wsMsg := models.WSMessage{
		Type: "room_updated",
		Data: gin.H{
			"room_id": roomID,
			"action":  "room_settings_changed",
			"room":    room,
		},
	}
	msgBytes, _ := json.Marshal(wsMsg)
	h.Hub.BroadcastToRoom(roomID, msgBytes)

	utils.SuccessResponse(c, room, "Room updated successfully")
}

// DeleteRoom удаление комнаты (только хост)
func (h *RoomHandlers) DeleteRoom(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid room ID")
		return
	}

	userID := c.GetInt("user_id")

	// Проверяем, что пользователь является хостом
	var hostID int
	var status string
	err = h.DB.QueryRow(`
		SELECT host_id, status FROM rooms WHERE id = $1
	`, roomID).Scan(&hostID, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "Room not found")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	if hostID != userID {
		utils.ForbiddenResponse(c, "Only room host can delete room")
		return
	}

	if status == "in_progress" {
		utils.BadRequestResponse(c, "Cannot delete room with tournament in progress")
		return
	}

	// Уведомляем участников через WebSocket перед удалением
	wsMsg := models.WSMessage{
		Type: "room_deleted",
		Data: gin.H{
			"room_id": roomID,
			"message": "Room has been deleted by host",
		},
	}
	msgBytes, _ := json.Marshal(wsMsg)
	h.Hub.BroadcastToRoom(roomID, msgBytes)

	// Удаляем комнату (каскадное удаление участников)
	_, err = h.DB.Exec(`DELETE FROM rooms WHERE id = $1`, roomID)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to delete room")
		return
	}

	utils.NoContentResponse(c, "Room deleted successfully")
}

// JoinRoom присоединение к комнате
func (h *RoomHandlers) JoinRoom(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid room ID")
		return
	}

	userID := c.GetInt("user_id")

	var req JoinRoomRequest
	c.ShouldBindJSON(&req)

	// Проверяем, что пользователь не в другой активной комнате
	var activeRoomCount int
	err = h.DB.Get(&activeRoomCount, `
		SELECT COUNT(*) 
		FROM room_participants rp
		JOIN rooms r ON rp.room_id = r.id
		WHERE rp.user_id = $1 AND r.status IN ('waiting', 'in_progress')
	`, userID)

	if err != nil {
		utils.InternalErrorResponse(c, "Database error")
		return
	}

	if activeRoomCount > 0 {
		utils.ConflictResponse(c, "You are already in an active room")
		return
	}

	// Получаем информацию о комнате
	var room models.Room
	err = h.DB.Get(&room, `
		SELECT id, max_players, current_count, status, is_private, password
		FROM rooms WHERE id = $1
	`, roomID)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "Room not found")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	if room.Status != "waiting" {
		utils.BadRequestResponse(c, "Room is not accepting new players")
		return
	}

	if room.CurrentCount >= room.MaxPlayers {
		utils.BadRequestResponse(c, "Room is full")
		return
	}

	if room.IsPrivate && room.Password != req.Password {
		utils.UnauthorizedResponse(c, "Invalid room password")
		return
	}

	// Проверяем, не находится ли пользователь уже в комнате
	var exists bool
	err = h.DB.Get(&exists, `
		SELECT EXISTS(SELECT 1 FROM room_participants WHERE room_id = $1 AND user_id = $2)
	`, roomID, userID)

	if err != nil || exists {
		utils.ConflictResponse(c, "Already in room")
		return
	}

	// Начинаем транзакцию
	tx, err := h.DB.Beginx()
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Добавляем пользователя в комнату
	_, err = tx.Exec(`
		INSERT INTO room_participants (room_id, user_id)
		VALUES ($1, $2)
	`, roomID, userID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to join room")
		return
	}

	// Обновляем счетчик участников
	_, err = tx.Exec(`
		UPDATE rooms SET current_count = current_count + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, roomID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to update room count")
		return
	}

	// Коммитим транзакцию
	if err = tx.Commit(); err != nil {
		utils.InternalErrorResponse(c, "Failed to commit transaction")
		return
	}

	// Получаем информацию о пользователе для уведомления
	var user models.User
	err = h.DB.Get(&user, `
		SELECT id, username, rating FROM users WHERE id = $1
	`, userID)

	// Отправляем обновление через WebSocket
	wsMsg := models.WSMessage{
		Type: "room_updated",
		Data: gin.H{
			"room_id": roomID,
			"action":  "user_joined",
			"user":    user,
		},
	}
	msgBytes, _ := json.Marshal(wsMsg)
	h.Hub.BroadcastToRoom(roomID, msgBytes)

	utils.SuccessResponse(c, gin.H{
		"message": "Joined room successfully",
		"room_id": roomID,
	})
}

// LeaveRoom покидание комнаты
func (h *RoomHandlers) LeaveRoom(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid room ID")
		return
	}

	userID := c.GetInt("user_id")

	// Проверяем, является ли пользователь хостом
	var hostID int
	var status string
	err = h.DB.QueryRow(`
		SELECT host_id, status FROM rooms WHERE id = $1
	`, roomID).Scan(&hostID, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "Room not found")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	if status == "in_progress" {
		utils.BadRequestResponse(c, "Cannot leave room with tournament in progress")
		return
	}

	// Начинаем транзакцию
	tx, err := h.DB.Beginx()
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Удаляем пользователя из комнаты
	result, err := tx.Exec(`
		DELETE FROM room_participants
		WHERE room_id = $1 AND user_id = $2
	`, roomID, userID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to leave room")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		utils.BadRequestResponse(c, "Not in this room")
		return
	}

	// Обновляем счетчик участников
	_, err = tx.Exec(`
		UPDATE rooms SET current_count = current_count - 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, roomID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to update room count")
		return
	}

	// Если хост покидает комнату, передаем права другому участнику или удаляем комнату
	if hostID == userID {
		var newHostID sql.NullInt64
		err = tx.Get(&newHostID, `
			SELECT user_id FROM room_participants 
			WHERE room_id = $1 
			ORDER BY joined_at ASC 
			LIMIT 1
		`, roomID)

		if err == nil && newHostID.Valid {
			// Передаем права хоста
			_, err = tx.Exec(`
				UPDATE rooms SET host_id = $1, updated_at = CURRENT_TIMESTAMP
				WHERE id = $2
			`, newHostID.Int64, roomID)

			if err != nil {
				utils.InternalErrorResponse(c, "Failed to transfer host rights")
				return
			}
		} else {
			// Удаляем пустую комнату
			_, err = tx.Exec(`DELETE FROM rooms WHERE id = $1`, roomID)
			if err != nil {
				utils.InternalErrorResponse(c, "Failed to delete empty room")
				return
			}
		}
	}

	// Коммитим транзакцию
	if err = tx.Commit(); err != nil {
		utils.InternalErrorResponse(c, "Failed to commit transaction")
		return
	}

	// Получаем информацию о пользователе для уведомления
	var user models.User
	err = h.DB.Get(&user, `
		SELECT id, username, rating FROM users WHERE id = $1
	`, userID)

	// Отправляем обновление через WebSocket
	wsMsg := models.WSMessage{
		Type: "room_updated",
		Data: gin.H{
			"room_id": roomID,
			"action":  "user_left",
			"user":    user,
		},
	}
	msgBytes, _ := json.Marshal(wsMsg)
	h.Hub.BroadcastToRoom(roomID, msgBytes)

	utils.SuccessResponse(c, gin.H{
		"message": "Left room successfully",
	})
}

// KickPlayer исключение игрока (только хост)
func (h *RoomHandlers) KickPlayer(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid room ID")
		return
	}

	userID := c.GetInt("user_id")

	// Проверяем, что пользователь является хостом
	var hostID int
	var status string
	err = h.DB.QueryRow(`
		SELECT host_id, status FROM rooms WHERE id = $1
	`, roomID).Scan(&hostID, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "Room not found")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	if hostID != userID {
		utils.ForbiddenResponse(c, "Only room host can kick players")
		return
	}

	if status != "waiting" {
		utils.BadRequestResponse(c, "Cannot kick players when tournament is in progress")
		return
	}

	var req KickPlayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	if req.UserID == userID {
		utils.BadRequestResponse(c, "Cannot kick yourself")
		return
	}

	// Проверяем, что игрок находится в комнате
	var exists bool
	err = h.DB.Get(&exists, `
		SELECT EXISTS(SELECT 1 FROM room_participants WHERE room_id = $1 AND user_id = $2)
	`, roomID, req.UserID)

	if err != nil {
		utils.InternalErrorResponse(c, "Database error")
		return
	}

	if !exists {
		utils.BadRequestResponse(c, "Player is not in this room")
		return
	}

	// Начинаем транзакцию
	tx, err := h.DB.Beginx()
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Удаляем игрока из комнаты
	_, err = tx.Exec(`
		DELETE FROM room_participants
		WHERE room_id = $1 AND user_id = $2
	`, roomID, req.UserID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to kick player")
		return
	}

	// Обновляем счетчик участников
	_, err = tx.Exec(`
		UPDATE rooms SET current_count = current_count - 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, roomID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to update room count")
		return
	}

	// Коммитим транзакцию
	if err = tx.Commit(); err != nil {
		utils.InternalErrorResponse(c, "Failed to commit transaction")
		return
	}

	// Получаем информацию о исключенном игроке
	var kickedUser models.User
	err = h.DB.Get(&kickedUser, `
		SELECT id, username, rating FROM users WHERE id = $1
	`, req.UserID)

	// Отправляем уведомления через WebSocket
	wsMsg := models.WSMessage{
		Type: "room_updated",
		Data: gin.H{
			"room_id": roomID,
			"action":  "user_kicked",
			"user":    kickedUser,
		},
	}
	msgBytes, _ := json.Marshal(wsMsg)
	h.Hub.BroadcastToRoom(roomID, msgBytes)

	// Отправляем персональное уведомление исключенному игроку
	// TODO: В будущем можно добавить отправку персональных сообщений через WebSocket

	utils.SuccessResponse(c, gin.H{
		"message": "Player kicked successfully",
		"user":    kickedUser,
	})
}

// SetRoomPassword установка/изменение пароля комнаты (только хост)
func (h *RoomHandlers) SetRoomPassword(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid room ID")
		return
	}

	userID := c.GetInt("user_id")

	// Проверяем, что пользователь является хостом
	var hostID int
	var status string
	err = h.DB.QueryRow(`
		SELECT host_id, status FROM rooms WHERE id = $1
	`, roomID).Scan(&hostID, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "Room not found")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	if hostID != userID {
		utils.ForbiddenResponse(c, "Only room host can change password")
		return
	}

	if status != "waiting" {
		utils.BadRequestResponse(c, "Cannot change password when tournament is in progress")
		return
	}

	var req struct {
		Password  string `json:"password"`
		IsPrivate bool   `json:"is_private"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Валидация: если комната приватная, пароль обязателен
	if req.IsPrivate && req.Password == "" {
		utils.BadRequestResponse(c, "Private rooms must have a password")
		return
	}

	// Обновляем пароль и статус приватности
	_, err = h.DB.Exec(`
		UPDATE rooms 
		SET password = $1, is_private = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`, req.Password, req.IsPrivate, roomID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to update room password")
		return
	}

	// Отправляем обновление через WebSocket
	wsMsg := models.WSMessage{
		Type: "room_updated",
		Data: gin.H{
			"room_id":    roomID,
			"action":     "password_changed",
			"is_private": req.IsPrivate,
		},
	}
	msgBytes, _ := json.Marshal(wsMsg)
	h.Hub.BroadcastToRoom(roomID, msgBytes)

	utils.SuccessResponse(c, gin.H{
		"message":    "Room password updated successfully",
		"is_private": req.IsPrivate,
	})
}

// GetRoomParticipants получение списка участников комнаты
func (h *RoomHandlers) GetRoomParticipants(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid room ID")
		return
	}

	// Проверяем существование комнаты
	var exists bool
	err = h.DB.Get(&exists, `SELECT EXISTS(SELECT 1 FROM rooms WHERE id = $1)`, roomID)
	if err != nil {
		utils.InternalErrorResponse(c, "Database error")
		return
	}
	if !exists {
		utils.NotFoundResponse(c, "Room not found")
		return
	}

	// Получаем участников с дополнительной информацией
	type ParticipantInfo struct {
		models.User
		JoinedAt time.Time `db:"joined_at" json:"joined_at"`
		IsHost   bool      `json:"is_host"`
	}

	var participants []ParticipantInfo
	err = h.DB.Select(&participants, `
		SELECT u.id, u.username, u.rating, u.wins, u.losses, u.created_at,
		       rp.joined_at, (u.id = r.host_id) as is_host
		FROM users u
		JOIN room_participants rp ON u.id = rp.user_id
		JOIN rooms r ON rp.room_id = r.id
		WHERE rp.room_id = $1
		ORDER BY rp.joined_at ASC
	`, roomID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to fetch participants")
		return
	}

	utils.SuccessResponse(c, participants, "Room participants fetched successfully")
}
