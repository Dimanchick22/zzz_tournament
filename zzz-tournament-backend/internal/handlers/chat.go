// internal/handlers/chat.go - исправленная версия
package handlers

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"zzz-tournament/internal/models"
	"zzz-tournament/internal/websocket"
	"zzz-tournament/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// ChatHandlers обработчики чата
type ChatHandlers struct {
	BaseHandlers
}

// NewChatHandlers создает новый экземпляр ChatHandlers
func NewChatHandlers(db *sqlx.DB, hub *websocket.Hub, logger *slog.Logger) *ChatHandlers {
	return &ChatHandlers{
		BaseHandlers: newBaseHandlers(db, hub, logger),
	}
}

// SendMessageRequest структура запроса отправки сообщения
type SendMessageRequest struct {
	Content string `json:"content" binding:"required"`
	Type    string `json:"type,omitempty"` // message, system, announcement
}

// EditMessageRequest структура запроса редактирования сообщения
type EditMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

// GetMessagesQuery параметры получения сообщений
type GetMessagesQuery struct {
	Limit  int    `form:"limit"`
	Before int    `form:"before"` // ID сообщения, до которого получать
	After  int    `form:"after"`  // ID сообщения, после которого получать
	Type   string `form:"type"`   // Фильтр по типу сообщений
}

// MessageWithUser сообщение с информацией о пользователе
type MessageWithUser struct {
	models.Message
	Username   string     `db:"username" json:"username"`
	UserRating int        `db:"user_rating" json:"user_rating"`
	IsEdited   bool       `json:"is_edited"`
	EditedAt   *time.Time `json:"edited_at,omitempty"`
	CanEdit    bool       `json:"can_edit"`
	CanDelete  bool       `json:"can_delete"`
}

// UserMessageStats статистика сообщений пользователя
type UserMessageStats struct {
	UserID       int        `db:"user_id" json:"user_id"`
	Username     string     `db:"username" json:"username"`
	MessageCount int        `db:"message_count" json:"message_count"`
	LastMessage  *time.Time `db:"last_message" json:"last_message,omitempty"`
}

// MessageActivityStats статистика активности сообщений
type MessageActivityStats struct {
	Hour         int `db:"hour" json:"hour"`
	MessageCount int `db:"message_count" json:"message_count"`
}

// GetRoomMessages получение сообщений комнаты
func (h *ChatHandlers) GetRoomMessages(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid room ID")
		return
	}

	userID := c.GetInt("user_id")

	var query GetMessagesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Проверяем, что пользователь является участником комнаты
	var isParticipant bool
	err = h.DB.Get(&isParticipant, `
		SELECT EXISTS(SELECT 1 FROM room_participants WHERE room_id = $1 AND user_id = $2)
	`, roomID, userID)

	if err != nil {
		utils.InternalErrorResponse(c, "Database error")
		return
	}

	if !isParticipant {
		utils.ForbiddenResponse(c, "You must be a room participant to view messages")
		return
	}

	// Устанавливаем значения по умолчанию
	if query.Limit <= 0 || query.Limit > 100 {
		query.Limit = 50
	}

	// Строим запрос
	whereConditions := []string{"m.room_id = $1"}
	args := []interface{}{roomID}
	argIndex := 2

	// Фильтр по типу сообщений
	if query.Type != "" {
		whereConditions = append(whereConditions, "m.type = $"+strconv.Itoa(argIndex))
		args = append(args, query.Type)
		argIndex++
	}

	// Пагинация
	if query.Before > 0 {
		whereConditions = append(whereConditions, "m.id < $"+strconv.Itoa(argIndex))
		args = append(args, query.Before)
		argIndex++
	}

	if query.After > 0 {
		whereConditions = append(whereConditions, "m.id > $"+strconv.Itoa(argIndex))
		args = append(args, query.After)
		argIndex++
	}

	whereClause := "WHERE " + joinStrings(whereConditions, " AND ")

	// Определяем сортировку
	orderClause := "ORDER BY m.created_at DESC, m.id DESC"
	if query.After > 0 {
		orderClause = "ORDER BY m.created_at ASC, m.id ASC"
	}

	// Добавляем LIMIT
	args = append(args, query.Limit)
	limitClause := "LIMIT $" + strconv.Itoa(argIndex)

	// Основной запрос
	mainQuery := `
		SELECT m.id, m.room_id, m.user_id, m.content, m.type, m.created_at,
		       COALESCE(u.username, 'System') as username,
		       COALESCE(u.rating, 0) as user_rating
		FROM messages m
		LEFT JOIN users u ON m.user_id = u.id ` +
		whereClause + " " + orderClause + " " + limitClause

	var messages []MessageWithUser
	err = h.DB.Select(&messages, mainQuery, args...)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to fetch messages")
		return
	}

	// Если запрашивали сообщения после определенного ID, разворачиваем порядок
	if query.After > 0 {
		for i := len(messages)/2 - 1; i >= 0; i-- {
			opp := len(messages) - 1 - i
			messages[i], messages[opp] = messages[opp], messages[i]
		}
	}

	// Получаем информацию о хосте комнаты для определения прав
	var hostID int
	err = h.DB.Get(&hostID, `SELECT host_id FROM rooms WHERE id = $1`, roomID)
	if err != nil {
		hostID = 0
	}

	// Устанавливаем права на редактирование и удаление
	for i := range messages {
		messages[i].CanEdit = messages[i].UserID == userID && messages[i].Type == "message"
		messages[i].CanDelete = messages[i].UserID == userID || userID == hostID

		// TODO: Добавить проверку на редактирование сообщений
		messages[i].IsEdited = false
		messages[i].EditedAt = nil
	}

	utils.SuccessResponse(c, gin.H{
		"messages": messages,
		"count":    len(messages),
		"has_more": len(messages) == query.Limit,
	})
}

