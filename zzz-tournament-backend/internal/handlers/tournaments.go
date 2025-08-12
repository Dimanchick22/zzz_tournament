// internal/handlers/tournaments.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"zzz-tournament/internal/models"
	"zzz-tournament/internal/websocket"
	"zzz-tournament/pkg/rating"
	"zzz-tournament/pkg/tournament"
	"zzz-tournament/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// TournamentHandlers обработчики турниров
type TournamentHandlers struct {
	BaseHandlers
}

// NewTournamentHandlers создает новый экземпляр TournamentHandlers
func NewTournamentHandlers(db *sqlx.DB, hub *websocket.Hub) *TournamentHandlers {
	return &TournamentHandlers{
		BaseHandlers: newBaseHandlers(db, hub),
	}
}

// StartTournamentRequest структура запроса запуска турнира
type StartTournamentRequest struct {
	Name   string `json:"name,omitempty"`
	Seeded bool   `json:"seeded"` // Использовать посевную сетку
}

// SubmitMatchResultRequest структура запроса результата матча
type SubmitMatchResultRequest struct {
	WinnerID int    `json:"winner_id" binding:"required"`
	Details  string `json:"details,omitempty"` // Дополнительные детали матча
}

// GetTournamentsQuery параметры фильтрации турниров
type GetTournamentsQuery struct {
	Status   string `form:"status"`
	UserID   int    `form:"user_id"`
	Page     int    `form:"page"`
	PerPage  int    `form:"per_page"`
	SortBy   string `form:"sort_by"`
	SortDesc bool   `form:"sort_desc"`
}

