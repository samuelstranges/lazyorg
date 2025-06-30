package views

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/HubertBel/lazyorg/internal/calendar"
	"github.com/HubertBel/lazyorg/internal/config"
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
	Config       *config.Config
	
	colorPickerEvent  *EventView
	colorPickerActive bool
	copiedEvent       *calendar.Event
	
	// Search functionality
	searchQuery       string
	searchMatches     []*calendar.Event
	currentMatchIndex int
	isSearchActive    bool
}

func NewAppView(g *gocui.Gui, db *database.Database, cfg *config.Config) *AppView {
	now := time.Now()
	t := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())

	c := calendar.NewCalendar(calendar.NewDay(t))
	em := eventmanager.NewEventManager(db)

	av := &AppView{
		BaseView:     NewBaseView("app"),
		Database:     db,
		EventManager: em,
		Calendar:     c,
		Config:       cfg,
	}
	

	av.AddChild("title", NewTitleView(c))
	av.AddChild("popup", NewEvenPopup(g, c, db, av.EventManager))
	av.AddChild("main", NewMainView(c))
	
	// Set up error handler for EventManager after popup is created
	av.setupErrorHandler(g)
	
	// Only add side view if it will have children
	if !cfg.HideDayOnStartup || !cfg.HideNotesOnStartup {
		av.AddChild("side", NewSideView(c, db, cfg))
	}
	
	av.AddChild("keybinds", NewKeybindsView())

	return av
}

// setupErrorHandler configures the EventManager to show error messages via popup
func (av *AppView) setupErrorHandler(g *gocui.Gui) {
	av.EventManager.SetErrorHandler(func(title, message string) {
		if popup, ok := av.GetChild("popup"); ok {
			if popupView, ok := popup.(*EventPopupView); ok {
				popupView.ShowErrorMessage(g, title, message)
			}
		}
	})
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

	if err = av.UpdateCurrentView(g); err != nil {
		return err
	}
	
	if err = av.showSearchStatus(g); err != nil {
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

	// Only add side view if it will have children
	if !av.Config.HideDayOnStartup || !av.Config.HideNotesOnStartup {
		av.AddChild("side", NewSideView(av.Calendar, av.Database, av.Config))
	}

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
	av.UpdateCurrentView(g)
}

func (av *AppView) UpdateToPrevDay(g *gocui.Gui) {
	av.Calendar.UpdateToPrevDay()
	av.UpdateCurrentView(g)
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

func (av *AppView) JumpToNextEvent() {
	allEvents := av.getAllEventsFromWeek()
	if len(allEvents) == 0 {
		return
	}

	currentTime := av.Calendar.CurrentDay.Date
	
	// Find the next event after current time
	for _, event := range allEvents {
		if event.Time.After(currentTime) {
			av.Calendar.CurrentDay.Date = event.Time
			av.Calendar.UpdateWeek()
			return
		}
	}
	
	// If no event found after current time, wrap to first event
	if len(allEvents) > 0 {
		av.Calendar.CurrentDay.Date = allEvents[0].Time
		av.Calendar.UpdateWeek()
	}
}

func (av *AppView) JumpToPrevEvent() {
	allEvents := av.getAllEventsFromWeek()
	if len(allEvents) == 0 {
		return
	}

	currentTime := av.Calendar.CurrentDay.Date
	
	// Find the previous event before current time (iterate backwards)
	for i := len(allEvents) - 1; i >= 0; i-- {
		event := allEvents[i]
		if event.Time.Before(currentTime) {
			av.Calendar.CurrentDay.Date = event.Time
			av.Calendar.UpdateWeek()
			return
		}
	}
	
	// If no event found before current time, wrap to last event
	if len(allEvents) > 0 {
		av.Calendar.CurrentDay.Date = allEvents[len(allEvents)-1].Time
		av.Calendar.UpdateWeek()
	}
}

func (av *AppView) getAllEventsFromWeek() []*calendar.Event {
	var allEvents []*calendar.Event
	
	// Collect all events from all days in the week
	for _, day := range av.Calendar.CurrentWeek.Days {
		allEvents = append(allEvents, day.Events...)
	}
	
	// Sort by time
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].Time.Before(allEvents[j].Time)
	})
	
	return allEvents
}

func (av *AppView) StartSearch(g *gocui.Gui) error {
	if popup, ok := av.GetChild("popup"); ok {
		if popupView, ok := popup.(*EventPopupView); ok {
			// Set up the search callback
			popupView.SearchCallback = func(query string) error {
				return av.executeSearchQuery(query)
			}
			
			popup.SetProperties(
				av.X+(av.W-PopupWidth)/2,
				av.Y+(av.H-PopupHeight)/2,
				PopupWidth,
				PopupHeight,
			)
			return popupView.ShowSearchPopup(g)
		}
	}
	return nil
}

func (av *AppView) executeSearchQuery(query string) error {
	av.searchQuery = query
	av.searchMatches = av.findMatches(query)
	av.currentMatchIndex = 0
	av.isSearchActive = true
	
	if len(av.searchMatches) > 0 {
		// Jump to first match
		firstMatch := av.searchMatches[0]
		av.Calendar.CurrentDay.Date = firstMatch.Time
		av.Calendar.UpdateWeek()
	}
	
	return nil
}

