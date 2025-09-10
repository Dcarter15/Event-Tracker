package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"srd-calendar-project/backend/internal/models"
	"srd-calendar-project/backend/internal/repository"
	"strconv"

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

// GetDivisionsForExercise (existing handler, unchanged for now)
func GetDivisionsForExercise(w http.ResponseWriter, r *http.Request) {
	divisions := []models.Division{
		{
			ID:   1,
			Name: "Division A",
			Teams: []models.Team{
				{ID: 1, Name: "Team 1", DivisionID: 1, POC: "John Doe", Status: "green", Comments: "All systems go."}, 
				{ID: 2, Name: "Team 2", DivisionID: 1, POC: "Jane Smith", Status: "yellow", Comments: "Minor issues with comms."}, 
			},
		},
		{
			ID:   2,
			Name: "Division B",
			Teams: []models.Team{
				{ID: 3, Name: "Team 3", DivisionID: 2, POC: "Peter Jones", Status: "green", Comments: ""},
				{ID: 4, Name: "Team 4", DivisionID: 2, POC: "Mary Williams", Status: "red", Comments: "Critical system failure."}, 
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(divisions)
}

// UpdateTeam (existing handler, unchanged for now)
func UpdateTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var team models.Team
	err := json.NewDecoder(r.Body).Decode(&team)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Received update for team: %+v\n", team)

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


	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Team updated successfully"))
}

