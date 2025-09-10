package models

import "time"

type Exercise struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

type Division struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Teams []Team `json:"teams"`
}

type Team struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	DivisionID int    `json:"division_id"`
	POC      string `json:"poc"`
	Status   string `json:"status"` // "green", "yellow", "red"
	Comments string `json:"comments"`
}
