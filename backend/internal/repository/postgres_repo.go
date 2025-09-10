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
		       COALESCE(aoc_involvement, ''), COALESCE(srd_poc, ''), COALESCE(cpd_poc, '')
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
		var desc, aoc, srdPoc, cpdPoc sql.NullString
		
		err := rows.Scan(&ex.ID, &ex.Name, &ex.StartDate, &ex.EndDate, 
			&desc, &aoc, &srdPoc, &cpdPoc)
		if err != nil {
			log.Printf("Error scanning exercise: %v", err)
			continue
		}

		ex.Description = desc.String
		ex.AOCInvolvement = aoc.String
		ex.SRDPOC = srdPoc.String
		ex.CPDPOC = cpdPoc.String

		// Load divisions for this exercise
		ex.Divisions = r.GetDivisionsForExercise(ex.ID)
		
		// Load tasked divisions
		ex.TaskedDivisions = r.GetTaskedDivisions(ex.ID)

		exercises = append(exercises, ex)
	}

	return exercises
}

// GetExerciseByIDDB returns a single exercise by ID from the database
func (r *PostgresRepository) GetExerciseByIDDB(id int) (models.Exercise, bool) {
	var ex models.Exercise
	var desc, aoc, srdPoc, cpdPoc sql.NullString

	query := `
		SELECT id, name, start_date, end_date, description, 
		       COALESCE(aoc_involvement, ''), COALESCE(srd_poc, ''), COALESCE(cpd_poc, '')
		FROM exercises
		WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(&ex.ID, &ex.Name, &ex.StartDate, &ex.EndDate,
		&desc, &aoc, &srdPoc, &cpdPoc)
	if err != nil {
		if err == sql.ErrNoRows {
			return ex, false
		}
		log.Printf("Error fetching exercise by ID: %v", err)
		return ex, false
	}

	ex.Description = desc.String
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
		INSERT INTO exercises (name, start_date, end_date, description, aoc_involvement, srd_poc, cpd_poc)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	err = tx.QueryRow(query, exercise.Name, exercise.StartDate, exercise.EndDate,
		exercise.Description, exercise.AOCInvolvement, exercise.SRDPOC, exercise.CPDPOC).Scan(&exercise.ID)
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
		SET name = $2, start_date = $3, end_date = $4, description = $5, 
		    aoc_involvement = $6, srd_poc = $7, cpd_poc = $8, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := tx.Exec(query, exercise.ID, exercise.Name, exercise.StartDate, exercise.EndDate,
		exercise.Description, exercise.AOCInvolvement, exercise.SRDPOC, exercise.CPDPOC)
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
		SELECT id, name
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
		div.ExerciseID = exerciseID
		
		err := rows.Scan(&div.ID, &div.Name)
		if err != nil {
			log.Printf("Error scanning division: %v", err)
			continue
		}

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
	divisions := []models.Division{
		{
			Name: "Division A",
			Teams: []models.Team{
				{Name: "OAT Team", POC: "John Doe", Status: "green", Comments: "All systems operational"},
				{Name: "Intel Team", POC: "Jane Smith", Status: "yellow", Comments: "Minor communications issues"},
			},
		},
		{
			Name: "Division B",
			Teams: []models.Team{
				{Name: "Ops Team", POC: "Peter Jones", Status: "green", Comments: "Ready for operations"},
				{Name: "Support Team", POC: "Mary Williams", Status: "red", Comments: "Equipment maintenance required"},
			},
		},
	}

	for i, division := range divisions {
		divisions[i] = r.createDivision(tx, exerciseID, division)
	}

	return divisions
}

// createDivision creates a division with its teams
func (r *PostgresRepository) createDivision(tx *sql.Tx, exerciseID int, division models.Division) models.Division {
	var divID int
	err := tx.QueryRow("INSERT INTO divisions (exercise_id, name) VALUES ($1, $2) RETURNING id",
		exerciseID, division.Name).Scan(&divID)
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
		log.Println("Initializing database with sample data...")
		
		exercise1 := models.Exercise{
			Name:        "Initial Exercise 1",
			StartDate:   time.Now(),
			EndDate:     time.Now().AddDate(0, 0, 7),
			Description: "First sample exercise",
		}
		r.CreateExerciseDB(exercise1)

		exercise2 := models.Exercise{
			Name:        "Initial Exercise 2",
			StartDate:   time.Now().AddDate(0, 0, 10),
			EndDate:     time.Now().AddDate(0, 0, 17),
			Description: "Second sample exercise",
		}
		r.CreateExerciseDB(exercise2)
		
		log.Println("Sample data created successfully")
	}
}