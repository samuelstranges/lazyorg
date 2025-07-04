package database

import (
	"github.com/samuelstranges/chronos/internal/calendar"
)

// AddEvent inserts a new event into the database
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

// DeleteEventById removes an event by its ID
func (database *Database) DeleteEventById(id int) error {
	_, err := database.db.Exec("DELETE FROM events WHERE id = ?", id)
	return err
}

// DeleteEventsByName removes all events with a specific name
func (database *Database) DeleteEventsByName(name string) error {
	_, err := database.db.Exec("DELETE FROM events WHERE name = ?", name)
	return err
}

// UpdateEventById updates an existing event by its ID
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

// UpdateEventByName is a placeholder function (currently unused)
func (database *Database) UpdateEventByName(name string) error {
	return nil
}