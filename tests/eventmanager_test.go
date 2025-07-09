package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/samuelstranges/chronos/internal/database"
	"github.com/samuelstranges/chronos/internal/eventmanager"
	"github.com/samuelstranges/chronos/internal/utils"
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

func TestEventOverlapDetection(t *testing.T) {
	em, db := setupTestEventManager(t)
	defer db.CloseDatabase()

	// Create base event: 10:00 AM - 11:00 AM
	baseEvent := createTestEvent("Base Event", "Original event", "Location", 0)
	baseEvent.Time = time.Date(2024, 1, 15, 10, 0, 0, 0, time.Local)
	baseEvent.DurationHour = 1.0

	_, success := em.AddEvent(baseEvent)
	if !success {
		t.Fatalf("Failed to add base event")
	}

	// Test 1: Exact overlap - should fail
	overlapEvent1 := createTestEvent("Overlap 1", "Exact same time", "Location", 0)
	overlapEvent1.Time = time.Date(2024, 1, 15, 10, 0, 0, 0, time.Local)
	overlapEvent1.DurationHour = 1.0

	_, success = em.AddEvent(overlapEvent1)
	if success {
		t.Error("Expected failure for exact overlap, but succeeded")
	}

	// Test 2: Partial overlap (starts during existing event) - should fail
	overlapEvent2 := createTestEvent("Overlap 2", "Starts during existing", "Location", 0)
	overlapEvent2.Time = time.Date(2024, 1, 15, 10, 30, 0, 0, time.Local)
	overlapEvent2.DurationHour = 1.0

	_, success = em.AddEvent(overlapEvent2)
	if success {
		t.Error("Expected failure for partial overlap, but succeeded")
	}

	// Test 3: Encompassing overlap (longer event that contains existing) - should fail
	overlapEvent3 := createTestEvent("Overlap 3", "Encompasses existing", "Location", 0)
	overlapEvent3.Time = time.Date(2024, 1, 15, 9, 30, 0, 0, time.Local)
	overlapEvent3.DurationHour = 2.0

	_, success = em.AddEvent(overlapEvent3)
	if success {
		t.Error("Expected failure for encompassing overlap, but succeeded")
	}

	// Test 4: Adjacent events (touching) - should succeed
	adjacentEvent := createTestEvent("Adjacent", "Starts when other ends", "Location", 0)
	adjacentEvent.Time = time.Date(2024, 1, 15, 11, 0, 0, 0, time.Local)
	adjacentEvent.DurationHour = 1.0

	_, success = em.AddEvent(adjacentEvent)
	if !success {
		t.Error("Expected success for adjacent events, but failed")
	}

	// Test 5: Non-overlapping event - should succeed
	nonOverlapEvent := createTestEvent("Non-overlap", "Different time", "Location", 0)
	nonOverlapEvent.Time = time.Date(2024, 1, 15, 13, 0, 0, 0, time.Local)
	nonOverlapEvent.DurationHour = 1.0

	_, success = em.AddEvent(nonOverlapEvent)
	if !success {
		t.Error("Expected success for non-overlapping event, but failed")
	}

	// Test 6: Different day - should succeed
	differentDayEvent := createTestEvent("Different Day", "Same time, different day", "Location", 0)
	differentDayEvent.Time = time.Date(2024, 1, 16, 10, 0, 0, 0, time.Local)
	differentDayEvent.DurationHour = 1.0

	_, success = em.AddEvent(differentDayEvent)
	if !success {
		t.Error("Expected success for different day event, but failed")
	}
}

