// internal/websocket/hub.go - исправленная версия
package websocket

import (
	"encoding/json"
	"log"
	"net/http"

	"zzz-tournament/internal/models"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: implement proper origin check
	},
}

type Client struct {
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan []byte
	UserID int
	RoomID int
}

type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	Rooms      map[int]map[*Client]bool // room_id -> clients
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
			h.Clients[client] = true
			log.Printf("Client registered: %d", client.UserID)

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)

				// Remove from room
				if client.RoomID > 0 {
					if room, exists := h.Rooms[client.RoomID]; exists {
						delete(room, client)
						if len(room) == 0 {
							delete(h.Rooms, client.RoomID)
						}
					}
				}
				log.Printf("Client unregistered: %d", client.UserID)
			}

		case message := <-h.Broadcast:
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}

func (h *Hub) BroadcastToRoom(roomID int, message []byte) {
	if room, exists := h.Rooms[roomID]; exists {
		for client := range room {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(h.Clients, client)
				delete(room, client)
			}
		}
	}
}

func (h *Hub) JoinRoom(client *Client, roomID int) {
	// Leave current room if any
	if client.RoomID > 0 {
		h.LeaveRoom(client)
	}

	// Join new room
	if h.Rooms[roomID] == nil {
		h.Rooms[roomID] = make(map[*Client]bool)
	}
	h.Rooms[roomID][client] = true
	client.RoomID = roomID
}

func (h *Hub) LeaveRoom(client *Client) {
	if client.RoomID > 0 {
		if room, exists := h.Rooms[client.RoomID]; exists {
			delete(room, client)
			if len(room) == 0 {
				delete(h.Rooms, client.RoomID)
			}
		}
		client.RoomID = 0
	}
}

func HandleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// TODO: Extract user ID from JWT token in query params or headers
	userID := 1 // Placeholder

	client := &Client{
		Hub:    hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		UserID: userID,
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
			break
		}

		var wsMsg models.WSMessage
		if err := json.Unmarshal(messageBytes, &wsMsg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
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

// функция handleMessage с безопасными приведениями типов
func (c *Client) handleMessage(msg models.WSMessage) {
	// Проверяем что Data не nil
	if msg.Data == nil {
		log.Printf("Received message with nil data, type: %s", msg.Type)
		return
	}

	// Безопасно приводим к map[string]interface{}
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		log.Printf("Message data is not map[string]interface{}, type: %s", msg.Type)
		return
	}

	switch msg.Type {
	case "join_room":
		roomIDRaw, exists := data["room_id"]
		if !exists || roomIDRaw == nil {
			log.Printf("join_room: room_id not found or nil")
			return
		}

		roomIDFloat, ok := roomIDRaw.(float64)
		if !ok {
			log.Printf("join_room: room_id is not a number, got %T", roomIDRaw)
			return
		}

		roomID := int(roomIDFloat)
		log.Printf("Client %d joining room %d", c.UserID, roomID)
		c.Hub.JoinRoom(c, roomID)

	case "leave_room":
		log.Printf("Client %d leaving room %d", c.UserID, c.RoomID)
		c.Hub.LeaveRoom(c)

	case "chat_message":
		roomIDRaw, roomExists := data["room_id"]
		contentRaw, contentExists := data["content"]

		if !roomExists || roomIDRaw == nil {
			log.Printf("chat_message: room_id not found or nil")
			return
		}

		if !contentExists || contentRaw == nil {
			log.Printf("chat_message: content not found or nil")
			return
		}

		roomIDFloat, ok := roomIDRaw.(float64)
		if !ok {
			log.Printf("chat_message: room_id is not a number, got %T", roomIDRaw)
			return
		}

		content, ok := contentRaw.(string)
		if !ok {
			log.Printf("chat_message: content is not a string, got %T", contentRaw)
			return
		}

		roomID := int(roomIDFloat)

		// TODO: Save message to database and broadcast to room
		response := models.WSMessage{
			Type: "chat_message",
			Data: map[string]interface{}{
				"room_id": roomID,
				"user_id": c.UserID,
				"content": content,
			},
		}

		responseBytes, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshaling chat message response: %v", err)
			return
		}

		c.Hub.BroadcastToRoom(roomID, responseBytes)

	case "heartbeat":
		// Отвечаем на heartbeat
		response := models.WSMessage{
			Type: "heartbeat",
			Data: map[string]interface{}{
				"timestamp": msg.Data,
			},
		}
		responseBytes, _ := json.Marshal(response)
		select {
		case c.Send <- responseBytes:
		default:
			log.Printf("Failed to send heartbeat response to client %d", c.UserID)
		}

	default:
		log.Printf("Unknown message type: %s from client %d", msg.Type, c.UserID)
	}
}