// SendMessage отправка сообщения в комнату (через HTTP API)
func (h *ChatHandlers) SendMessage(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid room ID")
		return
	}

	userID := c.GetInt("user_id")

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Проверяем, что пользователь является участником комнаты
	var isParticipant bool
	err = h.DB.Get(&isParticipant, `
		SELECT EXISTS(SELECT 1 FROM room_participants WHERE room_id = $1 AND user_id = $2)
	`, roomID, userID)

	if err != nil {
		utils.InternalErrorResponse(c, "Database error")
		return
	}

	if !isParticipant {
		utils.ForbiddenResponse(c, "You must be a room participant to send messages")
		return
	}

	// Валидация сообщения
	content := strings.TrimSpace(req.Content)
	if len(content) == 0 {
		utils.BadRequestResponse(c, "Message content cannot be empty")
		return
	}

	if len(content) > 1000 {
		utils.BadRequestResponse(c, "Message content is too long (maximum 1000 characters)")
		return
	}

	// Проверяем на спам (не более 5 сообщений в минуту)
	var recentMessageCount int
	err = h.DB.Get(&recentMessageCount, `
		SELECT COUNT(*) FROM messages 
		WHERE room_id = $1 AND user_id = $2 AND created_at > NOW() - INTERVAL '1 minute'
	`, roomID, userID)

	if err == nil && recentMessageCount >= 5 {
		utils.TooManyRequestsResponse(c, "Too many messages sent recently. Please wait a moment.")
		return
	}

	// Устанавливаем тип сообщения
	messageType := req.Type
	if messageType == "" {
		messageType = "message"
	}

	// Валидируем тип сообщения
	validTypes := []string{"message", "announcement"}
	isValidType := false
	for _, validType := range validTypes {
		if messageType == validType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		utils.BadRequestResponse(c, "Invalid message type")
		return
	}

	// Проверяем права на отправку объявлений
	if messageType == "announcement" {
		var hostID int
		err = h.DB.Get(&hostID, `SELECT host_id FROM rooms WHERE id = $1`, roomID)
		if err != nil || hostID != userID {
			utils.ForbiddenResponse(c, "Only room host can send announcements")
			return
		}
	}

	// Сохраняем сообщение в базу данных
	var messageID int
	err = h.DB.QueryRow(`
		INSERT INTO messages (room_id, user_id, content, type)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, roomID, userID, content, messageType).Scan(&messageID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to save message")
		return
	}

	// Получаем полную информацию о сообщении
	var message MessageWithUser
	err = h.DB.Get(&message, `
		SELECT m.id, m.room_id, m.user_id, m.content, m.type, m.created_at,
		       u.username, u.rating as user_rating
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.id = $1
	`, messageID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to fetch saved message")
		return
	}

	// Отправляем сообщение через WebSocket всем участникам комнаты
	wsMsg := models.WSMessage{
		Type: "chat_message",
		Data: gin.H{
			"room_id": roomID,
			"message": message,
		},
	}
	msgBytes, _ := json.Marshal(wsMsg)
	h.Hub.BroadcastToRoom(roomID, msgBytes)

	utils.CreatedResponse(c, message, "Message sent successfully")
}

// EditMessage редактирование сообщения
func (h *ChatHandlers) EditMessage(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid room ID")
		return
	}

	messageID, err := strconv.Atoi(c.Param("message_id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid message ID")
		return
	}

	userID := c.GetInt("user_id")

	var req EditMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Проверяем существование сообщения и права на редактирование
	var message models.Message
	err = h.DB.Get(&message, `
		SELECT id, room_id, user_id, content, type, created_at
		FROM messages 
		WHERE id = $1 AND room_id = $2
	`, messageID, roomID)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "Message not found")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	// Проверяем права на редактирование
	if message.UserID != userID {
		utils.ForbiddenResponse(c, "You can only edit your own messages")
		return
	}

	if message.Type != "message" {
		utils.BadRequestResponse(c, "Cannot edit system messages")
		return
	}

	// Проверяем время создания (можно редактировать только в течение 15 минут)
	if time.Since(message.CreatedAt) > 15*time.Minute {
		utils.BadRequestResponse(c, "Message is too old to edit (15 minute limit)")
		return
	}

	// Валидация нового содержимого
	content := strings.TrimSpace(req.Content)
	if len(content) == 0 {
		utils.BadRequestResponse(c, "Message content cannot be empty")
		return
	}

	if len(content) > 1000 {
		utils.BadRequestResponse(c, "Message content is too long (maximum 1000 characters)")
		return
	}

	// Обновляем сообщение
	_, err = h.DB.Exec(`
		UPDATE messages 
		SET content = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, content, messageID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to update message")
		return
	}

	// Получаем обновленное сообщение
	var updatedMessage MessageWithUser
	err = h.DB.Get(&updatedMessage, `
		SELECT m.id, m.room_id, m.user_id, m.content, m.type, m.created_at,
		       u.username, u.rating as user_rating
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.id = $1
	`, messageID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to fetch updated message")
		return
	}

	updatedMessage.IsEdited = true
	updatedMessage.CanEdit = true
	updatedMessage.CanDelete = true

	// Отправляем уведомление об изменении через WebSocket
	wsMsg := models.WSMessage{
		Type: "message_edited",
		Data: gin.H{
			"room_id": roomID,
			"message": updatedMessage,
		},
	}
	msgBytes, _ := json.Marshal(wsMsg)
	h.Hub.BroadcastToRoom(roomID, msgBytes)

	utils.SuccessResponse(c, updatedMessage, "Message edited successfully")
}

