package views

import (
	"fmt"
	"os"

	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/jroimartin/gocui"
)

type EventView struct {
	*BaseView

	Event               *calendar.Event
	ShowTopBorder       bool
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
	
	// Set text color based on whether this is current time event
	if ev.IsCurrentTimeEvent {
		v.FgColor = calendar.ColorCustomPurple
	} else {
		v.FgColor = gocui.ColorBlack
	}
	
	v.Frame = false
	v.Clear()
	
	if ev.ShowTopBorder {
		// Debug logging
		if f, err := os.OpenFile("/tmp/chronos_underline_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintf(f, "RENDER: Event '%s' ShowTopBorder=true, H=%d, W=%d\n", ev.Event.Name, ev.H, ev.W)
			f.Close()
		}
		
		if ev.H > 1 {
			// For multi-line events, add underline on the first line
			lineChars := ""
			for i := 0; i < ev.W; i++ {
				lineChars += "━" // Using thick box-drawing character
			}
			fmt.Fprint(v, lineChars)
			fmt.Fprint(v, "\n")
			// Then write the event name on the second line
			fmt.Fprint(v, ev.Event.Name)
		} else {
			// For single-line events, add underlined characters before the name
			underlinedChars := calendar.WrapTextWithUnderline("━━━━━") // Line chars only
			fmt.Fprint(v, underlinedChars)
			fmt.Fprint(v, " ")
			fmt.Fprint(v, ev.Event.Name)
		}
	} else {
		// Normal event without top border - just write the event name
		fmt.Fprint(v, ev.Event.Name)
	}

	return nil
}
