package models

import "time"

type Exercise struct {
	ID               int                `json:"id"`
	Name             string             `json:"name"`
	StartDate        time.Time          `json:"start_date"`
	EndDate          time.Time          `json:"end_date"`
	Description      string             `json:"description"`
	Priority         string             `json:"priority"`         // "high", "medium", "low"
	ExerciseEventPOC string             `json:"exercise_event_poc"`
	TaskedDivisions  []string           `json:"tasked_divisions"`
	AOCInvolvement   string             `json:"aoc_involvement"`
	SRDPOC           string             `json:"srd_poc"`
	CPDPOC           string             `json:"cpd_poc"`
	Divisions        []Division         `json:"divisions"`
	Events           []Event            `json:"events"`
}

type Division struct {
	ID                 int    `json:"id"`
	ExerciseID         int    `json:"exercise_id"`
	Name               string `json:"name"`
	LearningObjectives string `json:"learning_objectives"`
	Teams              []Team `json:"teams"`
}

type Team struct {
	ID         int       `json:"id"`
	ExerciseID int       `json:"exercise_id"`
	Name       string    `json:"name"`
	DivisionID int       `json:"division_id"`
	POC        string    `json:"poc"`
	Status     string    `json:"status"` // "green", "yellow", "red"
	StatusStart time.Time `json:"status_start"`
	StatusEnd   time.Time `json:"status_end"`
	Comments   string    `json:"comments"`
}

type Event struct {
	ID         int       `json:"id"`
	ExerciseID int       `json:"exercise_id"`
	Name       string    `json:"name"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	Type       string    `json:"type"`       // "milestone", "phase", "meeting", etc.
	Priority   string    `json:"priority"`   // "high", "medium", "low"
	POC        string    `json:"poc"`
	Status     string    `json:"status"`     // "planned", "in-progress", "completed", "cancelled"
	Description string   `json:"description"`
	Location   string    `json:"location"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
