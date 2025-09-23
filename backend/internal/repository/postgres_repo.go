package repository

import (
	"database/sql"
	"log"
	"srd-calendar-project/backend/internal/database"
	"srd-calendar-project/backend/internal/models"
	"time"
)

// PostgresRepository implements database operations using PostgreSQL
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository() *PostgresRepository {
	return &PostgresRepository{
		db: database.DB,
	}
}

// GetAllExercisesDB returns all exercises from the database
func (r *PostgresRepository) GetAllExercisesDB() []models.Exercise {
	query := `
		SELECT id, name, start_date, end_date, description, 
		       COALESCE(priority, 'medium'), COALESCE(exercise_event_poc, ''), COALESCE(aoc_involvement, ''), COALESCE(srd_poc, ''), COALESCE(cpd_poc, '')
		FROM exercises
		ORDER BY start_date
	`

	rows, err := r.db.Query(query)
	if err != nil {
		log.Printf("Error fetching exercises: %v", err)
		return []models.Exercise{}
	}
	defer rows.Close()

	var exercises []models.Exercise
	for rows.Next() {
		var ex models.Exercise
		var desc, priority, eventPoc, aoc, srdPoc, cpdPoc sql.NullString
		
		err := rows.Scan(&ex.ID, &ex.Name, &ex.StartDate, &ex.EndDate, 
			&desc, &priority, &eventPoc, &aoc, &srdPoc, &cpdPoc)
		if err != nil {
			log.Printf("Error scanning exercise: %v", err)
			continue
		}

		ex.Description = desc.String
		ex.Priority = priority.String
		ex.ExerciseEventPOC = eventPoc.String
		ex.AOCInvolvement = aoc.String
		ex.SRDPOC = srdPoc.String
		ex.CPDPOC = cpdPoc.String

		// Load divisions for this exercise
		ex.Divisions = r.GetDivisionsForExercise(ex.ID)
		
		// Load tasked divisions
		ex.TaskedDivisions = r.GetTaskedDivisions(ex.ID)
		
		// Load events for this exercise
		ex.Events = r.GetEventsForExercise(ex.ID)

		exercises = append(exercises, ex)
	}

	return exercises
}

// GetExerciseByIDDB returns a single exercise by ID from the database
func (r *PostgresRepository) GetExerciseByIDDB(id int) (models.Exercise, bool) {
	var ex models.Exercise
	var desc, eventPoc, aoc, srdPoc, cpdPoc sql.NullString

	query := `
		SELECT id, name, start_date, end_date, description, 
		       COALESCE(priority, 'medium'), COALESCE(exercise_event_poc, ''), COALESCE(aoc_involvement, ''), COALESCE(srd_poc, ''), COALESCE(cpd_poc, '')
		FROM exercises
		WHERE id = $1
	`

	var priority sql.NullString
	err := r.db.QueryRow(query, id).Scan(&ex.ID, &ex.Name, &ex.StartDate, &ex.EndDate,
		&desc, &priority, &eventPoc, &aoc, &srdPoc, &cpdPoc)
	if err != nil {
		if err == sql.ErrNoRows {
			return ex, false
		}
		log.Printf("Error fetching exercise by ID: %v", err)
		return ex, false
	}

	ex.Description = desc.String
	ex.Priority = priority.String
	ex.ExerciseEventPOC = eventPoc.String
	ex.AOCInvolvement = aoc.String
	ex.SRDPOC = srdPoc.String
	ex.CPDPOC = cpdPoc.String

	// Load divisions for this exercise
	ex.Divisions = r.GetDivisionsForExercise(ex.ID)
	
	// Load tasked divisions
	ex.TaskedDivisions = r.GetTaskedDivisions(ex.ID)

	return ex, true
}

