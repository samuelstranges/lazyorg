package eventmanager

import (
	"errors"
	"strconv"
	"time"

	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/samuelstranges/chronos/internal/database"
)

type ActionType string

const (
	ActionAdd    ActionType = "add"
	ActionDelete ActionType = "delete"
	ActionEdit   ActionType = "edit"
	ActionBulkDelete ActionType = "bulk_delete"
)

type UndoAction struct {
	Type        ActionType
	EventBefore *calendar.Event   // State before action (nil for add)
	EventAfter  *calendar.Event   // State after action (nil for delete)
	EventIds    []int             // For bulk operations (legacy)
	Events      []*calendar.Event // Full events for bulk operations
}

type EventManager struct {
	database   *database.Database
	undoStack  []UndoAction
	redoStack  []UndoAction
	maxUndos   int
	errorHandler func(title, message string) // Callback for displaying errors
}

func NewEventManager(db *database.Database) *EventManager {
	return &EventManager{
		database:  db,
		undoStack: make([]UndoAction, 0),
		redoStack: make([]UndoAction, 0),
		maxUndos:  50, // Keep last 50 actions
	}
}

// SetErrorHandler sets the error display callback
func (em *EventManager) SetErrorHandler(handler func(title, message string)) {
	em.errorHandler = handler
}

// toUTC converts an event's time to UTC for database storage
func (em *EventManager) toUTC(event *calendar.Event) *calendar.Event {
	utcEvent := *event
	utcEvent.Time = event.Time.UTC()
	return &utcEvent
}

// toLocal converts an event's time from UTC to local time for display
func (em *EventManager) toLocal(event *calendar.Event) *calendar.Event {
	localEvent := *event
	localEvent.Time = event.Time.In(time.Local)
	return &localEvent
}

// showError displays an error using the registered error handler
func (em *EventManager) showError(title, message string) {
	if em.errorHandler != nil {
		em.errorHandler(title, message)
	}
}

// AddEvent adds a new event and records it for undo
func (em *EventManager) AddEvent(event calendar.Event) (*calendar.Event, bool) {
	// Convert to UTC for database storage
	utcEvent := em.toUTC(&event)
	
	// Check for overlaps before adding (using UTC time for consistency)
	hasOverlap, err := em.database.CheckEventOverlap(*utcEvent)
	if err != nil {
		em.showError("Database Error", "Failed to check for overlapping events: "+err.Error())
		return nil, false
	}
	if hasOverlap {
		em.showError("Cannot Add Event", "This event overlaps with an existing event")
		return nil, false
	}

	newEventId, err := em.database.AddEvent(*utcEvent)
	if err != nil {
		em.showError("Cannot Add Event", "Failed to save event: "+err.Error())
		return nil, false
	}

	// Get the full event from database and convert back to local time
	newEvent, err := em.database.GetEventById(newEventId)
	if err != nil {
		em.showError("Database Error", "Failed to retrieve saved event: "+err.Error())
		return nil, false
	}
	localEvent := em.toLocal(newEvent)

	// Record undo action (store in local time for consistency with UI)
	em.pushUndoAction(UndoAction{
		Type:       ActionAdd,
		EventAfter: localEvent,
	})

	return localEvent, true
}

// DeleteEvent deletes an event and records it for undo
func (em *EventManager) DeleteEvent(eventId int) error {
	// Get the event before deleting for undo (convert from UTC to local)
	eventBefore, err := em.database.GetEventById(eventId)
	if err != nil {
		return err
	}
	
	// Check if event exists
	if eventBefore == nil {
		return errors.New("event not found: cannot delete non-existent event")
	}
	localEventBefore := em.toLocal(eventBefore)

	err = em.database.DeleteEventById(eventId)
	if err != nil {
		return err
	}

	// Record undo action (store in local time for consistency with UI)
	em.pushUndoAction(UndoAction{
		Type:        ActionDelete,
		EventBefore: localEventBefore,
	})

	return nil
}