// StartTournament запуск турнира в комнате (только хост)
func (h *TournamentHandlers) StartTournament(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid room ID")
		return
	}

	userID := c.GetInt("user_id")

	var req StartTournamentRequest
	c.ShouldBindJSON(&req)

	// Проверяем, что пользователь является хостом комнаты
	var hostID int
	var roomStatus string
	err = h.DB.QueryRow(`
		SELECT host_id, status FROM rooms WHERE id = $1
	`, roomID).Scan(&hostID, &roomStatus)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "Room not found")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	if hostID != userID {
		utils.ForbiddenResponse(c, "Only room host can start tournament")
		return
	}

	if roomStatus != "waiting" {
		utils.BadRequestResponse(c, "Room is not in waiting status")
		return
	}

	// Получаем участников комнаты
	var participants []models.User
	err = h.DB.Select(&participants, `
		SELECT u.id, u.username, u.rating
		FROM users u
		JOIN room_participants rp ON u.id = rp.user_id
		WHERE rp.room_id = $1
		ORDER BY rp.joined_at
	`, roomID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to get participants")
		return
	}

	if len(participants) < 2 {
		utils.BadRequestResponse(c, "Need at least 2 participants to start tournament")
		return
	}

	// Проверяем, не существует ли уже турнир для этой комнаты
	var existingTournament bool
	err = h.DB.Get(&existingTournament, `
		SELECT EXISTS(SELECT 1 FROM tournaments WHERE room_id = $1)
	`, roomID)

	if err == nil && existingTournament {
		utils.ConflictResponse(c, "Tournament already exists for this room")
		return
	}

	// Начинаем транзакцию
	tx, err := h.DB.Beginx()
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Создаем турнир
	tournamentName := req.Name
	if tournamentName == "" {
		tournamentName = "Tournament for Room " + strconv.Itoa(roomID)
	}

	var tournamentID int
	err = tx.QueryRow(`
		INSERT INTO tournaments (room_id, name, status, created_at)
		VALUES ($1, $2, 'started', CURRENT_TIMESTAMP)
		RETURNING id
	`, roomID, tournamentName).Scan(&tournamentID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to create tournament")
		return
	}

	// Конвертируем участников в формат для генерации сетки
	players := make([]tournament.Player, len(participants))
	for i, p := range participants {
		players[i] = tournament.Player{
			ID:       p.ID,
			Username: p.Username,
			Rating:   p.Rating,
		}
	}

	// Генерируем турнирную сетку
	var bracket *tournament.Bracket
	if req.Seeded {
		bracket, err = tournament.GenerateSeededBracket(players)
	} else {
		bracket, err = tournament.GenerateBracket(players)
	}

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to generate tournament bracket")
		return
	}

	// Сохраняем сетку в JSON
	bracketJSON, err := json.Marshal(bracket)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to serialize bracket")
		return
	}

	// Обновляем турнир с сеткой
	_, err = tx.Exec(`
		UPDATE tournaments SET bracket = $1 WHERE id = $2
	`, bracketJSON, tournamentID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to save bracket")
		return
	}

	// Создаем матчи
	for _, match := range bracket.Matches {
		_, err = tx.Exec(`
			INSERT INTO matches (tournament_id, round, player1_id, player2_id, status)
			VALUES ($1, $2, $3, $4, 'pending')
		`, tournamentID, match.Round, match.Player1ID, match.Player2ID)

		if err != nil {
			utils.InternalErrorResponse(c, "Failed to create matches")
			return
		}
	}

	// Обновляем статус комнаты
	_, err = tx.Exec(`
		UPDATE rooms SET status = 'in_progress', updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, roomID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to update room status")
		return
	}

	// Коммитим транзакцию
	if err = tx.Commit(); err != nil {
		utils.InternalErrorResponse(c, "Failed to commit transaction")
		return
	}

	// Отправляем уведомление через WebSocket
	wsMsg := models.WSMessage{
		Type: "tournament_started",
		Data: gin.H{
			"room_id":       roomID,
			"tournament_id": tournamentID,
			"bracket":       bracket,
		},
	}
	msgBytes, _ := json.Marshal(wsMsg)
	h.Hub.BroadcastToRoom(roomID, msgBytes)

	utils.SuccessResponse(c, gin.H{
		"tournament_id": tournamentID,
		"bracket":       bracket,
		"message":       "Tournament started successfully",
	})
}

// GetTournament получение турнира по ID
func (h *TournamentHandlers) GetTournament(c *gin.Context) {
	tournamentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid tournament ID")
		return
	}

	// Получаем основную информацию о турнире
	var tournament models.Tournament
	err = h.DB.Get(&tournament, `
		SELECT id, room_id, name, status, bracket, winner_id, created_at, updated_at
		FROM tournaments WHERE id = $1
	`, tournamentID)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "Tournament not found")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	// Получаем матчи турнира
	var matches []models.Match
	err = h.DB.Select(&matches, `
		SELECT m.id, m.tournament_id, m.round, m.player1_id, m.player2_id, 
		       m.winner_id, m.status, m.created_at, m.updated_at,
		       p1.username as player1_username, p1.rating as player1_rating,
		       p2.username as player2_username, p2.rating as player2_rating,
		       w.username as winner_username
		FROM matches m
		LEFT JOIN users p1 ON m.player1_id = p1.id
		LEFT JOIN users p2 ON m.player2_id = p2.id
		LEFT JOIN users w ON m.winner_id = w.id
		WHERE m.tournament_id = $1
		ORDER BY m.round, m.id
	`, tournamentID)

	if err == nil {
		// Заполняем информацию об игроках
		for i := range matches {
			if matches[i].Player1ID > 0 {
				matches[i].Player1 = &models.User{
					ID:       matches[i].Player1ID,
					Username: "", // Заполнится из запроса выше
					Rating:   0,  // Заполнится из запроса выше
				}
			}
			if matches[i].Player2ID > 0 {
				matches[i].Player2 = &models.User{
					ID:       matches[i].Player2ID,
					Username: "", // Заполнится из запроса выше
					Rating:   0,  // Заполнится из запроса выше
				}
			}
			if matches[i].WinnerID != nil {
				matches[i].Winner = &models.User{
					ID:       *matches[i].WinnerID,
					Username: "", // Заполнится из запроса выше
				}
			}
		}
		tournament.Matches = matches
	}

	// Получаем информацию о победителе
	if tournament.WinnerID != nil {
		var winner models.User
		err = h.DB.Get(&winner, `
			SELECT id, username, rating, wins, losses
			FROM users WHERE id = $1
		`, *tournament.WinnerID)

		if err == nil {
			tournament.Winner = &winner
		}
	}

	// Добавляем статистику прогресса
	var progress map[string]interface{}
	if len(matches) > 0 {
		totalMatches := len(matches)
		finishedMatches := 0
		for _, match := range matches {
			if match.Status == "finished" {
				finishedMatches++
			}
		}

		progressPercent := float64(finishedMatches) / float64(totalMatches) * 100

		progress = map[string]interface{}{
			"total_matches":    totalMatches,
			"finished_matches": finishedMatches,
			"progress":         progressPercent,
		}
	}

	utils.SuccessResponse(c, gin.H{
		"tournament": tournament,
		"progress":   progress,
	})
}

// GetTournaments получение списка турниров с фильтрацией
func (h *TournamentHandlers) GetTournaments(c *gin.Context) {
	var query GetTournamentsQuery
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

	// Фильтр по статусу
	if query.Status != "" {
		whereConditions = append(whereConditions, "t.status = $"+strconv.Itoa(argIndex))
		args = append(args, query.Status)
		argIndex++
	}

	// Фильтр по участнику
	if query.UserID > 0 {
		whereConditions = append(whereConditions,
			"EXISTS(SELECT 1 FROM room_participants rp WHERE rp.room_id = t.room_id AND rp.user_id = $"+strconv.Itoa(argIndex)+")")
		args = append(args, query.UserID)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + joinStrings(whereConditions, " AND ")
	}

	// Получаем общее количество
	countQuery := "SELECT COUNT(*) FROM tournaments t " + whereClause
	var total int
	err := h.DB.Get(&total, countQuery, args...)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to count tournaments")
		return
	}

	// Строим ORDER BY
	validSortFields := map[string]string{
		"created_at": "t.created_at",
		"updated_at": "t.updated_at",
		"name":       "t.name",
		"status":     "t.status",
	}

	sortField, exists := validSortFields[query.SortBy]
	if !exists {
		sortField = "t.created_at"
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
		SELECT t.id, t.room_id, t.name, t.status, t.winner_id, t.created_at, t.updated_at,
		       r.name as room_name, u.username as winner_username
		FROM tournaments t
		JOIN rooms r ON t.room_id = r.id
		LEFT JOIN users u ON t.winner_id = u.id ` +
		whereClause + " " + orderClause + " " + limitClause

	type TournamentInfo struct {
		models.Tournament
		RoomName       string `db:"room_name" json:"room_name"`
		WinnerUsername string `db:"winner_username" json:"winner_username,omitempty"`
	}

	var tournaments []TournamentInfo
	err = h.DB.Select(&tournaments, mainQuery, args...)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to fetch tournaments")
		return
	}

	pagination := utils.NewPaginationMeta(query.Page, query.PerPage, total)
	utils.PaginatedSuccessResponse(c, tournaments, pagination, "Tournaments fetched successfully")
}

