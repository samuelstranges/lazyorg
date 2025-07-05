package database

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/samuelstranges/chronos/internal/calendar"
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


	return nil
}


// CheckEventOverlap checks if a new event would overlap with any existing events
// Returns true if there's an overlap, false if no overlap
func (database *Database) CheckEventOverlap(newEvent calendar.Event, excludeEventId ...int) (bool, error) {
	// Calculate new event's time range
	newStartTime := newEvent.Time
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
			
			existingStartTime := existingEvent.Time
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
		
		existingStartTime := existingEvent.Time
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
			os.WriteFile("/tmp/chronos_debug.txt", []byte(debugInfo), 0644)
			return true, nil // Found an overlap
		}
	}
	
	debugInfo += fmt.Sprintf("  RESULT: NO OVERLAP\n")
	os.WriteFile("/tmp/chronos_debug.txt", []byte(debugInfo), 0644)
	
	return false, nil // No overlap found
}


func (database *Database) CloseDatabase() error {
	if database.db == nil {
		return nil
	}

	return database.db.Close()
}
