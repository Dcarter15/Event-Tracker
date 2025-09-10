package main

import (
	"log"
	"net/http"
	"srd-calendar-project/backend/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
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
	r.Put("/api/team/update", handlers.UpdateTeam)

	// Chatbot endpoint
	r.Post("/api/chatbot", handlers.ChatbotHandler)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}
