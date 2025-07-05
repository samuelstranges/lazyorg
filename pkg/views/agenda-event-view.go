package views

import (
	"fmt"
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
	location := aev.truncateField(event.Location, 15)
	description := aev.truncateField(event.Description, 25)
	
	// Format: "09:00-10:30 Event Name         Location       Description                1.5h"
	return fmt.Sprintf("%-11s %-20s %-15s %-25s %s", 
		startTime+"-"+endTimeStr, 
		name, 
		location, 
		description, 
		durationStr)
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