package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"srd-calendar-project/backend/internal/models"
	"srd-calendar-project/backend/internal/repository"
	"strconv"
	"strings"
	"time"
)

// Enhanced ChatbotHandler with better natural language processing
func EnhancedChatbotHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Message string `json:"message"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userMessage := strings.TrimSpace(requestBody.Message)
	reply := processCommand(userMessage)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"reply": reply})
}

func processCommand(message string) string {
	lowerMessage := strings.ToLower(message)

	// Help command
	if containsAny(lowerMessage, []string{"help", "what can you do", "commands", "?"}) {
		return getHelpMessage()
	}

	// List exercises
	if containsAny(lowerMessage, []string{"list exercise", "show exercise", "get exercise", "all exercise", "view exercise"}) {
		return listExercises()
	}

	// Add exercise with more flexible parsing
	if containsAny(lowerMessage, []string{"add exercise", "create exercise", "new exercise", "schedule exercise"}) {
		return addExercise(message)
	}

	// Update exercise
	if containsAny(lowerMessage, []string{"update exercise", "modify exercise", "change exercise", "edit exercise"}) {
		return updateExercise(message)
	}

	// Delete exercise
	if containsAny(lowerMessage, []string{"delete exercise", "remove exercise", "cancel exercise"}) {
		return deleteExercise(message)
	}

	// Get specific exercise details
	if containsAny(lowerMessage, []string{"show exercise", "get exercise", "details of exercise", "info about exercise"}) && containsNumber(lowerMessage) {
		return getExerciseDetails(message)
	}

	// Division/team related queries
	if containsAny(lowerMessage, []string{"division", "team"}) {
		return handleDivisionQuery(message)
	}

	// Date-related queries
	if containsAny(lowerMessage, []string{"today", "this week", "next week", "this month", "upcoming"}) {
		return getExercisesByTimeframe(message)
	}

	return "I'm not sure what you're asking. Type 'help' to see what I can do, or try commands like 'list exercises', 'add exercise', or 'show exercise details'."
}

func containsAny(text string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

func containsNumber(text string) bool {
	re := regexp.MustCompile(`\d+`)
	return re.MatchString(text)
}

func getHelpMessage() string {
	return `I can help you manage exercises and teams. Here are some things you can ask me:

📋 **Exercise Management:**
• "List exercises" - Show all exercises
• "Add exercise [name] from [date] to [date]" - Create a new exercise
• "Update exercise [ID] [field] to [value]" - Modify an exercise
• "Delete exercise [ID]" - Remove an exercise
• "Show exercise [ID]" - Get detailed info about an exercise

📅 **Time-based Queries:**
• "Show exercises this week/month"
• "What's happening today?"
• "Show upcoming exercises"

👥 **Division & Team Info:**
• "Show divisions"
• "Show team status"