// UpdateEvent updates an event and records it for undo
func (em *EventManager) UpdateEvent(eventId int, newEvent *calendar.Event) bool {
	// Get the event before updating for undo (convert from UTC to local)
	eventBefore, err := em.database.GetEventById(eventId)
	if err != nil {
		em.showError("Database Error", "Failed to retrieve original event: "+err.Error())
		return false
	}
	
	// Check if event exists
	if eventBefore == nil {
		em.showError("Event Not Found", "Cannot update event: event does not exist")
		return false
	}
	localEventBefore := em.toLocal(eventBefore)

	// Convert new event to UTC for database storage
	utcNewEvent := em.toUTC(newEvent)

	// Check for overlaps before updating (exclude current event, using UTC time)
	hasOverlap, err := em.database.CheckEventOverlap(*utcNewEvent, eventId)
	if err != nil {
		em.showError("Database Error", "Failed to check for overlapping events: "+err.Error())
		return false
	}
	if hasOverlap {
		em.showError("Cannot Edit Event", "Updated event would overlap with an existing event")
		return false
	}

	err = em.database.UpdateEventById(eventId, utcNewEvent)
	if err != nil {
		em.showError("Cannot Edit Event", "Failed to save changes: "+err.Error())
		return false
	}

	// Record undo action (store in local time for consistency with UI)
	em.pushUndoAction(UndoAction{
		Type:        ActionEdit,
		EventBefore: localEventBefore,
		EventAfter:  newEvent, // newEvent is already in local time from UI
	})

	return true
}

// DeleteEventsByName deletes all events with the same name and records it for undo
func (em *EventManager) DeleteEventsByName(name string) error {
	// Get all events with this name before deleting (convert from UTC to local)
	events, err := em.database.GetEventsByName(name)
	if err != nil {
		return err
	}

	// If no events found, nothing to delete
	if len(events) == 0 {
		return nil
	}

	// Convert all events to local time for undo storage
	localEvents := make([]*calendar.Event, len(events))
	for i, event := range events {
		localEvents[i] = em.toLocal(event)
	}

	err = em.database.DeleteEventsByName(name)
	if err != nil {
		return err
	}

	// Record undo action for bulk delete with full event data (in local time)
	em.pushUndoAction(UndoAction{
		Type:   ActionBulkDelete,
		Events: localEvents,
	})

	return nil
}

// Undo reverts the last action
func (em *EventManager) Undo() error {
	if len(em.undoStack) == 0 {
		return errors.New("nothing to undo")
	}

	// Pop the last action
	lastAction := em.undoStack[len(em.undoStack)-1]
	em.undoStack = em.undoStack[:len(em.undoStack)-1]

	// Push to redo stack before reverting
	em.redoStack = append(em.redoStack, lastAction)
	
	// Limit redo stack size
	if len(em.redoStack) > em.maxUndos {
		em.redoStack = em.redoStack[1:]
	}

	// Revert the action
	switch lastAction.Type {
	case ActionAdd:
		// Undo add by deleting the event (but don't record this delete)
		return em.database.DeleteEventById(lastAction.EventAfter.Id)

	case ActionDelete:
		// Undo delete by re-adding the event (convert to UTC for storage)
		utcEvent := em.toUTC(lastAction.EventBefore)
		_, err := em.database.AddEvent(*utcEvent)
		return err

	case ActionEdit:
		// Undo edit by restoring the old event state (convert to UTC for storage)
		utcEvent := em.toUTC(lastAction.EventBefore)
		return em.database.UpdateEventById(lastAction.EventBefore.Id, utcEvent)

	case ActionBulkDelete:
		// Undo bulk delete by re-adding all the deleted events (convert to UTC for storage)
		for _, event := range lastAction.Events {
			utcEvent := em.toUTC(event)
			_, err := em.database.AddEvent(*utcEvent)
			if err != nil {
				return err
			}
		}
		return nil

	default:
		return errors.New("unknown action type")
	}
}

