package notifications

import (
	"database/sql"
	"fmt"
	"log"
)

// NotificationService handles creating and sending notifications
type NotificationService struct {
	hub *Hub
	db  *sql.DB
}

// NewNotificationService creates a new notification service
func NewNotificationService(hub *Hub, db *sql.DB) *NotificationService {
	return &NotificationService{
		hub: hub,
		db:  db,
	}
}

// storeNotification stores a notification in the database
func (ns *NotificationService) storeNotification(notification Notification) {
	if ns.db == nil {
		log.Printf("Database not available for storing notification")
		return
	}

	query := `
		INSERT INTO activity_log (activity_type, entity_type, entity_id, entity_name, action, description, user_id, priority, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP)
	`

	_, err := ns.db.Exec(
		query,
		notification.Type,      // activity_type
		notification.Type,      // entity_type
		notification.EntityID,  // entity_id
		notification.EntityName, // entity_name
		notification.Action,    // action
		notification.Message,   // description
		notification.UserID,    // user_id
		notification.Priority,  // priority
	)

	if err != nil {
		log.Printf("Error storing notification in database: %v", err)
	}
}

// NotifyExerciseCreated sends notification for new exercise
func (ns *NotificationService) NotifyExerciseCreated(exerciseID int, exerciseName string, userID string) {
	notification := Notification{
		Type:       "exercise",
		Action:     "created",
		EntityID:   exerciseID,
		EntityName: exerciseName,
		Message:    fmt.Sprintf("🆕 New exercise '%s' created", exerciseName),
		UserID:     userID,
		Priority:   "normal",
	}

	// Store in database for persistence
	ns.storeNotification(notification)

	// Broadcast to connected clients
	ns.hub.BroadcastNotification(notification)
}

// NotifyExerciseUpdated sends notification for exercise update
func (ns *NotificationService) NotifyExerciseUpdated(exerciseID int, exerciseName string, userID string, changes string) {
	message := fmt.Sprintf("🔄 Exercise '%s' updated", exerciseName)
	if changes != "" {
		message = fmt.Sprintf("🔄 Exercise '%s' updated - %s", exerciseName, changes)
	}

	notification := Notification{
		Type:       "exercise",
		Action:     "updated",
		EntityID:   exerciseID,
		EntityName: exerciseName,
		Message:    message,
		UserID:     userID,
		Priority:   "normal",
	}

	// Store in database for persistence
	ns.storeNotification(notification)

	// Broadcast to connected clients
	ns.hub.BroadcastNotification(notification)
}

// NotifyExerciseDeleted sends notification for exercise deletion
func (ns *NotificationService) NotifyExerciseDeleted(exerciseID int, exerciseName string, userID string) {
	notification := Notification{
		Type:       "exercise",
		Action:     "deleted",
		EntityID:   exerciseID,
		EntityName: exerciseName,
		Message:    fmt.Sprintf("🗑️ Exercise '%s' deleted", exerciseName),
		UserID:     userID,
		Priority:   "normal",
	}

	// Store in database for persistence
	ns.storeNotification(notification)

	// Broadcast to connected clients
	ns.hub.BroadcastNotification(notification)
}

// NotifyEventCreated sends notification for new event
func (ns *NotificationService) NotifyEventCreated(eventID int, eventName string, exerciseName string, userID string) {
	notification := Notification{
		Type:       "event",
		Action:     "created",
		EntityID:   eventID,
		EntityName: eventName,
		Message:    fmt.Sprintf("📅 New event '%s' added to %s", eventName, exerciseName),
		UserID:     userID,
		Priority:   "normal",
	}

	// Store in database for persistence
	ns.storeNotification(notification)

	// Broadcast to connected clients
	ns.hub.BroadcastNotification(notification)
}

// NotifyEventUpdated sends notification for event update
func (ns *NotificationService) NotifyEventUpdated(eventID int, eventName string, exerciseName string, userID string) {
	notification := Notification{
		Type:       "event",
		Action:     "updated",
		EntityID:   eventID,
		EntityName: eventName,
		Message:    fmt.Sprintf("🔄 Event '%s' updated in %s", eventName, exerciseName),
		UserID:     userID,
		Priority:   "normal",
	}

	// Store in database for persistence
	ns.storeNotification(notification)

	// Broadcast to connected clients
	ns.hub.BroadcastNotification(notification)
}

// NotifyEventDeleted sends notification for event deletion
func (ns *NotificationService) NotifyEventDeleted(eventID int, eventName string, exerciseName string, userID string) {
	notification := Notification{
		Type:       "event",
		Action:     "deleted",
		EntityID:   eventID,
		EntityName: eventName,
		Message:    fmt.Sprintf("🗑️ Event '%s' deleted from %s", eventName, exerciseName),
		UserID:     userID,
		Priority:   "normal",
	}

	// Store in database for persistence
	ns.storeNotification(notification)

	// Broadcast to connected clients
	ns.hub.BroadcastNotification(notification)
}

// NotifyTaskAssigned sends notification for task assignment
func (ns *NotificationService) NotifyTaskAssigned(taskID int, taskName string, exerciseName string, divisionName string, teamName string, userID string) {
	notification := Notification{
		Type:       "task",
		Action:     "assigned",
		EntityID:   taskID,
		EntityName: taskName,
		Message:    fmt.Sprintf("📋 [%s] Task '%s' assigned to %s - %s", exerciseName, taskName, divisionName, teamName),
		UserID:     userID,
		Priority:   "normal",
	}

	// Store in database for persistence
	ns.storeNotification(notification)

	// Broadcast to connected clients
	ns.hub.BroadcastNotification(notification)
}

// NotifyTaskUnassigned sends notification for task unassignment
func (ns *NotificationService) NotifyTaskUnassigned(taskID int, taskName string, teamName string, userID string) {
	notification := Notification{
		Type:       "task",
		Action:     "unassigned",
		EntityID:   taskID,
		EntityName: taskName,
		Message:    fmt.Sprintf("📋 Task '%s' unassigned from %s", taskName, teamName),
		UserID:     userID,
		Priority:   "normal",
	}

	// Store in database for persistence
	ns.storeNotification(notification)

	// Broadcast to connected clients
	ns.hub.BroadcastNotification(notification)
}

// NotifyTeamStatusChanged sends notification for team status change
func (ns *NotificationService) NotifyTeamStatusChanged(teamID int, teamName string, exerciseName string, divisionName string, oldStatus string, newStatus string, userID string) {
	var emoji string
	priority := "normal"

	switch newStatus {
	case "red":
		emoji = "🔴"
		priority = "critical"
	case "yellow":
		emoji = "🟡"
	case "green":
		emoji = "🟢"
	default:
		emoji = "📊"
	}

	notification := Notification{
		Type:       "team",
		Action:     "status_changed",
		EntityID:   teamID,
		EntityName: teamName,
		Message:    fmt.Sprintf("%s [%s] %s - %s status changed from %s to %s", emoji, exerciseName, divisionName, teamName, oldStatus, newStatus),
		UserID:     userID,
		Priority:   priority,
	}

	// Store in database for persistence
	ns.storeNotification(notification)

	// Broadcast to connected clients
	ns.hub.BroadcastNotification(notification)
}

// BroadcastNotificationCountUpdate triggers a broadcast of updated notification counts to all connected clients
func (ns *NotificationService) BroadcastNotificationCountUpdate() {
	if ns.hub != nil {
		ns.hub.broadcastNotificationCount()
	}
}