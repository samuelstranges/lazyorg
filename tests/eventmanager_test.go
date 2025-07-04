package tests

import (
	"testing"
	"time"

	"github.com/HubertBel/lazyorg/internal/calendar"
	"github.com/HubertBel/lazyorg/internal/database"
	"github.com/HubertBel/lazyorg/internal/eventmanager"
)

// setupTestDB creates an in-memory database for testing
func setupTestDB(t *testing.T) *database.Database {
	db := &database.Database{}
	err := db.InitDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	return db
}

// setupTestEventManager creates an event manager with test database
func setupTestEventManager(t *testing.T) (*eventmanager.EventManager, *database.Database) {
	db := setupTestDB(t)
	em := eventmanager.NewEventManager(db)
	return em, db
}

// createTestEvent creates a test event with given parameters
func createTestEvent(name, description, location string, timeOffset time.Duration) calendar.Event {
	testTime := time.Now().Add(timeOffset)
	return calendar.Event{
		Name:         name,
		Description:  description,
		Location:     location,
		Time:         testTime,
		DurationHour: 1.0,
	}
}

func TestBulkDeleteUndoRedo(t *testing.T) {
	em, db := setupTestEventManager(t)
	defer db.CloseDatabase()

	// Create test events
	event1 := createTestEvent("Test Event", "First test event", "Location 1", 0)
	event2 := createTestEvent("Test Event", "Second test event", "Location 2", time.Hour)
	event3 := createTestEvent("Different Event", "Different event", "Location 3", 2*time.Hour)

	// Add events
	_, success := em.AddEvent(event1)
	if !success {
		t.Fatalf("Failed to add event1")
	}

	_, success = em.AddEvent(event2)
	if !success {
		t.Fatalf("Failed to add event2")
	}

	_, success = em.AddEvent(event3)
	if !success {
		t.Fatalf("Failed to add event3")
	}

	// Verify initial count
	events, err := em.GetEventsByName("Test Event")
	if err != nil {
		t.Fatalf("Failed to get events by name: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("Expected 2 'Test Event' events, got %d", len(events))
	}

	// Perform bulk delete
	err = em.DeleteEventsByName("Test Event")
	if err != nil {
		t.Fatalf("Failed to delete events by name: %v", err)
	}

	// Verify count after delete
	events, err = em.GetEventsByName("Test Event")
	if err != nil {
		t.Fatalf("Failed to get events after delete: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("Expected 0 events after delete, got %d", len(events))
	}

	// Test undo functionality
	if !em.CanUndo() {
		t.Error("Should be able to undo after bulk delete")
	}

	undoDesc := em.GetUndoDescription()
	if undoDesc == "" {
		t.Error("Undo description should not be empty")
	}

	err = em.Undo()
	if err != nil {
		t.Fatalf("Undo failed: %v", err)
	}

	// Verify count after undo
	events, err = em.GetEventsByName("Test Event")
	if err != nil {
		t.Fatalf("Failed to get events after undo: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("Expected 2 events after undo, got %d", len(events))
	}

	// Test redo functionality
	if !em.CanRedo() {
		t.Error("Should be able to redo after undo")
	}

	redoDesc := em.GetRedoDescription()
	if redoDesc == "" {
		t.Error("Redo description should not be empty")
	}

	err = em.Redo()
	if err != nil {
		t.Fatalf("Redo failed: %v", err)
	}

	// Verify final count after redo
	events, err = em.GetEventsByName("Test Event")
	if err != nil {
		t.Fatalf("Failed to get events after redo: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("Expected 0 events after redo, got %d", len(events))
	}

	// Verify that "Different Event" was not affected
	differentEvents, err := em.GetEventsByName("Different Event")
	if err != nil {
		t.Fatalf("Failed to get different events: %v", err)
	}
	if len(differentEvents) != 1 {
		t.Errorf("Expected 1 'Different Event', got %d", len(differentEvents))
	}
}

func TestAddEventUndoRedo(t *testing.T) {
	em, db := setupTestEventManager(t)
	defer db.CloseDatabase()

	event := createTestEvent("Add Test", "Test add event", "Test Location", 0)

	// Add event
	_, success := em.AddEvent(event)
	if !success {
		t.Fatalf("Failed to add event")
	}

	// Verify event was added
	events, err := em.GetEventsByName("Add Test")
	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("Expected 1 event after add, got %d", len(events))
	}

	// Test undo add
	err = em.Undo()
	if err != nil {
		t.Fatalf("Failed to undo add: %v", err)
	}

	events, err = em.GetEventsByName("Add Test")
	if err != nil {
		t.Fatalf("Failed to get events after undo: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("Expected 0 events after undo add, got %d", len(events))
	}

	// Test redo add
	err = em.Redo()
	if err != nil {
		t.Fatalf("Failed to redo add: %v", err)
	}

	events, err = em.GetEventsByName("Add Test")
	if err != nil {
		t.Fatalf("Failed to get events after redo: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("Expected 1 event after redo add, got %d", len(events))
	}

	// Verify the event content is correct
	if events[0].Description != "Test add event" {
		t.Errorf("Expected description 'Test add event', got '%s'", events[0].Description)
	}
}

func TestDeleteEventUndoRedo(t *testing.T) {
	em, db := setupTestEventManager(t)
	defer db.CloseDatabase()

	event := createTestEvent("Delete Test", "Test delete event", "Test Location", 0)

	// Add event first
	addedEvent, success := em.AddEvent(event)
	if !success {
		t.Fatalf("Failed to add event")
	}

	// Delete the event
	err := em.DeleteEvent(addedEvent.Id)
	if err != nil {
		t.Fatalf("Failed to delete event: %v", err)
	}

	// Verify event was deleted
	events, err := em.GetEventsByName("Delete Test")
	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("Expected 0 events after delete, got %d", len(events))
	}

	// Test undo delete
	err = em.Undo()
	if err != nil {
		t.Fatalf("Failed to undo delete: %v", err)
	}

	events, err = em.GetEventsByName("Delete Test")
	if err != nil {
		t.Fatalf("Failed to get events after undo: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("Expected 1 event after undo delete, got %d", len(events))
	}

	// Test redo delete
	err = em.Redo()
	if err != nil {
		t.Fatalf("Failed to redo delete: %v", err)
	}

	events, err = em.GetEventsByName("Delete Test")
	if err != nil {
		t.Fatalf("Failed to get events after redo: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("Expected 0 events after redo delete, got %d", len(events))
	}
}

func TestUndoRedoStackLimits(t *testing.T) {
	em, db := setupTestEventManager(t)
	defer db.CloseDatabase()

	// Test that undo/redo stacks work properly
	if em.CanUndo() {
		t.Error("Should not be able to undo with empty stack")
	}

	if em.CanRedo() {
		t.Error("Should not be able to redo with empty stack")
	}

	// Add an event
	event := createTestEvent("Stack Test", "Test stack limits", "Test Location", 0)
	_, success := em.AddEvent(event)
	if !success {
		t.Fatalf("Failed to add event")
	}

	// Now should be able to undo
	if !em.CanUndo() {
		t.Error("Should be able to undo after adding event")
	}

	// But not redo yet
	if em.CanRedo() {
		t.Error("Should not be able to redo before any undo")
	}

	// Undo the add
	err := em.Undo()
	if err != nil {
		t.Fatalf("Failed to undo: %v", err)
	}

	// Now should be able to redo
	if !em.CanRedo() {
		t.Error("Should be able to redo after undo")
	}

	// New action should clear redo stack
	event2 := createTestEvent("New Action", "Clears redo stack", "Test Location", 0)
	_, success = em.AddEvent(event2)
	if !success {
		t.Fatalf("Failed to add second event")
	}

	// Redo stack should be cleared
	if em.CanRedo() {
		t.Error("Redo stack should be cleared after new action")
	}
}