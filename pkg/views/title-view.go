package views

import (
	"fmt"
	"time"

	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/samuelstranges/chronos/internal/utils"
	"github.com/jroimartin/gocui"
)

type TitleView struct {
	*BaseView

	Calendar *calendar.Calendar
	ViewMode string
	CurrentDate time.Time
	WeatherData string
}

func NewTitleView(c *calendar.Calendar) *TitleView {
	tv := &TitleView{
		BaseView: NewBaseView("title"),
		Calendar: c,
	}

	return tv
}

func (tv *TitleView) Update(g *gocui.Gui) error {
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
		v.FgColor = gocui.AttrBold | gocui.ColorCyan
		v.Wrap = true
	}

	tv.updateBody(v)

	return nil
}

func (tv *TitleView) updateBody(v *gocui.View) {
	now := time.Now()
	tzAbbr, offset := now.Zone()
	offsetHours := offset / 3600

	// Line 1: Current date and time with timezone offset and weather
	line1 := fmt.Sprintf("Current date/time: %s %d, %d - %s (%s, UTC%+d)",
		now.Month().String(),
		now.Day(),
		now.Year(),
		utils.FormatHourFromTime(now),
		tzAbbr,
		offsetHours)
	
	// Add weather data to line 1 if available
	if tv.WeatherData != "" {
		line1 += " | " + tv.WeatherData
	}
	
	// Line 2: View context information (will be updated by AppView)
	line2 := tv.getContextualInfo()

	v.Clear()
	fmt.Fprintln(v, line1)
	fmt.Fprintln(v, line2)
}

func (tv *TitleView) getContextualInfo() string {
	switch tv.ViewMode {
	case "month":
		currentDate := tv.Calendar.CurrentDay.Date
		return fmt.Sprintf("Showing month: %s %d", currentDate.Month().String(), currentDate.Year())
	case "agenda":
		currentDate := tv.Calendar.CurrentDay.Date
		return fmt.Sprintf("Showing agenda: %s %d, %d", currentDate.Month().String(), currentDate.Day(), currentDate.Year())
	default: // "week" or empty
		selectedWeek := tv.Calendar.FormatWeekBody()
		startDate := tv.Calendar.CurrentWeek.StartDate
		return fmt.Sprintf("Showing week: %s, %d", selectedWeek, startDate.Year())
	}
}

// SetViewMode sets the current view mode for contextual display
func (tv *TitleView) SetViewMode(mode string) {
	tv.ViewMode = mode
}

// SetWeatherData sets the weather data to display in the title
func (tv *TitleView) SetWeatherData(weatherData string) {
	tv.WeatherData = weatherData
}
