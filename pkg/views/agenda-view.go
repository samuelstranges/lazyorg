package views

import (
	"fmt"
	"sort"
	"time"

	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/samuelstranges/chronos/internal/eventmanager"
	"github.com/samuelstranges/chronos/internal/utils"
	"github.com/jroimartin/gocui"
)

type AgendaView struct {
	*BaseView
	
	Calendar       *calendar.Calendar
	EventManager   *eventmanager.EventManager
	CurrentDate    time.Time
	Events         []*calendar.Event
	SelectedIndex  int
	HeaderHeight   int
}

func NewAgendaView(c *calendar.Calendar, em *eventmanager.EventManager) *AgendaView {
	av := &AgendaView{
		BaseView:      NewBaseView("agenda"),
		Calendar:      c,
		EventManager:  em,
		CurrentDate:   c.CurrentDay.Date,
		Events:        make([]*calendar.Event, 0),
		SelectedIndex: 0,
		HeaderHeight:  3, // Space for header and column titles
	}
	
	return av
}

func (av *AgendaView) Update(g *gocui.Gui) error {
	// Skip if dimensions are invalid
	if av.W <= 0 || av.H <= 0 {
		return nil
	}
	
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
	}
	
	v.Frame = true
	v.Clear()
	
	// Draw header
	av.drawHeader(v)
	
	// Load events for current date
	if err := av.loadEventsForDate(); err != nil {
		fmt.Fprintf(v, "\nError loading events: %v", err)
		return nil
	}
	
	// Calculate available space for events
	maxEvents := av.H - av.HeaderHeight - 1 // Leave space for header and potential bottom border
	if maxEvents < 0 {
		maxEvents = 0
	}
	
	// Determine which events to show and if overflow message is needed
	var eventsToShow []*calendar.Event
	var hasOverflow bool
	
	if len(av.Events) > maxEvents && maxEvents > 0 {
		// More events than can fit - reserve last line for overflow message
		eventsToShow = av.Events[:maxEvents-1]
		hasOverflow = true
	} else {
		// All events fit, or no space at all
		eventsToShow = av.Events
		hasOverflow = false
	}
	
	// Write events directly to main view with full details
	for i, event := range eventsToShow {
		startTime := utils.FormatHourFromTime(event.Time)
		duration := time.Duration(event.DurationHour * float64(time.Hour))
		endTime := event.Time.Add(duration)
		endTimeStr := utils.FormatHourFromTime(endTime)
		
		// Format duration
		durationStr := fmt.Sprintf("%.1fh", event.DurationHour)
		if event.DurationHour == float64(int(event.DurationHour)) {
			durationStr = fmt.Sprintf("%.0fh", event.DurationHour)
		}
		
		// Truncate fields to fit on screen with new column sizes
		name := av.truncateField(event.Name, 20)
		location := av.truncateField(event.Location, 37)  // 2.5x larger (15 * 2.5 ≈ 37)
		description := av.truncateField(event.Description, 25)
		
		// Apply ANSI color to the event name
		coloredName := calendar.WrapTextWithColor(name, event.Color)
		
		// Format the complete event line with reordered columns: Time, Event, Duration, Location, Description
		// Note: We need to pad the colored name manually since printf can't handle ANSI codes in width calculations
		paddedColoredName := coloredName
		// Add padding to reach 20 characters (visible length)
		namePadding := 20 - len(name) // Use original name length for padding calculation
		for j := 0; j < namePadding; j++ {
			paddedColoredName += " "
		}
		
		eventLine := fmt.Sprintf(" %-11s %s %-8s %-37s %s", 
			fmt.Sprintf("%s-%s", startTime, endTimeStr),
			paddedColoredName,
			durationStr,
			location,
			description)
		
		// Add selection indicator at the front
		if i == av.SelectedIndex {
			eventLine = fmt.Sprintf("→%s", eventLine[1:])  // Replace first space with arrow
		}
		
		fmt.Fprintln(v, eventLine)
	}
	
	// If there are more events than can be displayed, show count
	if hasOverflow {
		remaining := len(av.Events) - len(eventsToShow)
		fmt.Fprintf(v, " & %d more events", remaining)
	}
	
	return nil
}