Try asking: "Add exercise REFORPAC IPC from 2025-10-01 to 2025-10-15"`
}

func listExercises() string {
	exercises := repository.GetAllExercises()
	if len(exercises) == 0 {
		return "There are no exercises currently scheduled. You can add one by saying 'Add exercise [name] from [date] to [date]'."
	}

	reply := fmt.Sprintf("📅 **Current Exercises** (%d total):\n\n", len(exercises))
	for _, ex := range exercises {
		status := getExerciseStatus(ex)
		reply += fmt.Sprintf("%s **%s** (ID: %d)\n", status, ex.Name, ex.ID)
		reply += fmt.Sprintf("   📆 %s to %s\n", ex.StartDate.Format("Jan 2, 2006"), ex.EndDate.Format("Jan 2, 2006"))
		
		if ex.Description != "" {
			reply += fmt.Sprintf("   📝 %s\n", ex.Description)
		}
		if ex.SRDPOC != "" || ex.CPDPOC != "" {
			reply += "   👤 POCs: "
			if ex.SRDPOC != "" {
				reply += fmt.Sprintf("SRD: %s ", ex.SRDPOC)
			}
			if ex.CPDPOC != "" {
				reply += fmt.Sprintf("CPD: %s", ex.CPDPOC)
			}
			reply += "\n"
		}
		reply += "\n"
	}
	return reply
}

func getExerciseStatus(ex models.Exercise) string {
	now := time.Now()
	if now.Before(ex.StartDate) {
		return "🔵" // Future
	} else if now.After(ex.EndDate) {
		return "⚫" // Completed
	}
	return "🟢" // Active
}

func addExercise(message string) string {
	// Try to parse exercise details from the message
	// Patterns: "add exercise [name] from [date] to [date]"
	// "create exercise called [name] starting [date] ending [date]"
	
	// Extract name
	namePattern := regexp.MustCompile(`(?:called|named|:|exercise)\s+([A-Za-z0-9\s\-]+?)(?:\s+from|\s+starting|\s+on|,|$)`)
	nameMatch := namePattern.FindStringSubmatch(message)
	
	name := "New Exercise"
	if len(nameMatch) > 1 {
		name = strings.TrimSpace(nameMatch[1])
	}

	// Extract dates
	datePattern := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}|\d{1,2}/\d{1,2}/\d{4}|\w+\s+\d{1,2},?\s+\d{4})`)
	dates := datePattern.FindAllString(message, -1)
	
	var startDate, endDate time.Time
	if len(dates) >= 2 {
		startDate, _ = parseDate(dates[0])
		endDate, _ = parseDate(dates[1])
	} else if len(dates) == 1 {
		startDate, _ = parseDate(dates[0])
		endDate = startDate.AddDate(0, 0, 7) // Default to 1 week duration
	} else {
		// Default dates if none provided
		startDate = time.Now()
		endDate = time.Now().AddDate(0, 0, 7)
	}

	// Extract description if provided
	descPattern := regexp.MustCompile(`(?:description:|desc:|about:)\s*(.+?)(?:\.|$)`)
	descMatch := descPattern.FindStringSubmatch(message)
	description := ""
	if len(descMatch) > 1 {
		description = strings.TrimSpace(descMatch[1])
	}

	newExercise := models.Exercise{
		Name:        name,
		StartDate:   startDate,
		EndDate:     endDate,
		Description: description,
	}

	created := repository.CreateExercise(newExercise)
	return fmt.Sprintf("✅ Successfully created exercise:\n\n**%s** (ID: %d)\n📅 %s to %s\n\nYou can update it by saying 'Update exercise %d [field] to [value]'",
		created.Name, created.ID, 
		created.StartDate.Format("Jan 2, 2006"), 
		created.EndDate.Format("Jan 2, 2006"),
		created.ID)
}

