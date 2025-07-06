package views

import (
	"fmt"
	"os"
	"time"

	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/samuelstranges/chronos/internal/utils"
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
	
	// No individual frames - use shared grid
	v.Frame = false
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
	maxEvents := mdv.H - 2 // Leave space for day number AND potential grid line/border
	if maxEvents < 0 {
		maxEvents = 0
	}
	
	
	var eventsToShow []*calendar.Event
	var hasOverflow bool
	
	if len(mdv.Events) > maxEvents && maxEvents > 0 {
		// More events than can fit - reserve last line for overflow message
		eventsToShow = mdv.Events[:maxEvents-1]
		hasOverflow = true
		
		// Debug overflow logic
		if f, err := os.OpenFile("/tmp/chronos_overflow_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintf(f, "OVERFLOW: showing %d events, hiding %d events\n", len(eventsToShow), len(mdv.Events)-len(eventsToShow))
			f.Close()
		}
	} else {
		// All events fit, or no space at all
		eventsToShow = mdv.Events
		hasOverflow = false
		
		// Debug no overflow
		if len(mdv.Events) > 0 {
			if f, err := os.OpenFile("/tmp/chronos_overflow_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				fmt.Fprintf(f, "NO OVERFLOW: showing all %d events\n", len(eventsToShow))
				f.Close()
			}
		}
	}
	
	for _, event := range eventsToShow {
		// Use ANSI color codes for the event text
		eventColor := event.Color
		if eventColor == gocui.ColorDefault || eventColor == gocui.ColorBlack {
			// Fallback to a visible color if somehow the event has no color
			eventColor = gocui.ColorBlue
		}
		
		eventTime := utils.FormatHourFromTime(event.Time)
		coloredEventName := calendar.WrapTextWithColor(mdv.truncateEventName(event.Name), eventColor)
		eventLine := fmt.Sprintf("%s %s", eventTime, coloredEventName)
		fmt.Fprintf(v, "%s\n", eventLine)
	}
	
	// If there are more events than can be displayed, show count
	if hasOverflow {
		remaining := len(mdv.Events) - len(eventsToShow)
		fmt.Fprintf(v, "& %d more", remaining)
		
		// Debug overflow message
		if f, err := os.OpenFile("/tmp/chronos_overflow_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintf(f, "PRINTED OVERFLOW MESSAGE: 'and %d more'\n", remaining)
			f.Close()
		}
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
	maxWidth := mdv.W - 6 // Leave space for time (5 chars) and padding (1 char)
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