// CreateExerciseDB creates a new exercise in the database
func (r *PostgresRepository) CreateExerciseDB(exercise models.Exercise) models.Exercise {
	tx, err := r.db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return exercise
	}
	defer tx.Rollback()

	// Insert exercise
	query := `
		INSERT INTO exercises (name, start_date, end_date, description, priority, exercise_event_poc, aoc_involvement, srd_poc, cpd_poc)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	err = tx.QueryRow(query, exercise.Name, exercise.StartDate, exercise.EndDate,
		exercise.Description, exercise.Priority, exercise.ExerciseEventPOC, exercise.AOCInvolvement, exercise.SRDPOC, exercise.CPDPOC).Scan(&exercise.ID)
	if err != nil {
		log.Printf("Error creating exercise: %v", err)
		return exercise
	}

	// Create default divisions if none provided
	if len(exercise.Divisions) == 0 {
		exercise.Divisions = r.createDefaultDivisions(tx, exercise.ID)
	} else {
		// Save provided divisions
		for _, division := range exercise.Divisions {
			r.createDivision(tx, exercise.ID, division)
		}
	}

	// Save tasked divisions
	for _, divName := range exercise.TaskedDivisions {
		_, err = tx.Exec("INSERT INTO tasked_divisions (exercise_id, division_name) VALUES ($1, $2)", 
			exercise.ID, divName)
		if err != nil {
			log.Printf("Error saving tasked division: %v", err)
		}
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return exercise
	}

	return exercise
}

// UpdateExerciseDB updates an exercise in the database
func (r *PostgresRepository) UpdateExerciseDB(exercise models.Exercise) bool {
	tx, err := r.db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return false
	}
	defer tx.Rollback()

	query := `
		UPDATE exercises 
		SET name = $2, start_date = $3, end_date = $4, description = $5, priority = $6,
		    exercise_event_poc = $7, aoc_involvement = $8, srd_poc = $9, cpd_poc = $10, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := tx.Exec(query, exercise.ID, exercise.Name, exercise.StartDate, exercise.EndDate,
		exercise.Description, exercise.Priority, exercise.ExerciseEventPOC, exercise.AOCInvolvement, exercise.SRDPOC, exercise.CPDPOC)
	if err != nil {
		log.Printf("Error updating exercise: %v", err)
		return false
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return false
	}

	// Update divisions and teams if provided
	if len(exercise.Divisions) > 0 {
		for _, division := range exercise.Divisions {
			for _, team := range division.Teams {
				r.updateTeam(tx, team)
			}
		}
	}

	// Update tasked divisions
	_, err = tx.Exec("DELETE FROM tasked_divisions WHERE exercise_id = $1", exercise.ID)
	if err != nil {
		log.Printf("Error deleting old tasked divisions: %v", err)
	}
	
	for _, divName := range exercise.TaskedDivisions {
		_, err = tx.Exec("INSERT INTO tasked_divisions (exercise_id, division_name) VALUES ($1, $2)", 
			exercise.ID, divName)
		if err != nil {
			log.Printf("Error saving tasked division: %v", err)
		}
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return false
	}

	return true
}

