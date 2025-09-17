package main

import (
	"log"
	"net/http"
	"srd-calendar-project/backend/internal/database"
	"srd-calendar-project/backend/internal/handlers"
	"srd-calendar-project/backend/internal/repository"

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

	// Chatbot endpoint
	r.Post("/api/chatbot", handlers.EnhancedChatbotHandler)

	log.Println("Starting server on :8081")
	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}
