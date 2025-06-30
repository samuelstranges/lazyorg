package database

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/HubertBel/lazyorg/internal/calendar"
	"github.com/jroimartin/gocui"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db        *sql.DB
	DebugMode bool
}

func (database *Database) InitDatabase(path string) error {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return err
	}
	database.db = db

	return database.createTables()
}

func (database *Database) createTables() error {
	_, err := database.db.Exec(`
        CREATE TABLE IF NOT EXISTS events (
        id INTEGER NOT NULL PRIMARY KEY,
        name TEXT NOT NULL,
        description TEXT,
        location TEXT,
        time DATETIME NOT NULL,
        duration REAL NOT NULL,
        frequency INTEGER,
        occurence INTEGER,
        color INTEGER DEFAULT 0
    )`)
	if err != nil {
		return err
	}

	// Add color column to existing tables if it doesn't exist
	_, err = database.db.Exec(`
        ALTER TABLE events ADD COLUMN color INTEGER DEFAULT 0
    `)
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return err
	}

	_, err = database.db.Exec(`
        CREATE TABLE IF NOT EXISTS notes (
        id INTEGER NOT NULL PRIMARY KEY,
        content TEXT NOT NULL,
        updated_at DATETIME NOT NULL
    )`)
	if err != nil {
		return err
	}

	return nil
}