func TestUpdateEventOverlapPrevention(t *testing.T) {
	em, db := setupTestEventManager(t)
	defer db.CloseDatabase()

	// Create two events
	event1 := createTestEvent("Event 1", "First event", "Location", 0)
	event1.Time = time.Date(2024, 1, 15, 10, 0, 0, 0, time.Local)
	event1.DurationHour = 1.0

	event2 := createTestEvent("Event 2", "Second event", "Location", 0)
	event2.Time = time.Date(2024, 1, 15, 12, 0, 0, 0, time.Local)
	event2.DurationHour = 1.0

	_, success := em.AddEvent(event1)
	if !success {
		t.Fatalf("Failed to add first event")
	}

	addedEvent2, success := em.AddEvent(event2)
	if !success {
		t.Fatalf("Failed to add second event")
	}

	// Try to update event2 to overlap with event1 - should fail
	updatedEvent2 := *addedEvent2
	updatedEvent2.Time = time.Date(2024, 1, 15, 10, 30, 0, 0, time.Local)
	updatedEvent2.DurationHour = 1.0

	success = em.UpdateEvent(addedEvent2.Id, &updatedEvent2)
	if success {
		t.Error("Expected failure when updating to create overlap, but succeeded")
	}

	// Try to update event2 to non-overlapping time - should succeed
	updatedEvent2.Time = time.Date(2024, 1, 15, 14, 0, 0, 0, time.Local)
	success = em.UpdateEvent(addedEvent2.Id, &updatedEvent2)
	if !success {
		t.Error("Expected success when updating to non-overlapping time, but failed")
	}

	// Verify the update worked
	events, err := em.GetEventsByName("Event 2")
	if err != nil {
		t.Fatalf("Failed to retrieve updated event: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	if events[0].Time.Hour() != 14 {
		t.Errorf("Expected event time to be 14:00, got %d:00", events[0].Time.Hour())
	}
}

func TestUpdateEventUndoRedo(t *testing.T) {
	em, db := setupTestEventManager(t)
	defer db.CloseDatabase()

	// Create and add an event
	event := createTestEvent("Update Test", "Original description", "Original Location", 0)
	event.Time = time.Date(2024, 1, 15, 10, 0, 0, 0, time.Local)
	event.DurationHour = 1.0

	addedEvent, success := em.AddEvent(event)
	if !success {
		t.Fatalf("Failed to add event")
	}

	// Update the event
	updatedEvent := *addedEvent
	updatedEvent.Description = "Updated description"
	updatedEvent.Location = "Updated Location"
	updatedEvent.Time = time.Date(2024, 1, 15, 11, 0, 0, 0, time.Local)
	updatedEvent.DurationHour = 2.0

	success = em.UpdateEvent(addedEvent.Id, &updatedEvent)
	if !success {
		t.Fatalf("Failed to update event")
	}

	// Verify the update
	events, err := em.GetEventsByName("Update Test")
	if err != nil {
		t.Fatalf("Failed to get updated event: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	if events[0].Description != "Updated description" {
		t.Errorf("Expected description 'Updated description', got '%s'", events[0].Description)
	}
	if events[0].Location != "Updated Location" {
		t.Errorf("Expected location 'Updated Location', got '%s'", events[0].Location)
	}
	if events[0].Time.Hour() != 11 {
		t.Errorf("Expected time 11:00, got %d:00", events[0].Time.Hour())
	}
	if events[0].DurationHour != 2.0 {
		t.Errorf("Expected duration 2.0, got %f", events[0].DurationHour)
	}

	// Test undo
	err = em.Undo()
	if err != nil {
		t.Fatalf("Failed to undo update: %v", err)
	}

	// Verify original state is restored
	events, err = em.GetEventsByName("Update Test")
	if err != nil {
		t.Fatalf("Failed to get event after undo: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("Expected 1 event after undo, got %d", len(events))
	}
	if events[0].Description != "Original description" {
		t.Errorf("Expected original description after undo, got '%s'", events[0].Description)
	}
	if events[0].Location != "Original Location" {
		t.Errorf("Expected original location after undo, got '%s'", events[0].Location)
	}
	if events[0].Time.Hour() != 10 {
		t.Errorf("Expected original time 10:00 after undo, got %d:00", events[0].Time.Hour())
	}
	if events[0].DurationHour != 1.0 {
		t.Errorf("Expected original duration 1.0 after undo, got %f", events[0].DurationHour)
	}

	// Test redo
	err = em.Redo()
	if err != nil {
		t.Fatalf("Failed to redo update: %v", err)
	}

	// Verify updated state is restored
	events, err = em.GetEventsByName("Update Test")
	if err != nil {
		t.Fatalf("Failed to get event after redo: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("Expected 1 event after redo, got %d", len(events))
	}
	if events[0].Description != "Updated description" {
		t.Errorf("Expected updated description after redo, got '%s'", events[0].Description)
	}
}

func TestDatabaseDirectOperations(t *testing.T) {
	db := setupTestDB(t)
	defer db.CloseDatabase()

	// Test GetEventsByDate
	testDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.Local)
	
	// Initially should be empty
	events, err := db.GetEventsByDate(testDate)
	if err != nil {
		t.Fatalf("Failed to get events by date: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("Expected 0 events initially, got %d", len(events))
	}

	// Add events on different dates
	event1 := createTestEvent("Event 1", "First event", "Location", 0)
	event1.Time = time.Date(2024, 1, 15, 10, 0, 0, 0, time.Local)
	
	event2 := createTestEvent("Event 2", "Second event", "Location", 0)
	event2.Time = time.Date(2024, 1, 15, 12, 0, 0, 0, time.Local)
	
	event3 := createTestEvent("Event 3", "Different day", "Location", 0)
	event3.Time = time.Date(2024, 1, 16, 10, 0, 0, 0, time.Local)

	_, err = db.AddEvent(event1)
	if err != nil {
		t.Fatalf("Failed to add event1: %v", err)
	}
	
	_, err = db.AddEvent(event2)
	if err != nil {
		t.Fatalf("Failed to add event2: %v", err)
	}
	
	_, err = db.AddEvent(event3)
	if err != nil {
		t.Fatalf("Failed to add event3: %v", err)
	}

	// Test GetEventsByDate for 2024-01-15
	events, err = db.GetEventsByDate(testDate)
	if err != nil {
		t.Fatalf("Failed to get events by date: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("Expected 2 events for 2024-01-15, got %d", len(events))
	}

	// Test GetAllEvents
	allEvents, err := db.GetAllEvents()
	if err != nil {
		t.Fatalf("Failed to get all events: %v", err)
	}
	if len(allEvents) != 3 {
		t.Errorf("Expected 3 total events, got %d", len(allEvents))
	}

	// Test SearchEvents
	searchResults, err := db.SearchEvents("First")
	if err != nil {
		t.Fatalf("Failed to search events: %v", err)
	}
	if len(searchResults) != 1 {
		t.Errorf("Expected 1 search result for 'First', got %d", len(searchResults))
	}
	if searchResults[0].Name != "Event 1" {
		t.Errorf("Expected search result name 'Event 1', got '%s'", searchResults[0].Name)
	}

	// Test search with no results
	searchResults, err = db.SearchEvents("NonExistent")
	if err != nil {
		t.Fatalf("Failed to search for non-existent events: %v", err)
	}
	if len(searchResults) != 0 {
		t.Errorf("Expected 0 search results for 'NonExistent', got %d", len(searchResults))
	}
}

func TestErrorHandling(t *testing.T) {
	em, db := setupTestEventManager(t)
	defer db.CloseDatabase()

	// Test updating non-existent event - should fail because event doesn't exist
	nonExistentEvent := createTestEvent("Non-existent", "This doesn't exist", "Location", 0)
	success := em.UpdateEvent(99999, &nonExistentEvent)
	if success {
		t.Error("Expected failure when updating non-existent event, but succeeded")
	}

	// Test deleting non-existent event - should now return proper error
	err := em.DeleteEvent(99999)
	if err == nil {
		t.Error("Expected error when deleting non-existent event, but got nil")
	}
	expectedMsg := "event not found: cannot delete non-existent event"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	// Test getting non-existent event - should return nil, nil (not an error)
	event, err := em.GetEventById(99999)
	if err != nil {
		t.Errorf("Expected no error when getting non-existent event, but got: %v", err)
	}
	if event != nil {
		t.Error("Expected nil event when getting non-existent event")
	}

	// Test undo with empty stack - this should be safe and not crash
	if em.CanUndo() {
		t.Error("Expected CanUndo to return false with empty stack")
	}

	// Test redo with empty stack - this should be safe and not crash
	if em.CanRedo() {
		t.Error("Expected CanRedo to return false with empty stack")
	}
}

func TestLargeStackOperations(t *testing.T) {
	em, db := setupTestEventManager(t)
	defer db.CloseDatabase()

	// Add many events to test stack limits
	const numEvents = 60 // More than the 50 limit
	
	for i := 0; i < numEvents; i++ {
		event := createTestEvent(fmt.Sprintf("Event %d", i), fmt.Sprintf("Description %d", i), "Location", time.Duration(i)*time.Hour)
		_, success := em.AddEvent(event)
		if !success {
			t.Fatalf("Failed to add event %d", i)
		}
	}

	// Count how many undos we can do (should be limited to 50)
	undoCount := 0
	for em.CanUndo() {
		err := em.Undo()
		if err != nil {
			t.Fatalf("Failed to undo at count %d: %v", undoCount, err)
		}
		undoCount++
		if undoCount > 60 { // Safety check to prevent infinite loop
			break
		}
	}

	if undoCount != 50 {
		t.Errorf("Expected to undo 50 operations (stack limit), but undid %d", undoCount)
	}
}

func TestEventColorGeneration(t *testing.T) {
	_, db := setupTestEventManager(t)
	defer db.CloseDatabase()

	// Test that events with same name get same color
	event1 := createTestEvent("Same Name", "First event", "Location", 0)
	event2 := createTestEvent("Same Name", "Second event", "Location", time.Hour)

	id1, err := db.AddEvent(event1)
	if err != nil {
		t.Fatalf("Failed to add event1: %v", err)
	}

	id2, err := db.AddEvent(event2)
	if err != nil {
		t.Fatalf("Failed to add event2: %v", err)
	}

	// Retrieve events and check colors
	retrievedEvent1, err := db.GetEventById(id1)
	if err != nil {
		t.Fatalf("Failed to retrieve event1: %v", err)
	}

	retrievedEvent2, err := db.GetEventById(id2)
	if err != nil {
		t.Fatalf("Failed to retrieve event2: %v", err)
	}

	if retrievedEvent1.Color != retrievedEvent2.Color {
		t.Errorf("Expected events with same name to have same color, but got %v and %v", retrievedEvent1.Color, retrievedEvent2.Color)
	}

	// Test that events with different names get different colors (most of the time)
	event3 := createTestEvent("Different Name", "Third event", "Location", 2*time.Hour)
	id3, err := db.AddEvent(event3)
	if err != nil {
		t.Fatalf("Failed to add event3: %v", err)
	}

	retrievedEvent3, err := db.GetEventById(id3)
	if err != nil {
		t.Fatalf("Failed to retrieve event3: %v", err)
	}

	// Note: This test might occasionally fail due to hash collisions, but it's very unlikely
	if retrievedEvent1.Color == retrievedEvent3.Color {
		t.Logf("Warning: Events with different names got same color (rare hash collision)")
	}
}

func TestOptionalValidationFunctions(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
		testFunc func(string) bool
	}{
		// ValidateOptionalEventDate tests
		{"Empty date should be valid", "", true, utils.ValidateOptionalDate},
		{"Valid date should be valid", "20241225", true, utils.ValidateOptionalDate},
		{"Invalid date should be invalid", "invalid", false, utils.ValidateOptionalDate},
		{"Wrong date format should be invalid", "25-12-2024", false, utils.ValidateOptionalDate},
		{"Date with spaces should be valid when trimmed", "  20241225  ", true, utils.ValidateOptionalDate},
		
		// ValidateOptionalEventTime tests
		{"Empty time should be valid", "", true, utils.ValidateOptionalEventTime},
		{"Valid time should be valid", "14:30", true, utils.ValidateOptionalEventTime},
		{"Invalid time (wrong minutes) should be invalid", "14:15", false, utils.ValidateOptionalEventTime},
		{"Invalid time format should be invalid", "2:30 PM", false, utils.ValidateOptionalEventTime},
		{"Time with spaces should be valid when trimmed", "  14:00  ", true, utils.ValidateOptionalEventTime},
		{"Invalid hour should be invalid", "25:30", false, utils.ValidateOptionalEventTime},
		
		// Test 't' shortcut for date
		{"'t' should be valid for date", "t", true, utils.ValidateOptionalEventDate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.testFunc(tt.value)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for input '%s'", tt.expected, result, tt.value)
			}
		})
	}
}

func TestSearchEventsWithFilters(t *testing.T) {
	db := setupTestDB(t)
	defer db.CloseDatabase()

	// Add test events with known dates/times
	testEvents := []calendar.Event{
		*calendar.NewEvent("Morning Meeting", "Team standup", "Office", time.Date(2024, 12, 25, 9, 0, 0, 0, time.Local), 1.0, 0, 1, 0),
		*calendar.NewEvent("Lunch", "Team lunch", "Restaurant", time.Date(2024, 12, 25, 12, 30, 0, 0, time.Local), 1.5, 0, 1, 0),
		*calendar.NewEvent("Afternoon Meeting", "Project review", "Office", time.Date(2024, 12, 25, 15, 0, 0, 0, time.Local), 2.0, 0, 1, 0),
		*calendar.NewEvent("Dinner", "Family dinner", "Home", time.Date(2024, 12, 26, 18, 0, 0, 0, time.Local), 2.0, 0, 1, 0),
	}

	for _, event := range testEvents {
		if _, err := db.AddEvent(event); err != nil {
			t.Fatalf("Failed to add test event %s: %v", event.Name, err)
		}
	}

	tests := []struct {
		name           string
		criteria       database.SearchCriteria
		expectedCount  int
		expectedEvents []string // Names of expected events
	}{
		{
			name: "Search by text only",
			criteria: database.SearchCriteria{
				Query: "meeting",
			},
			expectedCount:  2,
			expectedEvents: []string{"Morning Meeting", "Afternoon Meeting"},
		},
		{
			name: "Search by date only (single day)",
			criteria: database.SearchCriteria{
				StartDate: "20241225",
				EndDate:   "20241225",
			},
			expectedCount:  3,
			expectedEvents: []string{"Morning Meeting", "Lunch", "Afternoon Meeting"},
		},
		{
			name: "Search by time range within a day",
			criteria: database.SearchCriteria{
				StartDate: "20241225",
				StartTime: "12:00",
				EndDate:   "20241225",
				EndTime:   "16:00",
			},
			expectedCount:  2,
			expectedEvents: []string{"Lunch", "Afternoon Meeting"},
		},
		{
			name: "Search by text with date filter",
			criteria: database.SearchCriteria{
				Query:     "lunch",
				StartDate: "20241225",
				EndDate:   "20241225",
			},
			expectedCount:  1,
			expectedEvents: []string{"Lunch"},
		},
		{
			name: "Search across multiple days",
			criteria: database.SearchCriteria{
				StartDate: "20241225",
				EndDate:   "20241226",
			},
			expectedCount:  4,
			expectedEvents: []string{"Morning Meeting", "Lunch", "Afternoon Meeting", "Dinner"},
		},
		{
			name: "Search with start date only",
			criteria: database.SearchCriteria{
				StartDate: "20241226",
			},
			expectedCount:  1,
			expectedEvents: []string{"Dinner"},
		},
		{
			name: "Search with end date only",
			criteria: database.SearchCriteria{
				EndDate: "20241225",
			},
			expectedCount:  3,
			expectedEvents: []string{"Morning Meeting", "Lunch", "Afternoon Meeting"},
		},
		{
			name: "Search with no results",
			criteria: database.SearchCriteria{
				Query: "nonexistent",
			},
			expectedCount:  0,
			expectedEvents: []string{},
		},
		{
			name: "Search with date range that has no events",
			criteria: database.SearchCriteria{
				StartDate: "20241224",
				EndDate:   "20241224",
			},
			expectedCount:  0,
			expectedEvents: []string{},
		},
		{
			name: "Empty criteria should return no results",
			criteria: database.SearchCriteria{
				Query:     "",
				StartDate: "",
				EndDate:   "",
			},
			expectedCount:  0,
			expectedEvents: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := db.SearchEventsWithFilters(tt.criteria)
			if err != nil {
				t.Fatalf("SearchEventsWithFilters failed: %v", err)
			}

			if len(results) != tt.expectedCount {
				t.Errorf("Expected %d results, got %d", tt.expectedCount, len(results))
			}

			// Check that we got the expected events
			if len(tt.expectedEvents) > 0 {
				resultNames := make(map[string]bool)
				for _, event := range results {
					resultNames[event.Name] = true
				}

				for _, expectedName := range tt.expectedEvents {
					if !resultNames[expectedName] {
						t.Errorf("Expected event '%s' not found in results", expectedName)
					}
				}
			}
		})
	}
}

func TestSearchEventsWithFiltersEdgeCases(t *testing.T) {
	db := setupTestDB(t)
	defer db.CloseDatabase()

	// Add events at specific times to test edge cases
	testEvents := []calendar.Event{
		*calendar.NewEvent("Midnight Event", "At midnight", "Home", time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local), 1.0, 0, 1, 0),
		*calendar.NewEvent("Late Night Event", "Before midnight", "Home", time.Date(2024, 1, 1, 23, 30, 0, 0, time.Local), 1.0, 0, 1, 0),
		*calendar.NewEvent("Early Morning", "Early start", "Office", time.Date(2024, 1, 2, 6, 0, 0, 0, time.Local), 1.0, 0, 1, 0),
	}

	for _, event := range testEvents {
		if _, err := db.AddEvent(event); err != nil {
			t.Fatalf("Failed to add test event %s: %v", event.Name, err)
		}
	}

	tests := []struct {
		name          string
		criteria      database.SearchCriteria
		expectedCount int
		description   string
	}{
		{
			name: "Default start time should be 00:00",
			criteria: database.SearchCriteria{
				StartDate: "20240101",
				EndDate:   "20240101",
			},
			expectedCount: 2, // Should include midnight event and late night event
			description:   "When no start time specified, should default to 00:00",
		},
		{
			name: "Default end time should be 23:59",
			criteria: database.SearchCriteria{
				StartDate: "20240101",
				EndDate:   "20240101",
			},
			expectedCount: 2, // Should include events until end of day
			description:   "When no end time specified, should default to 23:59",
		},
		{
			name: "Exact time match",
			criteria: database.SearchCriteria{
				StartDate: "20240101",
				StartTime: "23:30",
				EndDate:   "20240101",
				EndTime:   "23:30",
			},
			expectedCount: 1, // Should find exactly the 23:30 event
			description:   "Should find events at exact specified time",
		},
		{
			name: "Cross-day boundary search",
			criteria: database.SearchCriteria{
				StartDate: "20240101",
				StartTime: "23:00",
				EndDate:   "20240102",
				EndTime:   "07:00",
			},
			expectedCount: 2, // Late night + early morning
			description:   "Should handle searches that cross day boundaries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := db.SearchEventsWithFilters(tt.criteria)
			if err != nil {
				t.Fatalf("SearchEventsWithFilters failed: %v", err)
			}

			if len(results) != tt.expectedCount {
				t.Errorf("%s: Expected %d results, got %d", tt.description, tt.expectedCount, len(results))
				for i, event := range results {
					t.Logf("  Result %d: %s at %s", i+1, event.Name, event.Time.Format("2006-01-02 15:04"))
				}
			}
		})
	}
}

