package views

import (
	"fmt"
	"os"

	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/samuelstranges/chronos/internal/database"
	"github.com/jroimartin/gocui"
)

type CalendarView struct {
	*BaseView
	
	Calendar *calendar.Calendar
	Database *database.Database
	ViewMode string // "week" or "month"
	
	TimeView  *TimeView
	WeekView  *WeekView
	MonthView *MonthView
}

func NewCalendarView(c *calendar.Calendar, db *database.Database) *CalendarView {
	cv := &CalendarView{
		BaseView: NewBaseView("calendar"),
		Calendar: c,
		Database: db,
		ViewMode: "week", // Default to week view
	}
	
	// Create the time view (used by week view)
	cv.TimeView = NewTimeView()
	
	// Create both views but only add the active one as a child
	cv.WeekView = NewWeekView(c, cv.TimeView)
	cv.MonthView = NewMonthView(c, db)
	
	// Start with week view
	cv.AddChild("time", cv.TimeView)
	cv.AddChild("active", cv.WeekView)
	
	return cv
}

func (cv *CalendarView) Update(g *gocui.Gui) error {
	// Debug logging
	if f, err := os.OpenFile("/tmp/chronos_month_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "CalendarView.Update: X=%d, Y=%d, W=%d, H=%d, ViewMode=%s\n", cv.X, cv.Y, cv.W, cv.H, cv.ViewMode)
		f.Close()
	}
	
	// No need to create a view for the container itself
	cv.updateChildViewProperties()
	
	if err := cv.UpdateChildren(g); err != nil {
		return err
	}
	
	return nil
}

func (cv *CalendarView) updateChildViewProperties() {
	if cv.ViewMode == "week" {
		// Position time view and week view
		if cv.TimeView != nil {
			cv.TimeView.SetProperties(
				cv.X+1,
				cv.Y+1,
				TimeViewWidth,
				cv.H-2,
			)
		}
		
		if cv.WeekView != nil {
			cv.WeekView.SetProperties(
				cv.X+TimeViewWidth+1,
				cv.Y,
				cv.W-TimeViewWidth-1,
				cv.H,
			)
		}
	} else if cv.ViewMode == "month" {
		// Month view takes the full area
		if cv.MonthView != nil {
			cv.MonthView.SetProperties(
				cv.X,
				cv.Y,
				cv.W,
				cv.H,
			)
		}
	}
}

func (cv *CalendarView) SwitchToWeekView(g *gocui.Gui) error {
	if cv.ViewMode == "week" {
		return nil // Already in week view
	}
	
	// Debug logging
	if f, err := os.OpenFile("/tmp/chronos_month_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "Switching from %s to week view\n", cv.ViewMode)
		f.Close()
	}
	
	cv.ViewMode = "week"
	
	// Remove month view from children and add week/time views
	cv.children.Delete("active")
	cv.AddChild("time", cv.TimeView)
	cv.AddChild("active", cv.WeekView)
	
	return nil
}

func (cv *CalendarView) SwitchToMonthView(g *gocui.Gui) error {
	if cv.ViewMode == "month" {
		return nil // Already in month view
	}
	
	// Debug logging
	if f, err := os.OpenFile("/tmp/chronos_month_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "Switching from %s to month view\n", cv.ViewMode)
		f.Close()
	}
	
	cv.ViewMode = "month"
	
	// Remove week/time views from children and add month view
	cv.children.Delete("time")
	cv.children.Delete("active")
	cv.AddChild("active", cv.MonthView)
	
	return nil
}