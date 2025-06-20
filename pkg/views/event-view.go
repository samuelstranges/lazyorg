package views

import (
	"fmt"

	"github.com/HubertBel/lazyorg/internal/calendar"
	"github.com/jroimartin/gocui"
)

type EventView struct {
	*BaseView

	Event            *calendar.Event
	ShowBottomBorder bool
}

func NewEvenView(name string, e *calendar.Event) *EventView {
	return &EventView{
		BaseView: NewBaseView(name),

		Event: e,
	}
}

func (ev *EventView) Update(g *gocui.Gui) error {
	v, err := g.SetView(
		ev.Name,
		ev.X,
		ev.Y,
		ev.X+ev.W,
		ev.Y+ev.H,
	)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}

	v.BgColor = ev.Event.Color
	v.FgColor = gocui.ColorBlack
	v.Frame = false
	v.Clear()
	
	if ev.ShowBottomBorder {
		// Write event name normally
		fmt.Fprint(v, ev.Event.Name)
		
		// If the event is tall enough, add underscores on the second-to-last row
		if ev.H > 2 {
			// Move to the second-to-last row (account for the +1 in height calculation)
			for i := 1; i < ev.H-1; i++ {
				fmt.Fprint(v, "\n")
			}
			// Add underscores across the width
			for i := 0; i < ev.W; i++ {
				fmt.Fprint(v, "_")
			}
		} else {
			// For short events, add underscores right after name
			fmt.Fprint(v, " _____")
		}
	} else {
		fmt.Fprint(v, ev.Event.Name)
	}

	return nil
}
