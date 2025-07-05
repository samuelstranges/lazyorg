package views

import (
	"fmt"
	"os"
	
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
	if f, err := os.OpenFile("/tmp/chronos_switch_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "NewWeekView: Creating new week view\n")
		f.Close()
	}
	
	wv := &WeekView{
		BaseView: NewBaseView("week"),
		Calendar: c,
		TimeView: tv,
	}

	for i, dayName := range WeekdayNames {
		dayData := c.CurrentWeek.Days[i]
		if f, err := os.OpenFile("/tmp/chronos_switch_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintf(f, "NewWeekView: Creating DayView for %s (%s) with %d events\n", 
				dayName, dayData.Date.Format("2006-01-02"), len(dayData.Events))
			for j, event := range dayData.Events {
				fmt.Fprintf(f, "  Event %d: %s at %s\n", j, event.Name, event.Time.Format("15:04"))
			}
			f.Close()
		}
		wv.AddChild(dayName, NewDayView(dayName, dayData, tv))
	}

	if f, err := os.OpenFile("/tmp/chronos_switch_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "NewWeekView: Completed creating week view\n")
		f.Close()
	}

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
