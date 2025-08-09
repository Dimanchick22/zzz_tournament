// internal/handlers/handlers.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"zzz-tournament/internal/models"
	"zzz-tournament/internal/websocket"
	"zzz-tournament/pkg/auth"
	"zzz-tournament/pkg/rating"
	"zzz-tournament/pkg/tournament"
	"zzz-tournament/pkg/utils"
	"zzz-tournament/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type Handlers struct {
	DB  *sqlx.DB
	Hub *websocket.Hub
}

func New(db *sqlx.DB, hub *websocket.Hub) *Handlers {
	return &Handlers{
		DB:  db,
		Hub: hub,
	}
}

// Auth handlers
func (h *Handlers) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Валидация
	var errors validator.ValidationErrors

	if err := validator.ValidateUsername(req.Username); err != nil {
		errors = append(errors, *err)
	}

	if err := validator.ValidateEmail(req.Email); err != nil {
		errors = append(errors, *err)
	}

	if err := validator.ValidatePassword(req.Password); err != nil {
		errors = append(errors, *err)
	}

	if len(errors) > 0 {
		utils.BadRequestResponse(c, errors.Error())
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to hash password")
		return
	}

	var userID int
	err = h.DB.QueryRow(`
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id
	`, req.Username, req.Email, string(hashedPassword)).Scan(&userID)

	if err != nil {
		utils.BadRequestResponse(c, "Username or email already exists")
		return
	}

	token, err := auth.GenerateToken(userID, req.Username)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to generate token")
		return
	}

	utils.CreatedResponse(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":       userID,
			"username": req.Username,
			"email":    req.Email,
		},
	}, "User registered successfully")
}

func (h *Handlers) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	var user models.User
	err := h.DB.Get(&user, `
		SELECT id, username, email, password_hash, rating, wins, losses
		FROM users WHERE username = $1
	`, req.Username)

	if err != nil {
		utils.UnauthorizedResponse(c, "Invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.UnauthorizedResponse(c, "Invalid credentials")
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to generate token")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"token": token,
		"user":  user,
	}, "Login successful")
}

func (h *Handlers) RefreshToken(c *gin.Context) {
	utils.ErrorResponseWithDetails(c, http.StatusNotImplemented, "Not implemented")
}

// User handlers
func (h *Handlers) GetProfile(c *gin.Context) {
	userID := c.GetInt("user_id")

	var user models.User
	err := h.DB.Get(&user, `
		SELECT id, username, email, rating, wins, losses, created_at, updated_at
		FROM users WHERE id = $1
	`, userID)

	if err != nil {
		utils.NotFoundResponse(c, "User not found")
		return
	}

	utils.SuccessResponse(c, user)
}

func (h *Handlers) UpdateProfile(c *gin.Context) {
	var req struct {
		Username string `json:"username,omitempty"`
		Email    string `json:"email,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Profile updated successfully"})
}

func (h *Handlers) GetLeaderboard(c *gin.Context) {
	var users []models.User
	err := h.DB.Select(&users, `
		SELECT id, username, rating, wins, losses
		FROM users
		ORDER BY rating DESC
		LIMIT 100
	`)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to fetch leaderboard")
		return
	}

	utils.SuccessResponse(c, users)
}

// Hero handlers
func (h *Handlers) GetHeroes(c *gin.Context) {
	var heroes []models.Hero
	err := h.DB.Select(&heroes, `
		SELECT id, name, element, rarity, role, description, image_url, is_active
		FROM heroes
		WHERE is_active = true
		ORDER BY rarity DESC, name ASC
	`)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to fetch heroes")
		return
	}

	utils.SuccessResponse(c, heroes)
}

func (h *Handlers) CreateHero(c *gin.Context) {
	var req models.Hero
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	var heroID int
	err := h.DB.QueryRow(`
		INSERT INTO heroes (name, element, rarity, role, description, image_url)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, req.Name, req.Element, req.Rarity, req.Role, req.Description, req.ImageURL).Scan(&heroID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to create hero")
		return
	}

	req.ID = heroID
	utils.CreatedResponse(c, req)
}

func (h *Handlers) UpdateHero(c *gin.Context) {
	heroID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid hero ID")
		return
	}

	var req models.Hero
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	_, err = h.DB.Exec(`
		UPDATE heroes
		SET name = $1, element = $2, rarity = $3, role = $4, description = $5, image_url = $6
		WHERE id = $7
	`, req.Name, req.Element, req.Rarity, req.Role, req.Description, req.ImageURL, heroID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to update hero")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Hero updated successfully"})
}

