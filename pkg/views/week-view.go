package views

import (
	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/jroimartin/gocui"
)

var WeekdayNames = []string{
	"Sunday",
	"Monday",
	"Tuesday",
	"Wednesday",
	"Thursday",
	"Friday",
	"Saturday",
}

type WeekView struct {
	*BaseView

	Calendar *calendar.Calendar

	TimeView *TimeView
}

func NewWeekView(c *calendar.Calendar, tv *TimeView) *WeekView {
	wv := &WeekView{
		BaseView: NewBaseView("week"),
		Calendar: c,
		TimeView: tv,
	}

	wv.AddChild(WeekdayNames[0], NewDayView(WeekdayNames[0], c.CurrentWeek.Days[0], tv))
	wv.AddChild(WeekdayNames[1], NewDayView(WeekdayNames[1], c.CurrentWeek.Days[1], tv))
	wv.AddChild(WeekdayNames[2], NewDayView(WeekdayNames[2], c.CurrentWeek.Days[2], tv))
	wv.AddChild(WeekdayNames[3], NewDayView(WeekdayNames[3], c.CurrentWeek.Days[3], tv))
	wv.AddChild(WeekdayNames[4], NewDayView(WeekdayNames[4], c.CurrentWeek.Days[4], tv))
	wv.AddChild(WeekdayNames[5], NewDayView(WeekdayNames[5], c.CurrentWeek.Days[5], tv))
	wv.AddChild(WeekdayNames[6], NewDayView(WeekdayNames[6], c.CurrentWeek.Days[6], tv))

	return wv
}

func (wv *WeekView) Update(g *gocui.Gui) error {
	v, err := g.SetView(
		wv.Name,
		wv.X,
		wv.Y,
		wv.X+wv.W,
		wv.Y+wv.H,
	)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
	}

	wv.updateChildViewProperties()

	if err = wv.UpdateChildren(g); err != nil {
		return err
	}

	return nil
}

func (wv *WeekView) updateChildViewProperties() {
	x := wv.X
	w := wv.W/7 - Padding

	for _, weekday := range WeekdayNames {
		if dayView, ok := wv.GetChild(weekday); ok {

			dayView.SetProperties(
				x,
				wv.Y+1,
				w,
				wv.H-2,
			)
		}

		x += w + Padding
	}
}
