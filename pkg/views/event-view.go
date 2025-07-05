package views

import (
	"fmt"

	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/jroimartin/gocui"
)

type EventView struct {
	*BaseView

	Event               *calendar.Event
	ShowBottomBorder    bool
	IsCurrentTimeEvent  bool
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

	// Always set the event's background color (ensure it's not default/black)
	eventColor := ev.Event.Color
	if eventColor == gocui.ColorDefault || eventColor == gocui.ColorBlack {
		// Fallback to a visible color if somehow the event has no color
		eventColor = gocui.ColorBlue
	}
	v.BgColor = eventColor
	
	
	v.Frame = false
	v.Clear()
	
	if ev.ShowBottomBorder {
		// ANSI escape codes for underlining and resetting formatting.
		const ansiUnderline = "\x1b[4m" // Start underline
		const ansiBlackFg = "\x1b[30m"  // Set foreground color to black
		const ansiReset = "\x1b[0m"    // Reset all attributes
		
		// If the event is tall enough, add underscores on the second-to-last row
		if ev.H > 2 {
			fmt.Fprint(v, ev.Event.Name)
			// Move to the second-to-last row (account for the +1 in height calculation)
			for i := 1; i < ev.H-1; i++ { fmt.Fprint(v, "\n") }

			fmt.Fprint(v, ansiBlackFg)
			fmt.Fprint(v, ansiUnderline)
			for i := 0; i < ev.W; i++ { fmt.Fprint(v, "	") } // a non 'space character'
			fmt.Fprint(v, ansiReset)
		} else {
			// event is 30 mins long... must have text on same line as underline
			fmt.Fprint(v, ansiBlackFg)
			fmt.Fprint(v, ansiUnderline)
			fmt.Fprint(v, ev.Event.Name)
			for i := 0; i < ev.W; i++ { fmt.Fprint(v, "	") }
			fmt.Fprint(v, ansiReset)
		}
	} else {
		fmt.Fprint(v, ev.Event.Name)
	}

	// Set text color based on whether this is current time event
	if ev.IsCurrentTimeEvent {
		v.FgColor = calendar.ColorCustomPurple
	} else {
		v.FgColor = gocui.ColorBlack
	}

	return nil
}
