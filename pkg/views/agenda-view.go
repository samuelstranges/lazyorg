package views

import (
	"fmt"
	"sort"
	"time"

	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/samuelstranges/chronos/internal/database"
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
	
	// Draw header
	av.drawHeader(v)
	
	// Load events for current date
	if err := av.loadEventsForDate(); err != nil {
		fmt.Fprintf(v, "\nError loading events: %v", err)
		return nil
	}
	
	// Update child event views
	if err := av.updateEventViews(g); err != nil {
		return err
	}
	
	// Update children
	if err := av.UpdateChildren(g); err != nil {
		return err
	}
	
	return nil
}

func (av *AgendaView) drawHeader(v *gocui.View) {
	// Line 1: Empty for spacing
	fmt.Fprintln(v, "")
	
	// Line 2: Column headers
	fmt.Fprintf(v, "%-11s %-20s %-15s %-25s %s\n", 
		"Time", "Event", "Location", "Description", "Duration")
	
	// Line 3: Separator
	separator := ""
	for i := 0; i < av.W-2; i++ {
		separator += "-"
	}
	fmt.Fprintln(v, separator)
}

func (av *AgendaView) loadEventsForDate() error {
	// Get events for the current date
	events, err := av.Database.GetEventsByDate(av.CurrentDate)
	if err != nil {
		return err
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
	// Clear existing event views
	av.clearEventViews(g)
	
	// Create event views for each event
	for i, event := range av.Events {
		eventViewName := fmt.Sprintf("agenda_event_%d", i)
		
		// Calculate position
		x := av.X + 1
		y := av.Y + av.HeaderHeight + i
		w := av.W - 2
		h := 1
		
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
	if len(av.Events) == 0 {
		return av.Name
	}
	return fmt.Sprintf("agenda_event_%d", av.SelectedIndex)
}