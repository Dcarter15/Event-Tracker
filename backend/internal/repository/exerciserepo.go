package repository

import (
	"srd-calendar-project/backend/internal/models"
	"sync"
	"time"
)

// In-memory store for exercises
var (
	exercises = make(map[int]models.Exercise)
	nextID    = 1
	mu        sync.Mutex // Mutex to protect access to exercises map and nextID
)

func init() {
	// Initialize with some dummy data
	mu.Lock()
	defer mu.Unlock()
	exercises[nextID] = models.Exercise{ID: nextID, Name: "Initial Exercise 1", StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 7)}
	nextID++
	exercises[nextID] = models.Exercise{ID: nextID, Name: "Initial Exercise 2", StartDate: time.Now().AddDate(0, 0, 10), EndDate: time.Now().AddDate(0, 0, 17)}
	nextID++
}

// GetAllExercises returns all exercises from the in-memory store.
func GetAllExercises() []models.Exercise {
	mu.Lock()
	defer mu.Unlock()
	list := make([]models.Exercise, 0, len(exercises))
	for _, ex := range exercises {
		list = append(list, ex)
	}
	return list
}

// GetExerciseByID returns a single exercise by its ID.
func GetExerciseByID(id int) (models.Exercise, bool) {
	mu.Lock()
	defer mu.Unlock()
	ex, ok := exercises[id]
	return ex, ok
}

// CreateExercise adds a new exercise to the store and returns it with a new ID.
func CreateExercise(exercise models.Exercise) models.Exercise {
	mu.Lock()
	defer mu.Unlock()
	exercise.ID = nextID
	exercises[nextID] = exercise
	nextID++
	return exercise
}

// UpdateExercise updates an existing exercise. Returns true if updated, false if not found.
func UpdateExercise(exercise models.Exercise) bool {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := exercises[exercise.ID]; ok {
		exercises[exercise.ID] = exercise
		return true
	}
	return false
}

// DeleteExercise removes an exercise by its ID. Returns true if deleted, false if not found.
func DeleteExercise(id int) bool {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := exercises[id]; ok {
		delete(exercises, id)
		return true
	}
	return false
}