func TestSearchEventsWithTodayShortcut(t *testing.T) {
	db := setupTestDB(t)
	defer db.CloseDatabase()

	// Get today's date for comparison
	today := time.Now().Format("20060102")

	// Add test events - one for yesterday, today, and tomorrow
	testEvents := []calendar.Event{
		*calendar.NewEvent("Yesterday Event", "Yesterday's event", "Location", time.Now().AddDate(0, 0, -1), 1.0, 0, 1, 0),
		*calendar.NewEvent("Today Event", "Today's event", "Location", time.Now(), 1.0, 0, 1, 0),
		*calendar.NewEvent("Tomorrow Event", "Tomorrow's event", "Location", time.Now().AddDate(0, 0, 1), 1.0, 0, 1, 0),
	}

	for _, event := range testEvents {
		if _, err := db.AddEvent(event); err != nil {
			t.Fatalf("Failed to add test event %s: %v", event.Name, err)
		}
	}

	tests := []struct {
		name           string
		criteria       database.SearchCriteria
		expectedCount  int
		description    string
	}{
		{
			name: "Search from 't' (today) should find today and tomorrow events",
			criteria: database.SearchCriteria{
				StartDate: "t",
			},
			expectedCount: 2, // Today and tomorrow events
			description:   "Using 't' as start date should find events from today onwards",
		},
		{
			name: "Search to 't' (today) should find yesterday and today events",
			criteria: database.SearchCriteria{
				EndDate: "t",
			},
			expectedCount: 2, // Yesterday and today events
			description:   "Using 't' as end date should find events up to today",
		},
		{
			name: "Search exactly 't' to 't' should find only today's events",
			criteria: database.SearchCriteria{
				StartDate: "t",
				EndDate:   "t",
			},
			expectedCount: 1, // Only today's event
			description:   "Using 't' for both dates should find only today's events",
		},
		{
			name: "Compare 't' with explicit today date should be equivalent",
			criteria: database.SearchCriteria{
				StartDate: today,
				EndDate:   today,
			},
			expectedCount: 1, // Should match the 't' to 't' result
			description:   "Explicit today date should work same as 't' shortcut",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := db.SearchEventsWithFilters(tt.criteria)
			if err != nil {
				t.Fatalf("SearchEventsWithFilters failed: %v", err)
			}

			if len(results) != tt.expectedCount {
				t.Errorf("%s: Expected %d results, got %d", tt.description, tt.expectedCount, len(results))
				for i, event := range results {
					t.Logf("  Result %d: %s at %s", i+1, event.Name, event.Time.Format("2006-01-02 15:04"))
				}
			}
		})
	}
}// Additional tests for weekday recurrence functionality

