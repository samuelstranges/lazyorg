package views

import (
	"fmt"
	"os"
	
	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/samuelstranges/chronos/internal/database"
	"github.com/samuelstranges/chronos/internal/eventmanager"
	"github.com/samuelstranges/chronos/internal/utils"
	"github.com/jroimartin/gocui"
)

type MainView struct {
	*BaseView

	Calendar     *calendar.Calendar
	Database     *database.Database
	CalendarView *CalendarView
}

func NewMainView(c *calendar.Calendar, db *database.Database, em *eventmanager.EventManager) *MainView {
	mv := &MainView{
		BaseView:     NewBaseView("main"),
		Calendar:     c,
		Database:     db,
		CalendarView: NewCalendarView(c, db, em),
	}

	mv.AddChild("calendar", mv.CalendarView)

	return mv
}

func (mv *MainView) Update(g *gocui.Gui) error {
	v, err := g.SetView(
		mv.Name,
		mv.X,
		mv.Y,
		mv.X+mv.W,
		mv.Y+mv.H,
	)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.FgColor = gocui.AttrBold
	}

	mv.updateChildViewProperties()

	if err = mv.UpdateChildren(g); err != nil {
		return err
	}

	return nil
}

func (mv *MainView) updateChildViewProperties() {
	// Update calendar view to take full main view area
	if mv.CalendarView != nil {
		mv.CalendarView.SetProperties(
			mv.X,
			mv.Y,
			mv.W,
			mv.H,
		)
	}

	// Auto-adjust viewport for responsive behavior in week view
	if mv.CalendarView != nil && mv.CalendarView.ViewMode == "week" && mv.CalendarView.TimeView != nil {
		// Debug logging
		if f, err := os.OpenFile("/tmp/chronos_mainview_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintf(f, "=== MainView.updateChildViewProperties (Week Mode) ===\n")
			fmt.Fprintf(f, "MainView dimensions: X=%d, Y=%d, W=%d, H=%d\n", mv.X, mv.Y, mv.W, mv.H)
			fmt.Fprintf(f, "Current day time: %s\n", mv.Calendar.CurrentDay.Date.Format("2006-01-02 15:04"))
			f.Close()
		}
		
		// First, auto-adjust the viewport based on calendar cursor time and available space
		mv.CalendarView.TimeView.AutoAdjustViewport(mv.Calendar.CurrentDay.Date)
		
		// Then calculate and set cursor position within the adjusted viewport
		y := utils.TimeToPositionWithViewport(mv.Calendar.CurrentDay.Date, mv.CalendarView.TimeView.GetViewportStart())
		
		// Debug logging
		if f, err := os.OpenFile("/tmp/chronos_mainview_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintf(f, "Calculated cursor position: y=%d (ViewportStart=%d, visibleSlots=%d)\n", y, mv.CalendarView.TimeView.GetViewportStart(), mv.CalendarView.TimeView.GetVisibleSlots())
			if y >= mv.CalendarView.TimeView.GetVisibleSlots() {
				fmt.Fprintf(f, "WARNING: Cursor position %d is beyond visible slots %d!\n", y, mv.CalendarView.TimeView.GetVisibleSlots())
			}
			if y < 0 {
				fmt.Fprintf(f, "WARNING: Cursor position %d is negative!\n", y)
			}
			f.Close()
		}
		
		// Since we centered the viewport around the calendar time, cursor should always be visible
		mv.CalendarView.TimeView.SetCursor(y)
		
		// Debug logging end
		if f, err := os.OpenFile("/tmp/chronos_mainview_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintf(f, "=== MainView.updateChildViewProperties END ===\n\n")
			f.Close()
		}
	}

	if titleView, ok := mv.GetChild("title"); ok {
		titleView.SetProperties(
			mv.X,
			mv.Y,
			mv.W,
			TitleViewHeight,
		)
	}
}

func (mv *MainView) SwitchToWeekView(g *gocui.Gui) error {
	if mv.CalendarView != nil {
		return mv.CalendarView.SwitchToWeekView(g)
	}
	return nil
}

func (mv *MainView) SwitchToMonthView(g *gocui.Gui) error {
	if mv.CalendarView != nil {
		return mv.CalendarView.SwitchToMonthView(g)
	}
	return nil
}

func (mv *MainView) SwitchToAgendaView(g *gocui.Gui) error {
	if mv.CalendarView != nil {
		return mv.CalendarView.SwitchToAgendaView(g)
	}
	return nil
}