// DeleteExerciseDB deletes an exercise from the database
func (r *PostgresRepository) DeleteExerciseDB(id int) bool {
	query := "DELETE FROM exercises WHERE id = $1"
	result, err := r.db.Exec(query, id)
	if err != nil {
		log.Printf("Error deleting exercise: %v", err)
		return false
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0
}

// GetDivisionsForExercise gets all divisions for an exercise
func (r *PostgresRepository) GetDivisionsForExercise(exerciseID int) []models.Division {
	query := `
		SELECT id, name, COALESCE(learning_objectives, '')
		FROM divisions
		WHERE exercise_id = $1
		ORDER BY id
	`

	rows, err := r.db.Query(query, exerciseID)
	if err != nil {
		log.Printf("Error fetching divisions: %v", err)
		return []models.Division{}
	}
	defer rows.Close()

	var divisions []models.Division
	for rows.Next() {
		var div models.Division
		var learningObjectives sql.NullString
		div.ExerciseID = exerciseID
		
		err := rows.Scan(&div.ID, &div.Name, &learningObjectives)
		if err != nil {
			log.Printf("Error scanning division: %v", err)
			continue
		}
		
		div.LearningObjectives = learningObjectives.String

		// Load teams for this division
		div.Teams = r.GetTeamsForDivision(exerciseID, div.ID)
		divisions = append(divisions, div)
	}

	return divisions
}

// GetDivisionsForExerciseByName returns only divisions that match the specified name for an exercise
func (r *PostgresRepository) GetDivisionsForExerciseByName(exerciseID int, divisionName string) []models.Division {
	log.Printf("DEBUG: GetDivisionsForExerciseByName called with exerciseID=%d, divisionName='%s'", exerciseID, divisionName)
	query := `
		SELECT id, name, COALESCE(learning_objectives, '')
		FROM divisions
		WHERE exercise_id = $1 AND name = $2
		ORDER BY id
	`

	rows, err := r.db.Query(query, exerciseID, divisionName)
	if err != nil {
		log.Printf("Error fetching divisions by name for exercise: %v", err)
		return []models.Division{}
	}
	defer rows.Close()

	var divisions []models.Division
	for rows.Next() {
		var div models.Division
		var learningObjectives sql.NullString

		err := rows.Scan(&div.ID, &div.Name, &learningObjectives)
		if err != nil {
			log.Printf("Error scanning division: %v", err)
			continue
		}

		div.ExerciseID = exerciseID
		div.LearningObjectives = learningObjectives.String

		// Load teams for this division
		div.Teams = r.GetTeamsForDivision(exerciseID, div.ID)

		divisions = append(divisions, div)
	}

	return divisions
}

// GetTeamsForDivision gets all teams for a division
func (r *PostgresRepository) GetTeamsForDivision(exerciseID, divisionID int) []models.Team {
	query := `
		SELECT id, name, poc, status, status_start, status_end, comments
		FROM teams
		WHERE exercise_id = $1 AND division_id = $2
		ORDER BY id
	`

	rows, err := r.db.Query(query, exerciseID, divisionID)
	if err != nil {
		log.Printf("Error fetching teams: %v", err)
		return []models.Team{}
	}
	defer rows.Close()

	var teams []models.Team
	for rows.Next() {
		var team models.Team
		var poc, status, comments sql.NullString
		var statusStart, statusEnd sql.NullTime
		
		team.ExerciseID = exerciseID
		team.DivisionID = divisionID
		
		err := rows.Scan(&team.ID, &team.Name, &poc, &status, &statusStart, &statusEnd, &comments)
		if err != nil {
			log.Printf("Error scanning team: %v", err)
			continue
		}

		team.POC = poc.String
		team.Status = status.String
		if team.Status == "" {
			team.Status = "green"
		}
		team.Comments = comments.String
		
		if statusStart.Valid {
			team.StatusStart = statusStart.Time
		}
		if statusEnd.Valid {
			team.StatusEnd = statusEnd.Time
		}

		teams = append(teams, team)
	}

	return teams
}

// GetTaskedDivisions gets the tasked divisions for an exercise
func (r *PostgresRepository) GetTaskedDivisions(exerciseID int) []string {
	query := "SELECT division_name FROM tasked_divisions WHERE exercise_id = $1"
	
	rows, err := r.db.Query(query, exerciseID)
	if err != nil {
		log.Printf("Error fetching tasked divisions: %v", err)
		return []string{}
	}
	defer rows.Close()

	var divisions []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			divisions = append(divisions, name)
		}
	}
	
	return divisions
}

// createDefaultDivisions creates default divisions and teams for a new exercise
func (r *PostgresRepository) createDefaultDivisions(tx *sql.Tx, exerciseID int) []models.Division {
	divisions := r.createStandardDivisions()

	for i, division := range divisions {
		divisions[i] = r.createDivision(tx, exerciseID, division)
	}

	return divisions
}

