// internal/websocket/hub.go - исправленная версия
package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"zzz-tournament/internal/models"
	"zzz-tournament/pkg/auth"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		// В продакшене нужно проверять origins
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://127.0.0.1:3000",
		}

		for _, allowed := range allowedOrigins {
			if origin == allowed {
				return true
			}
		}

		// В режиме разработки разрешаем localhost
		return strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1")
	},
}

type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	UserID   int
	Username string
	RoomID   int
	mu       sync.RWMutex // Защита от race conditions
}

type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	Rooms      map[int]map[*Client]bool // room_id -> clients
	mu         sync.RWMutex             // Защита от race conditions
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Rooms:      make(map[int]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.mu.Unlock()
			log.Printf("Client registered: %d (%s)", client.UserID, client.Username)

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)

				// Remove from room safely
				if client.RoomID > 0 {
					if room, exists := h.Rooms[client.RoomID]; exists {
						delete(room, client)
						if len(room) == 0 {
							delete(h.Rooms, client.RoomID)
						}
					}
				}
				log.Printf("Client unregistered: %d (%s)", client.UserID, client.Username)
			}
			h.mu.Unlock()

		case message := <-h.Broadcast:
			h.mu.RLock()
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) BroadcastToRoom(roomID int, message []byte) {
	h.mu.RLock()
	room, exists := h.Rooms[roomID]
	if !exists {
		h.mu.RUnlock()
		return
	}

	// Создаем копию клиентов для безопасной итерации
	clients := make([]*Client, 0, len(room))
	for client := range room {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	// Отправляем сообщения
	for _, client := range clients {
		select {
		case client.Send <- message:
		default:
			// Клиент недоступен, удаляем его
			h.removeClientFromRoom(client, roomID)
		}
	}
}

func (h *Hub) removeClientFromRoom(client *Client, roomID int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, exists := h.Rooms[roomID]; exists {
		delete(room, client)
		if len(room) == 0 {
			delete(h.Rooms, roomID)
		}
	}

	delete(h.Clients, client)
	close(client.Send)
}

func (h *Hub) JoinRoom(client *Client, roomID int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Leave current room if any
	if client.RoomID > 0 {
		if room, exists := h.Rooms[client.RoomID]; exists {
			delete(room, client)
			if len(room) == 0 {
				delete(h.Rooms, client.RoomID)
			}
		}
	}

	// Join new room
	if h.Rooms[roomID] == nil {
		h.Rooms[roomID] = make(map[*Client]bool)
	}
	h.Rooms[roomID][client] = true

	client.mu.Lock()
	client.RoomID = roomID
	client.mu.Unlock()

	log.Printf("Client %d joined room %d", client.UserID, roomID)
}

func (h *Hub) LeaveRoom(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client.mu.RLock()
	roomID := client.RoomID
	client.mu.RUnlock()

	if roomID > 0 {
		if room, exists := h.Rooms[roomID]; exists {
			delete(room, client)
			if len(room) == 0 {
				delete(h.Rooms, roomID)
			}
		}

		client.mu.Lock()
		client.RoomID = 0
		client.mu.Unlock()

		log.Printf("Client %d left room %d", client.UserID, roomID)
	}
}

func HandleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Извлекаем JWT токен из query параметров или заголовков
	token := r.URL.Query().Get("token")
	if token == "" {
		token = r.Header.Get("Authorization")
		if strings.HasPrefix(token, "Bearer ") {
			token = strings.TrimPrefix(token, "Bearer ")
		}
	}

	if token == "" {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Валидируем токен
	userID, username, err := auth.GetUserFromToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		Hub:      hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		UserID:   userID,
		Username: username,
	}

	client.Hub.Register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var wsMsg models.WSMessage
		if err := json.Unmarshal(messageBytes, &wsMsg); err != nil {
			log.Printf("Error unmarshaling message from user %d: %v", c.UserID, err)
			continue
		}

		c.handleMessage(wsMsg)
	}
}

