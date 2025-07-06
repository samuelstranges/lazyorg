package views

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/samuelstranges/chronos/internal/utils"
	"github.com/jroimartin/gocui"
	"github.com/nsf/termbox-go"
)

type DayView struct {
	*BaseView

	Day      *calendar.Day
	TimeView *TimeView
	WeatherIcon string
	WeatherMaxTemp string
}

func NewDayView(name string, d *calendar.Day, tv *TimeView) *DayView {
	dv := &DayView{
		BaseView: NewBaseView(name),
		Day:      d,
		TimeView: tv,
	}

	return dv
}

func (dv *DayView) Update(g *gocui.Gui) error {
	v, err := g.SetView(
		dv.Name,
		dv.X,
		dv.Y,
		dv.X+dv.W,
		dv.Y+dv.H,
	)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}

	dv.updateBgColor(v)

	// Create title with weather if available
	title := dv.Day.FormatTitle()
	if dv.WeatherIcon != "" {
		if dv.WeatherMaxTemp != "" {
			title += " " + dv.WeatherMaxTemp + "°" + dv.WeatherIcon
		} else {
			title += " " + dv.WeatherIcon
		}
	}
	v.Title = title

	// Update current time highlighting (add or remove)
	if err = dv.updateCurrentTimeHighlight(g); err != nil {
		return err
	}

	if err = dv.updateChildViewProperties(g); err != nil {
		return err
	}

	if err = dv.UpdateChildren(g); err != nil {
		return err
	}

	return nil
}

func (dv *DayView) updateBgColor(v *gocui.View) {
	now := time.Now()
	if dv.Day.Date.Year() == now.Year() && dv.Day.Date.Month() == now.Month() && dv.Day.Date.Day() == now.Day() {
		v.BgColor = gocui.Attribute(termbox.ColorDarkGray)
	} else {
		v.BgColor = gocui.ColorDefault
	}
}