// createStandardDivisions creates the standard AOC divisions and teams structure
func (r *PostgresRepository) createStandardDivisions() []models.Division {
	return []models.Division{
		{
			Name: "COD",
			LearningObjectives: "Learning objectives for COD division",
			Teams: []models.Team{
				{Name: "Team 1", POC: "Team Leader", Status: "green", Comments: "Demo team"},
				{Name: "Team 2", POC: "Team Leader", Status: "green", Comments: "Demo team"},
				{Name: "Team 3", POC: "Team Leader", Status: "green", Comments: "Demo team"},
				{Name: "Team 4", POC: "Team Leader", Status: "green", Comments: "Demo team"},
			},
		},
		{
			Name: "CPD",
			LearningObjectives: "Learning objectives for CPD division",
			Teams: []models.Team{
				{Name: "Team 1", POC: "Team Leader", Status: "green", Comments: "Demo team"},
				{Name: "Team 2", POC: "Team Leader", Status: "green", Comments: "Demo team"},
				{Name: "Team 3", POC: "Team Leader", Status: "green", Comments: "Demo team"},
				{Name: "Team 4", POC: "Team Leader", Status: "green", Comments: "Demo team"},
			},
		},
		{
			Name: "SRD",
			LearningObjectives: "Learning objectives for SRD division",
			Teams: []models.Team{
				{Name: "Team 1", POC: "Team Leader", Status: "green", Comments: "Demo team"},
				{Name: "Team 2", POC: "Team Leader", Status: "green", Comments: "Demo team"},
				{Name: "Team 3", POC: "Team Leader", Status: "green", Comments: "Demo team"},
				{Name: "Team 4", POC: "Team Leader", Status: "green", Comments: "Demo team"},
			},
		},
		{
			Name: "ISRD",
			LearningObjectives: "Learning objectives for ISRD division",
			Teams: []models.Team{
				{Name: "Team 1", POC: "Team Leader", Status: "green", Comments: "Demo team"},
				{Name: "Team 2", POC: "Team Leader", Status: "green", Comments: "Demo team"},
				{Name: "Team 3", POC: "Team Leader", Status: "green", Comments: "Demo team"},
				{Name: "Team 4", POC: "Team Leader", Status: "green", Comments: "Demo team"},
			},
		},
		{
			Name: "AMD",
			LearningObjectives: "Learning objectives for AMD division",
			Teams: []models.Team{
				{Name: "Team 1", POC: "Team Leader", Status: "green", Comments: "Demo team"},
				{Name: "Team 2", POC: "Team Leader", Status: "green", Comments: "Demo team"},
				{Name: "Team 3", POC: "Team Leader", Status: "green", Comments: "Demo team"},
				{Name: "Team 4", POC: "Team Leader", Status: "green", Comments: "Demo team"},
			},
		},
	}
}

