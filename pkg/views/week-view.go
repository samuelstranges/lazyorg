package views

import (
	"fmt"
	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/samuelstranges/chronos/internal/config"
	"github.com/samuelstranges/chronos/internal/weather"
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

	for i, dayName := range WeekdayNames {
		wv.AddChild(dayName, NewDayView(dayName, c.CurrentWeek.Days[i], tv))
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

// UpdateWeatherData updates weather information for all day columns in week view
func (wv *WeekView) UpdateWeatherData(cfg *config.Config, weatherCache *weather.WeatherCache) error {
	if !config.IsWeatherEnabled(cfg) {
		return nil
	}
	
	location := config.GetWeatherLocation(cfg)
	if location == "" {
		return nil
	}
	
	// Get 3-day weather forecast
	forecast, err := weatherCache.GetWeatherForecast(location)
	if err != nil {
		return fmt.Errorf("failed to get weather forecast: %w", err)
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
	
	// Update weather data for all day views in the week
	for _, weekday := range WeekdayNames {
		if dayViewInterface, ok := wv.GetChild(weekday); ok {
			if dayView, ok := dayViewInterface.(*DayView); ok {
				dateStr := dayView.Day.Date.Format("2006-01-02")
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
