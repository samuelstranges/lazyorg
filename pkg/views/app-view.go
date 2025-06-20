package views

import (
	"fmt"
	"os"
	"time"

	"github.com/HubertBel/lazyorg/internal/calendar"
	"github.com/HubertBel/lazyorg/internal/database"
	"github.com/HubertBel/lazyorg/internal/eventmanager"
	"github.com/HubertBel/lazyorg/internal/utils"
	"github.com/jroimartin/gocui"
	"github.com/nsf/termbox-go"
)

var (
	MainViewWidthRatio = 0.8
	SideViewWidthRatio = 0.2
)

type AppView struct {
	*BaseView

	Database     *database.Database
	EventManager *eventmanager.EventManager
	Calendar     *calendar.Calendar
	
	colorPickerEvent  *EventView
	colorPickerActive bool
	copiedEvent       *calendar.Event
}

func NewAppView(g *gocui.Gui, db *database.Database) *AppView {
	now := time.Now()
	t := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())

	c := calendar.NewCalendar(calendar.NewDay(t))
	em := eventmanager.NewEventManager(db)

	av := &AppView{
		BaseView:     NewBaseView("app"),
		Database:     db,
		EventManager: em,
		Calendar:     c,
	}

	av.AddChild("title", NewTitleView(c))
	av.AddChild("popup", NewEvenPopup(g, c, db))
	av.AddChild("main", NewMainView(c))
	av.AddChild("side", NewSideView(c, db))
	av.AddChild("keybinds", NewKeybindsView())

	return av
}

func (av *AppView) Layout(g *gocui.Gui) error {
	return av.Update(g)
}

func (av *AppView) Update(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	av.SetProperties(0, 1, maxX-1, maxY-1)

	v, err := g.SetView(
		av.Name,
		av.X,
		av.Y,
		av.X+av.W,
		av.Y+av.H,
	)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
	}

	if err = av.updateEventsFromDatabase(); err != nil {
		return err
	}

	av.updateChildViewProperties()

	if err = av.UpdateChildren(g); err != nil {
		return err
	}

	if err = av.updateCurrentView(g); err != nil {
		return err
	}

	return nil
}

func (av *AppView) updateEventsFromDatabase() error {
	for _, v := range av.Calendar.CurrentWeek.Days {
		clear(v.Events)

		var err error
		events, err := av.Database.GetEventsByDate(v.Date)
		if err != nil {
			return err
		}

		v.Events = events
		v.SortEventsByTime()
	}

	return nil
}

func (av *AppView) ShowOrHideSideView(g *gocui.Gui) error {
	if sideView, ok := av.GetChild("side"); ok {
		if err := sideView.ClearChildren(g); err != nil {
			return err
		}
		SideViewWidthRatio = 0.0
		MainViewWidthRatio = 1.0

		av.children.Delete("side")
		return g.DeleteView("side")
	}

	SideViewWidthRatio = 0.2
	MainViewWidthRatio = 0.8

	av.AddChild("side", NewSideView(av.Calendar, av.Database))

	return nil
}

func (av *AppView) JumpToToday() {
	av.Calendar.JumpToToday()
}

func (av *AppView) UpdateToNextWeek() {
	av.Calendar.UpdateToNextWeek()
}

func (av *AppView) UpdateToPrevWeek() {
	av.Calendar.UpdateToPrevWeek()
}

func (av *AppView) UpdateToNextDay(g *gocui.Gui) {
	av.Calendar.UpdateToNextDay()
	av.updateCurrentView(g)
}

func (av *AppView) UpdateToPrevDay(g *gocui.Gui) {
	av.Calendar.UpdateToPrevDay()
	av.updateCurrentView(g)
}

func (av *AppView) UpdateToNextTime(g *gocui.Gui) {
	_, height := g.CurrentView().Size()
	if _, y := g.CurrentView().Cursor(); y < height-1 {
		av.Calendar.UpdateToNextTime()
	}
}

func (av *AppView) UpdateToPrevTime(g *gocui.Gui) {
	if _, y := g.CurrentView().Cursor(); y > 0 {
		av.Calendar.UpdateToPrevTime()
	}
}