// TestIsWeekday tests the IsWeekday utility function
func TestIsWeekday(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		expected bool
	}{
		{
			name:     "Monday is weekday",
			date:     time.Date(2024, 1, 1, 10, 0, 0, 0, time.Local), // Monday
			expected: true,
		},
		{
			name:     "Tuesday is weekday",
			date:     time.Date(2024, 1, 2, 10, 0, 0, 0, time.Local), // Tuesday
			expected: true,
		},
		{
			name:     "Wednesday is weekday",
			date:     time.Date(2024, 1, 3, 10, 0, 0, 0, time.Local), // Wednesday
			expected: true,
		},
		{
			name:     "Thursday is weekday",
			date:     time.Date(2024, 1, 4, 10, 0, 0, 0, time.Local), // Thursday
			expected: true,
		},
		{
			name:     "Friday is weekday",
			date:     time.Date(2024, 1, 5, 10, 0, 0, 0, time.Local), // Friday
			expected: true,
		},
		{
			name:     "Saturday is not weekday",
			date:     time.Date(2024, 1, 6, 10, 0, 0, 0, time.Local), // Saturday
			expected: false,
		},
		{
			name:     "Sunday is not weekday",
			date:     time.Date(2024, 1, 7, 10, 0, 0, 0, time.Local), // Sunday
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.IsWeekday(tt.date)
			if result != tt.expected {
				t.Errorf("IsWeekday(%s) = %v, expected %v", tt.date.Format("Monday"), result, tt.expected)
			}
		})
	}
}

