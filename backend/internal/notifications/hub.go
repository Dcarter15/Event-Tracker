package notifications

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Notification represents a system notification
type Notification struct {
	Type      string    `json:"type"`        // "exercise", "event", "task", "team"
	Action    string    `json:"action"`      // "created", "updated", "deleted"
	EntityID  int       `json:"entity_id"`
	EntityName string   `json:"entity_name"`
	Message   string    `json:"message"`
	UserID    string    `json:"user_id"`     // For when auth is added
	Timestamp time.Time `json:"timestamp"`
	Priority  string    `json:"priority"`    // "critical", "normal", "low"
}

// Client represents a connected WebSocket client
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Database connection for querying notification count
	db *sql.DB
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin for now
		// In production, you should check the origin
		return true
	},
}

// NewHub creates a new Hub
func NewHub(db *sql.DB) *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		db:         db,
	}
}

// sendNotificationCountToClient sends the current notification count to a specific client
func (h *Hub) sendNotificationCountToClient(client *Client) {
	if h.db == nil {
		return
	}

	var count int
	err := h.db.QueryRow("SELECT COUNT(*) FROM activity_log").Scan(&count)
	if err != nil {
		log.Printf("Error getting notification count: %v", err)
		return
	}

	countMessage := map[string]interface{}{
		"type":  "notification_count",
		"count": count,
	}

	data, err := json.Marshal(countMessage)
	if err != nil {
		log.Printf("Error marshaling notification count: %v", err)
		return
	}

	select {
	case client.send <- data:
		log.Printf("Sent notification count (%d) to newly connected client", count)
	default:
		log.Printf("Failed to send notification count to client")
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client connected. Total clients: %d", len(h.clients))

			// Send current notification count to newly connected client
			go h.sendNotificationCountToClient(client)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client disconnected. Total clients: %d", len(h.clients))
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// BroadcastNotification sends a notification to all connected clients
func (h *Hub) BroadcastNotification(notification Notification) {
	notification.Timestamp = time.Now()

	data, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Error marshaling notification: %v", err)
		return
	}

	select {
	case h.broadcast <- data:
		log.Printf("Broadcasting notification: %s", notification.Message)
	default:
		log.Printf("No clients connected to receive notification")
	}

	// Also broadcast the updated notification count
	go h.broadcastNotificationCount()
}

// broadcastNotificationCount sends the current notification count to all connected clients
func (h *Hub) broadcastNotificationCount() {
	if h.db == nil {
		return
	}

	var count int
	err := h.db.QueryRow("SELECT COUNT(*) FROM activity_log").Scan(&count)
	if err != nil {
		log.Printf("Error getting notification count for broadcast: %v", err)
		return
	}

	countMessage := map[string]interface{}{
		"type":  "notification_count",
		"count": count,
	}

	data, err := json.Marshal(countMessage)
	if err != nil {
		log.Printf("Error marshaling notification count for broadcast: %v", err)
		return
	}

	select {
	case h.broadcast <- data:
		log.Printf("Broadcasting updated notification count: %d", count)
	default:
		log.Printf("No clients connected to receive notification count update")
	}
}

// HandleWebSocket handles websocket requests from clients
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
	}

	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in new goroutines
	go client.writePump()
	go client.readPump()
}

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}