func (av *AppView) findMatches(query string) []*calendar.Event {
	var matches []*calendar.Event
	allEvents := av.getAllEventsFromWeek()
	
	query = strings.ToLower(query)
	
	for _, event := range allEvents {
		if strings.Contains(strings.ToLower(event.Name), query) ||
		   strings.Contains(strings.ToLower(event.Description), query) ||
		   strings.Contains(strings.ToLower(event.Location), query) {
			matches = append(matches, event)
		}
	}
	
	return matches
}

func (av *AppView) GoToNextMatch() error {
	if !av.isSearchActive || len(av.searchMatches) == 0 {
		return nil
	}
	
	av.currentMatchIndex = (av.currentMatchIndex + 1) % len(av.searchMatches)
	match := av.searchMatches[av.currentMatchIndex]
	av.Calendar.CurrentDay.Date = match.Time
	av.Calendar.UpdateWeek()
	
	return nil
}

func (av *AppView) GoToPrevMatch() error {
	if !av.isSearchActive || len(av.searchMatches) == 0 {
		return nil
	}
	
	av.currentMatchIndex = (av.currentMatchIndex - 1 + len(av.searchMatches)) % len(av.searchMatches)
	match := av.searchMatches[av.currentMatchIndex]
	av.Calendar.CurrentDay.Date = match.Time
	av.Calendar.UpdateWeek()
	
	return nil
}

func (av *AppView) ClearSearch() {
	av.isSearchActive = false
	av.searchQuery = ""
	av.searchMatches = nil
	av.currentMatchIndex = 0
}

func (av *AppView) GetSearchStatus() string {
	if !av.isSearchActive || len(av.searchMatches) == 0 {
		return ""
	}
	
	return fmt.Sprintf("%d/%d matches for '%s'", av.currentMatchIndex+1, len(av.searchMatches), av.searchQuery)
}

func (av *AppView) showSearchStatus(g *gocui.Gui) error {
	if !av.isSearchActive || len(av.searchMatches) == 0 {
		// Hide search status if not active
		if err := g.DeleteView("search-status"); err != nil && err != gocui.ErrUnknownView {
			return err
		}
		return nil
	}
	
	status := av.GetSearchStatus()
	width := len(status) + 4
	height := 3
	maxX, maxY := g.Size()
	
	// Position in bottom-right corner
	x := maxX - width - 2
	y := maxY - height - 2
	
	v, err := g.SetView("search-status", x, y, x+width, y+height)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	
	v.Title = ""
	v.Frame = true
	v.Clear()
	v.Write([]byte(status))
	
	// Set on top so it's visible
	g.SetViewOnTop("search-status")
	
	return nil
}

func (av *AppView) ShowGotoPopup(g *gocui.Gui) error {
	if popup, ok := av.FindChildView("popup"); ok {
		if popupView, ok := popup.(*EventPopupView); ok {
			popup.SetProperties(
				av.X+(av.W-PopupWidth)/2,
				av.Y+(av.H-PopupHeight)/2,
				PopupWidth,
				PopupHeight,
			)
			return popupView.ShowGotoPopup(g)
		}
	}
	return nil
}

func (av *AppView) ChangeToNotepadView(g *gocui.Gui) error {
	// Check if notepad exists, if not create the side view and notepad
	if _, ok := av.FindChildView("notepad"); !ok {
		// Ensure side view exists
		if _, ok := av.GetChild("side"); !ok {
			SideViewWidthRatio = 0.2
			MainViewWidthRatio = 0.8
			av.AddChild("side", NewSideView(av.Calendar, av.Database, av.Config))
		}
		
		// Add notepad to side view if it doesn't exist
		if sideView, ok := av.GetChild("side"); ok {
			if _, ok := sideView.(*SideView).GetChild("notepad"); !ok {
				sideView.(*SideView).AddChild("notepad", NewNotepadView(av.Calendar, av.Database))
			}
		}
	}
	
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
	return av.UpdateCurrentView(g)
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


func (av *AppView) IsColorPickerActive() bool {
	return av.colorPickerActive
}


func (av *AppView) CloseColorPicker(g *gocui.Gui) error {
	av.colorPickerActive = false
	av.colorPickerEvent = nil
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

// refreshCurrentTimeHighlighting updates the current time highlighting for today's column
func (av *AppView) refreshCurrentTimeHighlighting(g *gocui.Gui) {
	// Find today's day view and refresh its current time highlighting
	if view, ok := av.FindChildView(WeekdayNames[time.Now().Weekday()]); ok {
		if dayView, ok := view.(*DayView); ok {
			// Force refresh of current time highlighting
			dayView.addCurrentTimeHighlight(g)
		}
	}
}

// Override UpdateChildren to automatically refresh current time highlighting
func (av *AppView) UpdateChildren(g *gocui.Gui) error {
	// Call the base UpdateChildren implementation
	err := av.BaseView.UpdateChildren(g)
	if err != nil {
		return err
	}
	
	// Automatically refresh current time highlighting after any child update
	av.refreshCurrentTimeHighlighting(g)
	
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
	sideViewWidth := 0
	if sideView, ok := av.GetChild("side"); ok {
		if sideView.Children().Len() > 0 {
			sideViewWidth = int(float64(av.W) * SideViewWidthRatio)
		}
	}
	mainViewWidth := av.W - sideViewWidth - 1

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

func (av *AppView) UpdateCurrentView(g *gocui.Gui) error {
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
			pos := utils.TimeToPosition(av.Calendar.CurrentDay.Date, timeView.Body)
			if pos >= 0 {
				y = pos
			}
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