func (av *AppView) ChangeToNotepadView(g *gocui.Gui) error {
	_, err := g.SetCurrentView("notepad")
	if err != nil {
		return err
	}
	if view, ok := av.FindChildView("notepad"); ok {
		if notepadView, ok := view.(*NotepadView); ok {
			notepadView.IsActive = true
		}
	}

	return nil
}

func (av *AppView) ClearNotepadContent(g *gocui.Gui) error {
	if view, ok := av.FindChildView("notepad"); ok {
		if notepadView, ok := view.(*NotepadView); ok {
			return notepadView.ClearContent(g)
		}
	}

	return nil
}

func (av *AppView) SaveNotepadContent(g *gocui.Gui) error {
	if view, ok := av.FindChildView("notepad"); ok {
		if notepadView, ok := view.(*NotepadView); ok {
			return notepadView.SaveContent(g)
		}
	}

	return nil
}

func (av *AppView) ReturnToMainView(g *gocui.Gui) error {
	if err := av.SaveNotepadContent(g); err != nil {
		return err
	}
	if view, ok := av.FindChildView("notepad"); ok {
		if notepadView, ok := view.(*NotepadView); ok {
			notepadView.IsActive = false
		}
	}

	viewName := WeekdayNames[av.Calendar.CurrentDay.Date.Weekday()]
	g.SetCurrentView(viewName)
	return av.updateCurrentView(g)
}

func (av *AppView) DeleteEvent(g *gocui.Gui) {
	_, y := g.CurrentView().Cursor()

	if view, ok := av.FindChildView(WeekdayNames[av.Calendar.CurrentDay.Date.Weekday()]); ok {
		if dayView, ok := view.(*DayView); ok {
			if eventView, ok := dayView.IsOnEvent(y); ok {
				// Copy event to yank buffer before deleting (vim-like behavior)
				copiedEvent := *eventView.Event
				av.copiedEvent = &copiedEvent
				
				// Delete the event
				av.EventManager.DeleteEvent(eventView.Event.Id)
			}
		}
	}
}

func (av *AppView) DeleteEvents(g *gocui.Gui) {
	_, y := g.CurrentView().Cursor()

	if view, ok := av.FindChildView(WeekdayNames[av.Calendar.CurrentDay.Date.Weekday()]); ok {
		if dayView, ok := view.(*DayView); ok {
			if eventView, ok := dayView.IsOnEvent(y); ok {
				av.EventManager.DeleteEventsByName(eventView.Event.Name)
			}
		}
	}
}

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

func (av *AppView) ShowColorPicker(g *gocui.Gui) error {
	hoveredView := av.GetHoveredOnView(g)
	if eventView, ok := hoveredView.(*EventView); ok {
		av.colorPickerEvent = eventView
		av.colorPickerActive = true
		if err := av.showColorPickerPopup(g); err != nil {
			return err
		}
		return av.setColorPickerKeybindings(g)
	}
	return nil
}

func (av *AppView) showColorPickerPopup(g *gocui.Gui) error {
	v, err := g.SetView("colorpicker", av.W/2-15, av.H/2-5, av.W/2+15, av.H/2+5)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	
	v.Title = "Color Picker"
	v.Clear()
	v.Write([]byte("Select color:\n\n"))
	v.Write([]byte("r - Red\n"))
	v.Write([]byte("g - Green\n"))
	v.Write([]byte("y - Yellow\n"))
	v.Write([]byte("b - Blue\n"))
	v.Write([]byte("m - Magenta\n"))
	v.Write([]byte("c - Cyan\n"))
	v.Write([]byte("w - White\n"))
	v.Write([]byte("\nEsc - Cancel"))
	
	// Set the view on top so it receives input
	g.SetViewOnTop("colorpicker")
	
	// Disable cursor for this view
	g.Cursor = false
	
	// Make colorpicker the current view
	if _, err := g.SetCurrentView("colorpicker"); err != nil {
		return err
	}
	
	return nil
}

func (av *AppView) IsColorPickerActive() bool {
	return av.colorPickerActive
}

