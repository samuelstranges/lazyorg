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
	// Create start and end of day in local timezone to avoid timezone issues
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.AddDate(0, 0, 1)

	var debugInfo string
	if database.DebugMode {
		// DEBUG: Log the query parameters
		debugInfo = fmt.Sprintf("GET_EVENTS_BY_DATE DEBUG:\n")
		debugInfo += fmt.Sprintf("  Query Date: %s\n", date.Format("2006-01-02 15:04:05"))
		debugInfo += fmt.Sprintf("  Start of Day: %s (Unix: %d)\n", startOfDay.Format("2006-01-02 15:04:05"), startOfDay.Unix())
		debugInfo += fmt.Sprintf("  End of Day: %s (Unix: %d)\n", endOfDay.Format("2006-01-02 15:04:05"), endOfDay.Unix())
	}
	
	rows, err := database.db.Query(`
        SELECT * FROM events WHERE time >= ? AND time < ?`,
		startOfDay.Format("2006-01-02 15:04:05"),
		endOfDay.Format("2006-01-02 15:04:05"),
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