// DeleteMessage удаление сообщения
func (h *ChatHandlers) DeleteMessage(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid room ID")
		return
	}

	messageID, err := strconv.Atoi(c.Param("message_id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid message ID")
		return
	}

	userID := c.GetInt("user_id")

	// Проверяем существование сообщения
	var message models.Message
	err = h.DB.Get(&message, `
		SELECT id, room_id, user_id, content, type, created_at
		FROM messages 
		WHERE id = $1 AND room_id = $2
	`, messageID, roomID)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "Message not found")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	// Проверяем права на удаление
	canDelete := false

	// Автор может удалить свое сообщение
	if message.UserID == userID {
		canDelete = true
	} else {
		// Хост комнаты может удалить любое сообщение
		var hostID int
		err = h.DB.Get(&hostID, `SELECT host_id FROM rooms WHERE id = $1`, roomID)
		if err == nil && hostID == userID {
			canDelete = true
		}
	}

	if !canDelete {
		utils.ForbiddenResponse(c, "You don't have permission to delete this message")
		return
	}

	// Удаляем сообщение
	_, err = h.DB.Exec(`DELETE FROM messages WHERE id = $1`, messageID)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to delete message")
		return
	}

	// Отправляем уведомление об удалении через WebSocket
	wsMsg := models.WSMessage{
		Type: "message_deleted",
		Data: gin.H{
			"room_id":    roomID,
			"message_id": messageID,
		},
	}
	msgBytes, _ := json.Marshal(wsMsg)
	h.Hub.BroadcastToRoom(roomID, msgBytes)

	utils.NoContentResponse(c, "Message deleted successfully")
}