// createDivision creates a division with its teams
func (r *PostgresRepository) createDivision(tx *sql.Tx, exerciseID int, division models.Division) models.Division {
	var divID int
	err := tx.QueryRow("INSERT INTO divisions (exercise_id, name, learning_objectives) VALUES ($1, $2, $3) RETURNING id",
		exerciseID, division.Name, division.LearningObjectives).Scan(&divID)
	if err != nil {
		log.Printf("Error creating division: %v", err)
		return division
	}

	division.ID = divID
	division.ExerciseID = exerciseID

	// Create teams for this division
	for j, team := range division.Teams {
		var teamID int
		err = tx.QueryRow(`
			INSERT INTO teams (exercise_id, division_id, name, poc, status, comments)
			VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
			exerciseID, divID, team.Name, team.POC, team.Status, team.Comments).Scan(&teamID)
		
		if err != nil {
			log.Printf("Error creating team: %v", err)
			continue
		}
		
		division.Teams[j].ID = teamID
		division.Teams[j].ExerciseID = exerciseID
		division.Teams[j].DivisionID = divID
	}

	return division
}

// CreateDivisionDB creates a new division in the database
func (r *PostgresRepository) CreateDivisionDB(division models.Division) models.Division {
	query := `
		INSERT INTO divisions (exercise_id, name, learning_objectives)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	err := r.db.QueryRow(query, division.ExerciseID, division.Name, division.LearningObjectives).Scan(&division.ID)
	if err != nil {
		log.Printf("Error creating division: %v", err)
		return division
	}

	// Initialize empty teams slice
	division.Teams = []models.Team{}
	return division
}

// UpdateDivisionDB updates a division's information including learning objectives
func (r *PostgresRepository) UpdateDivisionDB(division models.Division) bool {
	query := `
		UPDATE divisions 
		SET name = $2, learning_objectives = $3
		WHERE id = $1
	`

	result, err := r.db.Exec(query, division.ID, division.Name, division.LearningObjectives)
	if err != nil {
		log.Printf("Error updating division: %v", err)
		return false
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0
}

// CreateTeamDB creates a new team in the database
func (r *PostgresRepository) CreateTeamDB(team models.Team) models.Team {
	query := `
		INSERT INTO teams (exercise_id, division_id, name, poc, status, comments)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	// Set default status if empty
	if team.Status == "" {
		team.Status = "green"
	}

	err := r.db.QueryRow(query, team.ExerciseID, team.DivisionID, team.Name, team.POC, team.Status, team.Comments).Scan(&team.ID)
	if err != nil {
		log.Printf("Error creating team: %v", err)
		return team
	}

	return team
}

// updateTeam updates a team in the database
func (r *PostgresRepository) updateTeam(tx *sql.Tx, team models.Team) error {
	query := `
		UPDATE teams 
		SET poc = $2, status = $3, status_start = $4, status_end = $5, 
		    comments = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	var statusStart, statusEnd interface{}
	if !team.StatusStart.IsZero() {
		statusStart = team.StatusStart
	} else {
		statusStart = nil
	}
	if !team.StatusEnd.IsZero() {
		statusEnd = team.StatusEnd
	} else {
		statusEnd = nil
	}

	_, err := tx.Exec(query, team.ID, team.POC, team.Status, statusStart, statusEnd, team.Comments)
	return err
}

// InitializeDatabase initializes the database with sample data if empty
func (r *PostgresRepository) InitializeDatabase() {
	// Check if there are any exercises
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM exercises").Scan(&count)
	if err != nil {
		log.Printf("Error checking exercise count: %v", err)
		return
	}

	// If no exercises exist, create initial data
	if count == 0 {
		log.Println("Initializing database with real exercise data...")
		
		// Create REFORPAC exercise
		reforpac := models.Exercise{
			Name:        "REFORPAC",
			StartDate:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC),
			Description: "Reformation of the Pacific Exercise",
			Priority:    "high",
			Divisions: r.createStandardDivisions(),
		}
		r.CreateExerciseDB(reforpac)

		// Create KEEN EDGE exercise  
		keenEdgeDivisions := r.createStandardDivisions()
		// Add some variation to KEEN EDGE
		keenEdgeDivisions[0].Teams[0].Status = "yellow"
		keenEdge := models.Exercise{
			Name:        "KEEN EDGE",
			StartDate:   time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
			Description: "Keen Edge Exercise",
			Priority:    "medium",
			ExerciseEventPOC: "Mike",
			Divisions: keenEdgeDivisions,
		}
		r.CreateExerciseDB(keenEdge)

		// Create BALIKATAN exercise
		balicatanDivisions := r.createStandardDivisions()
		// Add some variation to BALIKATAN
		balicatanDivisions[0].Teams[0].Status = "red"
		balikatan := models.Exercise{
			Name:        "BALIKATAN",
			StartDate:   time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC),
			Description: "Balikatan Exercise",
			Priority:    "low",
			Divisions: balicatanDivisions,
		}
		r.CreateExerciseDB(balikatan)
		
		log.Println("Real exercise data created successfully")
	}
}

// GetEventsForExercise gets all events for an exercise
func (r *PostgresRepository) GetEventsForExercise(exerciseID int) []models.Event {
	query := `
		SELECT id, exercise_id, name, start_date, end_date, type, priority, poc, status, description, location, created_at, updated_at
		FROM events
		WHERE exercise_id = $1
		ORDER BY start_date, id
	`

	rows, err := r.db.Query(query, exerciseID)
	if err != nil {
		log.Printf("Error fetching events: %v", err)
		return []models.Event{}
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		var poc, description, location sql.NullString
		
		err := rows.Scan(&event.ID, &event.ExerciseID, &event.Name, &event.StartDate, 
			&event.EndDate, &event.Type, &event.Priority, &poc, &event.Status, 
			&description, &location, &event.CreatedAt, &event.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning event: %v", err)
			continue
		}
		
		event.POC = poc.String
		event.Description = description.String
		event.Location = location.String
		events = append(events, event)
	}

	return events
}

// CreateEventDB creates a new event in the database
func (r *PostgresRepository) CreateEventDB(event models.Event) models.Event {
	query := `
		INSERT INTO events (exercise_id, name, start_date, end_date, type, priority, poc, status, description, location)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(query, event.ExerciseID, event.Name, event.StartDate, event.EndDate,
		event.Type, event.Priority, event.POC, event.Status, event.Description, event.Location).Scan(
		&event.ID, &event.CreatedAt, &event.UpdatedAt)
	if err != nil {
		log.Printf("Error creating event: %v", err)
		return event
	}

	return event
}

// UpdateEventDB updates an event in the database
func (r *PostgresRepository) UpdateEventDB(event models.Event) bool {
	query := `
		UPDATE events 
		SET name = $2, start_date = $3, end_date = $4, type = $5, priority = $6, 
		    poc = $7, status = $8, description = $9, location = $10, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.db.Exec(query, event.ID, event.Name, event.StartDate, event.EndDate,
		event.Type, event.Priority, event.POC, event.Status, event.Description, event.Location)
	if err != nil {
		log.Printf("Error updating event: %v", err)
		return false
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0
}

// DeleteEventDB deletes an event from the database
func (r *PostgresRepository) DeleteEventDB(id int) bool {
	query := "DELETE FROM events WHERE id = $1"
	result, err := r.db.Exec(query, id)
	if err != nil {
		log.Printf("Error deleting event: %v", err)
		return false
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0
}

// GetExercisesByDivisionIDDB returns exercises that contain the specified division
func (r *PostgresRepository) GetExercisesByDivisionIDDB(divisionID int) []models.Exercise {
	query := `
		SELECT DISTINCT e.id, e.name, e.start_date, e.end_date, e.description,
		       COALESCE(e.priority, 'medium'), COALESCE(e.exercise_event_poc, ''), COALESCE(e.aoc_involvement, ''), COALESCE(e.srd_poc, ''), COALESCE(e.cpd_poc, '')
		FROM exercises e
		INNER JOIN divisions d ON e.id = d.exercise_id
		WHERE d.id = $1
		ORDER BY e.start_date
	`

	rows, err := r.db.Query(query, divisionID)
	if err != nil {
		log.Printf("Error fetching exercises by division ID: %v", err)
		return []models.Exercise{}
	}
	defer rows.Close()

	var exercises []models.Exercise
	for rows.Next() {
		var ex models.Exercise
		var desc, priority, eventPoc, aoc, srdPoc, cpdPoc sql.NullString

		err := rows.Scan(&ex.ID, &ex.Name, &ex.StartDate, &ex.EndDate,
			&desc, &priority, &eventPoc, &aoc, &srdPoc, &cpdPoc)
		if err != nil {
			log.Printf("Error scanning exercise: %v", err)
			continue
		}

		ex.Description = desc.String
		ex.Priority = priority.String
		ex.ExerciseEventPOC = eventPoc.String
		ex.AOCInvolvement = aoc.String
		ex.SRDPOC = srdPoc.String
		ex.CPDPOC = cpdPoc.String

		// Load divisions for this exercise
		ex.Divisions = r.GetDivisionsForExercise(ex.ID)

		// Load tasked divisions
		ex.TaskedDivisions = r.GetTaskedDivisions(ex.ID)

		// Load events for this exercise
		ex.Events = r.GetEventsForExercise(ex.ID)

		exercises = append(exercises, ex)
	}

	return exercises
}

// GetExercisesByTeamIDDB returns exercises that contain the specified team
func (r *PostgresRepository) GetExercisesByTeamIDDB(teamID int) []models.Exercise {
	query := `
		SELECT DISTINCT e.id, e.name, e.start_date, e.end_date, e.description,
		       COALESCE(e.priority, 'medium'), COALESCE(e.exercise_event_poc, ''), COALESCE(e.aoc_involvement, ''), COALESCE(e.srd_poc, ''), COALESCE(e.cpd_poc, '')
		FROM exercises e
		INNER JOIN divisions d ON e.id = d.exercise_id
		INNER JOIN teams t ON d.id = t.division_id AND e.id = t.exercise_id
		WHERE t.id = $1
		ORDER BY e.start_date
	`

	rows, err := r.db.Query(query, teamID)
	if err != nil {
		log.Printf("Error fetching exercises by team ID: %v", err)
		return []models.Exercise{}
	}
	defer rows.Close()

	var exercises []models.Exercise
	for rows.Next() {
		var ex models.Exercise
		var desc, priority, eventPoc, aoc, srdPoc, cpdPoc sql.NullString

		err := rows.Scan(&ex.ID, &ex.Name, &ex.StartDate, &ex.EndDate,
			&desc, &priority, &eventPoc, &aoc, &srdPoc, &cpdPoc)
		if err != nil {
			log.Printf("Error scanning exercise: %v", err)
			continue
		}

		ex.Description = desc.String
		ex.Priority = priority.String
		ex.ExerciseEventPOC = eventPoc.String
		ex.AOCInvolvement = aoc.String
		ex.SRDPOC = srdPoc.String
		ex.CPDPOC = cpdPoc.String

		// Load divisions for this exercise
		ex.Divisions = r.GetDivisionsForExercise(ex.ID)

		// Load tasked divisions
		ex.TaskedDivisions = r.GetTaskedDivisions(ex.ID)

		// Load events for this exercise
		ex.Events = r.GetEventsForExercise(ex.ID)

		exercises = append(exercises, ex)
	}

	return exercises
}
// DeleteDivisionDB deletes a division and all its teams from the database
func (r *PostgresRepository) DeleteDivisionDB(id int) bool {
	tx, err := r.db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return false
	}
	defer tx.Rollback()

	// First delete all teams in this division
	_, err = tx.Exec("DELETE FROM teams WHERE division_id = $1", id)
	if err != nil {
		log.Printf("Error deleting teams for division %d: %v", id, err)
		return false
	}

	// Then delete the division itself
	result, err := tx.Exec("DELETE FROM divisions WHERE id = $1", id)
	if err != nil {
		log.Printf("Error deleting division %d: %v", id, err)
		return false
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return false
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return false
	}

	return true
}

// DeleteTeamDB deletes a team from the database
func (r *PostgresRepository) DeleteTeamDB(id int) bool {
	// Also need to remove any task assignments for this team
	tx, err := r.db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return false
	}
	defer tx.Rollback()

	// First, unassign all tasks from this team
	_, err = tx.Exec("UPDATE tasks SET team_id = NULL WHERE team_id = $1", id)
	if err != nil {
		log.Printf("Error unassigning tasks from team %d: %v", id, err)
		return false
	}

	// Delete team from team_tasks junction table (if it exists)
	// First check if the table exists
	var tableExists bool
	err = tx.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'team_tasks')").Scan(&tableExists)
	if err == nil && tableExists {
		_, err = tx.Exec("DELETE FROM team_tasks WHERE team_id = $1", id)
		if err != nil {
			log.Printf("Error deleting from team_tasks table: %v", err)
			return false
		}
	}

	// Then delete the team itself
	result, err := tx.Exec("DELETE FROM teams WHERE id = $1", id)
	if err != nil {
		log.Printf("Error deleting team %d: %v", id, err)
		return false
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return false
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return false
	}

	return true
}