func parseDate(dateStr string) (time.Time, error) {
	// Try different date formats
	formats := []string{
		"2006-01-02",
		"1/2/2006",
		"01/02/2006",
		"January 2, 2006",
		"Jan 2, 2006",
		"2 January 2006",
		"2 Jan 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	// If no format works, return current date
	return time.Now(), fmt.Errorf("could not parse date: %s", dateStr)
}

func updateExercise(message string) string {
	// Extract exercise ID
	idPattern := regexp.MustCompile(`(?:exercise|id)\s*(\d+)`)
	idMatch := idPattern.FindStringSubmatch(message)
	if len(idMatch) < 2 {
		return "Please specify the exercise ID. Example: 'Update exercise 1 name to REFORPAC IPC'"
	}

	id, err := strconv.Atoi(idMatch[1])
	if err != nil {
		return "Invalid exercise ID. Please use a number."
	}

	exercise, found := repository.GetExerciseByID(id)
	if !found {
		return fmt.Sprintf("Exercise with ID %d not found.", id)
	}

	lowerMessage := strings.ToLower(message)
	updated := false

	// Update name
	if strings.Contains(lowerMessage, "name to") || strings.Contains(lowerMessage, "rename") {
		namePattern := regexp.MustCompile(`(?:name to|rename to|called)\s+(.+?)(?:\.|$)`)
		nameMatch := namePattern.FindStringSubmatch(message)
		if len(nameMatch) > 1 {
			exercise.Name = strings.TrimSpace(nameMatch[1])
			updated = true
		}
	}

	// Update description
	if strings.Contains(lowerMessage, "description") {
		descPattern := regexp.MustCompile(`(?:description to|description:)\s+(.+?)(?:\.|$)`)
		descMatch := descPattern.FindStringSubmatch(message)
		if len(descMatch) > 1 {
			exercise.Description = strings.TrimSpace(descMatch[1])
			updated = true
		}
	}

	// Update POCs
	if strings.Contains(lowerMessage, "srd poc") {
		pocPattern := regexp.MustCompile(`(?:srd poc to|srd poc:)\s+(.+?)(?:\.|,|$)`)
		pocMatch := pocPattern.FindStringSubmatch(message)
		if len(pocMatch) > 1 {
			exercise.SRDPOC = strings.TrimSpace(pocMatch[1])
			updated = true
		}
	}

	if strings.Contains(lowerMessage, "cpd poc") {
		pocPattern := regexp.MustCompile(`(?:cpd poc to|cpd poc:)\s+(.+?)(?:\.|,|$)`)
		pocMatch := pocPattern.FindStringSubmatch(message)
		if len(pocMatch) > 1 {
			exercise.CPDPOC = strings.TrimSpace(pocMatch[1])
			updated = true
		}
	}

	// Update dates
	if strings.Contains(lowerMessage, "start date") || strings.Contains(lowerMessage, "starting") {
		datePattern := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}|\d{1,2}/\d{1,2}/\d{4})`)
		dates := datePattern.FindAllString(message, -1)
		if len(dates) > 0 {
			if newDate, err := parseDate(dates[0]); err == nil {
				exercise.StartDate = newDate
				updated = true
			}
		}
	}

	if strings.Contains(lowerMessage, "end date") || strings.Contains(lowerMessage, "ending") {
		datePattern := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}|\d{1,2}/\d{1,2}/\d{4})`)
		dates := datePattern.FindAllString(message, -1)
		if len(dates) > 0 {
			dateToUse := dates[len(dates)-1] // Use last date found
			if newDate, err := parseDate(dateToUse); err == nil {
				exercise.EndDate = newDate
				updated = true
			}
		}
	}

	if !updated {
		return "No updates were made. Try: 'Update exercise 1 name to New Name' or 'Update exercise 1 description to New Description'"
	}

	if repository.UpdateExercise(exercise) {
		return fmt.Sprintf("✅ Successfully updated exercise %d:\n\n**%s**\n%s", 
			id, exercise.Name, getExerciseDetailsString(exercise))
	}

	return "Failed to update exercise."
}

func deleteExercise(message string) string {
	// Extract ID
	idPattern := regexp.MustCompile(`\d+`)
	idMatch := idPattern.FindString(message)
	if idMatch == "" {
		return "Please specify the exercise ID to delete. Example: 'Delete exercise 1'"
	}

	id, err := strconv.Atoi(idMatch)
	if err != nil {
		return "Invalid exercise ID."
	}

	// Get exercise details before deleting
	exercise, found := repository.GetExerciseByID(id)
	if !found {
		return fmt.Sprintf("Exercise with ID %d not found.", id)
	}

	if repository.DeleteExercise(id) {
		return fmt.Sprintf("✅ Successfully deleted exercise:\n**%s** (ID: %d)", exercise.Name, id)
	}

	return "Failed to delete exercise."
}

func getExerciseDetails(message string) string {
	// Extract ID
	idPattern := regexp.MustCompile(`\d+`)
	idMatch := idPattern.FindString(message)
	if idMatch == "" {
		return "Please specify the exercise ID. Example: 'Show exercise 1'"
	}

	id, err := strconv.Atoi(idMatch)
	if err != nil {
		return "Invalid exercise ID."
	}

	exercise, found := repository.GetExerciseByID(id)
	if !found {
		return fmt.Sprintf("Exercise with ID %d not found.", id)
	}

	return fmt.Sprintf("📋 **Exercise Details:**\n\n**%s** (ID: %d)\n%s", 
		exercise.Name, exercise.ID, getExerciseDetailsString(exercise))
}

