package views

import (
	"fmt"

	"github.com/HubertBel/lazyorg/internal/utils"
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
	// Calculate starting time, but constrain to valid 24-hour range
	initialTime := 12 - tv.H/4
	if initialTime < 0 {
		initialTime = 0
	}
	
	tv.Body = ""

	for i := range tv.H {
		var time string
		currentHour := initialTime

		if i%2 == 0 {
			// Skip if hour is beyond 23:00
			if currentHour > 23 {
				break
			}
			hour := utils.FormatHour(currentHour, 0)
			time = fmt.Sprintf(" %s - \n", hour)
		} else {
			// Skip if hour is beyond 23:30
			if currentHour > 23 {
				break
			}
			hour := utils.FormatHour(currentHour, 30)
			time = fmt.Sprintf(" %s \n", hour)
			initialTime++
		}

		if i == tv.Cursor {
			runes := []rune(time)
			runes[0] = '>'
			time = string(runes)
		}

		tv.Body += time
	}

	v.Clear()
	fmt.Fprintln(v, tv.Body)
}

func (tv *TimeView) SetCursor(y int) {
	tv.Cursor = y
}
