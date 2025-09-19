package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	connStr := "user=postgres password=password dbname=sports_data port=150 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Standard divisions and teams structure
	standardDivisions := []string{"COD", "CPD", "SRD", "ISRD", "AMD"}
	standardTeams := []string{"Team 1", "Team 2", "Team 3", "Team 4"}

	// Exercises to update
	exercisesToUpdate := []string{"VALIANT SHIELD", "PACSENTRY"}

	for _, exerciseName := range exercisesToUpdate {
		log.Printf("Updating %s...", exerciseName)

		// Get exercise ID
		var exerciseID int
		err := db.QueryRow("SELECT id FROM exercises WHERE name = $1", exerciseName).Scan(&exerciseID)
		if err != nil {
			log.Printf("Error getting exercise ID for %s: %v", exerciseName, err)
			continue
		}

		// Delete existing divisions and teams for this exercise
		_, err = db.Exec("DELETE FROM divisions WHERE exercise_id = $1", exerciseID)
		if err != nil {
			log.Printf("Error deleting divisions for %s: %v", exerciseName, err)
			continue
		}

		// Insert standard divisions and teams
		for _, divisionName := range standardDivisions {
			var divisionID int
			err = db.QueryRow(`
				INSERT INTO divisions (exercise_id, name, learning_objectives, created_at) 
				VALUES ($1, $2, '', CURRENT_TIMESTAMP) 
				RETURNING id
			`, exerciseID, divisionName).Scan(&divisionID)
			if err != nil {
				log.Printf("Error inserting division %s: %v", divisionName, err)
				continue
			}

			// Insert teams for this division
			for _, teamName := range standardTeams {
				_, err = db.Exec(`
					INSERT INTO teams (exercise_id, division_id, name, poc, status, comments, created_at, updated_at)
					VALUES ($1, $2, $3, '', 'green', '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
				`, exerciseID, divisionID, teamName)
				if err != nil {
					log.Printf("Error inserting team %s for division %s: %v", teamName, divisionName, err)
				}
			}
		}

		log.Printf("Successfully updated %s with standard divisions and teams", exerciseName)
	}

	log.Println("Standardization complete!")
}