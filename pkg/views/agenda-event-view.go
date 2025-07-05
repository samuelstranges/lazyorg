package views

import (
	"fmt"
	"os"
	"time"

	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/samuelstranges/chronos/internal/utils"
	"github.com/jroimartin/gocui"
)

type AgendaEventView struct {
	*BaseView
	
	Event    *calendar.Event
	Selected bool
}

func NewAgendaEventView(name string, event *calendar.Event) *AgendaEventView {
	return &AgendaEventView{
		BaseView: NewBaseView(name),
		Event:    event,
		Selected: false,
	}
}

func (aev *AgendaEventView) Update(g *gocui.Gui) error {
	// Skip if dimensions are invalid
	if aev.W <= 0 || aev.H <= 0 {
		return nil
	}
	
	v, err := g.SetView(
		aev.Name,
		aev.X,
		aev.Y,
		aev.X+aev.W,
		aev.Y+aev.H,
	)
	if err != nil {
		if err != gocui.ErrUnknownView {
			// Debug logging for view creation errors
			if f, err := os.OpenFile("/tmp/chronos_agenda_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				fmt.Fprintf(f, "ERROR creating view %s: %v\n", aev.Name, err)
				f.Close()
			}
			return err
		}
	}
	
	// Set colors based on selection state
	if aev.Selected {
		v.BgColor = gocui.ColorBlack
		v.FgColor = gocui.ColorWhite
	} else {
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorDefault
	}
	
	v.Frame = false
	v.Clear()
	
	// Format the event information
	eventText := aev.formatEventLine()
	
	// Debug logging
	if f, err := os.OpenFile("/tmp/chronos_agenda_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "AgendaEventView %s Update: text='%s', selected=%t, dims=(%d,%d,%d,%d)\n", 
			aev.Name, eventText, aev.Selected, aev.X, aev.Y, aev.W, aev.H)
		f.Close()
	}
	
	fmt.Fprint(v, eventText)
	
	return nil
}

func (aev *AgendaEventView) formatEventLine() string {
	if aev.Event == nil {
		return ""
	}
	
	event := aev.Event
	
	// Calculate start and end times
	startTime := utils.FormatHourFromTime(event.Time)
	duration := time.Duration(event.DurationHour * float64(time.Hour))
	endTime := event.Time.Add(duration)
	endTimeStr := utils.FormatHourFromTime(endTime)
	
	// Format duration (e.g., "1.5h")
	durationStr := fmt.Sprintf("%.1fh", event.DurationHour)
	if event.DurationHour == float64(int(event.DurationHour)) {
		durationStr = fmt.Sprintf("%.0fh", event.DurationHour)
	}
	
	// Truncate fields to fit the line
	name := aev.truncateField(event.Name, 20)
	
	// Temporary simplified format for debugging
	return fmt.Sprintf("%s-%s %s (%s)", 
		startTime, endTimeStr, name, durationStr)
}

func (aev *AgendaEventView) truncateField(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}
	
	if maxWidth > 3 {
		return text[:maxWidth-3] + "..."
	}
	return text[:maxWidth]
}

func (aev *AgendaEventView) SetSelected(selected bool) {
	aev.Selected = selected
}