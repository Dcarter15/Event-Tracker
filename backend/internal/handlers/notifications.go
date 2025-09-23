package handlers

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	"srd-calendar-project/backend/internal/database"
	"srd-calendar-project/backend/internal/notifications"
)

var notificationService *notifications.NotificationService

// SetNotificationService sets the notification service instance
func SetNotificationService(service *notifications.NotificationService) {
	notificationService = service
}

// GetNotificationService returns the notification service instance
func GetNotificationService() *notifications.NotificationService {
	return notificationService
}

// NotificationResponse represents a notification in API responses
type NotificationResponse struct {
	ID          int    `json:"id"`
	Type        string `json:"type"`
	Action      string `json:"action"`
	EntityID    int    `json:"entity_id"`
	EntityName  string `json:"entity_name"`
	Message     string `json:"message"`
	UserID      string `json:"user_id"`
	Priority    string `json:"priority"`
	CreatedAt   string `json:"created_at"`
}

// getUserID extracts user ID from request - for now using a simple session approach
func getUserID(r *http.Request) string {
	// Check for session ID in header first (sent from frontend)
	if sessionID := r.Header.Get("X-Session-ID"); sessionID != "" {
		return sessionID
	}

	// Check for session ID in query parameter (for WebSocket connections)
	if sessionID := r.URL.Query().Get("sessionId"); sessionID != "" {
		return sessionID
	}

	// Fallback to IP-based identification for WebSocket connections
	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = forwarded
	}

	// Create a simple hash-based user ID but make it more stable
	// Use only IP for now to reduce variability
	return fmt.Sprintf("user_%x", md5.Sum([]byte(clientIP)))[:16]
}

// GetNotifications returns notifications not read by the current user with optional pagination
func GetNotifications(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Set defaults
	limit := 50 // Default limit
	offset := 0 // Default offset

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	userID := getUserID(r)

	// Query notifications that haven't been read by this user
	query := `
		SELECT al.id, al.activity_type, al.action, al.entity_id, al.entity_name, al.description, al.user_id, al.priority, al.created_at
		FROM activity_log al
		LEFT JOIN user_notifications un ON al.id = un.notification_id AND un.user_id = $1
		WHERE un.notification_id IS NULL
		ORDER BY al.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := database.DB.Query(query, userID, limit, offset)
	if err != nil {
		http.Error(w, "Error fetching notifications", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var notifications []NotificationResponse
	for rows.Next() {
		var n NotificationResponse
		var createdAt interface{}

		err := rows.Scan(&n.ID, &n.Type, &n.Action, &n.EntityID, &n.EntityName, &n.Message, &n.UserID, &n.Priority, &createdAt)
		if err != nil {
			continue
		}

		// Format timestamp
		if createdAt != nil {
			if t, ok := createdAt.(time.Time); ok {
				n.CreatedAt = t.Format(time.RFC3339)
			}
		}

		notifications = append(notifications, n)
	}

	if notifications == nil {
		notifications = []NotificationResponse{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}

// GetNotificationCount returns the count of unread notifications for the current user
func GetNotificationCount(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	// Count notifications that haven't been read by this user
	query := `
		SELECT COUNT(*)
		FROM activity_log al
		LEFT JOIN user_notifications un ON al.id = un.notification_id AND un.user_id = $1
		WHERE un.notification_id IS NULL
	`

	var count int
	err := database.DB.QueryRow(query, userID).Scan(&count)
	if err != nil {
		http.Error(w, "Error fetching notification count", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"count": count})
}

// ClearNotifications marks all current notifications as read for the current user
func ClearNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserID(r)
	log.Printf("ClearNotifications: userID = %s", userID)

	// Insert records for all current notifications that this user hasn't read yet
	query := `
		INSERT INTO user_notifications (user_id, notification_id, read_at)
		SELECT $1, al.id, CURRENT_TIMESTAMP
		FROM activity_log al
		LEFT JOIN user_notifications un ON al.id = un.notification_id AND un.user_id = $1
		WHERE un.notification_id IS NULL
		ON CONFLICT (user_id, notification_id) DO NOTHING
	`

	result, err := database.DB.Exec(query, userID)
	if err != nil {
		http.Error(w, "Error clearing notifications", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()

	// Trigger notification count update via WebSocket
	if notificationService != nil {
		notificationService.BroadcastNotificationCountUpdate()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"cleared": rowsAffected,
	})
}

// MarkNotificationAsRead marks a single notification as read for the current user
func MarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get notification ID from request body
	var request struct {
		NotificationID int `json:"notification_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := getUserID(r)

	// Insert record for this notification that this user has read
	query := `
		INSERT INTO user_notifications (user_id, notification_id, read_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, notification_id) DO NOTHING
	`

	result, err := database.DB.Exec(query, userID, request.NotificationID)
	if err != nil {
		http.Error(w, "Error marking notification as read", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()

	// Trigger notification count update via WebSocket
	if notificationService != nil {
		notificationService.BroadcastNotificationCountUpdate()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"marked":  rowsAffected > 0,
	})
}

// GetReadNotifications returns notifications that have been read by the current user with optional pagination
func GetReadNotifications(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Set defaults
	limit := 50 // Default limit
	offset := 0 // Default offset

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	userID := getUserID(r)

	// Query notifications that have been read by this user
	query := `
		SELECT al.id, al.activity_type, al.action, al.entity_id, al.entity_name, al.description, al.user_id, al.priority, al.created_at
		FROM activity_log al
		INNER JOIN user_notifications un ON al.id = un.notification_id AND un.user_id = $1
		ORDER BY un.read_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := database.DB.Query(query, userID, limit, offset)
	if err != nil {
		http.Error(w, "Error fetching read notifications", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var notifications []NotificationResponse
	for rows.Next() {
		var n NotificationResponse
		var createdAt interface{}

		err := rows.Scan(&n.ID, &n.Type, &n.Action, &n.EntityID, &n.EntityName, &n.Message, &n.UserID, &n.Priority, &createdAt)
		if err != nil {
			continue
		}

		// Format timestamp
		if createdAt != nil {
			if t, ok := createdAt.(time.Time); ok {
				n.CreatedAt = t.Format(time.RFC3339)
			}
		}

		notifications = append(notifications, n)
	}

	if notifications == nil {
		notifications = []NotificationResponse{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}