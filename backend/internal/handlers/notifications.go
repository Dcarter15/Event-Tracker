package handlers

import (
	"encoding/json"
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

// GetNotifications returns all notifications with optional pagination
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

	// Query notifications from database
	query := `
		SELECT id, activity_type, action, entity_id, entity_name, description, user_id, priority, created_at
		FROM activity_log
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := database.DB.Query(query, limit, offset)
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

// GetNotificationCount returns the total count of notifications
func GetNotificationCount(w http.ResponseWriter, r *http.Request) {
	query := "SELECT COUNT(*) FROM activity_log"

	var count int
	err := database.DB.QueryRow(query).Scan(&count)
	if err != nil {
		http.Error(w, "Error fetching notification count", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"count": count})
}