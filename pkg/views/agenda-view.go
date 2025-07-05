package views

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/samuelstranges/chronos/internal/database"
	"github.com/samuelstranges/chronos/internal/utils"
	"github.com/jroimartin/gocui"
)

type AgendaView struct {
	*BaseView
	
	Calendar       *calendar.Calendar
	Database       *database.Database
	CurrentDate    time.Time
	Events         []*calendar.Event
	SelectedIndex  int
	HeaderHeight   int
}

func NewAgendaView(c *calendar.Calendar, db *database.Database) *AgendaView {
	av := &AgendaView{
		BaseView:      NewBaseView("agenda"),
		Calendar:      c,
		Database:      db,
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
	v.Title = fmt.Sprintf(" Agenda - %s ", av.CurrentDate.Format("Monday, January 2, 2006"))
	v.Clear()
	
	// Debug logging for main view
	if f, err := os.OpenFile("/tmp/chronos_agenda_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "AgendaView main Update: dims=(%d,%d,%d,%d), events=%d, view=%p\n", 
			av.X, av.Y, av.W, av.H, len(av.Events), v)
		if v != nil {
			vw, vh := v.Size()
			fmt.Fprintf(f, "  View size: (%d,%d)\n", vw, vh)
		}
		f.Close()
	}
	
	// Draw header
	av.drawHeader(v)
	
	// Load events for current date
	if err := av.loadEventsForDate(); err != nil {
		fmt.Fprintf(v, "\nError loading events: %v", err)
		return nil
	}
	
	// Write events directly to main view with full details
	for i, event := range av.Events {
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
	
	// Debug logging
	if f, err := os.OpenFile("/tmp/chronos_agenda_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "TEST: Wrote %d events directly to main view\n", len(av.Events))
		f.Close()
	}
	
	return nil
}

func (av *AgendaView) drawHeader(v *gocui.View) {
	// Debug logging
	if f, err := os.OpenFile("/tmp/chronos_agenda_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "drawHeader called: view=%p, width=%d\n", v, av.W)
		f.Close()
	}
	
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
	
	// Debug logging after header
	if f, err := os.OpenFile("/tmp/chronos_agenda_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "drawHeader complete: separator length=%d\n", len(separator))
		f.Close()
	}
}

func (av *AgendaView) loadEventsForDate() error {
	// Debug logging
	if f, err := os.OpenFile("/tmp/chronos_agenda_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "loadEventsForDate called for: %s\n", av.CurrentDate.Format("2006-01-02"))
		f.Close()
	}
	
	// Get events for the current date
	events, err := av.Database.GetEventsByDate(av.CurrentDate)
	if err != nil {
		if f, err := os.OpenFile("/tmp/chronos_agenda_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintf(f, "ERROR loading events: %v\n", err)
			f.Close()
		}
		return err
	}
	
	// Debug logging
	if f, err := os.OpenFile("/tmp/chronos_agenda_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "Found %d events for %s\n", len(events), av.CurrentDate.Format("2006-01-02"))
		for i, event := range events {
			fmt.Fprintf(f, "  Event %d: %s at %s\n", i, event.Name, event.Time.Format("15:04"))
		}
		f.Close()
	}
	
	// Sort events by time
	sort.Slice(events, func(i, j int) bool {
		return events[i].Time.Before(events[j].Time)
	})
	
	av.Events = events
	
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
	// Debug logging
	if f, err := os.OpenFile("/tmp/chronos_agenda_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "updateEventViews called: %d events to display\n", len(av.Events))
		f.Close()
	}
	
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
		
		// Debug logging
		if f, err := os.OpenFile("/tmp/chronos_agenda_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintf(f, "Creating event view %s at (%d,%d) size (%d,%d) for event: %s\n", 
				eventViewName, x, y, w, h, event.Name)
			f.Close()
		}
		
		// Skip if outside view bounds
		if y >= av.Y+av.H-1 {
			if f, err := os.OpenFile("/tmp/chronos_agenda_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				fmt.Fprintf(f, "Skipping event %s - outside bounds: y=%d, maxY=%d\n", eventViewName, y, av.Y+av.H-1)
				f.Close()
			}
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