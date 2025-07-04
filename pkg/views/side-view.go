package views

import (
	"github.com/HubertBel/lazyorg/internal/calendar"
	"github.com/HubertBel/lazyorg/internal/config"
	"github.com/HubertBel/lazyorg/internal/database"
	"github.com/jroimartin/gocui"
)

type SideView struct {
	*BaseView

	Calendar *calendar.Calendar
}

func NewSideView(c *calendar.Calendar, db *database.Database, cfg *config.Config) *SideView {
	sv := &SideView{
		BaseView: NewBaseView("side"),
		Calendar: c,
	}

	if !cfg.HideDayOnStartup {
		sv.AddChild("hover", NewHoverView(c))
	}

	return sv
}

func (sv *SideView) Update(g *gocui.Gui) error {
	v, err := g.SetView(
		sv.Name,
		sv.X,
		sv.Y,
		sv.X+sv.W,
		sv.Y+sv.H,
	)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.FgColor = gocui.AttrBold
	}

	sv.updateChildViewProperties()

	if err = sv.UpdateChildren(g); err != nil {
		return err
	}

	return nil
}

func (sv *SideView) updateChildViewProperties() {
	if hoverView, ok := sv.GetChild("hover"); ok {
		hoverView.SetProperties(
			sv.X,
			sv.Y,
			sv.W,
			sv.H,
		)
	}
}
