package models

import "time"

type Exercise struct {
	ID            int                `json:"id"`
	Name          string             `json:"name"`
	StartDate     time.Time          `json:"start_date"`
	EndDate       time.Time          `json:"end_date"`
	Description   string             `json:"description"`
	TaskedDivisions []string         `json:"tasked_divisions"`
	AOCInvolvement string            `json:"aoc_involvement"`
	SRDPOC        string             `json:"srd_poc"`
	CPDPOC        string             `json:"cpd_poc"`
	Divisions     []Division         `json:"divisions"`
}

type Division struct {
	ID         int    `json:"id"`
	ExerciseID int    `json:"exercise_id"`
	Name       string `json:"name"`
	Teams      []Team `json:"teams"`
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
