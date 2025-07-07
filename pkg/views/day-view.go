package views

import (
	"fmt"
	"sort"
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
		// Use the new viewport-aware position calculation
		timePosition := utils.TimeToPositionWithViewport(event.Time, dv.TimeView.GetViewportStart())
		visibleSlots := dv.TimeView.GetVisibleSlots()
		
		// Calculate event end time and position
		eventEndTime := event.Time.Add(time.Duration(event.DurationHour * float64(time.Hour)))
		eventEndPosition := utils.TimeToPositionWithViewport(eventEndTime, dv.TimeView.GetViewportStart())
		
		// Skip events that are completely outside the viewport
		if timePosition < 0 && eventEndPosition <= 0 {
			continue // Event ends before viewport starts
		}
		if timePosition >= visibleSlots {
			continue // Event starts after viewport ends
		}
		
		// Calculate display position and height
		var y, h int
		if timePosition < 0 {
			// Event starts before viewport - show from top of viewport
			y = dv.Y
			// Calculate how much of the event is visible
			if eventEndPosition > visibleSlots {
				h = visibleSlots // Event extends past viewport end
			} else {
				h = eventEndPosition // Show until event ends
			}
		} else {
			// Event starts within viewport
			y = dv.Y + timePosition
			h = utils.DurationToHeight(event.DurationHour) + 1
			
			// Truncate event height if it extends beyond visible area
			if timePosition + h > visibleSlots {
				h = visibleSlots - timePosition
			}
		}
		
		w := dv.W
		
		
		// Ensure minimum height
		if h <= 0 {
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
		
		if existingView, exists := eventViews[viewName]; exists {
			existingView.X, existingView.Y, existingView.W, existingView.H = x, y, w, h
			existingView.Event = event
			existingView.ShowBottomBorder = showBottomBorder
			delete(eventViews, viewName)
		} else {
			ev := NewEvenView(viewName, event)
			ev.X, ev.Y, ev.W, ev.H = x, y, w, h
			ev.ShowBottomBorder = showBottomBorder
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
