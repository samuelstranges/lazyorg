package views

import (
	"fmt"
	"os"
	"sort"
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
	DebugMode    bool
	
	colorPickerEvent  *EventView
	colorPickerActive bool
	copiedEvent       *calendar.Event
	
	// Search functionality
	searchQuery       string
	searchMatches     []*calendar.Event
	currentMatchIndex int
	isSearchActive    bool
}

// Helper function for debug logging with append
func (av *AppView) appendDebugLog(filename, content string) {
	if !av.DebugMode {
		return
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	file.WriteString(content)
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
		DebugMode:    db.DebugMode,
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
	// Get all events from database instead of current week only
	allEvents, err := av.Database.GetAllEvents()
	if err != nil || len(allEvents) == 0 {
		return
	}

	currentTime := av.Calendar.CurrentDay.Date
	
	// Find the next event after current time
	for _, event := range allEvents {
		// Normalize both times to local timezone for consistent comparison
		var eventTime time.Time
		// TEMPORARY FIX: ALL events stored in UTC are wrong - use original time instead of converting
		if event.Time.Location().String() == "UTC" {
			eventTime = time.Date(event.Time.Year(), event.Time.Month(), event.Time.Day(), event.Time.Hour(), event.Time.Minute(), event.Time.Second(), event.Time.Nanosecond(), time.Local)
		} else {
			eventTime = event.Time.In(time.Local)
		}
		currentTimeLocal := currentTime.In(time.Local)
		
		if eventTime.After(currentTimeLocal) {
			// Use the normalized time for consistent navigation
			av.Calendar.CurrentDay.Date = eventTime
			av.Calendar.UpdateWeek()
			return
		}
	}
	
	// If no event found after current time, wrap to first event
	if len(allEvents) > 0 {
		var firstEventTime time.Time
		// TEMPORARY FIX: ALL events stored in UTC are wrong - use original time instead of converting
		if allEvents[0].Time.Location().String() == "UTC" {
			firstEventTime = time.Date(allEvents[0].Time.Year(), allEvents[0].Time.Month(), allEvents[0].Time.Day(), allEvents[0].Time.Hour(), allEvents[0].Time.Minute(), allEvents[0].Time.Second(), allEvents[0].Time.Nanosecond(), time.Local)
		} else {
			firstEventTime = allEvents[0].Time.In(time.Local)
		}
		av.Calendar.CurrentDay.Date = firstEventTime
		av.Calendar.UpdateWeek()
	}
}

func (av *AppView) JumpToPrevEvent() {
	// Get all events from database instead of current week only
	allEvents, err := av.Database.GetAllEvents()
	if err != nil || len(allEvents) == 0 {
		return
	}

	currentTime := av.Calendar.CurrentDay.Date
	
	// Find the previous event before current time (iterate backwards)
	for i := len(allEvents) - 1; i >= 0; i-- {
		event := allEvents[i]
		// Normalize both times to local timezone for consistent comparison
		var eventTime time.Time
		// TEMPORARY FIX: ALL events stored in UTC are wrong - use original time instead of converting
		if event.Time.Location().String() == "UTC" {
			eventTime = time.Date(event.Time.Year(), event.Time.Month(), event.Time.Day(), event.Time.Hour(), event.Time.Minute(), event.Time.Second(), event.Time.Nanosecond(), time.Local)
		} else {
			eventTime = event.Time.In(time.Local)
		}
		currentTimeLocal := currentTime.In(time.Local)
		
		if eventTime.Before(currentTimeLocal) {
			// Use the normalized time for consistent navigation
			av.Calendar.CurrentDay.Date = eventTime
			av.Calendar.UpdateWeek()
			return
		}
	}
	
	// If no event found before current time, wrap to last event
	if len(allEvents) > 0 {
		lastEvent := allEvents[len(allEvents)-1]
		var lastEventTime time.Time
		// TEMPORARY FIX: ALL events stored in UTC are wrong - use original time instead of converting
		if lastEvent.Time.Location().String() == "UTC" {
			lastEventTime = time.Date(lastEvent.Time.Year(), lastEvent.Time.Month(), lastEvent.Time.Day(), lastEvent.Time.Hour(), lastEvent.Time.Minute(), lastEvent.Time.Second(), lastEvent.Time.Nanosecond(), time.Local)
		} else {
			lastEventTime = lastEvent.Time.In(time.Local)
		}
		av.Calendar.CurrentDay.Date = lastEventTime
		av.Calendar.UpdateWeek()
	}
}

func (av *AppView) getAllEventsFromWeek() []*calendar.Event {
	var allEvents []*calendar.Event
	
	// Always show debug info for now
	debugInfo := fmt.Sprintf("\n*** ENTERING getAllEventsFromWeek() ***\n")
	debugInfo += fmt.Sprintf("DebugMode = %t\n", av.DebugMode)
	debugInfo += fmt.Sprintf("Current week start: %s\n", av.Calendar.CurrentWeek.Days[0].Date.Format("2006-01-02"))
	debugInfo += fmt.Sprintf("Current week end: %s\n", av.Calendar.CurrentWeek.Days[6].Date.Format("2006-01-02"))
	av.appendDebugLog("/tmp/lazyorg_nav_debug.txt", debugInfo)
	
	// Collect all events from all days in the week
	for dayIndex, day := range av.Calendar.CurrentWeek.Days {
		debugInfo := fmt.Sprintf("Day %d (%s): %d events\n", dayIndex, day.Date.Format("Mon 2006-01-02"), len(day.Events))
		for eventIndex, event := range day.Events {
			debugInfo += fmt.Sprintf("  Event %d: %s at %s\n", eventIndex, event.Name, event.Time.Format("2006-01-02 15:04:05"))
		}
		av.appendDebugLog("/tmp/lazyorg_nav_debug.txt", debugInfo)
		allEvents = append(allEvents, day.Events...)
	}
	
	debugInfo = fmt.Sprintf("\n\n--- getAllEventsFromWeek() SORTING DEBUG ---\n")
	debugInfo += fmt.Sprintf("Total events before sorting: %d\n", len(allEvents))
	debugInfo += fmt.Sprintf("Events BEFORE sorting:\n")
	for i, event := range allEvents {
		localTime := event.Time.In(time.Local)
		debugInfo += fmt.Sprintf("  %d: %s at %s (Local: %s, TZ: %s, Unix: %d)\n", i, event.Name, event.Time.Format("2006-01-02 15:04:05"), localTime.Format("2006-01-02 15:04:05"), event.Time.Location().String(), event.Time.Unix())
	}
	av.appendDebugLog("/tmp/lazyorg_nav_debug.txt", debugInfo)
	
	// Sort by time - normalize timezones to avoid mixed UTC/Local storage issues
	sort.Slice(allEvents, func(i, j int) bool {
		// Fix obviously wrong timezone conversions for Morning events
		timeI := allEvents[i].Time
		timeJ := allEvents[j].Time
		
		// TEMPORARY FIX: ALL events stored in UTC are wrong - use original time instead of converting
		if allEvents[i].Time.Location().String() == "UTC" {
			// Events stored in UTC should actually be Local time, not converted to Local+10hours
			timeI = time.Date(timeI.Year(), timeI.Month(), timeI.Day(), timeI.Hour(), timeI.Minute(), timeI.Second(), timeI.Nanosecond(), time.Local)
		} else {
			timeI = timeI.In(time.Local)
		}
		
		if allEvents[j].Time.Location().String() == "UTC" {
			// Events stored in UTC should actually be Local time, not converted to Local+10hours
			timeJ = time.Date(timeJ.Year(), timeJ.Month(), timeJ.Day(), timeJ.Hour(), timeJ.Minute(), timeJ.Second(), timeJ.Nanosecond(), time.Local)
		} else {
			timeJ = timeJ.In(time.Local)
		}
		
		return timeI.Before(timeJ)
	})
	
	debugInfo = fmt.Sprintf("Events after sorting by time:\n")
	for i, event := range allEvents {
		localTime := event.Time.In(time.Local)
		debugInfo += fmt.Sprintf("  %d: %s at %s (Local: %s, Unix: %d)\n", i, event.Name, event.Time.Format("2006-01-02 15:04:05"), localTime.Format("2006-01-02 15:04:05"), event.Time.Unix())
	}
	debugInfo += fmt.Sprintf("--- End getAllEventsFromWeek() ---\n")
	av.appendDebugLog("/tmp/lazyorg_nav_debug.txt", debugInfo)
	
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
		
		// Normalize time for consistent navigation
		var eventTime time.Time
		if firstMatch.Time.Location().String() == "UTC" {
			eventTime = time.Date(firstMatch.Time.Year(), firstMatch.Time.Month(), firstMatch.Time.Day(), firstMatch.Time.Hour(), firstMatch.Time.Minute(), firstMatch.Time.Second(), firstMatch.Time.Nanosecond(), time.Local)
		} else {
			eventTime = firstMatch.Time.In(time.Local)
		}
		
		av.Calendar.CurrentDay.Date = eventTime
		av.Calendar.UpdateWeek()
	}
	
	return nil
}

