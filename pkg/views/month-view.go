package views

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/samuelstranges/chronos/internal/config"
	"github.com/samuelstranges/chronos/internal/eventmanager"
	"github.com/samuelstranges/chronos/internal/weather"
	"github.com/jroimartin/gocui"
)

var MonthDayNames = []string{
	"Sunday",
	"Monday",
	"Tuesday", 
	"Wednesday",
	"Thursday",
	"Friday",
	"Saturday",
}

const dayHeaderHeight = 2

type MonthView struct {
	*BaseView
	
	Calendar     *calendar.Calendar
	EventManager *eventmanager.EventManager
	
	// Current month being displayed
	CurrentMonth time.Time
	
	// Grid dimensions
	CellWidth  int
	CellHeight int
	GridRows   int
	GridCols   int
	
	// Weather support
	Config       *config.Config
	WeatherCache *weather.WeatherCache
}

func NewMonthView(c *calendar.Calendar, em *eventmanager.EventManager) *MonthView {
	mv := &MonthView{
		BaseView:     NewBaseView("month"),
		Calendar:     c,
		EventManager: em,
		CurrentMonth: c.CurrentDay.Date,
		GridCols:     7, // 7 days of the week
		GridRows:     6, // Maximum 6 rows for a month
		Config:       nil, // Will be set by UpdateWeatherData
		WeatherCache: nil, // Will be set by UpdateWeatherData
	}
	
	// Initialize month day views
	mv.createMonthDayViews()
	
	return mv
}

func (mv *MonthView) createMonthDayViews() {
	// Get the first day of the month
	firstDay := time.Date(mv.CurrentMonth.Year(), mv.CurrentMonth.Month(), 1, 0, 0, 0, 0, mv.CurrentMonth.Location())
	
	// Find the Sunday of the week containing the first day
	startOfWeek := firstDay
	for startOfWeek.Weekday() != time.Sunday {
		startOfWeek = startOfWeek.AddDate(0, 0, -1)
	}
	
	// Create 42 day views (6 rows × 7 columns)
	for i := 0; i < 42; i++ {
		dayDate := startOfWeek.AddDate(0, 0, i)
		dayName := fmt.Sprintf("monthday_%d", i)
		
		// Create day view
		dayView := NewMonthDayView(dayName, dayDate, mv.CurrentMonth)
		mv.AddChild(dayName, dayView)
	}
}

func (mv *MonthView) Update(g *gocui.Gui) error {
	// Debug logging
	if f, err := os.OpenFile("/tmp/chronos_month_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "MonthView.Update called: X=%d, Y=%d, W=%d, H=%d\n", mv.X, mv.Y, mv.W, mv.H)
		f.Close()
	}
	
	// Skip if dimensions are invalid OR if view is hidden (positioned at -1000)
	if mv.W <= 0 || mv.H <= 0 || mv.X <= -500 || mv.Y <= -500 {
		if f, err := os.OpenFile("/tmp/chronos_month_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintf(f, "SKIPPING MonthView update - invalid dimensions or hidden: X=%d, Y=%d, W=%d, H=%d\n", mv.X, mv.Y, mv.W, mv.H)
			f.Close()
		}
		return nil
	}
	
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
		v.Frame = false
	}
	
	// Load events for current month
	if err := mv.loadEventsForMonth(); err != nil {
		return err
	}
	
	// Calculate cell dimensions based on available space
	dayHeaderHeight := 2 // Space for day names and separator line
	mv.CellWidth = mv.W / mv.GridCols
	if mv.CellWidth < 1 {
		mv.CellWidth = 1
	}
	mv.CellHeight = (mv.H - dayHeaderHeight) / mv.GridRows // Leave space for day headers
	if mv.CellHeight < 1 {
		mv.CellHeight = 1
	}
	
	// Debug cell calculations
	if f, err := os.OpenFile("/tmp/chronos_month_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "Cell calculations: W=%d, H=%d, CellW=%d, CellH=%d, GridCols=%d, GridRows=%d\n", 
			mv.W, mv.H, mv.CellWidth, mv.CellHeight, mv.GridCols, mv.GridRows)
		f.Close()
	}
	
	// Clear view for drawing
	v.Clear()
	
	// Draw compact grid with shared borders
	mv.drawCompactGrid(v)
	
	mv.updateChildViewProperties()
	
	if err = mv.UpdateChildren(g); err != nil {
		return err
	}
	
	return nil
}

