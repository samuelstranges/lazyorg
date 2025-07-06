package views

import (
	"fmt"
	"sort"
	"time"

	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/samuelstranges/chronos/internal/database"
	"github.com/jroimartin/gocui"
)

// JumpToNextEvent navigates to the next event chronologically
func (av *AppView) JumpToNextEvent() {
	// Get all events from EventManager and convert to local time
	allEvents, err := av.EventManager.GetAllEvents()
	if err != nil || len(allEvents) == 0 {
		return
	}

	// Convert UTC events to local time for comparison
	localEvents := make([]*calendar.Event, len(allEvents))
	for i, event := range allEvents {
		localEvent := *event
		localEvent.Time = event.Time.In(time.Local)
		localEvents[i] = &localEvent
	}

	currentTime := av.Calendar.CurrentDay.Date
	
	// Find the next event after current time
	for _, event := range localEvents {
		if event.Time.After(currentTime) {
			av.Calendar.CurrentDay.Date = event.Time
			av.Calendar.UpdateWeek()
			return
		}
	}
	
	// If no event found after current time, wrap to first event
	if len(localEvents) > 0 {
		av.Calendar.CurrentDay.Date = localEvents[0].Time
		av.Calendar.UpdateWeek()
	}
}

// JumpToPrevEvent navigates to the previous event chronologically
func (av *AppView) JumpToPrevEvent() {
	// Get all events from EventManager and convert to local time
	allEvents, err := av.EventManager.GetAllEvents()
	if err != nil || len(allEvents) == 0 {
		return
	}

	// Convert UTC events to local time for comparison
	localEvents := make([]*calendar.Event, len(allEvents))
	for i, event := range allEvents {
		localEvent := *event
		localEvent.Time = event.Time.In(time.Local)
		localEvents[i] = &localEvent
	}

	currentTime := av.Calendar.CurrentDay.Date
	
	// Find the previous event before current time (iterate backwards)
	for i := len(localEvents) - 1; i >= 0; i-- {
		event := localEvents[i]
		if event.Time.Before(currentTime) {
			av.Calendar.CurrentDay.Date = event.Time
			av.Calendar.UpdateWeek()
			return
		}
	}
	
	// If no event found before current time, wrap to last event
	if len(localEvents) > 0 {
		lastEvent := localEvents[len(localEvents)-1]
		av.Calendar.CurrentDay.Date = lastEvent.Time
		av.Calendar.UpdateWeek()
	}
}

// JumpToEndOfEvent navigates to end of current event, or end of next event if not in one
func (av *AppView) JumpToEndOfEvent() {
	// Get all events from EventManager and convert to local time
	allEvents, err := av.EventManager.GetAllEvents()
	if err != nil || len(allEvents) == 0 {
		return
	}

	// Convert UTC events to local time for comparison
	localEvents := make([]*calendar.Event, len(allEvents))
	for i, event := range allEvents {
		localEvent := *event
		localEvent.Time = event.Time.In(time.Local)
		localEvents[i] = &localEvent
	}

	currentTime := av.Calendar.CurrentDay.Date
	
	// Check if we're currently within an event
	for _, event := range localEvents {
		eventStart := event.Time
		eventEnd := eventStart.Add(time.Duration(event.DurationHour * float64(time.Hour)))
		// Move to the last 30-minute slot of the event, not beyond it
		eventLastSlot := eventEnd.Add(-30 * time.Minute)
		
		// If current time is within this event (inclusive of start, exclusive of end)
		// but NOT already at the end of the event
		if (currentTime.Equal(eventStart) || currentTime.After(eventStart)) && currentTime.Before(eventEnd) && !currentTime.Equal(eventLastSlot) {
			av.Calendar.CurrentDay.Date = eventLastSlot
			av.Calendar.UpdateWeek()
			return
		}
	}
	
	// If not within an event, jump to END of next event
	for _, event := range localEvents {
		if event.Time.After(currentTime) {
			eventEnd := event.Time.Add(time.Duration(event.DurationHour * float64(time.Hour)))
			// Move to the last 30-minute slot of the event, not beyond it
			eventLastSlot := eventEnd.Add(-30 * time.Minute)
			av.Calendar.CurrentDay.Date = eventLastSlot
			av.Calendar.UpdateWeek()
			return
		}
	}
	
	// If no event found after current time, wrap to END of first event
	if len(localEvents) > 0 {
		firstEventEnd := localEvents[0].Time.Add(time.Duration(localEvents[0].DurationHour * float64(time.Hour)))
		// Move to the last 30-minute slot of the event, not beyond it
		firstEventLastSlot := firstEventEnd.Add(-30 * time.Minute)
		av.Calendar.CurrentDay.Date = firstEventLastSlot
		av.Calendar.UpdateWeek()
	}
}