func (h *Handlers) DeleteHero(c *gin.Context) {
	heroID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid hero ID")
		return
	}

	_, err = h.DB.Exec(`UPDATE heroes SET is_active = false WHERE id = $1`, heroID)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to delete hero")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Hero deleted successfully"})
}

// Room handlers
func (h *Handlers) GetRooms(c *gin.Context) {
	var rooms []models.Room
	err := h.DB.Select(&rooms, `
		SELECT r.id, r.name, r.description, r.host_id, r.max_players, r.current_count, r.status, r.is_private, r.created_at, r.updated_at
		FROM rooms r
		WHERE r.status = 'waiting'
		ORDER BY r.created_at DESC
	`)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to fetch rooms")
		return
	}

	// Fetch participants for each room
	for i := range rooms {
		var participants []models.User
		err := h.DB.Select(&participants, `
			SELECT u.id, u.username, u.rating
			FROM users u
			JOIN room_participants rp ON u.id = rp.user_id
			WHERE rp.room_id = $1
		`, rooms[i].ID)

		if err == nil {
			rooms[i].Participants = participants
		}
	}

	utils.SuccessResponse(c, rooms)
}

func (h *Handlers) CreateRoom(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req struct {
		Name        string `json:"name" binding:"required,min=3,max=255"`
		Description string `json:"description"`
		MaxPlayers  int    `json:"max_players" binding:"required,min=2,max=16"`
		IsPrivate   bool   `json:"is_private"`
		Password    string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

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

	// Add host as participant
	_, err = tx.Exec(`
		INSERT INTO room_participants (room_id, user_id)
		VALUES ($1, $2)
	`, roomID, userID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to add host to room")
		return
	}

	if err = tx.Commit(); err != nil {
		utils.InternalErrorResponse(c, "Failed to commit transaction")
		return
	}

	utils.CreatedResponse(c, gin.H{
		"id":      roomID,
		"message": "Room created successfully",
	})
}

func (h *Handlers) GetRoom(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid room ID")
		return
	}

	var room models.Room
	err = h.DB.Get(&room, `
		SELECT id, name, description, host_id, max_players, current_count, status, is_private, created_at, updated_at
		FROM rooms WHERE id = $1
	`, roomID)

	if err != nil {
		utils.NotFoundResponse(c, "Room not found")
		return
	}

	// Fetch participants
	var participants []models.User
	err = h.DB.Select(&participants, `
		SELECT u.id, u.username, u.rating
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

func (h *Handlers) UpdateRoom(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid room ID")
		return
	}

	userID := c.GetInt("user_id")

	// Check if user is room host
	var hostID int
	err = h.DB.Get(&hostID, `SELECT host_id FROM rooms WHERE id = $1`, roomID)
	if err != nil {
		utils.NotFoundResponse(c, "Room not found")
		return
	}

	if hostID != userID {
		utils.ForbiddenResponse(c, "Only room host can update room")
		return
	}

	var req struct {
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
		MaxPlayers  int    `json:"max_players,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Room updated successfully"})
}

func (h *Handlers) DeleteRoom(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid room ID")
		return
	}

	userID := c.GetInt("user_id")

	// Check if user is room host
	var hostID int
	err = h.DB.Get(&hostID, `SELECT host_id FROM rooms WHERE id = $1`, roomID)
	if err != nil {
		utils.NotFoundResponse(c, "Room not found")
		return
	}

	if hostID != userID {
		utils.ForbiddenResponse(c, "Only room host can delete room")
		return
	}

	_, err = h.DB.Exec(`DELETE FROM rooms WHERE id = $1`, roomID)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to delete room")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Room deleted successfully"})
}