func (av *AppView) findMatches(query string) []*calendar.Event {
	// Use database search instead of current week only
	matches, err := av.Database.SearchEvents(query)
	if err != nil {
		return []*calendar.Event{}
	}
	
	return matches
}

func (av *AppView) GoToNextMatch() error {
	if !av.isSearchActive || len(av.searchMatches) == 0 {
		return nil
	}
	
	av.currentMatchIndex = (av.currentMatchIndex + 1) % len(av.searchMatches)
	match := av.searchMatches[av.currentMatchIndex]
	
	// Normalize time for consistent navigation
	var eventTime time.Time
	if match.Time.Location().String() == "UTC" {
		eventTime = time.Date(match.Time.Year(), match.Time.Month(), match.Time.Day(), match.Time.Hour(), match.Time.Minute(), match.Time.Second(), match.Time.Nanosecond(), time.Local)
	} else {
		eventTime = match.Time.In(time.Local)
	}
	
	av.Calendar.CurrentDay.Date = eventTime
	av.Calendar.UpdateWeek()
	
	return nil
}

func (av *AppView) GoToPrevMatch() error {
	if !av.isSearchActive || len(av.searchMatches) == 0 {
		return nil
	}
	
	av.currentMatchIndex = (av.currentMatchIndex - 1 + len(av.searchMatches)) % len(av.searchMatches)
	match := av.searchMatches[av.currentMatchIndex]
	
	// Normalize time for consistent navigation
	var eventTime time.Time
	if match.Time.Location().String() == "UTC" {
		eventTime = time.Date(match.Time.Year(), match.Time.Month(), match.Time.Day(), match.Time.Hour(), match.Time.Minute(), match.Time.Second(), match.Time.Nanosecond(), time.Local)
	} else {
		eventTime = match.Time.In(time.Local)
	}
	
	av.Calendar.CurrentDay.Date = eventTime
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
				os.WriteFile("/tmp/lazyorg_debug.txt", []byte(debugInfo + finalDebug), 0644)
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

// refreshCurrentTimeHighlighting updates the current time highlighting for today's column
func (av *AppView) refreshCurrentTimeHighlighting(g *gocui.Gui) {
	// Iterate through all day views and update their current time highlighting
	for _, day := range av.Calendar.CurrentWeek.Days {
		if view, ok := av.FindChildView(WeekdayNames[day.Date.Weekday()]); ok {
			if dayView, ok := view.(*DayView); ok {
				// Call the new update method on each day view
				dayView.updateCurrentTimeHighlight(g)
			}
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

