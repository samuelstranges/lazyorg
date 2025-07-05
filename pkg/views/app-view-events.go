package views

import (
	"fmt"
	"os"
	"time"

	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/jroimartin/gocui"
)

// DeleteEvent deletes a single event at the cursor position
func (av *AppView) DeleteEvent(g *gocui.Gui) {
	hoveredView := av.GetHoveredOnView(g)
	if eventView, ok := hoveredView.(*EventView); ok {
		// Copy event to yank buffer before deleting (vim-like behavior)
		copiedEvent := *eventView.Event
		av.copiedEvent = &copiedEvent
		
		// Delete the event
		av.EventManager.DeleteEvent(eventView.Event.Id)
	}
}

// DeleteEvents deletes all events with the same name as the event at cursor position
func (av *AppView) DeleteEvents(g *gocui.Gui) {
	hoveredView := av.GetHoveredOnView(g)
	if eventView, ok := hoveredView.(*EventView); ok {
		av.EventManager.DeleteEventsByName(eventView.Event.Name)
	}
}

// ShowNewEventPopup displays the new event creation popup
func (av *AppView) ShowNewEventPopup(g *gocui.Gui) error {
	if view, ok := av.GetChild("popup"); ok {
		if popupView, ok := view.(*EventPopupView); ok {
			view.SetProperties(
				av.X+(av.W-PopupWidth)/2,
				av.Y+(av.H-PopupHeight)/2,
				PopupWidth,
				PopupHeight,
			)
			return popupView.ShowNewEventPopup(g)
		}
	}
	return nil
}