func (c *Client) writePump() {
	defer c.Conn.Close()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

// Безопасная обработка сообщений с правильными типами
func (c *Client) handleMessage(msg models.WSMessage) {
	switch msg.Type {
	case "join_room":
		c.handleJoinRoom(msg.Data)
	case "leave_room":
		c.handleLeaveRoom()
	case "chat_message":
		c.handleChatMessage(msg.Data)
	case "heartbeat":
		c.handleHeartbeat(msg.Data)
	default:
		log.Printf("Unknown message type: %s from client %d", msg.Type, c.UserID)
	}
}

func (c *Client) handleJoinRoom(data interface{}) {
	roomID, ok := c.extractRoomID(data)
	if !ok {
		return
	}

	// TODO: Проверить, может ли пользователь присоединиться к комнате
	// (существует ли комната, не забанен ли пользователь и т.д.)

	c.Hub.JoinRoom(c, roomID)

	// Отправляем подтверждение
	response := models.WSMessage{
		Type: "room_joined",
		Data: map[string]interface{}{
			"room_id": roomID,
			"success": true,
		},
	}
	c.sendMessage(response)
}

func (c *Client) handleLeaveRoom() {
	c.Hub.LeaveRoom(c)

	response := models.WSMessage{
		Type: "room_left",
		Data: map[string]interface{}{
			"success": true,
		},
	}
	c.sendMessage(response)
}

func (c *Client) handleChatMessage(data interface{}) {
	roomID, content, ok := c.extractChatData(data)
	if !ok {
		return
	}

	// Проверяем, что клиент находится в этой комнате
	c.mu.RLock()
	clientRoomID := c.RoomID
	c.mu.RUnlock()

	if clientRoomID != roomID {
		log.Printf("Client %d tried to send message to room %d but is in room %d",
			c.UserID, roomID, clientRoomID)
		return
	}

	// TODO: Сохранить сообщение в базу данных
	// TODO: Проверить, не заглушен ли пользователь

	response := models.WSMessage{
		Type: "chat_message",
		Data: map[string]interface{}{
			"room_id":   roomID,
			"user_id":   c.UserID,
			"username":  c.Username,
			"content":   content,
			"timestamp": json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
		},
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling chat message response: %v", err)
		return
	}

	c.Hub.BroadcastToRoom(roomID, responseBytes)
}

func (c *Client) handleHeartbeat(data interface{}) {
	response := models.WSMessage{
		Type: "heartbeat",
		Data: map[string]interface{}{
			"timestamp":   data,
			"server_time": time.Now().Unix(),
		},
	}
	c.sendMessage(response)
}

// Вспомогательные функции для безопасного извлечения данных
func (c *Client) extractRoomID(data interface{}) (int, bool) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		log.Printf("join_room: data is not map[string]interface{}, got %T", data)
		return 0, false
	}

	roomIDRaw, exists := dataMap["room_id"]
	if !exists || roomIDRaw == nil {
		log.Printf("join_room: room_id not found or nil")
		return 0, false
	}

	// JSON numbers приходят как float64
	roomIDFloat, ok := roomIDRaw.(float64)
	if !ok {
		log.Printf("join_room: room_id is not a number, got %T", roomIDRaw)
		return 0, false
	}

	return int(roomIDFloat), true
}

func (c *Client) extractChatData(data interface{}) (int, string, bool) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		log.Printf("chat_message: data is not map[string]interface{}, got %T", data)
		return 0, "", false
	}

	roomIDRaw, roomExists := dataMap["room_id"]
	contentRaw, contentExists := dataMap["content"]

	if !roomExists || roomIDRaw == nil {
		log.Printf("chat_message: room_id not found or nil")
		return 0, "", false
	}

	if !contentExists || contentRaw == nil {
		log.Printf("chat_message: content not found or nil")
		return 0, "", false
	}

	roomIDFloat, ok := roomIDRaw.(float64)
	if !ok {
		log.Printf("chat_message: room_id is not a number, got %T", roomIDRaw)
		return 0, "", false
	}

	content, ok := contentRaw.(string)
	if !ok {
		log.Printf("chat_message: content is not a string, got %T", contentRaw)
		return 0, "", false
	}

	// Валидация содержимого
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		log.Printf("chat_message: empty content")
		return 0, "", false
	}

	if len(content) > 1000 {
		log.Printf("chat_message: content too long")
		return 0, "", false
	}

	return int(roomIDFloat), content, true
}

func (c *Client) sendMessage(msg models.WSMessage) {
	responseBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	select {
	case c.Send <- responseBytes:
	default:
		log.Printf("Failed to send message to client %d", c.UserID)
	}
}