func (database *Database) AddEvent(event calendar.Event) (int, error) {
	result, err := database.db.Exec(`
        INSERT INTO events (
            name, description, location, time, duration, frequency, occurence, color
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		event.Name,
		event.Description,
		event.Location,
		event.Time,
		event.DurationHour,
		event.FrequencyDay,
		event.Occurence,
		int(event.Color),
	)
	if err != nil {
		return -1, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}

	return int(id), err
}

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
		os.WriteFile("/tmp/lazyorg_getevents_debug.txt", []byte(debugInfo), 0644)
	}

	return events, nil
}

func (database *Database) DeleteEventById(id int) error {
	_, err := database.db.Exec("DELETE FROM events WHERE id = ?", id)
	return err
}

func (database *Database) DeleteEventsByName(name string) error {
	_, err := database.db.Exec("DELETE FROM events WHERE name = ?", name)
	return err
}

func (database *Database) UpdateEventById(id int, event *calendar.Event) error {
	_, err := database.db.Exec(
		`UPDATE events SET
            name = ?, 
            description = ?, 
            location = ?, 
            time = ?, 
            duration = ?, 
            frequency = ?, 
            occurence = ?,
            color = ?
        WHERE id = ?`,
		event.Name,
		event.Description,
		event.Location,
		event.Time,
		event.DurationHour,
		event.FrequencyDay,
		event.Occurence,
		int(event.Color),
		id,
	)

	return err
}

func (database *Database) UpdateEventByName(name string) error {
	return nil
}

func (database *Database) SaveNote(content string) error {
	_, err := database.db.Exec("DELETE FROM notes")
	if err != nil {
		return err
	}

	_, err = database.db.Exec(`INSERT INTO notes (
            content, updated_at
        ) VALUES (?, datetime('now'))`, content)

	return err
}

func (database *Database) GetLatestNote() (string, error) {
	var content string
	err := database.db.QueryRow(
		"SELECT content FROM notes ORDER BY updated_at DESC LIMIT 1",
	).Scan(&content)

	return content, err
}

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

// CheckEventOverlap checks if a new event would overlap with any existing events
// Returns true if there's an overlap, false if no overlap
func (database *Database) CheckEventOverlap(newEvent calendar.Event, excludeEventId ...int) (bool, error) {
	// Normalize new event's time to match GetEventsByDate timezone handling
	normalizedTime := time.Date(newEvent.Time.Year(), newEvent.Time.Month(), newEvent.Time.Day(), 
		newEvent.Time.Hour(), newEvent.Time.Minute(), newEvent.Time.Second(), 
		newEvent.Time.Nanosecond(), newEvent.Time.Location())
	
	// Calculate new event's time range using normalized time
	newStartTime := normalizedTime
	newEndTime := newStartTime.Add(time.Duration(newEvent.DurationHour * float64(time.Hour)))
	
	// Only log debug info if debug mode is enabled
	if !database.DebugMode {
		// Get all events for the same date
		existingEvents, err := database.GetEventsByDate(newEvent.Time)
		if err != nil {
			return false, err
		}
		
		// Check each existing event for overlap
		for _, existingEvent := range existingEvents {
			// Skip if this is the same event (for edits)
			if len(excludeEventId) > 0 && existingEvent.Id == excludeEventId[0] {
				continue
			}
			
			// Normalize existing event times to same timezone as new event
			existingStartTime := time.Date(existingEvent.Time.Year(), existingEvent.Time.Month(), existingEvent.Time.Day(), 
				existingEvent.Time.Hour(), existingEvent.Time.Minute(), existingEvent.Time.Second(), 
				existingEvent.Time.Nanosecond(), newEvent.Time.Location())
			existingEndTime := existingStartTime.Add(time.Duration(existingEvent.DurationHour * float64(time.Hour)))
			
			// Check for overlap: events overlap if one starts before the other ends
			// Adjacent events (one ends exactly when another starts) are allowed
			overlap := (newStartTime.Before(existingEndTime) && newEndTime.After(existingStartTime))
			
			if overlap {
				return true, nil // Found an overlap
			}
		}
		
		return false, nil // No overlap found
	}

	// DEBUG: Log overlap check details
	debugInfo := fmt.Sprintf("OVERLAP CHECK DEBUG:\n")
	debugInfo += fmt.Sprintf("  New Event: %s\n", newEvent.Name)
	debugInfo += fmt.Sprintf("  New Start: %s\n", newStartTime.Format("2006-01-02 15:04:05"))
	debugInfo += fmt.Sprintf("  New End: %s\n", newEndTime.Format("2006-01-02 15:04:05"))
	debugInfo += fmt.Sprintf("  New Duration (hours): %f\n", newEvent.DurationHour)
	debugInfo += fmt.Sprintf("  Duration calculation: %f * %d = %d nanoseconds\n", newEvent.DurationHour, int64(time.Hour), int64(newEvent.DurationHour * float64(time.Hour)))
	
	// Get all events for the same date
	debugInfo += fmt.Sprintf("  Calling GetEventsByDate with: %s\n", newEvent.Time.Format("2006-01-02 15:04:05"))
	existingEvents, err := database.GetEventsByDate(newEvent.Time)
	if err != nil {
		return false, err
	}
	
	debugInfo += fmt.Sprintf("  Found %d existing events:\n", len(existingEvents))
	
	// Check each existing event for overlap
	for _, existingEvent := range existingEvents {
		// Skip if this is the same event (for edits)
		if len(excludeEventId) > 0 && existingEvent.Id == excludeEventId[0] {
			continue
		}
		
		// Normalize existing event times to same timezone as new event
		existingStartTime := time.Date(existingEvent.Time.Year(), existingEvent.Time.Month(), existingEvent.Time.Day(), 
			existingEvent.Time.Hour(), existingEvent.Time.Minute(), existingEvent.Time.Second(), 
			existingEvent.Time.Nanosecond(), newEvent.Time.Location())
		existingEndTime := existingStartTime.Add(time.Duration(existingEvent.DurationHour * float64(time.Hour)))
		
		debugInfo += fmt.Sprintf("    Existing Event: %s\n", existingEvent.Name)
		debugInfo += fmt.Sprintf("      Start: %s (TZ: %s, Unix: %d)\n", existingStartTime.Format("2006-01-02 15:04:05"), existingStartTime.Location().String(), existingStartTime.Unix())
		debugInfo += fmt.Sprintf("      End: %s (TZ: %s, Unix: %d)\n", existingEndTime.Format("2006-01-02 15:04:05"), existingEndTime.Location().String(), existingEndTime.Unix())
		debugInfo += fmt.Sprintf("      Duration (hours): %f\n", existingEvent.DurationHour)
		
		// Check for overlap: events overlap if one starts before the other ends
		// Adjacent events (one ends exactly when another starts) are allowed
		newStartsBeforeExistingEnds := newStartTime.Before(existingEndTime)
		newEndsAfterExistingStarts := newEndTime.After(existingStartTime)
		overlap := newStartsBeforeExistingEnds && newEndsAfterExistingStarts
		
		debugInfo += fmt.Sprintf("      NewStart.Before(ExistingEnd): %v (%s < %s)\n", newStartsBeforeExistingEnds, newStartTime.Format("2006-01-02 15:04:05"), existingEndTime.Format("2006-01-02 15:04:05"))
		debugInfo += fmt.Sprintf("      NewEnd.After(ExistingStart): %v (%s > %s)\n", newEndsAfterExistingStarts, newEndTime.Format("2006-01-02 15:04:05"), existingStartTime.Format("2006-01-02 15:04:05"))
		debugInfo += fmt.Sprintf("      Raw comparison: %d vs %d (Unix seconds)\n", newStartTime.Unix(), existingEndTime.Unix())
		debugInfo += fmt.Sprintf("      OVERLAP: %v\n", overlap)
		
		if overlap {
			debugInfo += fmt.Sprintf("  RESULT: OVERLAP DETECTED\n")
			os.WriteFile("/tmp/lazyorg_debug.txt", []byte(debugInfo), 0644)
			return true, nil // Found an overlap
		}
	}
	
	debugInfo += fmt.Sprintf("  RESULT: NO OVERLAP\n")
	os.WriteFile("/tmp/lazyorg_debug.txt", []byte(debugInfo), 0644)
	
	return false, nil // No overlap found
}

func (database *Database) CloseDatabase() error {
	if database.db == nil {
		return nil
	}

	return database.db.Close()
}
