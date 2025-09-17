package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// InitDB initializes the database connection
func InitDB() error {
	// Get database configuration from environment variables with defaults
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "test_db")

	// First, connect to the default postgres database to create our database if needed
	defaultPsql := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		host, port, user, password)
	
	defaultDB, err := sql.Open("postgres", defaultPsql)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database: %w", err)
	}
	
	// Check if database exists, create if not
	var exists bool
	err = defaultDB.QueryRow("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)", dbname).Scan(&exists)
	if err != nil {
		defaultDB.Close()
		return fmt.Errorf("failed to check if database exists: %w", err)
	}
	
	if !exists {
		_, err = defaultDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
		if err != nil {
			defaultDB.Close()
			return fmt.Errorf("failed to create database: %w", err)
		}
		log.Printf("Created database %s", dbname)
	}
	defaultDB.Close()

	// Now connect to our specific database
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Successfully connected to database %s on %s:%s", dbname, host, port)
	
	// Create tables if they don't exist
	err = createTables()
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// createTables creates the database schema
func createTables() error {
	// Create each table separately to better handle errors
	tables := []string{
		`CREATE TABLE IF NOT EXISTS exercises (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			start_date TIMESTAMP NOT NULL,
			end_date TIMESTAMP NOT NULL,
			description TEXT,
			exercise_event_poc VARCHAR(255),
			aoc_involvement VARCHAR(255),
			srd_poc VARCHAR(255),
			cpd_poc VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tasked_divisions (
			id SERIAL PRIMARY KEY,
			exercise_id INTEGER REFERENCES exercises(id) ON DELETE CASCADE,
			division_name VARCHAR(255),
			UNIQUE(exercise_id, division_name)
		)`,
		`CREATE TABLE IF NOT EXISTS divisions (
			id SERIAL PRIMARY KEY,
			exercise_id INTEGER REFERENCES exercises(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			learning_objectives TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS teams (
			id SERIAL PRIMARY KEY,
			exercise_id INTEGER REFERENCES exercises(id) ON DELETE CASCADE,
			division_id INTEGER REFERENCES divisions(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			poc VARCHAR(255),
			status VARCHAR(50) DEFAULT 'green',
			status_start TIMESTAMP,
			status_end TIMESTAMP,
			comments TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS events (
			id SERIAL PRIMARY KEY,
			exercise_id INTEGER REFERENCES exercises(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			start_date TIMESTAMP NOT NULL,
			end_date TIMESTAMP NOT NULL,
			type VARCHAR(50) DEFAULT 'milestone',
			priority VARCHAR(20) DEFAULT 'medium',
			poc VARCHAR(255),
			status VARCHAR(50) DEFAULT 'planned',
			description TEXT,
			location VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	
	// Create indexes
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_exercises_dates ON exercises(start_date, end_date)`,
		`CREATE INDEX IF NOT EXISTS idx_divisions_exercise ON divisions(exercise_id)`,
		`CREATE INDEX IF NOT EXISTS idx_teams_exercise ON teams(exercise_id)`,
		`CREATE INDEX IF NOT EXISTS idx_teams_division ON teams(division_id)`,
		`CREATE INDEX IF NOT EXISTS idx_events_exercise ON events(exercise_id)`,
		`CREATE INDEX IF NOT EXISTS idx_events_dates ON events(start_date, end_date)`,
	}
	
	// Execute table creation
	for _, table := range tables {
		_, err := DB.Exec(table)
		if err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}
	
	// Execute index creation
	for _, index := range indexes {
		_, err := DB.Exec(index)
		if err != nil {
			// Log index creation errors but don't fail
			log.Printf("Warning: failed to create index: %v", err)
		}
	}

	// Add exercise_event_poc column if it doesn't exist
	_, err := DB.Exec(`
		DO $$ 
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'exercises' AND column_name = 'exercise_event_poc') THEN
				ALTER TABLE exercises ADD COLUMN exercise_event_poc VARCHAR(255);
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning: failed to add exercise_event_poc column: %v", err)
	}

	// Add learning_objectives column to divisions if it doesn't exist
	_, err = DB.Exec(`
		DO $$ 
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'divisions' AND column_name = 'learning_objectives') THEN
				ALTER TABLE divisions ADD COLUMN learning_objectives TEXT;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning: failed to add learning_objectives column: %v", err)
	}

	// Add priority column to exercises if it doesn't exist
	_, err = DB.Exec(`
		DO $$ 
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'exercises' AND column_name = 'priority') THEN
				ALTER TABLE exercises ADD COLUMN priority VARCHAR(20) DEFAULT 'medium';
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning: failed to add priority column: %v", err)
	}

	log.Println("Database schema created/verified successfully")
	return nil
}

// CloseDB closes the database connection
func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}