package views

import (
	"fmt"

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
	// If we're in week mode but don't have a WeekView, create it now (after events are loaded)
	if cv.ViewMode == "week" && cv.WeekView == nil {
		cv.WeekView = NewWeekView(cv.Calendar, cv.TimeView)
		cv.AddChild("active", cv.WeekView)
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
	
	// Delete month view and all its children from gocui
	if cv.MonthView != nil {
		if err := cv.deleteMonthViewFromGUI(g); err != nil {
			return err
		}
	}
	
	cv.ViewMode = "week"
	
	// Don't create WeekView here - mark that we need to recreate it
	// The main update cycle will handle creating it with current event data
	cv.WeekView = nil
	
	// Remove month view from children 
	cv.children.Delete("active")
	cv.AddChild("time", cv.TimeView)
	// Don't add WeekView yet - will be created after events are loaded
	
	return nil
}

func (cv *CalendarView) SwitchToMonthView(g *gocui.Gui) error {
	if cv.ViewMode == "month" {
		return nil // Already in month view
	}
	
	// Delete week view and time view from gocui
	if cv.WeekView != nil {
		if err := cv.deleteWeekViewFromGUI(g); err != nil {
			return err
		}
	}
	if cv.TimeView != nil {
		if err := g.DeleteView(cv.TimeView.Name); err != nil && err != gocui.ErrUnknownView {
			return err
		}
	}
	
	cv.ViewMode = "month"
	
	// Remove week/time views from children and add month view
	cv.children.Delete("time")
	cv.children.Delete("active")
	cv.AddChild("active", cv.MonthView)
	
	return nil
}

func (cv *CalendarView) deleteMonthViewFromGUI(g *gocui.Gui) error {
	// Delete the main month view
	if err := g.DeleteView(cv.MonthView.Name); err != nil && err != gocui.ErrUnknownView {
		return err
	}
	
	// Delete all month day views (monthday_0 to monthday_41)
	for i := 0; i < 42; i++ {
		dayViewName := fmt.Sprintf("monthday_%d", i)
		if err := g.DeleteView(dayViewName); err != nil && err != gocui.ErrUnknownView {
			// Continue deleting other views even if one fails
		}
	}
	
	return nil
}

func (cv *CalendarView) deleteWeekViewFromGUI(g *gocui.Gui) error {
	// Delete all event views from each day first
	if cv.WeekView != nil {
		weekdayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
		for _, dayName := range weekdayNames {
			if dayView, ok := cv.WeekView.GetChild(dayName); ok {
				if dv, ok := dayView.(*DayView); ok {
					// Delete all event views in this day view
					for pair := dv.children.Oldest(); pair != nil; pair = pair.Next() {
						if eventView, ok := pair.Value.(*EventView); ok {
							if err := g.DeleteView(eventView.Name); err != nil && err != gocui.ErrUnknownView {
								// Continue deleting other views even if one fails
							}
							// Also delete any border views
							borderViewName := eventView.Name + "_border"
							if err := g.DeleteView(borderViewName); err != nil && err != gocui.ErrUnknownView {
								// Continue deleting other views even if one fails
							}
						}
					}
				}
			}
		}
	}
	
	// Delete the main week view
	if err := g.DeleteView(cv.WeekView.Name); err != nil && err != gocui.ErrUnknownView {
		return err
	}
	
	// Delete all weekday views
	weekdayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	for _, dayName := range weekdayNames {
		if err := g.DeleteView(dayName); err != nil && err != gocui.ErrUnknownView {
			// Continue deleting other views even if one fails
		}
	}
	
	return nil
}