// getAllEventsFromWeek returns all events from the current week, sorted chronologically
func (av *AppView) getAllEventsFromWeek() []*calendar.Event {
	var allEvents []*calendar.Event
	
	// Always show debug info for now
	debugInfo := fmt.Sprintf("\n*** ENTERING getAllEventsFromWeek() ***\n")
	debugInfo += fmt.Sprintf("DebugMode = %t\n", av.DebugMode)
	debugInfo += fmt.Sprintf("Current week start: %s\n", av.Calendar.CurrentWeek.Days[0].Date.Format("2006-01-02"))
	debugInfo += fmt.Sprintf("Current week end: %s\n", av.Calendar.CurrentWeek.Days[6].Date.Format("2006-01-02"))
	av.appendDebugLog("/tmp/chronos_nav_debug.txt", debugInfo)
	
	// Collect all events from all days in the week
	for dayIndex, day := range av.Calendar.CurrentWeek.Days {
		debugInfo := fmt.Sprintf("Day %d (%s): %d events\n", dayIndex, day.Date.Format("Mon 2006-01-02"), len(day.Events))
		for eventIndex, event := range day.Events {
			debugInfo += fmt.Sprintf("  Event %d: %s at %s\n", eventIndex, event.Name, event.Time.Format("2006-01-02 15:04:05"))
		}
		av.appendDebugLog("/tmp/chronos_nav_debug.txt", debugInfo)
		allEvents = append(allEvents, day.Events...)
	}
	
	debugInfo = fmt.Sprintf("\n\n--- getAllEventsFromWeek() SORTING DEBUG ---\n")
	debugInfo += fmt.Sprintf("Total events before sorting: %d\n", len(allEvents))
	debugInfo += fmt.Sprintf("Events BEFORE sorting:\n")
	for i, event := range allEvents {
		localTime := event.Time.In(time.Local)
		debugInfo += fmt.Sprintf("  %d: %s at %s (Local: %s, TZ: %s, Unix: %d)\n", i, event.Name, event.Time.Format("2006-01-02 15:04:05"), localTime.Format("2006-01-02 15:04:05"), event.Time.Location().String(), event.Time.Unix())
	}
	av.appendDebugLog("/tmp/chronos_nav_debug.txt", debugInfo)
	
	// Sort by time
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].Time.Before(allEvents[j].Time)
	})
	
	debugInfo = fmt.Sprintf("Events after sorting by time:\n")
	for i, event := range allEvents {
		localTime := event.Time.In(time.Local)
		debugInfo += fmt.Sprintf("  %d: %s at %s (Local: %s, Unix: %d)\n", i, event.Name, event.Time.Format("2006-01-02 15:04:05"), localTime.Format("2006-01-02 15:04:05"), event.Time.Unix())
	}
	debugInfo += fmt.Sprintf("--- End getAllEventsFromWeek() ---\n")
	av.appendDebugLog("/tmp/chronos_nav_debug.txt", debugInfo)
	
	return allEvents
}

// StartSearch initiates a search for events
func (av *AppView) StartSearch(g *gocui.Gui) error {
	if popup, ok := av.GetChild("popup"); ok {
		if popupView, ok := popup.(*EventPopupView); ok {
			// Set up the search callback
			popupView.SearchCallback = func(criteria database.SearchCriteria) error {
				return av.executeSearchQuery(criteria)
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

// executeSearchQuery performs the actual search and navigates to first result
func (av *AppView) executeSearchQuery(criteria database.SearchCriteria) error {
	av.searchQuery = criteria.Query
	av.searchMatches = av.findMatches(criteria)
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

// findMatches searches for events matching the criteria
func (av *AppView) findMatches(criteria database.SearchCriteria) []*calendar.Event {
	// Use enhanced EventManager search with date/time filtering
	matches, err := av.EventManager.SearchEventsWithFilters(criteria)
	if err != nil {
		return []*calendar.Event{}
	}
	
	// Convert UTC events to local time for display
	localMatches := make([]*calendar.Event, len(matches))
	for i, event := range matches {
		localEvent := *event
		localEvent.Time = event.Time.In(time.Local)
		localMatches[i] = &localEvent
	}
	
	return localMatches
}

// GoToNextMatch navigates to the next search result
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

// GoToPrevMatch navigates to the previous search result
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

// ClearSearch clears the current search state
func (av *AppView) ClearSearch() {
	av.isSearchActive = false
	av.searchQuery = ""
	av.searchMatches = nil
	av.currentMatchIndex = 0
}

// GetSearchStatus returns the current search status string
func (av *AppView) GetSearchStatus() string {
	if !av.isSearchActive || len(av.searchMatches) == 0 {
		return ""
	}
	
	return fmt.Sprintf("%d/%d matches for '%s'", av.currentMatchIndex+1, len(av.searchMatches), av.searchQuery)
}

// showSearchStatus displays the search status in the UI
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

// JumpToStartOfDay moves the cursor to 00:00 of the current day
func (av *AppView) JumpToStartOfDay() {
	currentDate := av.Calendar.CurrentDay.Date
	startOfDay := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, currentDate.Location())
	av.Calendar.CurrentDay.Date = startOfDay
	av.Calendar.UpdateWeek()
}

// JumpToEndOfDay moves the cursor to 23:30 of the current day
func (av *AppView) JumpToEndOfDay() {
	currentDate := av.Calendar.CurrentDay.Date
	endOfDay := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 23, 30, 0, 0, currentDate.Location())
	av.Calendar.CurrentDay.Date = endOfDay
	av.Calendar.UpdateWeek()
}

// ShowGotoPopup displays the goto date/time popup
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