func (av *AppView) SelectColor(g *gocui.Gui, colorName string) error {
	if av.colorPickerEvent == nil {
		return av.CloseColorPicker(g)
	}
	
	color := calendar.ColorNameToAttribute(colorName)
	av.colorPickerEvent.Event.Color = color
	if err := av.EventManager.UpdateEvent(av.colorPickerEvent.Event.Id, av.colorPickerEvent.Event); err != nil {
		return err
	}
	
	return av.CloseColorPicker(g)
}

func (av *AppView) CloseColorPicker(g *gocui.Gui) error {
	// Remove color picker keybindings first
	if err := av.removeColorPickerKeybindings(g); err != nil {
		return err
	}
	
	// Reset state
	av.colorPickerActive = false
	av.colorPickerEvent = nil
	
	// Delete view
	if err := g.DeleteView("colorpicker"); err != nil && err != gocui.ErrUnknownView {
		return err
	}
	
	// Re-enable cursor
	g.Cursor = true
	
	// Return to main view
	viewName := WeekdayNames[av.Calendar.CurrentDay.Date.Weekday()]
	if _, err := g.SetCurrentView(viewName); err != nil {
		return err
	}
	return nil
}

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

func (av *AppView) PasteEvent(g *gocui.Gui) error {
	if av.copiedEvent == nil {
		return nil // Nothing to paste
	}

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
	debugInfo := fmt.Sprintf("PASTE DEBUG:\n")
	debugInfo += fmt.Sprintf("  Current View Name: %s\n", currentViewName)
	debugInfo += fmt.Sprintf("  Calendar Weekday(): %d (%s)\n", calendarWeekday, calendarWeekdayName)
	debugInfo += fmt.Sprintf("  Calendar Date: %s\n", calendarDate.Format("2006-01-02 15:04:05"))
	debugInfo += fmt.Sprintf("  Calendar Date Weekday: %s\n", calendarDate.Weekday().String())
	
	// Print all week days for reference
	debugInfo += fmt.Sprintf("  Week Days:\n")
	for i, day := range av.Calendar.CurrentWeek.Days {
		debugInfo += fmt.Sprintf("    [%d] %s: %s\n", i, WeekdayNames[i], day.Date.Format("2006-01-02"))
	}
	
	// Write to debug file
	os.WriteFile("/tmp/lazyorg_debug.txt", []byte(debugInfo), 0644)

	// Use the exact same logic as DeleteEvent
	if view, ok := av.FindChildView(WeekdayNames[av.Calendar.CurrentDay.Date.Weekday()]); ok {
		if _, ok := view.(*DayView); ok {
			// Create a new event based on the copied one
			newEvent := *av.copiedEvent
			newEvent.Id = 0 // Reset ID so database will assign a new one
			
			// Use the current calendar day and time directly
			newEvent.Time = av.Calendar.CurrentDay.Date
			
			// Add final time to debug
			finalDebug := fmt.Sprintf("\nFINAL EVENT TIME: %s\n", newEvent.Time.Format("2006-01-02 15:04:05"))
			os.WriteFile("/tmp/lazyorg_debug.txt", []byte(debugInfo + finalDebug), 0644)

			// Add to database
			if _, err := av.EventManager.AddEvent(newEvent); err != nil {
				return err
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

func (av *AppView) ShowKeybinds(g *gocui.Gui) error {
	if view, ok := av.GetChild("keybinds"); ok {
		if keybindsView, ok := view.(*KeybindsView); ok {
			if keybindsView.IsVisible {
				keybindsView.IsVisible = false
				return g.DeleteView(keybindsView.Name)
			}

			keybindsView.IsVisible = true
			keybindsView.SetProperties(
				av.X+(av.W-KeybindsWidth)/2,
				av.Y+(av.H-KeybindsHeight)/2,
				KeybindsWidth,
				KeybindsHeight,
			)

			return keybindsView.Update(g)
		}
	}

	return nil
}

func (av *AppView) updateChildViewProperties() {
	mainViewWidth := int(float64(av.W-1) * MainViewWidthRatio)
	sideViewWidth := int(float64(av.W) * SideViewWidthRatio)

	if titleView, ok := av.GetChild("title"); ok {
		titleView.SetProperties(
			av.X+sideViewWidth+1,
			av.Y,
			mainViewWidth,
			TitleViewHeight,
		)
	}

	if mainView, ok := av.GetChild("main"); ok {
		y := av.Y + TitleViewHeight + 1
		mainView.SetProperties(
			av.X+sideViewWidth+1,
			y,
			mainViewWidth,
			av.H-y,
		)
	}

	if sideView, ok := av.GetChild("side"); ok {
		sideView.SetProperties(
			av.X,
			av.Y,
			sideViewWidth,
			av.H,
		)
	}
}

func (av *AppView) updateCurrentView(g *gocui.Gui) error {
	if view, ok := av.GetChild("popup"); ok {
		if popupView, ok := view.(*EventPopupView); ok {
			if popupView.IsVisible {
				return nil
			}
		}
	}
	if view, ok := av.GetChild("keybinds"); ok {
		if keybindsView, ok := view.(*KeybindsView); ok {
			if keybindsView.IsVisible {
				g.Cursor = false
				g.SetCurrentView("keybinds")
				return nil
			}
		}
	}
	if g.CurrentView() != nil && g.CurrentView().Name() == "notepad" {
		return nil
	}

	g.Cursor = true
	if view, ok := av.FindChildView("hover"); ok {
		if hoverView, ok := view.(*HoverView); ok {
			hoverView.CurrentView = av.GetHoveredOnView(g)
			hoverView.Update(g)
		}
	}

	g.SetCurrentView(WeekdayNames[av.Calendar.CurrentDay.Date.Weekday()])
	g.CurrentView().BgColor = gocui.Attribute(termbox.ColorBlack)
	g.CurrentView().SetCursor(1, av.GetCursorY())

	return nil
}

func (av *AppView) GetHoveredOnView(g *gocui.Gui) View {
	viewName := WeekdayNames[av.Calendar.CurrentDay.Date.Weekday()]
	var hoveredView View

	if view, ok := av.FindChildView(viewName); ok {
		if dayView, ok := view.(*DayView); ok {
			if eventView, ok := dayView.IsOnEvent(av.GetCursorY()); ok {
				hoveredView = eventView
			} else {
				hoveredView = dayView
			}
		}
	}

	return hoveredView
}

func (av *AppView) GetCursorY() int {
	y := 0

	if view, ok := av.FindChildView("time"); ok {
		if timeView, ok := view.(*TimeView); ok {
			y = utils.TimeToPosition(av.Calendar.CurrentDay.Date, timeView.Body)
		}
	}

	return y
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

// setColorPickerKeybindings adds temporary keybindings for color picker mode
func (av *AppView) setColorPickerKeybindings(g *gocui.Gui) error {
	colorPickerKeybindings := map[rune]string{
		'r': "Red",
		'g': "Green", 
		'y': "Yellow",
		'b': "Blue",
		'm': "Magenta",
		'c': "Cyan",
		'w': "White",
	}
	
	// First, remove ALL keybindings from weekday views while color picker is active
	for _, viewName := range WeekdayNames {
		// Remove all main keybindings to prevent interference
		keysToRemove := []interface{}{'a', 'e', 'c', 'h', 'l', 'j', 'k', 'T', 'H', 'L', 'd', 'D', 'y', 'p', 'u', 'r', gocui.KeyCtrlN, gocui.KeyCtrlS, '?', 'q', gocui.KeyArrowLeft, gocui.KeyArrowRight, gocui.KeyArrowDown, gocui.KeyArrowUp}
		for _, key := range keysToRemove {
			g.DeleteKeybinding(viewName, key, gocui.ModNone)
		}
	}
	
	// Set global keybindings for color picker (empty string means global)
	for key, color := range colorPickerKeybindings {
		color := color // capture for closure
		if err := g.SetKeybinding("", key, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			if av.IsColorPickerActive() {
				return av.SelectColor(g, color)
			}
			return nil
		}); err != nil {
			return err
		}
	}
	
	// Add global escape key to close color picker
	if err := g.SetKeybinding("", gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if av.IsColorPickerActive() {
			return av.CloseColorPicker(g)
		}
		return nil
	}); err != nil {
		return err
	}
	
	return nil
}

// removeColorPickerKeybindings removes temporary keybindings for color picker mode
func (av *AppView) removeColorPickerKeybindings(g *gocui.Gui) error {
	// Remove global color picker keybindings
	keys := []interface{}{'r', 'g', 'y', 'b', 'm', 'c', 'w', gocui.KeyEsc}
	for _, key := range keys {
		if err := g.DeleteKeybinding("", key, gocui.ModNone); err != nil && err != gocui.ErrUnknownView {
			// Ignore errors if keybinding doesn't exist
		}
	}
	
	// Restore all original keybindings to weekday views
	for _, viewName := range WeekdayNames {
		// Restore main keybindings
		mainKeybindings := []struct {
			key     interface{}
			handler func(*gocui.Gui, *gocui.View) error
		}{
			{'a', func(g *gocui.Gui, v *gocui.View) error { return av.ShowNewEventPopup(g) }},
			{'e', func(g *gocui.Gui, v *gocui.View) error { return av.ShowEditEventPopup(g) }},
			{'c', func(g *gocui.Gui, v *gocui.View) error { return av.ShowColorPicker(g) }},
			{'h', func(g *gocui.Gui, v *gocui.View) error { av.UpdateToPrevDay(g); return nil }},
			{'l', func(g *gocui.Gui, v *gocui.View) error { av.UpdateToNextDay(g); return nil }},
			{'j', func(g *gocui.Gui, v *gocui.View) error { av.UpdateToNextTime(g); return nil }},
			{'k', func(g *gocui.Gui, v *gocui.View) error { av.UpdateToPrevTime(g); return nil }},
			{'T', func(g *gocui.Gui, v *gocui.View) error { av.JumpToToday(); return nil }},
			{'H', func(g *gocui.Gui, v *gocui.View) error { av.UpdateToPrevWeek(); return nil }},
			{'L', func(g *gocui.Gui, v *gocui.View) error { av.UpdateToNextWeek(); return nil }},
			{'d', func(g *gocui.Gui, v *gocui.View) error { av.DeleteEvent(g); return nil }},
			{'D', func(g *gocui.Gui, v *gocui.View) error { av.DeleteEvents(g); return nil }},
			{'y', func(g *gocui.Gui, v *gocui.View) error { av.CopyEvent(g); return nil }},
			{'p', func(g *gocui.Gui, v *gocui.View) error { return av.PasteEvent(g) }},
			{'u', func(g *gocui.Gui, v *gocui.View) error { return av.Undo(g) }},
			{'r', func(g *gocui.Gui, v *gocui.View) error { return av.Redo(g) }},
			{gocui.KeyCtrlN, func(g *gocui.Gui, v *gocui.View) error { return av.ChangeToNotepadView(g) }},
			{gocui.KeyCtrlS, func(g *gocui.Gui, v *gocui.View) error { return av.ShowOrHideSideView(g) }},
			{'?', func(g *gocui.Gui, v *gocui.View) error { return av.ShowKeybinds(g) }},
			{'q', func(g *gocui.Gui, v *gocui.View) error { return gocui.ErrQuit }},
			{gocui.KeyArrowLeft, func(g *gocui.Gui, v *gocui.View) error { av.UpdateToPrevDay(g); return nil }},
			{gocui.KeyArrowRight, func(g *gocui.Gui, v *gocui.View) error { av.UpdateToNextDay(g); return nil }},
			{gocui.KeyArrowDown, func(g *gocui.Gui, v *gocui.View) error { av.UpdateToNextTime(g); return nil }},
			{gocui.KeyArrowUp, func(g *gocui.Gui, v *gocui.View) error { av.UpdateToPrevTime(g); return nil }},
		}
		
		for _, kb := range mainKeybindings {
			if err := g.SetKeybinding(viewName, kb.key, gocui.ModNone, kb.handler); err != nil {
				return err
			}
		}
	}
	
	return nil
}
