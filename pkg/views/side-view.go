package views

import (
	"github.com/HubertBel/lazyorg/internal/calendar"
	"github.com/HubertBel/lazyorg/internal/database"
	"github.com/jroimartin/gocui"
)

type SideView struct {
	*BaseView

	Calendar *calendar.Calendar
}

func NewSideView(c *calendar.Calendar, db *database.Database) *SideView {
	sv := &SideView{
		BaseView: NewBaseView("side"),
		Calendar: c,
	}

	sv.AddChild("hover", NewHoverView(c))
	sv.AddChild("notepad", NewNotepadView(c, db))

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
	heightHover := int(float64(sv.H) * 0.5)
	heightNotepad := sv.H - heightHover - 2

	if hoverView, ok := sv.GetChild("hover"); ok {
		hoverView.SetProperties(
			sv.X,
			sv.Y,
			sv.W,
			heightHover,
		)
	}

	if notepadView, ok := sv.GetChild("notepad"); ok {
		notepadView.SetProperties(
			sv.X,
			sv.Y+heightHover+1,
			sv.W,
			heightNotepad,
		)
	}
}