func (mv *MonthView) updateChildViewProperties() {
	// Debug child view updates
	if f, err := os.OpenFile("/tmp/chronos_month_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "updateChildViewProperties: MonthView X=%d, Y=%d, W=%d, H=%d\n", mv.X, mv.Y, mv.W, mv.H)
		f.Close()
	}
	
	for i := 0; i < 42; i++ {
		dayName := fmt.Sprintf("monthday_%d", i)
		if dayView, ok := mv.GetChild(dayName); ok {
			row := i / mv.GridCols
			col := i % mv.GridCols
			
			// Calculate position for compact grid layout (no individual borders)
			x := mv.X + col*mv.CellWidth
			y := mv.Y + dayHeaderHeight + row*mv.CellHeight
			w := mv.CellWidth
			h := mv.CellHeight
			
			// Ensure we don't exceed bounds and have positive dimensions
			if x+w >= mv.X+mv.W {
				w = mv.X + mv.W - x - 1
			}
			if y+h >= mv.Y+mv.H {
				h = mv.Y + mv.H - y - 1
			}
			
			// Ensure minimum positive dimensions
			if w < 1 {
				w = 1
			}
			if h < 1 {
				h = 1
			}
			
			// Debug problematic dimensions
			if w <= 0 || h <= 0 || x < -500 || y < -500 {
				if f, err := os.OpenFile("/tmp/chronos_month_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
					fmt.Fprintf(f, "INVALID day view %d: x=%d, y=%d, w=%d, h=%d (row=%d, col=%d)\n", i, x, y, w, h, row, col)
					f.Close()
				}
			}
			
			dayView.SetProperties(x, y, w, h)
		}
	}
}

// Navigation methods
func (mv *MonthView) UpdateToNextMonth() {
	mv.CurrentMonth = mv.CurrentMonth.AddDate(0, 1, 0)
	mv.refreshMonthDayViews()
}

func (mv *MonthView) UpdateToPrevMonth() {
	mv.CurrentMonth = mv.CurrentMonth.AddDate(0, -1, 0)
	mv.refreshMonthDayViews()
}

func (mv *MonthView) refreshMonthDayViews() {
	// Debug logging
	if f, err := os.OpenFile("/tmp/chronos_month_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "refreshMonthDayViews called for %s %d\n", mv.CurrentMonth.Month().String(), mv.CurrentMonth.Year())
		f.Close()
	}
	
	// Clear existing children
	for i := 0; i < 42; i++ {
		dayName := fmt.Sprintf("monthday_%d", i)
		mv.children.Delete(dayName)
	}
	
	// Recreate month day views
	mv.createMonthDayViews()
	
	// Debug logging
	if f, err := os.OpenFile("/tmp/chronos_month_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "recreated %d day views for %s %d\n", mv.children.Len(), mv.CurrentMonth.Month().String(), mv.CurrentMonth.Year())
		f.Close()
	}
}

func (mv *MonthView) loadEventsForMonth() error {
	// Load events for the current month
	events, err := mv.EventManager.GetEventsByMonth(mv.CurrentMonth.Year(), mv.CurrentMonth.Month())
	if err != nil {
		return err
	}
	
	// Convert UTC events to local time for display
	localEvents := make([]*calendar.Event, len(events))
	for i, event := range events {
		localEvent := *event
		localEvent.Time = event.Time.In(time.Local)
		localEvents[i] = &localEvent
	}
	
	// Distribute events to appropriate day views
	for i := 0; i < 42; i++ {
		dayName := fmt.Sprintf("monthday_%d", i)
		if dayView, ok := mv.GetChild(dayName); ok {
			if monthDayView, ok := dayView.(*MonthDayView); ok {
				monthDayView.LoadEvents(localEvents)
			}
		}
	}
	
	return nil
}

