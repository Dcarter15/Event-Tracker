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

// CreateDivision creates a new division for an exercise
func CreateDivision(division models.Division) models.Division {
	if repo == nil {
		log.Println("Repository not initialized")
		return division
	}
	return repo.CreateDivisionDB(division)
}

// UpdateDivision updates a division's information including learning objectives
func UpdateDivision(division models.Division) bool {
	if repo == nil {
		log.Println("Repository not initialized")
		return false
	}
	return repo.UpdateDivisionDB(division)
}

// CreateTeam creates a new team within a division
func CreateTeam(team models.Team) models.Team {
	if repo == nil {
		log.Println("Repository not initialized")
		return team
	}
	return repo.CreateTeamDB(team)
}

// GetEventsForExercise returns all events for a specific exercise
func GetEventsForExercise(exerciseID int) []models.Event {
	if repo == nil {
		log.Println("Repository not initialized")
		return []models.Event{}
	}
	return repo.GetEventsForExercise(exerciseID)
}

// CreateEvent creates a new event for an exercise
func CreateEvent(event models.Event) models.Event {
	if repo == nil {
		log.Println("Repository not initialized")
		return event
	}
	return repo.CreateEventDB(event)
}

// UpdateEvent updates an existing event
func UpdateEvent(event models.Event) bool {
	if repo == nil {
		log.Println("Repository not initialized")
		return false
	}
	return repo.UpdateEventDB(event)
}

// DeleteEvent removes an event by its ID
func DeleteEvent(id int) bool {
	if repo == nil {
		log.Println("Repository not initialized")
		return false
	}
	return repo.DeleteEventDB(id)
}

// GetExercisesByDivisionID returns exercises that contain the specified division
func GetExercisesByDivisionID(divisionID int) []models.Exercise {
	if repo == nil {
		log.Println("Repository not initialized")
		return []models.Exercise{}
	}
	return repo.GetExercisesByDivisionIDDB(divisionID)
}

// GetExercisesByTeamID returns exercises that contain the specified team
func GetExercisesByTeamID(teamID int) []models.Exercise {
	if repo == nil {
		log.Println("Repository not initialized")
		return []models.Exercise{}
	}
	return repo.GetExercisesByTeamIDDB(teamID)
}

// DeleteDivision removes a division and all its teams by ID
func DeleteDivision(id int) bool {
	if repo == nil {
		log.Println("Repository not initialized")
		return false
	}
	return repo.DeleteDivisionDB(id)
}

// DeleteTeam removes a team by ID
func DeleteTeam(id int) bool {
	if repo == nil {
		log.Println("Repository not initialized")
		return false
	}
	return repo.DeleteTeamDB(id)
}

// GetExercisesByDivisionName returns exercises that contain a division with the specified name
func GetExercisesByDivisionName(divisionName string) []models.Exercise {
	if repo == nil {
		log.Println("Repository not initialized, returning empty list")
		return []models.Exercise{}
	}
	return repo.GetExercisesByDivisionNameDB(divisionName)
}

// GetExercisesByTeamName returns exercises that contain a team with the specified name
func GetExercisesByTeamName(teamName string) []models.Exercise {
	if repo == nil {
		log.Println("Repository not initialized, returning empty list")
		return []models.Exercise{}
	}
	return repo.GetExercisesByTeamNameDB(teamName)
}