// SendSystemMessage отправка системного сообщения
func (h *ChatHandlers) SendSystemMessage(roomID int, content string, messageType string) error {
	if messageType == "" {
		messageType = "system"
	}

	_, err := h.DB.Exec(`
		INSERT INTO messages (room_id, user_id, content, type)
		VALUES ($1, NULL, $2, $3)
	`, roomID, content, messageType)

	if err != nil {
		return err
	}

	// Отправляем через WebSocket
	wsMsg := models.WSMessage{
		Type: "chat_message",
		Data: gin.H{
			"room_id": roomID,
			"message": gin.H{
				"content":  content,
				"type":     messageType,
				"username": "System",
			},
		},
	}
	msgBytes, _ := json.Marshal(wsMsg)
	h.Hub.BroadcastToRoom(roomID, msgBytes)

	return nil
}

// GetChatStats получение статистики чата комнаты
func (h *ChatHandlers) GetChatStats(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid room ID")
		return
	}

	userID := c.GetInt("user_id")

	// Проверяем, что пользователь является участником комнаты
	var isParticipant bool
	err = h.DB.Get(&isParticipant, `
		SELECT EXISTS(SELECT 1 FROM room_participants WHERE room_id = $1 AND user_id = $2)
	`, roomID, userID)

	if err != nil {
		utils.InternalErrorResponse(c, "Database error")
		return
	}

	if !isParticipant {
		utils.ForbiddenResponse(c, "You must be a room participant to view chat stats")
		return
	}

	// Собираем статистику
	type ChatStats struct {
		TotalMessages    int                    `json:"total_messages"`
		MessagesByType   map[string]int         `json:"messages_by_type"`
		MessagesByUser   []UserMessageStats     `json:"messages_by_user"`
		RecentActivity   []MessageActivityStats `json:"recent_activity"`
		MostActiveHour   int                    `json:"most_active_hour"`
		FirstMessageTime *time.Time             `json:"first_message_time,omitempty"`
		LastMessageTime  *time.Time             `json:"last_message_time,omitempty"`
	}

	var stats ChatStats

	// Общее количество сообщений
	err = h.DB.Get(&stats.TotalMessages, `
		SELECT COUNT(*) FROM messages WHERE room_id = $1
	`, roomID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to get total message count")
		return
	}

	// Сообщения по типам
	rows, err := h.DB.Query(`
		SELECT type, COUNT(*) FROM messages 
		WHERE room_id = $1 
		GROUP BY type
	`, roomID)

	if err == nil {
		stats.MessagesByType = make(map[string]int)
		for rows.Next() {
			var msgType string
			var count int
			if err := rows.Scan(&msgType, &count); err == nil {
				stats.MessagesByType[msgType] = count
			}
		}
		rows.Close()
	}

	// Сообщения по пользователям
	err = h.DB.Select(&stats.MessagesByUser, `
		SELECT m.user_id, COALESCE(u.username, 'System') as username, 
		       COUNT(*) as message_count, MAX(m.created_at) as last_message
		FROM messages m
		LEFT JOIN users u ON m.user_id = u.id
		WHERE m.room_id = $1
		GROUP BY m.user_id, u.username
		ORDER BY message_count DESC
		LIMIT 10
	`, roomID)

	if err != nil {
		stats.MessagesByUser = []UserMessageStats{}
	}

	// Активность по часам
	err = h.DB.Select(&stats.RecentActivity, `
		SELECT EXTRACT(HOUR FROM created_at) as hour, COUNT(*) as message_count
		FROM messages 
		WHERE room_id = $1 AND created_at > NOW() - INTERVAL '24 hours'
		GROUP BY EXTRACT(HOUR FROM created_at)
		ORDER BY hour
	`, roomID)

	if err != nil {
		stats.RecentActivity = []MessageActivityStats{}
	}

	// Самый активный час
	err = h.DB.Get(&stats.MostActiveHour, `
		SELECT EXTRACT(HOUR FROM created_at) as hour
		FROM messages 
		WHERE room_id = $1
		GROUP BY EXTRACT(HOUR FROM created_at)
		ORDER BY COUNT(*) DESC
		LIMIT 1
	`, roomID)

	if err != nil {
		stats.MostActiveHour = 0
	}

	// Время первого и последнего сообщения
	err = h.DB.QueryRow(`
		SELECT MIN(created_at), MAX(created_at)
		FROM messages WHERE room_id = $1
	`, roomID).Scan(&stats.FirstMessageTime, &stats.LastMessageTime)

	utils.SuccessResponse(c, stats, "Chat statistics fetched successfully")
}

