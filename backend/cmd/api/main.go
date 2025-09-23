package main

import (
	"log"
	"net/http"
	"srd-calendar-project/backend/internal/database"
	"srd-calendar-project/backend/internal/handlers"
	"srd-calendar-project/backend/internal/repository"
	"srd-calendar-project/backend/internal/notifications"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Initialize database connection
	err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	// Initialize repository with database
	repository.Initialize()

	// Initialize WebSocket notification hub
	notificationHub := notifications.NewHub(database.DB)
	go notificationHub.Run()

	// Initialize notification service with database access
	notificationService := notifications.NewNotificationService(notificationHub, database.DB)

	// Make notification service available to handlers
	handlers.SetNotificationService(notificationService)

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Public routes
	r.Get("/api/exercises", handlers.GetExercises)
	r.Post("/api/exercises", handlers.CreateExerciseHandler)
	r.Put("/api/exercises/{id}", handlers.UpdateExerciseHandler)
	r.Delete("/api/exercises/{id}", handlers.DeleteExerciseHandler)

	r.Get("/api/divisions", handlers.GetDivisionsForExercise)
	r.Post("/api/divisions", handlers.CreateDivision)
	r.Put("/api/divisions/update", handlers.UpdateDivision)
	r.Post("/api/teams", handlers.CreateTeam)
	r.Put("/api/team/update", handlers.UpdateTeam)

	// Event endpoints
	r.Get("/api/events", handlers.GetEvents)
	r.Post("/api/events", handlers.CreateEvent)
	r.Put("/api/events/{id}", handlers.UpdateEvent)
	r.Delete("/api/events/{id}", handlers.DeleteEvent)

	// Task endpoints
	r.Get("/api/tasks", handlers.GetTasks)
	r.Post("/api/tasks", handlers.CreateTask)
	r.Put("/api/tasks/{id}", handlers.UpdateTask)
	r.Put("/api/tasks/{id}/assign", handlers.AssignTaskToTeam)
	r.Put("/api/tasks/{id}/assign-multiple", handlers.AssignTaskToMultipleTeams)
	r.Delete("/api/tasks/{id}", handlers.DeleteTask)

	// Chatbot endpoint
	r.Post("/api/chatbot", handlers.EnhancedChatbotHandler)

	// Notification endpoints
	r.Get("/api/notifications", handlers.GetNotifications)
	r.Get("/api/notifications/read", handlers.GetReadNotifications)
	r.Get("/api/notifications/count", handlers.GetNotificationCount)
	r.Post("/api/notifications/clear", handlers.ClearNotifications)
	r.Post("/api/notifications/mark-read", handlers.MarkNotificationAsRead)

	// WebSocket endpoint
	r.Get("/ws", notificationHub.HandleWebSocket)

	log.Println("Starting server on :8081")
	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}
