package repository

import (
	"log"
	"srd-calendar-project/backend/internal/models"
)

// Global repository instance
var repo *PostgresRepository

// Initialize sets up the repository with PostgreSQL
func Initialize() {
	repo = NewPostgresRepository()
	repo.InitializeDatabase()
}

// GetAllExercises returns all exercises from the database
func GetAllExercises() []models.Exercise {
	if repo == nil {
		log.Println("Repository not initialized, returning empty list")
		return []models.Exercise{}
	}
	return repo.GetAllExercisesDB()
}

// GetExerciseByID returns a single exercise by its ID
func GetExerciseByID(id int) (models.Exercise, bool) {
	if repo == nil {
		log.Println("Repository not initialized")
		return models.Exercise{}, false
	}
	return repo.GetExerciseByIDDB(id)
}

// CreateExercise adds a new exercise and returns it with a new ID
func CreateExercise(exercise models.Exercise) models.Exercise {
	if repo == nil {
		log.Println("Repository not initialized")
		return exercise
	}
	return repo.CreateExerciseDB(exercise)
}

// UpdateExercise updates an existing exercise
func UpdateExercise(exercise models.Exercise) bool {
	if repo == nil {
		log.Println("Repository not initialized")
		return false
	}
	return repo.UpdateExerciseDB(exercise)
}

// DeleteExercise removes an exercise by its ID
func DeleteExercise(id int) bool {
	if repo == nil {
		log.Println("Repository not initialized")
		return false
	}
	return repo.DeleteExerciseDB(id)
}