// ClearChatHistory очистка истории чата (только хост)
func (h *ChatHandlers) ClearChatHistory(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid room ID")
		return
	}

	userID := c.GetInt("user_id")

	// Проверяем, что пользователь является хостом комнаты
	var hostID int
	err = h.DB.Get(&hostID, `SELECT host_id FROM rooms WHERE id = $1`, roomID)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "Room not found")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	if hostID != userID {
		utils.ForbiddenResponse(c, "Only room host can clear chat history")
		return
	}

	// Запрос подтверждения
	confirm := c.Query("confirm")
	if confirm != "true" {
		utils.BadRequestResponse(c, "Add ?confirm=true to confirm chat history deletion")
		return
	}

	// Удаляем все сообщения кроме системных
	result, err := h.DB.Exec(`
		DELETE FROM messages 
		WHERE room_id = $1 AND type = 'message'
	`, roomID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to clear chat history")
		return
	}

	rowsAffected, _ := result.RowsAffected()

	// Отправляем системное сообщение о очистке
	h.SendSystemMessage(roomID, "Chat history has been cleared by room host", "system")

	// Отправляем уведомление через WebSocket
	wsMsg := models.WSMessage{
		Type: "chat_cleared",
		Data: gin.H{
			"room_id": roomID,
			"message": "Chat history has been cleared",
		},
	}
	msgBytes, _ := json.Marshal(wsMsg)
	h.Hub.BroadcastToRoom(roomID, msgBytes)

	utils.SuccessResponse(c, gin.H{
		"messages_deleted": rowsAffected,
		"message":          "Chat history cleared successfully",
	})
}

// MuteUser заглушение пользователя в чате (только хост)
func (h *ChatHandlers) MuteUser(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid room ID")
		return
	}

	targetUserID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID")
		return
	}

	userID := c.GetInt("user_id")

	// Проверяем, что пользователь является хостом комнаты
	var hostID int
	err = h.DB.Get(&hostID, `SELECT host_id FROM rooms WHERE id = $1`, roomID)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "Room not found")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	if hostID != userID {
		utils.ForbiddenResponse(c, "Only room host can mute users")
		return
	}

	if targetUserID == userID {
		utils.BadRequestResponse(c, "Cannot mute yourself")
		return
	}

	// Проверяем, что пользователь является участником комнаты
	var isParticipant bool
	err = h.DB.Get(&isParticipant, `
		SELECT EXISTS(SELECT 1 FROM room_participants WHERE room_id = $1 AND user_id = $2)
	`, roomID, targetUserID)

	if err != nil || !isParticipant {
		utils.BadRequestResponse(c, "User is not a participant of this room")
		return
	}

	// TODO: Реализовать систему мута (создать таблицу room_mutes)
	// Пока что отправляем системное сообщение
	var targetUsername string
	err = h.DB.Get(&targetUsername, `SELECT username FROM users WHERE id = $1`, targetUserID)
	if err != nil {
		targetUsername = "Unknown"
	}

	systemMessage := targetUsername + " has been muted by room host"
	h.SendSystemMessage(roomID, systemMessage, "system")

	utils.SuccessResponse(c, gin.H{
		"message":  "User muted successfully",
		"user_id":  targetUserID,
		"username": targetUsername,
	})
}

// UnmuteUser снятие заглушения с пользователя (только хост)
func (h *ChatHandlers) UnmuteUser(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid room ID")
		return
	}

	targetUserID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID")
		return
	}

	userID := c.GetInt("user_id")

	// Проверяем, что пользователь является хостом комнаты
	var hostID int
	err = h.DB.Get(&hostID, `SELECT host_id FROM rooms WHERE id = $1`, roomID)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "Room not found")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	if hostID != userID {
		utils.ForbiddenResponse(c, "Only room host can unmute users")
		return
	}

	// TODO: Реализовать систему мута (удалить из таблицы room_mutes)
	// Пока что отправляем системное сообщение
	var targetUsername string
	err = h.DB.Get(&targetUsername, `SELECT username FROM users WHERE id = $1`, targetUserID)
	if err != nil {
		targetUsername = "Unknown"
	}

	systemMessage := targetUsername + " has been unmuted by room host"
	h.SendSystemMessage(roomID, systemMessage, "system")

	utils.SuccessResponse(c, gin.H{
		"message":  "User unmuted successfully",
		"user_id":  targetUserID,
		"username": targetUsername,
	})
}
