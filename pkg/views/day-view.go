package views

import (
	"fmt"
	"sort"
	"time"

	"github.com/HubertBel/lazyorg/internal/calendar"
	"github.com/HubertBel/lazyorg/internal/utils"
	"github.com/jroimartin/gocui"
	"github.com/nsf/termbox-go"
)

type DayView struct {
	*BaseView

	Day      *calendar.Day
	TimeView *TimeView
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

	v.Title = dv.Day.FormatTitle()

	// Add current time highlighting if this is today
	if err = dv.addCurrentTimeHighlight(g); err != nil {
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
	if dv.Day.Date.YearDay() == time.Now().YearDay() {
		v.BgColor = gocui.Attribute(termbox.ColorDarkGray)
	} else {
		v.BgColor = gocui.ColorDefault
	}
}

func (dv *DayView) addCurrentTimeHighlight(g *gocui.Gui) error {
	// Only highlight if this is today
	if dv.Day.Date.YearDay() != time.Now().YearDay() {
		return nil
	}

	now := time.Now()
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
	
	// Find event at current time (if any)
	var currentTimeEvent *calendar.Event
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
			currentTimeEvent = event
			break
		}
	}
	
	// Remove existing highlight view if it exists
	highlightViewName := dv.Name + "_current_time_highlight"
	if err := g.DeleteView(highlightViewName); err != nil && err != gocui.ErrUnknownView {
		return err
	}
	
	// Case 1: No event at current time - show cyan hashes
	if currentTimeEvent == nil {
		x := dv.X
		y := dv.Y + timePosition
		w := dv.W
		h := 2 // Cover the half-hour slot (2 lines)
		
		// Create new highlight view with cyan hashes
		v, err := g.SetView(highlightViewName, x, y, x+w, y+h)
		if err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			// Set background to match the day view background (transparent)
			if dv.Day.Date.YearDay() == time.Now().YearDay() {
				v.BgColor = gocui.Attribute(termbox.ColorDarkGray)
			} else {
				v.BgColor = gocui.ColorDefault
			}
			v.FgColor = gocui.ColorCyan
			v.Frame = false
			
			// Fill with cyan hash characters
			hashLine := ""
			for i := 0; i < w-1; i++ {
				hashLine += "#"
			}
			fmt.Fprintf(v, "%s\n", hashLine)
			fmt.Fprintf(v, "%s\n", hashLine)
		}
	} else {
		// Case 2: Event exists at current time
		// Check if event is in the "title" position (first line of its time slot)
		eventPosition := utils.TimeToPosition(currentTimeEvent.Time, dv.TimeView.Body)
		isInTitlePosition := (eventPosition == timePosition)
		
		if isInTitlePosition {
			// Event is in title position - we'll handle this by modifying event color in updateChildViewProperties
			// For now, just mark that we need to color the event text cyan
			// This will be handled when the event view is created
		} else {
			// Event exists elsewhere - show cyan hashes
			x := dv.X
			y := dv.Y + timePosition
			w := dv.W
			h := 2
			
			v, err := g.SetView(highlightViewName, x, y, x+w, y+h)
			if err != nil {
				if err != gocui.ErrUnknownView {
					return err
				}
				// Set background to match the day view background (transparent)
				if dv.Day.Date.YearDay() == time.Now().YearDay() {
					v.BgColor = gocui.Attribute(termbox.ColorDarkGray)
				} else {
					v.BgColor = gocui.ColorDefault
				}
				v.FgColor = gocui.ColorCyan
				v.Frame = false
				
				// Fill with cyan hash characters
				hashLine := ""
				for i := 0; i < w-1; i++ {
					hashLine += "#"
				}
				fmt.Fprintf(v, "%s\n", hashLine)
				fmt.Fprintf(v, "%s\n", hashLine)
			}
		}
	}
	
	return nil
}

func (dv *DayView) isEventAtCurrentTime(event *calendar.Event) bool {
	// Only check if this is today
	if dv.Day.Date.YearDay() != time.Now().YearDay() {
		return false
	}
	
	now := time.Now()
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
		y := dv.Y + utils.TimeToPosition(event.Time, dv.TimeView.Body)
		w := dv.W
		h := utils.DurationToHeight(event.DurationHour) + 1

		if (y + h) >= (dv.Y + dv.H) {
			newHeight := (dv.Y + dv.H) - y
			if newHeight <= 0 {
				continue
			}
			h = newHeight
		}
		if y <= dv.Y {
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