// Redo re-applies the last undone action
func (em *EventManager) Redo() error {
	if len(em.redoStack) == 0 {
		return errors.New("nothing to redo")
	}

	// Pop the last undone action
	lastAction := em.redoStack[len(em.redoStack)-1]
	em.redoStack = em.redoStack[:len(em.redoStack)-1]

	// Push back to undo stack before re-applying (but don't clear redo stack)
	em.undoStack = append(em.undoStack, lastAction)
	
	// Limit undo stack size
	if len(em.undoStack) > em.maxUndos {
		em.undoStack = em.undoStack[1:]
	}

	// Re-apply the action
	switch lastAction.Type {
	case ActionAdd:
		// Redo add by re-adding the event (convert to UTC for storage)
		utcEvent := em.toUTC(lastAction.EventAfter)
		_, err := em.database.AddEvent(*utcEvent)
		return err

	case ActionDelete:
		// Redo delete by deleting the event again (but don't record this delete)
		return em.database.DeleteEventById(lastAction.EventBefore.Id)

	case ActionEdit:
		// Redo edit by applying the new event state (convert to UTC for storage)
		utcEvent := em.toUTC(lastAction.EventAfter)
		return em.database.UpdateEventById(lastAction.EventAfter.Id, utcEvent)

	case ActionBulkDelete:
		// Redo bulk delete by deleting all events with the same name again
		if len(lastAction.Events) > 0 {
			eventName := lastAction.Events[0].Name
			return em.database.DeleteEventsByName(eventName)
		}
		return nil

	default:
		return errors.New("unknown action type")
	}
}

// CanUndo returns true if there are actions that can be undone
func (em *EventManager) CanUndo() bool {
	return len(em.undoStack) > 0
}

// CanRedo returns true if there are actions that can be redone
func (em *EventManager) CanRedo() bool {
	return len(em.redoStack) > 0
}

// GetUndoDescription returns a description of what the next undo would do
func (em *EventManager) GetUndoDescription() string {
	if len(em.undoStack) == 0 {
		return "Nothing to undo"
	}

	lastAction := em.undoStack[len(em.undoStack)-1]
	switch lastAction.Type {
	case ActionAdd:
		return "Undo add: " + lastAction.EventAfter.Name
	case ActionDelete:
		return "Undo delete: " + lastAction.EventBefore.Name
	case ActionEdit:
		return "Undo edit: " + lastAction.EventBefore.Name
	case ActionBulkDelete:
		if len(lastAction.Events) > 0 {
			return "Undo bulk delete: " + lastAction.Events[0].Name + " (" + strconv.Itoa(len(lastAction.Events)) + " events)"
		}
		return "Undo bulk delete"
	default:
		return "Undo last action"
	}
}

// GetRedoDescription returns a description of what the next redo would do
func (em *EventManager) GetRedoDescription() string {
	if len(em.redoStack) == 0 {
		return "Nothing to redo"
	}

	lastAction := em.redoStack[len(em.redoStack)-1]
	switch lastAction.Type {
	case ActionAdd:
		return "Redo add: " + lastAction.EventAfter.Name
	case ActionDelete:
		return "Redo delete: " + lastAction.EventBefore.Name
	case ActionEdit:
		return "Redo edit: " + lastAction.EventAfter.Name
	case ActionBulkDelete:
		if len(lastAction.Events) > 0 {
			return "Redo bulk delete: " + lastAction.Events[0].Name + " (" + strconv.Itoa(len(lastAction.Events)) + " events)"
		}
		return "Redo bulk delete"
	default:
		return "Redo last action"
	}
}

// pushUndoAction adds an action to the undo stack
func (em *EventManager) pushUndoAction(action UndoAction) {
	em.undoStack = append(em.undoStack, action)
	
	// Clear redo stack when new action is performed
	em.redoStack = make([]UndoAction, 0)

	// Limit undo stack size
	if len(em.undoStack) > em.maxUndos {
		em.undoStack = em.undoStack[1:]
	}
}

// Pass-through methods for read operations (no undo needed)
func (em *EventManager) GetEventById(id int) (*calendar.Event, error) {
	return em.database.GetEventById(id)
}

func (em *EventManager) GetEventsByDate(date time.Time) ([]*calendar.Event, error) {
	return em.database.GetEventsByDate(date)
}

func (em *EventManager) GetEventsByName(name string) ([]*calendar.Event, error) {
	return em.database.GetEventsByName(name)
}

func (em *EventManager) GetEventsByMonth(year int, month time.Month) ([]*calendar.Event, error) {
	return em.database.GetEventsByMonth(year, month)
}

func (em *EventManager) GetAllEvents() ([]*calendar.Event, error) {
	return em.database.GetAllEvents()
}

func (em *EventManager) SearchEventsWithFilters(criteria database.SearchCriteria) ([]*calendar.Event, error) {
	return em.database.SearchEventsWithFilters(criteria)
}