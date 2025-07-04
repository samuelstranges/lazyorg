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

// JumpToPrevEvent navigates to the previous event chronologically
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

// findMatches searches for events matching the criteria
func (av *AppView) findMatches(criteria database.SearchCriteria) []*calendar.Event {
	// Use enhanced database search with date/time filtering
	matches, err := av.Database.SearchEventsWithFilters(criteria)
	if err != nil {
		return []*calendar.Event{}
	}
	
	return matches
}

// GoToNextMatch navigates to the next search result
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

// GoToPrevMatch navigates to the previous search result
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