package database

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/jroimartin/gocui"
)

// GetEventById retrieves a single event by its ID
func (database *Database) GetEventById(id int) (*calendar.Event, error) {
	rows, err := database.db.Query(`
        SELECT * FROM events WHERE id = ?`,
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var event calendar.Event
		var colorInt int
		if err := rows.Scan(
			&event.Id,
			&event.Name,
			&event.Description,
			&event.Location,
			&event.Time,
			&event.DurationHour,
			&event.FrequencyDay,
			&event.Occurence,
			&colorInt,
		); err != nil {
			return nil, err
		}
		if colorInt == 0 {
			event.Color = calendar.GenerateColorFromName(event.Name)
		} else {
			event.Color = gocui.Attribute(colorInt)
		}
		return &event, nil
	}

	return nil, nil
}

// GetEventsByDate retrieves all events for a specific date
func (database *Database) GetEventsByDate(date time.Time) ([]*calendar.Event, error) {
	// Create start and end of day in the input date's timezone, then convert to UTC for database comparison
	// This ensures we find all UTC-stored events that fall within the local day
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.AddDate(0, 0, 1)
	
	// Convert to UTC for database comparison since events are stored in UTC
	startOfDayUTC := startOfDay.UTC()
	endOfDayUTC := endOfDay.UTC()

	var debugInfo string
	if database.DebugMode {
		// DEBUG: Log the query parameters
		debugInfo = fmt.Sprintf("GET_EVENTS_BY_DATE DEBUG:\n")
		debugInfo += fmt.Sprintf("  Query Date: %s\n", date.Format("2006-01-02 15:04:05"))
		debugInfo += fmt.Sprintf("  Start of Day (Local): %s (Unix: %d)\n", startOfDay.Format("2006-01-02 15:04:05"), startOfDay.Unix())
		debugInfo += fmt.Sprintf("  End of Day (Local): %s (Unix: %d)\n", endOfDay.Format("2006-01-02 15:04:05"), endOfDay.Unix())
		debugInfo += fmt.Sprintf("  Start of Day (UTC): %s (Unix: %d)\n", startOfDayUTC.Format("2006-01-02 15:04:05"), startOfDayUTC.Unix())
		debugInfo += fmt.Sprintf("  End of Day (UTC): %s (Unix: %d)\n", endOfDayUTC.Format("2006-01-02 15:04:05"), endOfDayUTC.Unix())
	}
	
	rows, err := database.db.Query(`
        SELECT * FROM events WHERE time >= ? AND time < ?`,
		startOfDayUTC.Format("2006-01-02 15:04:05"),
		endOfDayUTC.Format("2006-01-02 15:04:05"),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*calendar.Event
	for rows.Next() {
		var event calendar.Event
		var colorInt int

		if err := rows.Scan(
			&event.Id,
			&event.Name,
			&event.Description,
			&event.Location,
			&event.Time,
			&event.DurationHour,
			&event.FrequencyDay,
			&event.Occurence,
			&colorInt,
		); err != nil {
			return nil, err
		}

		if database.DebugMode {
			// DEBUG: Log each event found
			debugInfo += fmt.Sprintf("  Found Event: %s at %s (Unix: %d)\n", event.Name, event.Time.Format("2006-01-02 15:04:05"), event.Time.Unix())
		}

		if colorInt == 0 {
			event.Color = calendar.GenerateColorFromName(event.Name)
		} else {
			event.Color = gocui.Attribute(colorInt)
		}
		events = append(events, &event)
	}

	// Write debug info only if in debug mode
	if database.DebugMode {
		os.WriteFile("/tmp/chronos_getevents_debug.txt", []byte(debugInfo), 0644)
	}

	return events, nil
}

// GetEventsByName retrieves all events with a specific name
func (database *Database) GetEventsByName(name string) ([]*calendar.Event, error) {
	rows, err := database.db.Query(`
        SELECT * FROM events WHERE name = ?`,
		name,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*calendar.Event
	for rows.Next() {
		var event calendar.Event
		var colorInt int

		if err := rows.Scan(
			&event.Id,
			&event.Name,
			&event.Description,
			&event.Location,
			&event.Time,
			&event.DurationHour,
			&event.FrequencyDay,
			&event.Occurence,
			&colorInt,
		); err != nil {
			return nil, err
		}

		if colorInt == 0 {
			event.Color = calendar.GenerateColorFromName(event.Name)
		} else {
			event.Color = gocui.Attribute(colorInt)
		}
		events = append(events, &event)
	}

	return events, nil
}

// GetAllEvents returns all events sorted by time
func (database *Database) GetAllEvents() ([]*calendar.Event, error) {
	rows, err := database.db.Query(`
        SELECT * FROM events ORDER BY time ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*calendar.Event
	for rows.Next() {
		var event calendar.Event
		var colorInt int

		if err := rows.Scan(
			&event.Id,
			&event.Name,
			&event.Description,
			&event.Location,
			&event.Time,
			&event.DurationHour,
			&event.FrequencyDay,
			&event.Occurence,
			&colorInt,
		); err != nil {
			return nil, err
		}

		if colorInt == 0 {
			event.Color = calendar.GenerateColorFromName(event.Name)
		} else {
			event.Color = gocui.Attribute(colorInt)
		}
		events = append(events, &event)
	}

	return events, nil
}

// SearchEvents searches for events by name, description, or location across all events
func (database *Database) SearchEvents(query string) ([]*calendar.Event, error) {
	query = strings.ToLower(query)
	searchPattern := "%" + query + "%"
	
	rows, err := database.db.Query(`
        SELECT * FROM events 
        WHERE LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(location) LIKE ?
        ORDER BY time ASC`,
		searchPattern, searchPattern, searchPattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*calendar.Event
	for rows.Next() {
		var event calendar.Event
		var colorInt int

		if err := rows.Scan(
			&event.Id,
			&event.Name,
			&event.Description,
			&event.Location,
			&event.Time,
			&event.DurationHour,
			&event.FrequencyDay,
			&event.Occurence,
			&colorInt,
		); err != nil {
			return nil, err
		}

		if colorInt == 0 {
			event.Color = calendar.GenerateColorFromName(event.Name)
		} else {
			event.Color = gocui.Attribute(colorInt)
		}
		events = append(events, &event)
	}

	return events, nil
}

// SearchCriteria holds all search parameters
type SearchCriteria struct {
	Query     string
	StartDate string
	StartTime string
	EndDate   string
	EndTime   string
}

// SearchEventsWithFilters searches for events with text query and optional date/time filters
func (database *Database) SearchEventsWithFilters(criteria SearchCriteria) ([]*calendar.Event, error) {
	var queryParts []string
	var args []interface{}
	
	// Add text search if provided
	if criteria.Query != "" {
		query := strings.ToLower(criteria.Query)
		searchPattern := "%" + query + "%"
		queryParts = append(queryParts, "(LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(location) LIKE ?)")
		args = append(args, searchPattern, searchPattern, searchPattern)
	}
	
	// Parse and add date/time filters
	var startDateTime, endDateTime *time.Time
	
	// Parse start date/time
	if criteria.StartDate != "" {
		startDate := criteria.StartDate
		// Handle 't' shortcut for today
		if startDate == "t" {
			startDate = time.Now().Format("20060102")
		}
		
		startTime := "00:00"
		if criteria.StartTime != "" {
			startTime = criteria.StartTime
		}
		if parsedStart, err := time.ParseInLocation("20060102 15:04", startDate+" "+startTime, time.Local); err == nil {
			startDateTime = &parsedStart
		}
	}
	
	// Parse end date/time
	if criteria.EndDate != "" {
		endDate := criteria.EndDate
		// Handle 't' shortcut for today
		if endDate == "t" {
			endDate = time.Now().Format("20060102")
		}
		
		endTime := "23:59"
		if criteria.EndTime != "" {
			endTime = criteria.EndTime
		}
		if parsedEnd, err := time.ParseInLocation("20060102 15:04", endDate+" "+endTime, time.Local); err == nil {
			// When doing exact time matches, add 59 seconds to be inclusive of the whole minute
			if criteria.EndTime != "" && criteria.StartTime == criteria.EndTime {
				parsedEnd = parsedEnd.Add(59 * time.Second)
			}
			endDateTime = &parsedEnd
		}
	}
	
	// Add date/time range filters
	if startDateTime != nil {
		queryParts = append(queryParts, "time >= ?")
		args = append(args, startDateTime.Format("2006-01-02 15:04:05"))
	}
	
	if endDateTime != nil {
		queryParts = append(queryParts, "time <= ?")
		args = append(args, endDateTime.Format("2006-01-02 15:04:05"))
	}
	
	// If no criteria provided, return empty results
	if len(queryParts) == 0 {
		return []*calendar.Event{}, nil
	}
	
	// Build the final query
	sqlQuery := "SELECT * FROM events WHERE " + strings.Join(queryParts, " AND ") + " ORDER BY time ASC"
	
	rows, err := database.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*calendar.Event
	for rows.Next() {
		var event calendar.Event
		var colorInt int

		if err := rows.Scan(
			&event.Id,
			&event.Name,
			&event.Description,
			&event.Location,
			&event.Time,
			&event.DurationHour,
			&event.FrequencyDay,
			&event.Occurence,
			&colorInt,
		); err != nil {
			return nil, err
		}

		if colorInt == 0 {
			event.Color = calendar.GenerateColorFromName(event.Name)
		} else {
			event.Color = gocui.Attribute(colorInt)
		}
		events = append(events, &event)
	}

	return events, nil
}
// GetEventsByMonth retrieves all events for a specific month
func (database *Database) GetEventsByMonth(year int, month time.Month) ([]*calendar.Event, error) {
	// Create start and end of month in local timezone, then convert to UTC for database comparison
	startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)
	
	// Convert to UTC for database comparison since events are stored in UTC
	startOfMonthUTC := startOfMonth.UTC()
	endOfMonthUTC := endOfMonth.UTC()
	
	rows, err := database.db.Query(`
        SELECT * FROM events WHERE time >= ? AND time < ? ORDER BY time ASC`,
		startOfMonthUTC.Format("2006-01-02 15:04:05"),
		endOfMonthUTC.Format("2006-01-02 15:04:05"),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*calendar.Event
	for rows.Next() {
		var event calendar.Event
		var colorInt int

		if err := rows.Scan(
			&event.Id,
			&event.Name,
			&event.Description,
			&event.Location,
			&event.Time,
			&event.DurationHour,
			&event.FrequencyDay,
			&event.Occurence,
			&colorInt,
		); err != nil {
			return nil, err
		}

		if colorInt == 0 {
			event.Color = calendar.GenerateColorFromName(event.Name)
		} else {
			event.Color = gocui.Attribute(colorInt)
		}
		events = append(events, &event)
	}

	return events, nil
}
// GetEventsByDateRange retrieves all events within a specific date range
// The startDate and endDate should be in local timezone and will be converted to UTC for database queries
func (database *Database) GetEventsByDateRange(startDate, endDate time.Time) ([]*calendar.Event, error) {
	// Convert input dates to UTC for database comparison since events are stored in UTC
	// This ensures we find all UTC-stored events that fall within the local time range
	startDateUTC := startDate.UTC()
	endDateUTC := endDate.UTC()
	
	rows, err := database.db.Query(`
        SELECT * FROM events WHERE time >= ? AND time < ? ORDER BY time ASC`,
		startDateUTC.Format("2006-01-02 15:04:05"),
		endDateUTC.Format("2006-01-02 15:04:05"),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*calendar.Event
	for rows.Next() {
		var event calendar.Event
		var colorInt int

		if err := rows.Scan(
			&event.Id,
			&event.Name,
			&event.Description,
			&event.Location,
			&event.Time,
			&event.DurationHour,
			&event.FrequencyDay,
			&event.Occurence,
			&colorInt,
		); err != nil {
			return nil, err
		}

		if colorInt == 0 {
			event.Color = calendar.GenerateColorFromName(event.Name)
		} else {
			event.Color = gocui.Attribute(colorInt)
		}
		events = append(events, &event)
	}

	return events, nil
}