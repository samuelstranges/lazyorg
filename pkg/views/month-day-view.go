package views

import (
	"fmt"
	"time"

	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/jroimartin/gocui"
	"github.com/nsf/termbox-go"
)

type MonthDayView struct {
	*BaseView
	
	Date         time.Time
	CurrentMonth time.Time
	Events       []*calendar.Event
	
	// Visual properties
	IsCurrentMonth bool
	IsToday        bool
	IsSelected     bool
}

func NewMonthDayView(name string, date time.Time, currentMonth time.Time) *MonthDayView {
	mdv := &MonthDayView{
		BaseView:       NewBaseView(name),
		Date:           date,
		CurrentMonth:   currentMonth,
		Events:         make([]*calendar.Event, 0),
		IsCurrentMonth: date.Month() == currentMonth.Month() && date.Year() == currentMonth.Year(),
		IsToday:        isToday(date),
	}
	
	return mdv
}

func (mdv *MonthDayView) Update(g *gocui.Gui) error {
	// Skip if dimensions are invalid
	if mdv.W <= 0 || mdv.H <= 0 {
		return nil
	}
	
	v, err := g.SetView(
		mdv.Name,
		mdv.X,
		mdv.Y,
		mdv.X+mdv.W,
		mdv.Y+mdv.H,
	)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	
	// Set frame and colors
	v.Frame = true
	mdv.updateColors(v)
	
	// Clear and draw content
	v.Clear()
	
	// Draw day number
	dayNum := mdv.Date.Day()
	dayStr := fmt.Sprintf("%2d", dayNum)
	if mdv.IsToday {
		dayStr += "•" // Add dot for today
	}
	fmt.Fprintf(v, "%s\n", dayStr)
	
	// Draw events (as bullet points)
	maxEvents := mdv.H - 2 // Leave space for day number and borders
	if maxEvents < 0 {
		maxEvents = 0
	}
	
	eventsToShow := mdv.Events
	if len(eventsToShow) > maxEvents {
		eventsToShow = eventsToShow[:maxEvents]
	}
	
	for _, event := range eventsToShow {
		eventLine := fmt.Sprintf("• %s", mdv.truncateEventName(event.Name))
		fmt.Fprintf(v, "%s\n", eventLine)
	}
	
	// If there are more events than can be displayed, show count
	if len(mdv.Events) > maxEvents {
		remaining := len(mdv.Events) - maxEvents
		fmt.Fprintf(v, "• +%d more\n", remaining)
	}
	
	return nil
}

func (mdv *MonthDayView) updateColors(v *gocui.View) {
	if mdv.IsToday {
		// Today - highlighted background
		v.BgColor = gocui.Attribute(termbox.ColorDarkGray)
		v.FgColor = gocui.ColorWhite
	} else if !mdv.IsCurrentMonth {
		// Different month - muted colors
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.Attribute(termbox.ColorDarkGray)
	} else {
		// Current month - normal colors
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorDefault
	}
}

func (mdv *MonthDayView) truncateEventName(name string) string {
	maxWidth := mdv.W - 3 // Leave space for bullet and borders
	if maxWidth < 1 {
		return ""
	}
	
	if len(name) <= maxWidth {
		return name
	}
	
	// Truncate with ellipsis
	if maxWidth > 3 {
		return name[:maxWidth-3] + "..."
	}
	return name[:maxWidth]
}

func (mdv *MonthDayView) LoadEvents(events []*calendar.Event) {
	mdv.Events = make([]*calendar.Event, 0)
	
	for _, event := range events {
		// Check if event is on this day
		if isSameDay(event.Time, mdv.Date) {
			mdv.Events = append(mdv.Events, event)
		}
	}
}

func isToday(date time.Time) bool {
	now := time.Now()
	return date.Year() == now.Year() && 
		   date.Month() == now.Month() && 
		   date.Day() == now.Day()
}

func isSameDay(date1, date2 time.Time) bool {
	return date1.Year() == date2.Year() && 
		   date1.Month() == date2.Month() && 
		   date1.Day() == date2.Day()
}