// TestValidateFrequency tests the ValidateFrequency utility function
func TestValidateFrequency(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid number frequency",
			input:    "7",
			expected: true,
		},
		{
			name:     "Valid single digit",
			input:    "1",
			expected: true,
		},
		{
			name:     "Valid large number",
			input:    "30",
			expected: true,
		},
		{
			name:     "Lowercase 'w' for weekdays",
			input:    "w",
			expected: true,
		},
		{
			name:     "Uppercase 'W' for weekdays",
			input:    "W",
			expected: true,
		},
		{
			name:     "Empty string invalid",
			input:    "",
			expected: false,
		},
		{
			name:     "Zero is invalid",
			input:    "0",
			expected: false,
		},
		{
			name:     "Negative number invalid",
			input:    "-1",
			expected: false,
		},
		{
			name:     "Non-numeric non-w string invalid",
			input:    "abc",
			expected: false,
		},
		{
			name:     "Whitespace around valid input",
			input:    "  7  ",
			expected: true,
		},
		{
			name:     "Whitespace around 'w'",
			input:    "  w  ",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ValidateFrequency(tt.input)
			if result != tt.expected {
				t.Errorf("ValidateFrequency(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestGetReccuringEventsWeekdayStartingMonday tests weekday recurrence starting on Monday
func TestGetReccuringEventsWeekdayStartingMonday(t *testing.T) {
	// Test weekday recurrence starting on Monday
	startDate := time.Date(2024, 1, 1, 9, 0, 0, 0, time.Local) // Monday
	event := calendar.NewEvent("Daily Standup", "Weekday standup", "Office", startDate, 1.0, -1, 5, 0)
	
	recurring := event.GetReccuringEvents()
	
	if len(recurring) != 5 {
		t.Errorf("Expected 5 recurring events, got %d", len(recurring))
	}
	
	// Check dates are correct (Mon, Tue, Wed, Thu, Fri)
	expectedDates := []time.Time{
		time.Date(2024, 1, 1, 9, 0, 0, 0, time.Local), // Monday
		time.Date(2024, 1, 2, 9, 0, 0, 0, time.Local), // Tuesday
		time.Date(2024, 1, 3, 9, 0, 0, 0, time.Local), // Wednesday
		time.Date(2024, 1, 4, 9, 0, 0, 0, time.Local), // Thursday
		time.Date(2024, 1, 5, 9, 0, 0, 0, time.Local), // Friday
	}
	
	for i, event := range recurring {
		if !event.Time.Equal(expectedDates[i]) {
			t.Errorf("Event %d: expected time %s, got %s", i, expectedDates[i].Format("2006-01-02 15:04"), event.Time.Format("2006-01-02 15:04"))
		}
		
		// Verify all events are on weekdays
		if !utils.IsWeekday(event.Time) {
			t.Errorf("Event %d is not on a weekday: %s", i, event.Time.Format("Monday"))
		}
	}
}

// TestGetReccuringEventsWeekdayStartingSaturday tests weekday recurrence starting on weekend
func TestGetReccuringEventsWeekdayStartingSaturday(t *testing.T) {
	// Test weekday recurrence starting on Saturday (should move to Monday)
	startDate := time.Date(2024, 1, 6, 9, 0, 0, 0, time.Local) // Saturday
	event := calendar.NewEvent("Daily Standup", "Weekday standup", "Office", startDate, 1.0, -1, 3, 0)
	
	recurring := event.GetReccuringEvents()
	
	if len(recurring) != 3 {
		t.Errorf("Expected 3 recurring events, got %d", len(recurring))
	}
	
	// Check dates are correct (should start on Monday, not Saturday)
	expectedDates := []time.Time{
		time.Date(2024, 1, 8, 9, 0, 0, 0, time.Local),  // Monday (moved from Saturday)
		time.Date(2024, 1, 9, 9, 0, 0, 0, time.Local),  // Tuesday
		time.Date(2024, 1, 10, 9, 0, 0, 0, time.Local), // Wednesday
	}
	
	for i, event := range recurring {
		if !event.Time.Equal(expectedDates[i]) {
			t.Errorf("Event %d: expected time %s, got %s", i, expectedDates[i].Format("2006-01-02 15:04"), event.Time.Format("2006-01-02 15:04"))
		}
		
		// Verify all events are on weekdays
		if !utils.IsWeekday(event.Time) {
			t.Errorf("Event %d is not on a weekday: %s", i, event.Time.Format("Monday"))
		}
	}
}