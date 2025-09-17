package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"srd-calendar-project/backend/internal/models"
	"srd-calendar-project/backend/internal/repository"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

// GetExercises returns all exercises from the repository.
func GetExercises(w http.ResponseWriter, r *http.Request) {
	exercises := repository.GetAllExercises()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exercises)
}

// CreateExerciseHandler creates a new exercise.
func CreateExerciseHandler(w http.ResponseWriter, r *http.Request) {
	var exercise models.Exercise
	err := json.NewDecoder(r.Body).Decode(&exercise)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdExercise := repository.CreateExercise(exercise)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdExercise)
}

// UpdateExerciseHandler updates an existing exercise.
func UpdateExerciseHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid exercise ID", http.StatusBadRequest)
		return
	}

	var exercise models.Exercise
	err = json.NewDecoder(r.Body).Decode(&exercise)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	exercise.ID = id // Ensure the ID from the URL is used

	if !repository.UpdateExercise(exercise) {
		http.Error(w, "Exercise not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(exercise) // Return the updated exercise
}

// DeleteExerciseHandler deletes an exercise by ID.
func DeleteExerciseHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid exercise ID", http.StatusBadRequest)
		return
	}

	if !repository.DeleteExercise(id) {
		http.Error(w, "Exercise not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// GetDivisionsForExercise returns divisions for a specific exercise
func GetDivisionsForExercise(w http.ResponseWriter, r *http.Request) {
	// Get exercise ID from query parameter
	exerciseIDStr := r.URL.Query().Get("exercise_id")
	if exerciseIDStr == "" {
		http.Error(w, "Exercise ID required", http.StatusBadRequest)
		return
	}
	
	exerciseID, err := strconv.Atoi(exerciseIDStr)
	if err != nil {
		http.Error(w, "Invalid exercise ID", http.StatusBadRequest)
		return
	}
	
	// Get the exercise and return its divisions
	exercise, found := repository.GetExerciseByID(exerciseID)
	if !found {
		http.Error(w, "Exercise not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exercise.Divisions)
}

// CreateDivision creates a new division for an exercise
func CreateDivision(w http.ResponseWriter, r *http.Request) {
	var division models.Division
	err := json.NewDecoder(r.Body).Decode(&division)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdDivision := repository.CreateDivision(division)
	if createdDivision.ID == 0 {
		http.Error(w, "Failed to create division", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdDivision)
}

// UpdateDivision updates a division's information including learning objectives
func UpdateDivision(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var division models.Division
	err := json.NewDecoder(r.Body).Decode(&division)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update the division in the repository
	if repository.UpdateDivision(division) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(division)
	} else {
		http.Error(w, "Failed to update division", http.StatusInternalServerError)
	}
}

// CreateTeam creates a new team within a division
func CreateTeam(w http.ResponseWriter, r *http.Request) {
	var team models.Team
	err := json.NewDecoder(r.Body).Decode(&team)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdTeam := repository.CreateTeam(team)
	if createdTeam.ID == 0 {
		http.Error(w, "Failed to create team", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdTeam)
}

// UpdateTeam updates a team within a specific exercise
func UpdateTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Use a custom struct to handle date strings
	var teamUpdate struct {
		ID          int    `json:"id"`
		ExerciseID  int    `json:"exercise_id"`
		Name        string `json:"name"`
		DivisionID  int    `json:"division_id"`
		POC         string `json:"poc"`
		Status      string `json:"status"`
		StatusStart string `json:"status_start"`
		StatusEnd   string `json:"status_end"`
		Comments    string `json:"comments"`
	}
	
	err := json.NewDecoder(r.Body).Decode(&teamUpdate)
	if err != nil {
		log.Printf("Error decoding team update: %v", err)
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Convert to models.Team with proper date handling
	team := models.Team{
		ID:         teamUpdate.ID,
		ExerciseID: teamUpdate.ExerciseID,
		Name:       teamUpdate.Name,
		DivisionID: teamUpdate.DivisionID,
		POC:        teamUpdate.POC,
		Status:     teamUpdate.Status,
		Comments:   teamUpdate.Comments,
	}
	
	// Parse dates if provided
	if teamUpdate.StatusStart != "" {
		if t, err := time.Parse("2006-01-02", teamUpdate.StatusStart); err == nil {
			team.StatusStart = t
		} else if t, err := time.Parse(time.RFC3339, teamUpdate.StatusStart); err == nil {
			team.StatusStart = t
		}
	}
	
	if teamUpdate.StatusEnd != "" {
		if t, err := time.Parse("2006-01-02", teamUpdate.StatusEnd); err == nil {
			team.StatusEnd = t
		} else if t, err := time.Parse(time.RFC3339, teamUpdate.StatusEnd); err == nil {
			team.StatusEnd = t
		}
	}

	log.Printf("Received update for team: %+v\n", team)
	log.Printf("Team ExerciseID: %d, DivisionID: %d, TeamID: %d\n", team.ExerciseID, team.DivisionID, team.ID)

	// Validate required fields
	if team.ExerciseID == 0 {
		log.Printf("Missing ExerciseID in team update")
		http.Error(w, "Exercise ID is required", http.StatusBadRequest)
		return
	}

	// Get the exercise
	exercise, found := repository.GetExerciseByID(team.ExerciseID)
	if !found {
		log.Printf("Exercise not found with ID: %d", team.ExerciseID)
		http.Error(w, "Exercise not found", http.StatusNotFound)
		return
	}

	// Update the team within the exercise's divisions
	updated := false
	for i, division := range exercise.Divisions {
		if division.ID == team.DivisionID {
			for j, t := range division.Teams {
				if t.ID == team.ID {
					exercise.Divisions[i].Teams[j] = team
					updated = true
					break
				}
			}
		}
		if updated {
			break
		}
	}

	if !updated {
		http.Error(w, "Team not found", http.StatusNotFound)
		return
	}

	// Save the updated exercise
	if !repository.UpdateExercise(exercise) {
		http.Error(w, "Failed to update exercise", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Team updated successfully"))
}

// ChatbotHandler processes natural language commands for the chatbot.
func ChatbotHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Message string `json:"message"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userMessage := strings.ToLower(strings.TrimSpace(requestBody.Message))
	var reply string

	if strings.Contains(userMessage, "list exercises") {
		exercises := repository.GetAllExercises()
		if len(exercises) == 0 {
			reply = "There are no exercises currently."
		} else {
			reply = "Here are the current exercises:\n"
			for _, ex := range exercises {
				reply += fmt.Sprintf("- ID: %d, Name: %s, Start: %s, End: %s\n", ex.ID, ex.Name, ex.StartDate.Format("2006-01-02"), ex.EndDate.Format("2006-01-02"))
			}
		}
	} else if strings.Contains(userMessage, "add exercise") {
		// For simplicity, we'll add a dummy exercise. In a real scenario, this would involve more prompts.
		newExercise := models.Exercise{
			Name:      "New Exercise from Chatbot",
			StartDate: time.Now(),
			EndDate:   time.Now().AddDate(0, 0, 5),
		}
		created := repository.CreateExercise(newExercise)
		reply = fmt.Sprintf("Added new exercise: ID %d, Name: %s.", created.ID, created.Name)
	} else if strings.Contains(userMessage, "change name of exercise") {
		// Expecting format like "change name of exercise 1 to New Name"
		parts := strings.Split(userMessage, " ")
		idStr := ""
		newName := ""
		foundID := false
		for i, part := range parts {
			if part == "exercise" && i+1 < len(parts) {
				idStr = parts[i+1]
				foundID = true
			} else if part == "to" && i+1 < len(parts) && foundID {
				newName = strings.Join(parts[i+1:], " ")
				break
			}
		}

		id, err := strconv.Atoi(idStr)
		if err != nil || newName == "" {
			reply = "Please specify the exercise ID and the new name. E.g., 'change name of exercise 1 to My New Exercise'."
		} else {
			existingEx, found := repository.GetExerciseByID(id)
			if !found {
				reply = fmt.Sprintf("Exercise with ID %d not found.", id)
			} else {
				existingEx.Name = newName
				if repository.UpdateExercise(existingEx) {
					reply = fmt.Sprintf("Successfully changed name of exercise %d to %s.", id, newName)
				} else {
					reply = "Failed to update exercise name."
				}
			}
		}
	} else if strings.Contains(userMessage, "delete exercise") {
		// Expecting format like "delete exercise 1"
		parts := strings.Split(userMessage, " ")
		idStr := ""
		for i, part := range parts {
			if part == "exercise" && i+1 < len(parts) {
				idStr = parts[i+1]
				break
			}
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			reply = "Please specify the exercise ID to delete. E.g., 'delete exercise 1'."
		} else {
			if repository.DeleteExercise(id) {
				reply = fmt.Sprintf("Successfully deleted exercise with ID %d.", id)
			} else {
				reply = fmt.Sprintf("Exercise with ID %d not found.", id)
			}
		}
	} else {
		reply = "I understand 'list exercises', 'add exercise', 'change name of exercise [ID] to [New Name]', and 'delete exercise [ID]'."
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"reply": reply})
}

// GetEvents returns all events for a specific exercise
func GetEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	// Get exercise ID from query parameter
	exerciseIDStr := r.URL.Query().Get("exercise_id")
	if exerciseIDStr == "" {
		http.Error(w, "exercise_id parameter is required", http.StatusBadRequest)
		return
	}

	exerciseID, err := strconv.Atoi(exerciseIDStr)
	if err != nil {
		http.Error(w, "Invalid exercise_id", http.StatusBadRequest)
		return
	}

	// Get events for the exercise using the repository
	events := repository.GetEventsForExercise(exerciseID)
	
	if err := json.NewEncoder(w).Encode(events); err != nil {
		http.Error(w, "Failed to encode events", http.StatusInternalServerError)
		return
	}
}

// CreateEvent creates a new event
func CreateEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	var event models.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Set default values if not provided
	if event.Type == "" {
		event.Type = "milestone"
	}
	if event.Priority == "" {
		event.Priority = "medium"
	}
	if event.Status == "" {
		event.Status = "planned"
	}

	// Create the event using the repository
	createdEvent := repository.CreateEvent(event)
	
	if err := json.NewEncoder(w).Encode(createdEvent); err != nil {
		http.Error(w, "Failed to encode event", http.StatusInternalServerError)
		return
	}
}

// UpdateEvent updates an existing event
func UpdateEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	var event models.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Update the event using the repository
	success := repository.UpdateEvent(event)
	
	if success {
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	} else {
		http.Error(w, "Failed to update event", http.StatusInternalServerError)
	}
}

// DeleteEvent deletes an event
func DeleteEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	// Extract event ID from URL path
	idStr := r.URL.Path[len("/api/events/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	// Delete the event using the repository
	success := repository.DeleteEvent(id)
	
	if success {
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	} else {
		http.Error(w, "Failed to delete event", http.StatusInternalServerError)
	}
}
