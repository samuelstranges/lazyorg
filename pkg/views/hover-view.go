package views

import (
	"fmt"

	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/jroimartin/gocui"
)

type HoverView struct {
	*BaseView

	Calendar    *calendar.Calendar
	CurrentView View
}

func NewHoverView(c *calendar.Calendar) *HoverView {
	hv := &HoverView{
		BaseView: NewBaseView("hover"),
		Calendar: c,
	}

	return hv
}

func (hv *HoverView) Update(g *gocui.Gui) error {
	v, err := g.SetView(
		hv.Name,
		hv.X,
		hv.Y,
		hv.X+hv.W,
		hv.Y+hv.H,
	)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = true
	}

	hv.updateTitle(v)
	hv.updateBody(v)

	return nil
}

func (hv *HoverView) updateTitle(v *gocui.View) {
	if view, ok := hv.CurrentView.(*DayView); ok {
		v.Title = " " + view.Day.FormatTimeAndHour() + " "
	} else if view, ok := hv.CurrentView.(*EventView); ok {
		v.Title = " " + view.Event.Name + " "
	}
}

func (hv *HoverView) updateBody(v *gocui.View) {
	v.Clear()
	if view, ok := hv.CurrentView.(*DayView); ok {
		v.Wrap = false
		v.FgColor = gocui.AttrBold | gocui.ColorYellow
		fmt.Fprintln(v, view.Day.FormatBody())
	} else if view, ok := hv.CurrentView.(*EventView); ok {
		v.Wrap = true
		v.FgColor = gocui.AttrBold | gocui.ColorMagenta
		fmt.Fprintln(v, view.Event.FormatBody())
	}
}