// updateCurrentTimeHighlight adds or removes the current time highlight based on whether it's today
func (dv *DayView) updateCurrentTimeHighlight(g *gocui.Gui) error {
	var (
		highlightViewName = dv.Name + "_current_time_highlight"
		now               = time.Now()
		v                 *gocui.View
		err               error
		f                 *os.File
	)
	
	// Check if any popup/form views exist - if so, don't create time highlighting
	// This prevents time highlighting from appearing over popups
	if views := g.Views(); views != nil {
		for _, view := range views {
			if strings.Contains(view.Name(), "Edit Event") || 
			   strings.Contains(view.Name(), "Add Event") ||
			   strings.Contains(view.Name(), "Goto") ||
			   strings.Contains(view.Name(), "Color") {
				// Remove any existing highlight and return early
				g.DeleteView(highlightViewName)
				return nil
			}
		}
	}

	// Debug logging
	if f, err = os.OpenFile("/tmp/chronos_currenttime_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "updateCurrentTimeHighlight: ViewDate=%s, Now=%s, SameYear=%t, SameMonth=%t, SameDay=%t\n",
			dv.Day.Date.Format("2006-01-02"),
			now.Format("2006-01-02"),
			dv.Day.Date.Year() == now.Year(),
			dv.Day.Date.Month() == now.Month(),
			dv.Day.Date.Day() == now.Day())
		f.Close()
	}

	// If this is not today's date, ensure the highlight view is removed
	if dv.Day.Date.Year() != now.Year() || dv.Day.Date.Month() != now.Month() || dv.Day.Date.Day() != now.Day() {
		if err = g.DeleteView(highlightViewName); err != nil && err != gocui.ErrUnknownView {
			return err
		}
		return nil
	}

	// If it is today, proceed to add/update the highlight
	currentHour := now.Hour()
	currentMinute := now.Minute()

	// Round to nearest half hour (0 or 30)
	currentHalfHour := 0
	if currentMinute >= 30 {
		currentHalfHour = 30
	}

	// Create a time for the current half-hour slot
	currentTime := time.Date(now.Year(), now.Month(), now.Day(), currentHour, currentHalfHour, 0, 0, now.Location())

	// Get the position of this time slot
	timePosition := utils.TimeToPosition(currentTime, dv.TimeView.Body)
	if timePosition < 0 {
		return nil // Time not in visible range
	}

	// Find event that starts exactly at current half-hour (for text coloring)
	var eventStartingNow *calendar.Event
	for _, event := range dv.Day.Events {
		eventTime := event.Time
		eventHour := eventTime.Hour()
		eventMinute := eventTime.Minute()
		eventHalfHour := 0
		if eventMinute >= 30 {
			eventHalfHour = 30
		}

		// Check if event starts exactly at current half-hour
		if eventHour == currentHour && eventHalfHour == currentHalfHour {
			eventStartingNow = event
			break
		}
	}

	// Find event that is currently running at this time (for background color)
	var runningEvent *calendar.Event
	for _, event := range dv.Day.Events {
		eventStartTime := event.Time
		eventEndTime := eventStartTime.Add(time.Duration(event.DurationHour * float64(time.Hour)))

		// Check if the current time falls within this event's duration
		if currentTime.After(eventStartTime) || currentTime.Equal(eventStartTime) {
			if currentTime.Before(eventEndTime) {
				runningEvent = event
				break
			}
		}
	}

	// Remove existing highlight view if it exists
	if err = g.DeleteView(highlightViewName); err != nil && err != gocui.ErrUnknownView {
		return err
	}

	x := dv.X
	y := dv.Y + timePosition
	w := dv.W
	h := 2 // Cover the half-hour slot (2 lines)

	// Create new highlight view
	v, err = g.SetView(highlightViewName, x, y, x+w, y+h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}

	v.Frame = false

	// Determine colors and content based on events
	if eventStartingNow != nil {
		// Event starts exactly at current time - use event's color for background, purple for text
		v.BgColor = eventStartingNow.Color
		v.FgColor = calendar.ColorCustomPurple
		// Display event name with purple foreground
		fmt.Fprintf(v, "%s\n", eventStartingNow.Name)
		fmt.Fprintf(v, "%s\n", eventStartingNow.Name)
	} else if runningEvent != nil {
		// Event is running at current time - use event's color for background, purple hashes
		v.BgColor = runningEvent.Color
		v.FgColor = calendar.ColorCustomPurple
		// Fill with purple hash characters
		hashLine := ""
		for i := 0; i < w-1; i++ {
			hashLine += "#"
		}
		fmt.Fprintf(v, "%s\n", hashLine)
		fmt.Fprintf(v, "%s\n", hashLine)
	} else {
		// No event at current time - use day view background, purple hashes
		isCurrentView := g.CurrentView() != nil && g.CurrentView().Name() == dv.Name
		if isCurrentView {
			// Today with cursor active - use black (matches cursor override)
			v.BgColor = gocui.ColorBlack
		} else {
			// Today without cursor - use grey
			v.BgColor = gocui.Attribute(termbox.ColorDarkGray)
		}
		v.FgColor = calendar.ColorCustomPurple
		// Fill with purple hash characters
		hashLine := ""
		for i := 0; i < w-1; i++ {
			hashLine += "#"
		}
		fmt.Fprintf(v, "%s\n", hashLine)
		fmt.Fprintf(v, "%s\n", hashLine)
	}

	return nil
}

func (dv *DayView) isEventAtCurrentTime(event *calendar.Event) bool {
	// Only check if this is today (same year, month, and day)
	now := time.Now()
	if dv.Day.Date.Year() != now.Year() || dv.Day.Date.Month() != now.Month() || dv.Day.Date.Day() != now.Day() {
		return false
	}
	currentHour := now.Hour()
	currentMinute := now.Minute()

	// Round to nearest half hour (0 or 30)
	currentHalfHour := 0
	if currentMinute >= 30 {
		currentHalfHour = 30
	}

	eventTime := event.Time
	eventHour := eventTime.Hour()
	eventMinute := eventTime.Minute()
	eventHalfHour := 0
	if eventMinute >= 30 {
		eventHalfHour = 30
	}

	return eventHour == currentHour && eventHalfHour == currentHalfHour
}