func getExerciseDetailsString(ex models.Exercise) string {
	details := fmt.Sprintf("📅 %s to %s\n", 
		ex.StartDate.Format("Jan 2, 2006"), 
		ex.EndDate.Format("Jan 2, 2006"))
	
	duration := ex.EndDate.Sub(ex.StartDate).Hours() / 24
	details += fmt.Sprintf("⏱️ Duration: %.0f days\n", duration)
	
	if ex.Description != "" {
		details += fmt.Sprintf("📝 Description: %s\n", ex.Description)
	}
	
	if ex.SRDPOC != "" {
		details += fmt.Sprintf("👤 SRD POC: %s\n", ex.SRDPOC)
	}
	
	if ex.CPDPOC != "" {
		details += fmt.Sprintf("👤 CPD POC: %s\n", ex.CPDPOC)
	}
	
	if ex.AOCInvolvement != "" {
		details += fmt.Sprintf("🎯 AOC Involvement: %s\n", ex.AOCInvolvement)
	}
	
	if len(ex.TaskedDivisions) > 0 {
		details += fmt.Sprintf("🏢 Tasked Divisions: %s\n", strings.Join(ex.TaskedDivisions, ", "))
	}
	
	return details
}

func handleDivisionQuery(message string) string {
	// This would connect to actual division/team data
	// For now, returning placeholder info
	return `📊 **Division Status Overview:**

**Division A** 🟡
• Team 1: 🟢 Green - All systems operational
• Team 2: 🟡 Yellow - Minor communications issues

**Division B** 🔴
• Team 3: 🟢 Green - Ready for operations  
• Team 4: 🔴 Red - Critical system failure, maintenance required

Use the main interface to update team statuses and add comments.`
}

func getExercisesByTimeframe(message string) string {
	exercises := repository.GetAllExercises()
	lowerMessage := strings.ToLower(message)
	now := time.Now()
	
	var filtered []models.Exercise
	var timeframeDesc string
	
	if strings.Contains(lowerMessage, "today") {
		timeframeDesc = "Today"
		for _, ex := range exercises {
			if now.After(ex.StartDate) && now.Before(ex.EndDate) {
				filtered = append(filtered, ex)
			}
		}
	} else if strings.Contains(lowerMessage, "this week") {
		timeframeDesc = "This Week"
		weekEnd := now.AddDate(0, 0, 7)
		for _, ex := range exercises {
			if ex.StartDate.Before(weekEnd) && ex.EndDate.After(now) {
				filtered = append(filtered, ex)
			}
		}
	} else if strings.Contains(lowerMessage, "next week") {
		timeframeDesc = "Next Week"
		weekStart := now.AddDate(0, 0, 7)
		weekEnd := now.AddDate(0, 0, 14)
		for _, ex := range exercises {
			if ex.StartDate.Before(weekEnd) && ex.EndDate.After(weekStart) {
				filtered = append(filtered, ex)
			}
		}
	} else if strings.Contains(lowerMessage, "this month") {
		timeframeDesc = "This Month"
		monthEnd := now.AddDate(0, 1, 0)
		for _, ex := range exercises {
			if ex.StartDate.Before(monthEnd) && ex.EndDate.After(now) {
				filtered = append(filtered, ex)
			}
		}
	} else if strings.Contains(lowerMessage, "upcoming") {
		timeframeDesc = "Upcoming"
		for _, ex := range exercises {
			if ex.StartDate.After(now) {
				filtered = append(filtered, ex)
			}
		}
	}
	
	if len(filtered) == 0 {
		return fmt.Sprintf("No exercises found for %s.", timeframeDesc)
	}
	
	reply := fmt.Sprintf("📅 **Exercises for %s:**\n\n", timeframeDesc)
	for _, ex := range filtered {
		status := getExerciseStatus(ex)
		reply += fmt.Sprintf("%s **%s** (ID: %d)\n", status, ex.Name, ex.ID)
		reply += fmt.Sprintf("   📆 %s to %s\n\n", 
			ex.StartDate.Format("Jan 2"), 
			ex.EndDate.Format("Jan 2"))
	}
	
	return reply
}