// GetExercisesByDivisionNameDB returns exercises that contain a division with the specified name
func (r *PostgresRepository) GetExercisesByDivisionNameDB(divisionName string) []models.Exercise {
	log.Printf("DEBUG: GetExercisesByDivisionNameDB called with divisionName='%s'", divisionName)
	query := `
		SELECT DISTINCT e.id, e.name, e.start_date, e.end_date, e.description,
		       COALESCE(e.priority, 'medium'), COALESCE(e.exercise_event_poc, ''), COALESCE(e.aoc_involvement, ''), COALESCE(e.srd_poc, ''), COALESCE(e.cpd_poc, '')
		FROM exercises e
		INNER JOIN divisions d ON e.id = d.exercise_id
		WHERE d.name = $1
		ORDER BY e.start_date
	`

	rows, err := r.db.Query(query, divisionName)
	if err != nil {
		log.Printf("Error fetching exercises by division name: %v", err)
		return []models.Exercise{}
	}
	defer rows.Close()

	var exercises []models.Exercise
	for rows.Next() {
		var ex models.Exercise
		var desc, priority, eventPoc, aoc, srdPoc, cpdPoc sql.NullString

		err := rows.Scan(&ex.ID, &ex.Name, &ex.StartDate, &ex.EndDate,
			&desc, &priority, &eventPoc, &aoc, &srdPoc, &cpdPoc)
		if err != nil {
			log.Printf("Error scanning exercise: %v", err)
			continue
		}

		ex.Description = desc.String
		ex.Priority = priority.String
		ex.ExerciseEventPOC = eventPoc.String
		ex.AOCInvolvement = aoc.String
		ex.SRDPOC = srdPoc.String
		ex.CPDPOC = cpdPoc.String

		// Load only the divisions that match the filter criteria
		ex.Divisions = r.GetDivisionsForExerciseByName(ex.ID, divisionName)
		log.Printf("DEBUG: Exercise %s (ID: %d) has %d matching divisions for name '%s'", ex.Name, ex.ID, len(ex.Divisions), divisionName)

		// Load tasked divisions
		ex.TaskedDivisions = r.GetTaskedDivisions(ex.ID)

		// Load events for this exercise
		ex.Events = r.GetEventsForExercise(ex.ID)

		exercises = append(exercises, ex)
	}

	return exercises
}