func (dv *DayView) updateChildViewProperties(g *gocui.Gui) error {

	eventViews := make(map[string]*EventView)
	for pair := dv.children.Oldest(); pair != nil; pair = pair.Next() {
		if eventView, ok := pair.Value.(*EventView); ok {
			eventViews[eventView.GetName()] = eventView
		}
	}

	// Sort events by time to check for consecutive events
	events := make([]*calendar.Event, len(dv.Day.Events))
	copy(events, dv.Day.Events)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Time.Before(events[j].Time)
	})
	
	for i, event := range events {
		x := dv.X
		timePosition := utils.TimeToPosition(event.Time, dv.TimeView.Body)
		y := dv.Y + timePosition
		w := dv.W
		h := utils.DurationToHeight(event.DurationHour) + 1
		

		if (y + h) >= (dv.Y + dv.H) {
			newHeight := (dv.Y + dv.H) - y
			if newHeight <= 0 {
				continue
			}
			h = newHeight
		}
		if y < dv.Y {
			continue
		}

		// Check if this event should have a bottom border
		showBottomBorder := false
		if i < len(events)-1 {
			nextEvent := events[i+1]
			eventEndTime := event.Time.Add(time.Duration(event.DurationHour * float64(time.Hour)))
			
			// Check if next event starts immediately after this one and has same color
			// Use a small tolerance for time comparison to handle minor precision differences
			timeDiff := nextEvent.Time.Sub(eventEndTime)
			if timeDiff >= 0 && timeDiff < time.Minute && nextEvent.Color == event.Color {
				showBottomBorder = true
			}
		}

		viewName := fmt.Sprintf("%s-%d", event.Name, event.Id)
		isCurrentTimeEvent := dv.isEventAtCurrentTime(event)
		
		if existingView, exists := eventViews[viewName]; exists {
			existingView.X, existingView.Y, existingView.W, existingView.H = x, y, w, h
			existingView.Event = event
			existingView.ShowBottomBorder = showBottomBorder
			existingView.IsCurrentTimeEvent = isCurrentTimeEvent
			delete(eventViews, viewName)
		} else {
			ev := NewEvenView(viewName, event)
			ev.X, ev.Y, ev.W, ev.H = x, y, w, h
			ev.ShowBottomBorder = showBottomBorder
			ev.IsCurrentTimeEvent = isCurrentTimeEvent
			dv.AddChild(viewName, ev)
		}
	}

	for viewName := range eventViews {
		if err := g.DeleteView(viewName); err != nil && err != gocui.ErrUnknownView {
			return err
		}
		// Also delete any border views
		borderViewName := viewName + "_border"
		if err := g.DeleteView(borderViewName); err != nil && err != gocui.ErrUnknownView {
			// Ignore error if border view doesn't exist
		}
		dv.children.Delete(viewName)
	}

	return nil
}

func (dv *DayView) IsOnEvent(y int) (*EventView, bool) {
	// Convert cursor position to absolute screen coordinates
	absoluteY := dv.Y + y
	for pair := dv.children.Newest(); pair != nil; pair = pair.Prev() {
		if eventView, ok := pair.Value.(*EventView); ok {
			// Subtract 1 from height to exclude the bottom padding/underline row from cursor detection
			detectableHeight := eventView.H - 1
			if detectableHeight < 1 {
				detectableHeight = 1 // Ensure at least 1 row is detectable
			}
			if absoluteY >= eventView.Y && absoluteY < (eventView.Y+detectableHeight) {
				return eventView, true
			}
		}
	}
	return nil, false
}

// SetWeatherData sets weather information for this day view
func (dv *DayView) SetWeatherData(icon, maxTemp string) {
	dv.WeatherIcon = icon
	dv.WeatherMaxTemp = maxTemp
}