// ShowEditEventPopup displays the edit event popup for the event at cursor position
func (av *AppView) ShowEditEventPopup(g *gocui.Gui) error {
	if view, ok := av.GetChild("popup"); ok {
		if popupView, ok := view.(*EventPopupView); ok {
			view.SetProperties(
				av.X+(av.W-PopupWidth)/2,
				av.Y+(av.H-PopupHeight)/2,
				PopupWidth,
				PopupHeight,
			)
			hoveredView := av.GetHoveredOnView(g)
			if eventView, ok := hoveredView.(*EventView); ok {
				err := popupView.ShowEditEventPopup(g, eventView)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// ShowColorPicker displays the color picker popup for the event at cursor position
func (av *AppView) ShowColorPicker(g *gocui.Gui) error {
	hoveredView := av.GetHoveredOnView(g)
	if eventView, ok := hoveredView.(*EventView); ok {
		// Ensure we have a valid event before showing color picker
		if eventView.Event == nil {
			return nil
		}
		
		av.colorPickerEvent = eventView
		av.colorPickerActive = true
		
		if popup, ok := av.FindChildView("popup"); ok {
			if popupView, ok := popup.(*EventPopupView); ok {
				// Set up callback for color selection
				popupView.ColorPickerCallback = func(colorName string) error {
					color := calendar.ColorNameToAttribute(colorName)
					av.colorPickerEvent.Event.Color = color
					if !av.EventManager.UpdateEvent(av.colorPickerEvent.Event.Id, av.colorPickerEvent.Event) {
						// Error is handled by EventManager internally
						return nil
					}
					av.CloseColorPicker(g)
					return nil
				}
				
				popup.SetProperties(
					av.X+(av.W-PopupWidth)/2,
					av.Y+(av.H-PopupHeight)/2,
					PopupWidth,
					PopupHeight,
				)
				return popupView.ShowColorPickerPopup(g)
			}
		}
	}
	return nil
}

// IsColorPickerActive returns whether the color picker is currently active
func (av *AppView) IsColorPickerActive() bool {
	return av.colorPickerActive
}

// CloseColorPicker closes the color picker and resets state
func (av *AppView) CloseColorPicker(g *gocui.Gui) error {
	av.colorPickerActive = false
	av.colorPickerEvent = nil
	return nil
}

// CopyEvent copies the event at cursor position to the yank buffer
func (av *AppView) CopyEvent(g *gocui.Gui) {
	_, y := g.CurrentView().Cursor()

	if view, ok := av.FindChildView(WeekdayNames[av.Calendar.CurrentDay.Date.Weekday()]); ok {
		if dayView, ok := view.(*DayView); ok {
			if eventView, ok := dayView.IsOnEvent(y); ok {
				// Create a copy of the event
				copiedEvent := *eventView.Event
				av.copiedEvent = &copiedEvent
			}
		}
	}
}

// PasteEvent pastes the copied event to the current day/time
func (av *AppView) PasteEvent(g *gocui.Gui) error {
	if av.copiedEvent == nil {
		return nil // Nothing to paste
	}

	var debugInfo string
	if av.DebugMode {
		// DEBUGGING: Let's see what's happening
		currentView := g.CurrentView()
		currentViewName := ""
		if currentView != nil {
			currentViewName = currentView.Name()
		}
		
		calendarWeekday := av.Calendar.CurrentDay.Date.Weekday()
		calendarWeekdayName := WeekdayNames[calendarWeekday]
		calendarDate := av.Calendar.CurrentDay.Date
		
		// Print debug info to a file
		debugInfo = fmt.Sprintf("PASTE DEBUG:\n")
		debugInfo += fmt.Sprintf("  Current View Name: %s\n", currentViewName)
		debugInfo += fmt.Sprintf("  Calendar Weekday(): %d (%s)\n", calendarWeekday, calendarWeekdayName)
		debugInfo += fmt.Sprintf("  Calendar Date: %s\n", calendarDate.Format("2006-01-02 15:04:05"))
		debugInfo += fmt.Sprintf("  Calendar Date Weekday: %s\n", calendarDate.Weekday().String())
		
		// Print all week days for reference
		debugInfo += fmt.Sprintf("  Week Days:\n")
		for i, day := range av.Calendar.CurrentWeek.Days {
			debugInfo += fmt.Sprintf("    [%d] %s: %s\n", i, WeekdayNames[i], day.Date.Format("2006-01-02"))
		}
	}

	// Use the exact same logic as DeleteEvent
	if view, ok := av.FindChildView(WeekdayNames[av.Calendar.CurrentDay.Date.Weekday()]); ok {
		if _, ok := view.(*DayView); ok {
			// Create a new event based on the copied one
			newEvent := *av.copiedEvent
			newEvent.Id = 0 // Reset ID so database will assign a new one
			
			// DEBUG: Check current view vs calendar date
			currentView := g.CurrentView()
			currentViewName := ""
			if currentView != nil {
				currentViewName = currentView.Name()
			}
			
			// Get the actual date from the current view name
			var targetDate time.Time
			for i, dayName := range WeekdayNames {
				if dayName == currentViewName {
					// Found the matching day, get the date from the week
					if i < len(av.Calendar.CurrentWeek.Days) {
						targetDate = av.Calendar.CurrentWeek.Days[i].Date
						break
					}
				}
			}
			
			// If we couldn't determine target date, fall back to calendar current day
			if targetDate.IsZero() {
				targetDate = av.Calendar.CurrentDay.Date
			}
			
			// Use the target date with current time
			newEvent.Time = targetDate
			
			if av.DebugMode {
				// Add final time to debug
				finalDebug := fmt.Sprintf("\nDEBUG CALENDAR vs VIEW:\n")
				finalDebug += fmt.Sprintf("  Current View: %s\n", currentViewName)
				finalDebug += fmt.Sprintf("  Calendar.CurrentDay.Date: %s\n", av.Calendar.CurrentDay.Date.Format("2006-01-02 15:04:05"))
				finalDebug += fmt.Sprintf("  Target Date from View: %s\n", targetDate.Format("2006-01-02 15:04:05"))
				finalDebug += fmt.Sprintf("  FINAL EVENT TIME: %s\n", newEvent.Time.Format("2006-01-02 15:04:05"))
				os.WriteFile("/tmp/chronos_debug.txt", []byte(debugInfo + finalDebug), 0644)
			}

			// Add to database
			if _, success := av.EventManager.AddEvent(newEvent); !success {
				// Error is handled by EventManager internally
				return nil
			}

			// Refresh events from database to ensure proper display
			if err := av.updateEventsFromDatabase(); err != nil {
				return err
			}

			// Force a view update to refresh the display
			if err := av.UpdateChildren(g); err != nil {
				return err
			}
		}
	}

	return nil
}

// Undo reverts the last action
func (av *AppView) Undo(g *gocui.Gui) error {
	err := av.EventManager.Undo()
	if err != nil && err.Error() == "nothing to undo" {
		// Silently ignore when there's nothing to undo
		return nil
	}
	
	return err
}

// Redo re-applies the last undone action
func (av *AppView) Redo(g *gocui.Gui) error {
	err := av.EventManager.Redo()
	if err != nil && err.Error() == "nothing to redo" {
		// Silently ignore when there's nothing to redo
		return nil
	}
	
	return err
}