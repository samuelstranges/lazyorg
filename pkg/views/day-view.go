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