// SubmitMatchResult отправка результата матча
func (h *TournamentHandlers) SubmitMatchResult(c *gin.Context) {
	tournamentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid tournament ID")
		return
	}

	matchID, err := strconv.Atoi(c.Param("match_id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid match ID")
		return
	}

	userID := c.GetInt("user_id")

	var req SubmitMatchResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Проверяем, что матч принадлежит турниру
	var match models.Match
	var roomID int
	err = h.DB.QueryRow(`
		SELECT m.id, m.tournament_id, m.round, m.player1_id, m.player2_id, m.status, t.room_id
		FROM matches m
		JOIN tournaments t ON m.tournament_id = t.id
		WHERE m.id = $1 AND m.tournament_id = $2
	`, matchID, tournamentID).Scan(&match.ID, &match.TournamentID, &match.Round,
		&match.Player1ID, &match.Player2ID, &match.Status, &roomID)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "Match not found")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	// Проверяем, что пользователь является участником матча или хостом комнаты
	var isParticipant bool
	var isHost bool

	if userID == match.Player1ID || userID == match.Player2ID {
		isParticipant = true
	}

	var hostID int
	err = h.DB.Get(&hostID, `SELECT host_id FROM rooms WHERE id = $1`, roomID)
	if err == nil && hostID == userID {
		isHost = true
	}

	if !isParticipant && !isHost {
		utils.ForbiddenResponse(c, "Only match participants or room host can submit results")
		return
	}

	if match.Status != "pending" {
		utils.BadRequestResponse(c, "Match is already completed")
		return
	}

	if req.WinnerID != match.Player1ID && req.WinnerID != match.Player2ID {
		utils.BadRequestResponse(c, "Winner must be one of the match participants")
		return
	}

	// Начинаем транзакцию
	tx, err := h.DB.Beginx()
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Обновляем результат матча
	_, err = tx.Exec(`
		UPDATE matches
		SET winner_id = $1, status = 'finished', updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, req.WinnerID, matchID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to update match")
		return
	}

	// Определяем проигравшего
	loserID := match.Player1ID
	if req.WinnerID == match.Player1ID {
		loserID = match.Player2ID
	}

	// Обновляем рейтинги игроков
	err = h.updatePlayerRatings(tx, req.WinnerID, loserID)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to update player ratings")
		return
	}

	// Продвигаем турнир (создаем следующий матч если нужно)
	err = h.advanceTournament(tx, tournamentID, matchID, req.WinnerID)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to advance tournament")
		return
	}

	// Проверяем, завершился ли турнир
	var isFinished bool
	var finalWinnerID sql.NullInt64
	err = tx.QueryRow(`
		SELECT COUNT(*) = 0, 
		       (SELECT winner_id FROM matches 
		        WHERE tournament_id = $1 AND round = (
		        	SELECT MAX(round) FROM matches WHERE tournament_id = $1
		        ) AND status = 'finished' LIMIT 1)
		FROM matches
		WHERE tournament_id = $1 AND status = 'pending'
	`, tournamentID).Scan(&isFinished, &finalWinnerID)

	if err == nil && isFinished && finalWinnerID.Valid {
		// Турнир завершен
		_, err = tx.Exec(`
			UPDATE tournaments
			SET status = 'finished', winner_id = $1, updated_at = CURRENT_TIMESTAMP
			WHERE id = $2
		`, finalWinnerID.Int64, tournamentID)

		if err != nil {
			utils.InternalErrorResponse(c, "Failed to finish tournament")
			return
		}

		// Обновляем статус комнаты
		_, err = tx.Exec(`
			UPDATE rooms SET status = 'finished', updated_at = CURRENT_TIMESTAMP
			WHERE id = $1
		`, roomID)

		if err != nil {
			utils.InternalErrorResponse(c, "Failed to update room status")
			return
		}
	}

	// Коммитим транзакцию
	if err = tx.Commit(); err != nil {
		utils.InternalErrorResponse(c, "Failed to commit transaction")
		return
	}

	// Отправляем уведомления через WebSocket
	wsMsg := models.WSMessage{
		Type: "match_result",
		Data: gin.H{
			"room_id":       roomID,
			"tournament_id": tournamentID,
			"match_id":      matchID,
			"winner_id":     req.WinnerID,
			"loser_id":      loserID,
		},
	}
	msgBytes, _ := json.Marshal(wsMsg)
	h.Hub.BroadcastToRoom(roomID, msgBytes)

	// Если турнир завершен, отправляем уведомление
	if isFinished && finalWinnerID.Valid {
		finishMsg := models.WSMessage{
			Type: "tournament_finished",
			Data: gin.H{
				"room_id":       roomID,
				"tournament_id": tournamentID,
				"winner_id":     finalWinnerID.Int64,
			},
		}
		finishMsgBytes, _ := json.Marshal(finishMsg)
		h.Hub.BroadcastToRoom(roomID, finishMsgBytes)
	}

	utils.SuccessResponse(c, gin.H{
		"message":     "Match result submitted successfully",
		"winner_id":   req.WinnerID,
		"is_finished": isFinished,
	})
}

// GetMatch получение матча по ID
func (h *TournamentHandlers) GetMatch(c *gin.Context) {
	matchID, err := strconv.Atoi(c.Param("match_id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid match ID")
		return
	}

	var match models.Match
	err = h.DB.Get(&match, `
		SELECT m.id, m.tournament_id, m.round, m.player1_id, m.player2_id, 
		       m.winner_id, m.status, m.created_at, m.updated_at
		FROM matches m
		WHERE m.id = $1
	`, matchID)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "Match not found")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	// Получаем информацию об игроках
	if match.Player1ID > 0 {
		var player1 models.User
		err = h.DB.Get(&player1, `
			SELECT id, username, rating, wins, losses
			FROM users WHERE id = $1
		`, match.Player1ID)
		if err == nil {
			match.Player1 = &player1
		}
	}

	if match.Player2ID > 0 {
		var player2 models.User
		err = h.DB.Get(&player2, `
			SELECT id, username, rating, wins, losses
			FROM users WHERE id = $1
		`, match.Player2ID)
		if err == nil {
			match.Player2 = &player2
		}
	}

	if match.WinnerID != nil {
		var winner models.User
		err = h.DB.Get(&winner, `
			SELECT id, username, rating, wins, losses
			FROM users WHERE id = $1
		`, *match.WinnerID)
		if err == nil {
			match.Winner = &winner
		}
	}

	utils.SuccessResponse(c, match)
}

// CancelTournament отмена турнира (только хост)
func (h *TournamentHandlers) CancelTournament(c *gin.Context) {
	tournamentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid tournament ID")
		return
	}

	userID := c.GetInt("user_id")

	// Проверяем, что пользователь является хостом комнаты
	var hostID int
	var roomID int
	var status string
	err = h.DB.QueryRow(`
		SELECT r.host_id, r.id, t.status
		FROM tournaments t
		JOIN rooms r ON t.room_id = r.id
		WHERE t.id = $1
	`, tournamentID).Scan(&hostID, &roomID, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "Tournament not found")
		} else {
			utils.InternalErrorResponse(c, "Database error")
		}
		return
	}

	if hostID != userID {
		utils.ForbiddenResponse(c, "Only room host can cancel tournament")
		return
	}

	if status == "finished" {
		utils.BadRequestResponse(c, "Cannot cancel finished tournament")
		return
	}

	// Начинаем транзакцию
	tx, err := h.DB.Beginx()
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Обновляем статус турнира
	_, err = tx.Exec(`
		UPDATE tournaments
		SET status = 'cancelled', updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, tournamentID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to cancel tournament")
		return
	}

	// Отмечаем все незавершенные матчи как отмененные
	_, err = tx.Exec(`
		UPDATE matches
		SET status = 'cancelled', updated_at = CURRENT_TIMESTAMP
		WHERE tournament_id = $1 AND status = 'pending'
	`, tournamentID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to cancel matches")
		return
	}

	// Возвращаем комнату в статус ожидания
	_, err = tx.Exec(`
		UPDATE rooms SET status = 'waiting', updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, roomID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to update room status")
		return
	}

	// Коммитим транзакцию
	if err = tx.Commit(); err != nil {
		utils.InternalErrorResponse(c, "Failed to commit transaction")
		return
	}

	// Отправляем уведомление через WebSocket
	wsMsg := models.WSMessage{
		Type: "tournament_cancelled",
		Data: gin.H{
			"room_id":       roomID,
			"tournament_id": tournamentID,
			"message":       "Tournament has been cancelled by host",
		},
	}
	msgBytes, _ := json.Marshal(wsMsg)
	h.Hub.BroadcastToRoom(roomID, msgBytes)

	utils.SuccessResponse(c, gin.H{
		"message": "Tournament cancelled successfully",
	})
}

// GetTournamentStats получение статистики турнира
func (h *TournamentHandlers) GetTournamentStats(c *gin.Context) {
	tournamentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid tournament ID")
		return
	}

	// Проверяем существование турнира
	var exists bool
	err = h.DB.Get(&exists, `
		SELECT EXISTS(SELECT 1 FROM tournaments WHERE id = $1)
	`, tournamentID)

	if err != nil {
		utils.InternalErrorResponse(c, "Database error")
		return
	}

	if !exists {
		utils.NotFoundResponse(c, "Tournament not found")
		return
	}

	// Собираем статистику
	type TournamentStats struct {
		TotalMatches    int                    `json:"total_matches"`
		FinishedMatches int                    `json:"finished_matches"`
		Progress        float64                `json:"progress"`
		ByRound         map[string]interface{} `json:"by_round"`
		Duration        *time.Duration         `json:"duration,omitempty"`
		AvgMatchTime    *time.Duration         `json:"avg_match_time,omitempty"`
	}

	var stats TournamentStats

	// Общая статистика матчей
	err = h.DB.QueryRow(`
		SELECT COUNT(*), COUNT(*) FILTER (WHERE status = 'finished')
		FROM matches WHERE tournament_id = $1
	`, tournamentID).Scan(&stats.TotalMatches, &stats.FinishedMatches)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to get match statistics")
		return
	}

	if stats.TotalMatches > 0 {
		stats.Progress = float64(stats.FinishedMatches) / float64(stats.TotalMatches) * 100
	}

	// Статистика по раундам
	type RoundStats struct {
		Round           int `db:"round"`
		TotalMatches    int `db:"total_matches"`
		FinishedMatches int `db:"finished_matches"`
	}

	var roundStats []RoundStats
	err = h.DB.Select(&roundStats, `
		SELECT round, 
		       COUNT(*) as total_matches,
		       COUNT(*) FILTER (WHERE status = 'finished') as finished_matches
		FROM matches 
		WHERE tournament_id = $1
		GROUP BY round
		ORDER BY round
	`, tournamentID)

	if err == nil {
		byRound := make(map[string]interface{})
		for _, rs := range roundStats {
			roundKey := "round_" + strconv.Itoa(rs.Round)
			byRound[roundKey] = map[string]interface{}{
				"total":    rs.TotalMatches,
				"finished": rs.FinishedMatches,
				"progress": float64(rs.FinishedMatches) / float64(rs.TotalMatches) * 100,
			}
		}
		stats.ByRound = byRound
	}

	// Рассчитываем длительность турнира
	var createdAt, updatedAt time.Time
	var status string
	err = h.DB.QueryRow(`
		SELECT created_at, updated_at, status FROM tournaments WHERE id = $1
	`, tournamentID).Scan(&createdAt, &updatedAt, &status)

	if err == nil {
		if status == "finished" {
			duration := updatedAt.Sub(createdAt)
			stats.Duration = &duration
		}

		// Средняя длительность матча
		if stats.FinishedMatches > 0 {
			avgDuration := updatedAt.Sub(createdAt) / time.Duration(stats.FinishedMatches)
			stats.AvgMatchTime = &avgDuration
		}
	}

	utils.SuccessResponse(c, stats, "Tournament statistics fetched successfully")
}

// Вспомогательные функции

// updatePlayerRatings обновляет рейтинги игроков после матча
func (h *TournamentHandlers) updatePlayerRatings(tx *sqlx.Tx, winnerID, loserID int) error {
	// Получаем текущие рейтинги и количество игр
	var winnerRating, loserRating, winnerGames, loserGames int

	err := tx.QueryRow(`
		SELECT rating, wins + losses FROM users WHERE id = $1
	`, winnerID).Scan(&winnerRating, &winnerGames)

	if err != nil {
		return err
	}

	err = tx.QueryRow(`
		SELECT rating, wins + losses FROM users WHERE id = $1
	`, loserID).Scan(&loserRating, &loserGames)

	if err != nil {
		return err
	}

	// Рассчитываем новые рейтинги
	winnerKFactor := rating.GetKFactor(winnerRating, winnerGames)
	loserKFactor := rating.GetKFactor(loserRating, loserGames)

	newWinnerRating, newLoserRating := rating.CalculateRatingChange(
		winnerRating, loserRating, winnerKFactor)

	// Применяем модификатор для турнирных матчей
	_, newLoserRating = rating.CalculateRatingChange(
		winnerRating, loserRating, loserKFactor)

	// Обновляем статистику победителя
	_, err = tx.Exec(`
		UPDATE users 
		SET wins = wins + 1, rating = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, newWinnerRating, winnerID)

	if err != nil {
		return err
	}

	// Обновляем статистику проигравшего
	_, err = tx.Exec(`
		UPDATE users 
		SET losses = losses + 1, rating = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, newLoserRating, loserID)

	return err
}

// advanceTournament продвигает турнир после завершения матча
func (h *TournamentHandlers) advanceTournament(tx *sqlx.Tx, tournamentID, matchID, winnerID int) error {
	// Получаем информацию о завершенном матче
	var round int
	err := tx.Get(&round, `
		SELECT round FROM matches WHERE id = $1
	`, matchID)

	if err != nil {
		return err
	}

	// Ищем следующий матч в следующем раунде где должен играть победитель
	var nextMatchID sql.NullInt64
	err = tx.QueryRow(`
		SELECT id FROM matches 
		WHERE tournament_id = $1 AND round = $2 AND status = 'pending'
		AND (player1_id = 0 OR player2_id = 0)
		ORDER BY id ASC
		LIMIT 1
	`, tournamentID, round+1).Scan(&nextMatchID)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// Если есть следующий матч, добавляем победителя
	if nextMatchID.Valid {
		// Проверяем, какое место свободно
		var player1ID, player2ID int
		err = tx.QueryRow(`
			SELECT player1_id, player2_id FROM matches WHERE id = $1
		`, nextMatchID.Int64).Scan(&player1ID, &player2ID)

		if err != nil {
			return err
		}

		if player1ID == 0 {
			_, err = tx.Exec(`
				UPDATE matches SET player1_id = $1 WHERE id = $2
			`, winnerID, nextMatchID.Int64)
		} else if player2ID == 0 {
			_, err = tx.Exec(`
				UPDATE matches SET player2_id = $1 WHERE id = $2
			`, winnerID, nextMatchID.Int64)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