// GetExercisesByTeamNameDB returns exercises that contain a team with the specified name
func (r *PostgresRepository) GetExercisesByTeamNameDB(teamName string) []models.Exercise {
	query := `
		SELECT DISTINCT e.id, e.name, e.start_date, e.end_date, e.description,
		       COALESCE(e.priority, 'medium'), COALESCE(e.exercise_event_poc, ''), COALESCE(e.aoc_involvement, ''), COALESCE(e.srd_poc, ''), COALESCE(e.cpd_poc, '')
		FROM exercises e
		INNER JOIN divisions d ON e.id = d.exercise_id
		INNER JOIN teams t ON d.id = t.division_id AND e.id = t.exercise_id
		WHERE t.name = $1
		ORDER BY e.start_date
	`

	rows, err := r.db.Query(query, teamName)
	if err != nil {
		log.Printf("Error fetching exercises by team name: %v", err)
		return []models.Exercise{}
	}
	defer rows.Close()

	var exercises []models.Exercise
	for rows.Next() {
		var ex models.Exercise
		var desc, priority, eventPoc, aoc, srdPoc, cpdPoc sql.NullString

		err := rows.Scan(&ex.ID, &ex.Name, &ex.StartDate, &ex.EndDate,
			&desc, &priority, &eventPoc, &aoc, &srdPoc, &cpdPoc)
		if err != nil {
			log.Printf("Error scanning exercise: %v", err)
			continue
		}

		ex.Description = desc.String
		ex.Priority = priority.String
		ex.ExerciseEventPOC = eventPoc.String
		ex.AOCInvolvement = aoc.String
		ex.SRDPOC = srdPoc.String
		ex.CPDPOC = cpdPoc.String

		// Load divisions for this exercise, filtered by team name
		ex.Divisions = r.GetDivisionsForExerciseByTeamName(ex.ID, teamName)

		// Load tasked divisions
		ex.TaskedDivisions = r.GetTaskedDivisions(ex.ID)

		// Load events for this exercise
		ex.Events = r.GetEventsForExercise(ex.ID)

		exercises = append(exercises, ex)
	}

	return exercises
}

