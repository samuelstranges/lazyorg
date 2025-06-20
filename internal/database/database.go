package database

import (
	"database/sql"
	"strings"
	"time"

	"github.com/HubertBel/lazyorg/internal/calendar"
	"github.com/jroimartin/gocui"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
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

		if colorInt == 0 {
			event.Color = calendar.GenerateColorFromName(event.Name)
		} else {
			event.Color = gocui.Attribute(colorInt)
		}
		events = append(events, &event)
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

func (database *Database) CloseDatabase() error {
	if database.db == nil {
		return nil
	}

	return database.db.Close()
}