func (h *Handlers) JoinRoom(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid room ID")
		return
	}

	userID := c.GetInt("user_id")

	var req struct {
		Password string `json:"password"`
	}
	c.ShouldBindJSON(&req)

	// Check if room exists and get details
	var room models.Room
	err = h.DB.Get(&room, `
		SELECT id, max_players, current_count, status, is_private, password
		FROM rooms WHERE id = $1
	`, roomID)

	if err != nil {
		utils.NotFoundResponse(c, "Room not found")
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

	// Check if user is already in room
	var exists bool
	err = h.DB.Get(&exists, `
		SELECT EXISTS(SELECT 1 FROM room_participants WHERE room_id = $1 AND user_id = $2)
	`, roomID, userID)

	if err != nil || exists {
		utils.BadRequestResponse(c, "Already in room")
		return
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Add user to room
	_, err = tx.Exec(`
		INSERT INTO room_participants (room_id, user_id)
		VALUES ($1, $2)
	`, roomID, userID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to join room")
		return
	}

	// Update room count
	_, err = tx.Exec(`
		UPDATE rooms SET current_count = current_count + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, roomID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to update room count")
		return
	}

	if err = tx.Commit(); err != nil {
		utils.InternalErrorResponse(c, "Failed to commit transaction")
		return
	}

	// Broadcast room update via WebSocket
	wsMsg := models.WSMessage{
		Type: "room_updated",
		Data: gin.H{
			"room_id": roomID,
			"action":  "user_joined",
			"user_id": userID,
		},
	}
	msgBytes, _ := json.Marshal(wsMsg)
	h.Hub.BroadcastToRoom(roomID, msgBytes)

	utils.SuccessResponse(c, gin.H{"message": "Joined room successfully"})
}

func (h *Handlers) LeaveRoom(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid room ID")
		return
	}

	userID := c.GetInt("user_id")

	tx, err := h.DB.Beginx()
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Remove user from room
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

	// Update room count
	_, err = tx.Exec(`
		UPDATE rooms SET current_count = current_count - 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, roomID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to update room count")
		return
	}

	// Check if room is empty and delete if necessary
	var count int
	err = tx.Get(&count, `SELECT current_count FROM rooms WHERE id = $1`, roomID)
	if err == nil && count == 0 {
		_, err = tx.Exec(`DELETE FROM rooms WHERE id = $1`, roomID)
	}

	if err = tx.Commit(); err != nil {
		utils.InternalErrorResponse(c, "Failed to commit transaction")
		return
	}

	// Broadcast room update via WebSocket
	wsMsg := models.WSMessage{
		Type: "room_updated",
		Data: gin.H{
			"room_id": roomID,
			"action":  "user_left",
			"user_id": userID,
		},
	}
	msgBytes, _ := json.Marshal(wsMsg)
	h.Hub.BroadcastToRoom(roomID, msgBytes)

	utils.SuccessResponse(c, gin.H{"message": "Left room successfully"})
}

// Tournament handlers
func (h *Handlers) StartTournament(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid room ID")
		return
	}

	userID := c.GetInt("user_id")

	// Check if user is room host
	var hostID int
	err = h.DB.Get(&hostID, `SELECT host_id FROM rooms WHERE id = $1`, roomID)
	if err != nil {
		utils.NotFoundResponse(c, "Room not found")
		return
	}

	if hostID != userID {
		utils.ForbiddenResponse(c, "Only room host can start tournament")
		return
	}

	// Get room participants
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

	// Check if tournament already exists
	var existingTournament bool
	err = h.DB.Get(&existingTournament, `
		SELECT EXISTS(SELECT 1 FROM tournaments WHERE room_id = $1)
	`, roomID)

	if err == nil && existingTournament {
		utils.BadRequestResponse(c, "Tournament already exists for this room")
		return
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Create tournament
	var tournamentID int
	err = tx.QueryRow(`
		INSERT INTO tournaments (room_id, name, status)
		VALUES ($1, $2, 'started')
		RETURNING id
	`, roomID, "Tournament for Room "+strconv.Itoa(roomID)).Scan(&tournamentID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to create tournament")
		return
	}

	// Convert models.User to tournament.Player
	players := make([]tournament.Player, len(participants))
	for i, p := range participants {
		players[i] = tournament.Player{
			ID:       p.ID,
			Username: p.Username,
			Rating:   p.Rating,
		}
	}

	// Generate bracket and matches
	bracket, err := tournament.GenerateBracket(players)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to generate tournament bracket")
		return
	}

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

	// Update room status
	_, err = tx.Exec(`
		UPDATE rooms SET status = 'in_progress', updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, roomID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to update room status")
		return
	}

	if err = tx.Commit(); err != nil {
		utils.InternalErrorResponse(c, "Failed to commit transaction")
		return
	}

	// Broadcast tournament start via WebSocket
	wsMsg := models.WSMessage{
		Type: "tournament_started",
		Data: gin.H{
			"room_id":       roomID,
			"tournament_id": tournamentID,
		},
	}
	msgBytes, _ := json.Marshal(wsMsg)
	h.Hub.BroadcastToRoom(roomID, msgBytes)

	utils.SuccessResponse(c, gin.H{
		"tournament_id": tournamentID,
		"message":       "Tournament started successfully",
	})
}

func (h *Handlers) GetTournament(c *gin.Context) {
	tournamentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid tournament ID")
		return
	}

	var tournament models.Tournament
	err = h.DB.Get(&tournament, `
		SELECT id, room_id, name, status, winner_id, created_at, updated_at
		FROM tournaments WHERE id = $1
	`, tournamentID)

	if err != nil {
		utils.NotFoundResponse(c, "Tournament not found")
		return
	}

	// Get matches
	var matches []models.Match
	err = h.DB.Select(&matches, `
		SELECT m.id, m.tournament_id, m.round, m.player1_id, m.player2_id, m.winner_id, m.status, m.created_at, m.updated_at
		FROM matches m
		WHERE m.tournament_id = $1
		ORDER BY m.round, m.id
	`, tournamentID)

	if err == nil {
		tournament.Matches = matches
	}

	utils.SuccessResponse(c, tournament)
}

func (h *Handlers) SubmitMatchResult(c *gin.Context) {
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

	var req struct {
		WinnerID int `json:"winner_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Verify match belongs to tournament
	var match models.Match
	err = h.DB.Get(&match, `
		SELECT id, tournament_id, round, player1_id, player2_id, status
		FROM matches
		WHERE id = $1 AND tournament_id = $2
	`, matchID, tournamentID)

	if err != nil {
		utils.NotFoundResponse(c, "Match not found")
		return
	}

	if match.Status != "pending" {
		utils.BadRequestResponse(c, "Match already completed")
		return
	}

	if req.WinnerID != match.Player1ID && req.WinnerID != match.Player2ID {
		utils.BadRequestResponse(c, "Winner must be one of the match participants")
		return
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Update match result
	_, err = tx.Exec(`
		UPDATE matches
		SET winner_id = $1, status = 'finished', updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, req.WinnerID, matchID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to update match")
		return
	}

	// Update user stats
	loserID := match.Player1ID
	if req.WinnerID == match.Player1ID {
		loserID = match.Player2ID
	}

	// Get current ratings
	var winnerRating, loserRating, winnerGames, loserGames int

	err = tx.Get(&winnerRating, `SELECT rating FROM users WHERE id = $1`, req.WinnerID)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to get winner rating")
		return
	}

	err = tx.Get(&loserRating, `SELECT rating FROM users WHERE id = $1`, loserID)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to get loser rating")
		return
	}

	err = tx.Get(&winnerGames, `SELECT wins + losses FROM users WHERE id = $1`, req.WinnerID)
	if err != nil {
		winnerGames = 0
	}

	err = tx.Get(&loserGames, `SELECT wins + losses FROM users WHERE id = $1`, loserID)
	if err != nil {
		loserGames = 0
	}

	// Calculate new ratings
	winnerKFactor := rating.GetKFactor(winnerRating, winnerGames)
	loserKFactor := rating.GetKFactor(loserRating, loserGames)

	newWinnerRating, _ := rating.CalculateRatingChange(winnerRating, loserRating, winnerKFactor)
	_, newLoserRating := rating.CalculateRatingChange(winnerRating, loserRating, loserKFactor)

	_, err = tx.Exec(`UPDATE users SET wins = wins + 1, rating = $1 WHERE id = $2`, newWinnerRating, req.WinnerID)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to update winner stats")
		return
	}

	_, err = tx.Exec(`UPDATE users SET losses = losses + 1, rating = $1 WHERE id = $2`, newLoserRating, loserID)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to update loser stats")
		return
	}

	// Check if tournament is finished
	var pendingMatches int
	err = tx.Get(&pendingMatches, `
		SELECT COUNT(*) FROM matches
		WHERE tournament_id = $1 AND status = 'pending'
	`, tournamentID)

	if err == nil && pendingMatches == 0 {
		// Tournament finished
		var finalWinner int
		err = tx.Get(&finalWinner, `
			SELECT winner_id FROM matches
			WHERE tournament_id = $1 AND round = (
				SELECT MAX(round) FROM matches WHERE tournament_id = $1
			)
		`, tournamentID)

		if err == nil {
			_, err = tx.Exec(`
				UPDATE tournaments
				SET status = 'finished', winner_id = $1, updated_at = CURRENT_TIMESTAMP
				WHERE id = $2
			`, finalWinner, tournamentID)
		}
	}

	if err = tx.Commit(); err != nil {
		utils.InternalErrorResponse(c, "Failed to commit transaction")
		return
	}

	utils.SuccessResponse(c, nil, "Match result submitted successfully")
}

// Chat handlers
func (h *Handlers) GetRoomMessages(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid room ID")
		return
	}

	var messages []models.Message
	err = h.DB.Select(&messages, `
		SELECT m.id, m.room_id, m.user_id, m.content, m.type, m.created_at
		FROM messages m
		WHERE m.room_id = $1
		ORDER BY m.created_at DESC
		LIMIT 50
	`, roomID)

	if err != nil {
		utils.InternalErrorResponse(c, "Failed to fetch messages")
		return
	}

	utils.SuccessResponse(c, messages)
}
