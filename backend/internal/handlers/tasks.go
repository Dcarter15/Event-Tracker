package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"srd-calendar-project/backend/internal/database"
	"srd-calendar-project/backend/internal/models"
)

// GetTasks returns all tasks for a given exercise
func GetTasks(w http.ResponseWriter, r *http.Request) {
	exerciseIDStr := r.URL.Query().Get("exercise_id")
	if exerciseIDStr == "" {
		http.Error(w, "exercise_id is required", http.StatusBadRequest)
		return
	}

	exerciseID, err := strconv.Atoi(exerciseIDStr)
	if err != nil {
		http.Error(w, "Invalid exercise_id", http.StatusBadRequest)
		return
	}

	query := `
		SELECT t.id, t.exercise_id, t.team_id, t.name, t.description, t.status,
		       t.due_date, t.assigned_to, t.completed_at, t.created_at, t.updated_at,
		       COALESCE(tm.name, '') as team_name,
		       COALESCE(d.name, '') as division_name
		FROM tasks t
		LEFT JOIN teams tm ON t.team_id = tm.id
		LEFT JOIN divisions d ON tm.division_id = d.id
		WHERE t.exercise_id = $1
		ORDER BY
			CASE t.status
				WHEN 'pending' THEN 1
				WHEN 'in-progress' THEN 2
				WHEN 'completed' THEN 3
			END,
			t.due_date ASC NULLS LAST,
			t.created_at DESC
	`

	rows, err := database.DB.Query(query, exerciseID)
	if err != nil {
		log.Printf("Error querying tasks: %v", err)
		http.Error(w, "Error fetching tasks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		var teamID sql.NullInt64
		var dueDate, completedAt sql.NullTime
		var description, assignedTo, teamName, divisionName sql.NullString

		err := rows.Scan(
			&task.ID,
			&task.ExerciseID,
			&teamID,
			&task.Name,
			&description,
			&task.Status,
			&dueDate,
			&assignedTo,
			&completedAt,
			&task.CreatedAt,
			&task.UpdatedAt,
			&teamName,
			&divisionName,
		)
		if err != nil {
			log.Printf("Error scanning task: %v", err)
			continue
		}

		task.Description = description.String
		task.AssignedTo = assignedTo.String
		task.TeamName = teamName.String
		task.DivisionName = divisionName.String
		if teamID.Valid {
			tid := int(teamID.Int64)
			task.TeamID = &tid
		}
		if dueDate.Valid {
			task.DueDate = &dueDate.Time
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}

		// Load all teams assigned to this task from task_teams table
		teamsQuery := `
			SELECT tt.team_id, tm.name, tm.poc, tm.status, tm.comments, d.name as division_name
			FROM task_teams tt
			JOIN teams tm ON tt.team_id = tm.id
			LEFT JOIN divisions d ON tm.division_id = d.id
			WHERE tt.task_id = $1
			ORDER BY tm.name
		`
		teamRows, err := database.DB.Query(teamsQuery, task.ID)
		if err != nil {
			log.Printf("Error loading teams for task %d: %v", task.ID, err)
		} else {
			var teamIDs []int
			var teams []models.Team
			for teamRows.Next() {
				var team models.Team
				var poc, status, comments, divisionName sql.NullString
				err := teamRows.Scan(&team.ID, &team.Name, &poc, &status, &comments, &divisionName)
				if err != nil {
					log.Printf("Error scanning team for task: %v", err)
					continue
				}
				team.POC = poc.String
				team.Status = status.String
				team.Comments = comments.String
				team.ExerciseID = exerciseID

				teamIDs = append(teamIDs, team.ID)
				teams = append(teams, team)
			}
			teamRows.Close()
			task.TeamIDs = teamIDs
			task.Teams = teams
		}

		tasks = append(tasks, task)
	}

	if tasks == nil {
		tasks = []models.Task{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// CreateTask creates a new task
func CreateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if task.Name == "" {
		http.Error(w, "Task name is required", http.StatusBadRequest)
		return
	}
	if task.ExerciseID == 0 {
		http.Error(w, "Exercise ID is required", http.StatusBadRequest)
		return
	}

	// Set default status if not provided
	if task.Status == "" {
		task.Status = "pending"
	}

	query := `
		INSERT INTO tasks (exercise_id, team_id, name, description, status, due_date, assigned_to, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`

	var teamID sql.NullInt64
	if task.TeamID != nil {
		teamID = sql.NullInt64{Int64: int64(*task.TeamID), Valid: true}
	}

	err := database.DB.QueryRow(
		query,
		task.ExerciseID,
		teamID,
		task.Name,
		sql.NullString{String: task.Description, Valid: task.Description != ""},
		task.Status,
		task.DueDate,
		sql.NullString{String: task.AssignedTo, Valid: task.AssignedTo != ""},
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)

	if err != nil {
		log.Printf("Error creating task: %v", err)
		http.Error(w, "Error creating task", http.StatusInternalServerError)
		return
	}

	// Handle multiple team assignments
	if len(task.TeamIDs) > 0 {
		for _, teamID := range task.TeamIDs {
			_, err := database.DB.Exec(
				"INSERT INTO task_teams (task_id, team_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
				task.ID, teamID)
			if err != nil {
				log.Printf("Error assigning task to team %d: %v", teamID, err)
			}
		}

		// Load the full team information for response
		teamsQuery := `
			SELECT tt.team_id, tm.name, tm.poc, tm.status, tm.comments
			FROM task_teams tt
			JOIN teams tm ON tt.team_id = tm.id
			WHERE tt.task_id = $1
			ORDER BY tm.name
		`
		teamRows, err := database.DB.Query(teamsQuery, task.ID)
		if err == nil {
			var teams []models.Team
			for teamRows.Next() {
				var team models.Team
				var poc, status, comments sql.NullString
				err := teamRows.Scan(&team.ID, &team.Name, &poc, &status, &comments)
				if err == nil {
					team.POC = poc.String
					team.Status = status.String
					team.Comments = comments.String
					team.ExerciseID = task.ExerciseID
					teams = append(teams, team)
				}
			}
			teamRows.Close()
			task.Teams = teams
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// UpdateTask updates an existing task
func UpdateTask(w http.ResponseWriter, r *http.Request) {
	taskIDStr := chi.URLParam(r, "id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	task.ID = taskID

	// Handle status change to completed
	var completedAt *time.Time
	if task.Status == "completed" {
		now := time.Now()
		completedAt = &now
	}

	query := `
		UPDATE tasks 
		SET name = $2, description = $3, status = $4, due_date = $5, 
		    assigned_to = $6, team_id = $7, completed_at = $8, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at
	`

	var teamID sql.NullInt64
	if task.TeamID != nil {
		teamID = sql.NullInt64{Int64: int64(*task.TeamID), Valid: true}
	}

	err = database.DB.QueryRow(
		query,
		task.ID,
		task.Name,
		sql.NullString{String: task.Description, Valid: task.Description != ""},
		task.Status,
		task.DueDate,
		sql.NullString{String: task.AssignedTo, Valid: task.AssignedTo != ""},
		teamID,
		completedAt,
	).Scan(&task.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			log.Printf("Error updating task: %v", err)
			http.Error(w, "Error updating task", http.StatusInternalServerError)
		}
		return
	}

	task.CompletedAt = completedAt

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// AssignTaskToTeam assigns or unassigns a task to/from a team
func AssignTaskToTeam(w http.ResponseWriter, r *http.Request) {
	taskIDStr := chi.URLParam(r, "id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var body struct {
		TeamID *int `json:"team_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `
		UPDATE tasks 
		SET team_id = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at
	`

	var teamID sql.NullInt64
	if body.TeamID != nil {
		teamID = sql.NullInt64{Int64: int64(*body.TeamID), Valid: true}
	}

	var updatedAt time.Time
	err = database.DB.QueryRow(query, taskID, teamID).Scan(&updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			log.Printf("Error assigning task to team: %v", err)
			http.Error(w, "Error assigning task", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Task assignment updated successfully",
		"updated_at": updatedAt,
	})
}

// AssignTaskToMultipleTeams assigns a task to multiple teams
func AssignTaskToMultipleTeams(w http.ResponseWriter, r *http.Request) {
	taskIDStr := chi.URLParam(r, "id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var body struct {
		TeamIDs []int `json:"team_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := database.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Error assigning teams", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Clear existing team assignments
	_, err = tx.Exec("DELETE FROM task_teams WHERE task_id = $1", taskID)
	if err != nil {
		log.Printf("Error clearing existing team assignments: %v", err)
		http.Error(w, "Error assigning teams", http.StatusInternalServerError)
		return
	}

	// Add new team assignments
	for _, teamID := range body.TeamIDs {
		_, err = tx.Exec("INSERT INTO task_teams (task_id, team_id) VALUES ($1, $2)", taskID, teamID)
		if err != nil {
			log.Printf("Error assigning task to team %d: %v", teamID, err)
			http.Error(w, "Error assigning teams", http.StatusInternalServerError)
			return
		}
	}

	// Update task's updated_at timestamp
	var updatedAt time.Time
	err = tx.QueryRow("UPDATE tasks SET updated_at = CURRENT_TIMESTAMP WHERE id = $1 RETURNING updated_at", taskID).Scan(&updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			log.Printf("Error updating task timestamp: %v", err)
			http.Error(w, "Error assigning teams", http.StatusInternalServerError)
		}
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Error assigning teams", http.StatusInternalServerError)
		return
	}

	// Load assigned teams for response
	teamsQuery := `
		SELECT tt.team_id, tm.name, tm.poc, tm.status, tm.comments, d.name as division_name
		FROM task_teams tt
		JOIN teams tm ON tt.team_id = tm.id
		LEFT JOIN divisions d ON tm.division_id = d.id
		WHERE tt.task_id = $1
		ORDER BY tm.name
	`
	teamRows, err := database.DB.Query(teamsQuery, taskID)
	var teams []models.Team
	if err == nil {
		for teamRows.Next() {
			var team models.Team
			var poc, status, comments, divisionName sql.NullString
			err := teamRows.Scan(&team.ID, &team.Name, &poc, &status, &comments, &divisionName)
			if err == nil {
				team.POC = poc.String
				team.Status = status.String
				team.Comments = comments.String
				teams = append(teams, team)
			}
		}
		teamRows.Close()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Task assigned to multiple teams successfully",
		"updated_at": updatedAt,
		"teams": teams,
	})
}

// DeleteTask deletes a task
func DeleteTask(w http.ResponseWriter, r *http.Request) {
	taskIDStr := chi.URLParam(r, "id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM tasks WHERE id = $1`
	result, err := database.DB.Exec(query, taskID)
	if err != nil {
		log.Printf("Error deleting task: %v", err)
		http.Error(w, "Error deleting task", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		http.Error(w, "Error deleting task", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}