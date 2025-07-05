package views

import (
	"fmt"

	"github.com/samuelstranges/chronos/internal/utils"
	"github.com/jroimartin/gocui"
)

type TimeView struct {
	*BaseView
	Body   string
	Cursor int
}

func NewTimeView() *TimeView {
	tv := &TimeView{
		BaseView: NewBaseView("time"),
		Cursor:   0,
	}

	return tv
}

func (tv *TimeView) Update(g *gocui.Gui) error {
	v, err := g.SetView(
		tv.Name,
		tv.X,
		tv.Y,
		tv.X+tv.W,
		tv.Y+tv.H,
	)

	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.FgColor = gocui.ColorGreen
	}

	tv.updateBody(v)

	return nil
}

func (tv *TimeView) updateBody(v *gocui.View) {
	// Always start from 00:00 to ensure all events (including 00:00 events) are displayable
	// This prevents events at 00:00 from being hidden due to TimeToPosition returning -1
	initialTime := 0
	
	tv.Body = ""

	for i := range tv.H {
		var timeStr string
		loopHour := initialTime

		if i%2 == 0 {
			// Skip if hour is beyond 23:00
			if loopHour > 23 {
				break
			}
			hour := utils.FormatHour(loopHour, 0)
			timeStr = fmt.Sprintf(" %s - \n", hour)
		} else {
			// Skip if hour is beyond 23:30
			if loopHour > 23 {
				break
			}
			hour := utils.FormatHour(loopHour, 30)
			timeStr = fmt.Sprintf(" %s \n", hour)
			initialTime++
		}


		if i == tv.Cursor {
			runes := []rune(timeStr)
			runes[0] = '>'
			timeStr = string(runes)
		}

		tv.Body += timeStr
	}

	v.Clear()
	fmt.Fprintln(v, tv.Body)
}

func (tv *TimeView) SetCursor(y int) {
	tv.Cursor = y
}