func (mv *MonthView) drawCompactGrid(v *gocui.View) {
	// Draw day headers with pipe separators
	dayHeaders := ""
	for i, dayName := range MonthDayNames {
		// Use 3-character abbreviation for day names
		dayAbbr := dayName[:3]
		paddedName := fmt.Sprintf("%-*s", mv.CellWidth-1, dayAbbr)
		dayHeaders += paddedName
		if i < len(MonthDayNames)-1 {
			dayHeaders += "│"
		}
	}
	fmt.Fprintf(v, "%s\n", dayHeaders)
	
	// Draw grid lines for each row
	for row := 0; row < mv.GridRows; row++ {
		// Draw horizontal separator line
		if row == 0 {
			// Top border
			separator := ""
			for col := 0; col < mv.GridCols; col++ {
				separator += strings.Repeat("─", mv.CellWidth-1)
				if col < mv.GridCols-1 {
					separator += "┬"
				}
			}
			fmt.Fprintf(v, "%s\n", separator)
		}
		
		// Draw the row content (will be filled by MonthDayView children)
		for line := 0; line < mv.CellHeight-1; line++ {
			rowLine := ""
			for col := 0; col < mv.GridCols; col++ {
				// Leave space for content (filled by child views)
				rowLine += strings.Repeat(" ", mv.CellWidth-1)
				if col < mv.GridCols-1 {
					rowLine += "│"
				}
			}
			fmt.Fprintf(v, "%s\n", rowLine)
		}
		
		// Draw horizontal separator (except for last row)
		if row < mv.GridRows-1 {
			separator := ""
			for col := 0; col < mv.GridCols; col++ {
				separator += strings.Repeat("─", mv.CellWidth-1)
				if col < mv.GridCols-1 {
					separator += "┼"
				}
			}
			fmt.Fprintf(v, "%s\n", separator)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// UpdateWeatherData updates weather icons for all visible days in the month view
func (mv *MonthView) UpdateWeatherData(cfg *config.Config, weatherCache *weather.WeatherCache) error {
	mv.Config = cfg
	mv.WeatherCache = weatherCache
	
	if !config.IsWeatherEnabled(cfg) {
		return nil
	}
	
	location := config.GetWeatherLocation(cfg)
	if location == "" {
		return nil
	}
	
	// Get cached weather forecast only - no API calls during UI updates
	forecast, exists := weatherCache.GetCachedWeatherForecast(location)
	if !exists {
		// No cached forecast data available yet - skip weather display for now
		return nil
	}
	
	// Create a map of date -> weather data for quick lookup
	weatherMap := make(map[string]struct{icon, maxTemp string})
	unit := config.GetWeatherUnit(cfg)
	
	for _, day := range forecast.Days {
		dateStr := day.Date.Format("2006-01-02")
		
		// Choose temperature based on unit preference
		maxTemp := day.MaxTempC
		if unit == "fahrenheit" {
			maxTemp = day.MaxTempF
		}
		
		weatherMap[dateStr] = struct{icon, maxTemp string}{
			icon: day.Icon, 
			maxTemp: maxTemp,
		}
	}
	
	// Update weather data for all day views
	for i := 0; i < 42; i++ {
		dayName := fmt.Sprintf("monthday_%d", i)
		if dayViewInterface, ok := mv.GetChild(dayName); ok {
			if dayView, ok := dayViewInterface.(*MonthDayView); ok {
				dateStr := dayView.Date.Format("2006-01-02")
				if weatherData, exists := weatherMap[dateStr]; exists {
					dayView.SetWeatherData(weatherData.icon, weatherData.maxTemp)
				} else {
					dayView.SetWeatherData("", "") // Clear weather data if none available
				}
			}
		}
	}
	
	return nil
}