func (av *AgendaView) drawHeader(v *gocui.View) {
	// Line 1: Empty for spacing
	fmt.Fprintln(v, "")
	
	// Line 2: Column headers - reordered: Time, Event, Duration, Location, Description
	// Add space at start for arrow positioning
	fmt.Fprintf(v, " %-11s %-20s %-8s %-37s %s\n", 
		"Time", "Event", "Duration", "Location", "Description")
	
	// Line 3: Separator
	separator := " "  // Start with space for arrow positioning
	for i := 0; i < av.W-3; i++ {
		separator += "-"
	}
	fmt.Fprintln(v, separator)
}

func (av *AgendaView) loadEventsForDate() error {
	// Get events for the current date
	events, err := av.EventManager.GetEventsByDate(av.CurrentDate)
	if err != nil {
		return err
	}
	
	// Convert UTC events to local time for display
	localEvents := make([]*calendar.Event, len(events))
	for i, event := range events {
		localEvent := *event
		localEvent.Time = event.Time.In(time.Local)
		localEvents[i] = &localEvent
	}
	
	// Sort events by time
	sort.Slice(localEvents, func(i, j int) bool {
		return localEvents[i].Time.Before(localEvents[j].Time)
	})
	
	av.Events = localEvents
	
	// Ensure selected index is valid
	if av.SelectedIndex >= len(av.Events) {
		av.SelectedIndex = len(av.Events) - 1
	}
	if av.SelectedIndex < 0 {
		av.SelectedIndex = 0
	}
	
	return nil
}

func (av *AgendaView) updateEventViews(g *gocui.Gui) error {
	// Clear existing event views
	av.clearEventViews(g)
	
	// Create event views for each event
	for i, event := range av.Events {
		eventViewName := fmt.Sprintf("agenda_event_%d", i)
		
		// Calculate position relative to parent view (not screen coordinates)
		// The event views should be positioned within the main agenda view
		x := av.X + 1  // 1 pixel from left edge of parent
		y := av.Y + av.HeaderHeight + i  // After header + event index
		w := av.W - 2  // 2 pixels narrower than parent (1 pixel margin on each side)
		h := 1         // 1 line height per event
		
		// Skip if outside view bounds
		if y >= av.Y+av.H-1 {
			break
		}
		
		// Create event view
		eventView := NewAgendaEventView(eventViewName, event)
		eventView.SetProperties(x, y, w, h)
		eventView.SetSelected(i == av.SelectedIndex)
		
		av.AddChild(eventViewName, eventView)
	}
	
	return nil
}

func (av *AgendaView) clearEventViews(g *gocui.Gui) {
	// Remove all agenda_event_* views
	childrenToRemove := make([]string, 0)
	
	for pair := av.children.Oldest(); pair != nil; pair = pair.Next() {
		viewName := pair.Key
		if len(viewName) >= 13 && viewName[:13] == "agenda_event_" {
			childrenToRemove = append(childrenToRemove, viewName)
			g.DeleteView(viewName)
		}
	}
	
	for _, viewName := range childrenToRemove {
		av.children.Delete(viewName)
	}
}

func (av *AgendaView) MoveSelection(direction int) {
	if len(av.Events) == 0 {
		return
	}
	
	newIndex := av.SelectedIndex + direction
	
	// Bounds checking
	if newIndex < 0 {
		newIndex = 0
	} else if newIndex >= len(av.Events) {
		newIndex = len(av.Events) - 1
	}
	
	av.SelectedIndex = newIndex
}

func (av *AgendaView) GetSelectedEvent() *calendar.Event {
	if av.SelectedIndex >= 0 && av.SelectedIndex < len(av.Events) {
		return av.Events[av.SelectedIndex]
	}
	return nil
}

func (av *AgendaView) SetCurrentDate(date time.Time) {
	av.CurrentDate = date
	av.SelectedIndex = 0 // Reset selection when date changes
}

func (av *AgendaView) GetSelectedEventViewName() string {
	// Since all events are displayed in the main agenda view (not as separate child views),
	// always return the main agenda view name for focus management
	return av.Name
}

func (av *AgendaView) truncateField(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}
	
	if maxWidth > 3 {
		return text[:maxWidth-3] + "..."
	}
	return text[:maxWidth]
}