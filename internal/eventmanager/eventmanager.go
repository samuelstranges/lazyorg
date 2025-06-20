package eventmanager

import (
	"errors"
	"time"

	"github.com/HubertBel/lazyorg/internal/calendar"
	"github.com/HubertBel/lazyorg/internal/database"
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
	EventBefore *calendar.Event // State before action (nil for add)
	EventAfter  *calendar.Event // State after action (nil for delete)
	EventIds    []int           // For bulk operations
}

type EventManager struct {
	database   *database.Database
	undoStack  []UndoAction
	redoStack  []UndoAction
	maxUndos   int
}

func NewEventManager(db *database.Database) *EventManager {
	return &EventManager{
		database:  db,
		undoStack: make([]UndoAction, 0),
		redoStack: make([]UndoAction, 0),
		maxUndos:  50, // Keep last 50 actions
	}
}

// AddEvent adds a new event and records it for undo
func (em *EventManager) AddEvent(event calendar.Event) (*calendar.Event, error) {
	newEventId, err := em.database.AddEvent(event)
	if err != nil {
		return nil, err
	}

	// Get the full event from database
	newEvent, err := em.database.GetEventById(newEventId)
	if err != nil {
		return nil, err
	}

	// Record undo action
	em.pushUndoAction(UndoAction{
		Type:       ActionAdd,
		EventAfter: newEvent,
	})

	return newEvent, nil
}

// DeleteEvent deletes an event and records it for undo
func (em *EventManager) DeleteEvent(eventId int) error {
	// Get the event before deleting for undo
	eventBefore, err := em.database.GetEventById(eventId)
	if err != nil {
		return err
	}

	err = em.database.DeleteEventById(eventId)
	if err != nil {
		return err
	}

	// Record undo action
	em.pushUndoAction(UndoAction{
		Type:        ActionDelete,
		EventBefore: eventBefore,
	})

	return nil
}

// UpdateEvent updates an event and records it for undo
func (em *EventManager) UpdateEvent(eventId int, newEvent *calendar.Event) error {
	// Get the event before updating for undo
	eventBefore, err := em.database.GetEventById(eventId)
	if err != nil {
		return err
	}

	err = em.database.UpdateEventById(eventId, newEvent)
	if err != nil {
		return err
	}

	// Record undo action
	em.pushUndoAction(UndoAction{
		Type:        ActionEdit,
		EventBefore: eventBefore,
		EventAfter:  newEvent,
	})

	return nil
}

// DeleteEventsByName deletes all events with the same name and records it for undo
func (em *EventManager) DeleteEventsByName(name string) error {
	// Get all events with this name before deleting
	events, err := em.database.GetEventsByName(name)
	if err != nil {
		return err
	}

	var eventIds []int
	for _, event := range events {
		eventIds = append(eventIds, event.Id)
	}

	err = em.database.DeleteEventsByName(name)
	if err != nil {
		return err
	}

	// Record undo action for bulk delete
	em.pushUndoAction(UndoAction{
		Type:     ActionBulkDelete,
		EventIds: eventIds,
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
		// Undo delete by re-adding the event (but don't record this add)
		_, err := em.database.AddEvent(*lastAction.EventBefore)
		return err

	case ActionEdit:
		// Undo edit by restoring the old event state (but don't record this update)
		return em.database.UpdateEventById(lastAction.EventBefore.Id, lastAction.EventBefore)

	case ActionBulkDelete:
		// This is more complex - we'd need to store the full events, not just IDs
		// For now, return an error
		return errors.New("bulk delete undo not fully implemented yet")

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
		// Redo add by re-adding the event (but don't record this add)
		_, err := em.database.AddEvent(*lastAction.EventAfter)
		return err

	case ActionDelete:
		// Redo delete by deleting the event again (but don't record this delete)
		return em.database.DeleteEventById(lastAction.EventBefore.Id)

	case ActionEdit:
		// Redo edit by applying the new event state (but don't record this update)
		return em.database.UpdateEventById(lastAction.EventAfter.Id, lastAction.EventAfter)

	case ActionBulkDelete:
		// This is more complex - we'd need to store the full events, not just IDs
		// For now, return an error
		return errors.New("bulk delete redo not fully implemented yet")

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