// GetDivisionsForExerciseByTeamName returns divisions for an exercise that contain a team with the specified name
func (r *PostgresRepository) GetDivisionsForExerciseByTeamName(exerciseID int, teamName string) []models.Division {
	query := `
		SELECT DISTINCT d.id, d.name, COALESCE(d.learning_objectives, '')
		FROM divisions d
		INNER JOIN teams t ON d.id = t.division_id
		WHERE d.exercise_id = $1 AND t.name = $2
		ORDER BY d.id
	`

	rows, err := r.db.Query(query, exerciseID, teamName)
	if err != nil {
		log.Printf("Error fetching divisions for exercise %d with team %s: %v", exerciseID, teamName, err)
		return []models.Division{}
	}
	defer rows.Close()

	var divisions []models.Division
	for rows.Next() {
		var division models.Division
		var learningObjectives sql.NullString

		err := rows.Scan(&division.ID, &division.Name, &learningObjectives)
		if err != nil {
			log.Printf("Error scanning division: %v", err)
			continue
		}

		division.ExerciseID = exerciseID
		division.LearningObjectives = learningObjectives.String

		// Load teams for this division, filtered by team name
		division.Teams = r.GetTeamsForDivisionByName(division.ID, teamName)

		divisions = append(divisions, division)
	}

	return divisions
}

// GetTeamsForDivisionByName returns teams for a division filtered by team name
func (r *PostgresRepository) GetTeamsForDivisionByName(divisionID int, teamName string) []models.Team {
	query := `
		SELECT id, name, COALESCE(poc, ''), COALESCE(status, 'green'),
		       COALESCE(status_start, CURRENT_TIMESTAMP), COALESCE(status_end, CURRENT_TIMESTAMP),
		       COALESCE(comments, ''), exercise_id
		FROM teams
		WHERE division_id = $1 AND name = $2
		ORDER BY id
	`

	rows, err := r.db.Query(query, divisionID, teamName)
	if err != nil {
		log.Printf("Error fetching teams for division %d with name %s: %v", divisionID, teamName, err)
		return []models.Team{}
	}
	defer rows.Close()

	var teams []models.Team
	for rows.Next() {
		var team models.Team
		var poc, status, comments sql.NullString

		err := rows.Scan(&team.ID, &team.Name, &poc, &status,
			&team.StatusStart, &team.StatusEnd, &comments, &team.ExerciseID)
		if err != nil {
			log.Printf("Error scanning team: %v", err)
			continue
		}

		team.DivisionID = divisionID
		team.POC = poc.String
		team.Status = status.String
		team.Comments = comments.String

		teams = append(teams, team)
	}

	return teams
}

