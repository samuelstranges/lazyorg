package views

import (
	"fmt"
	"time"
	
	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/samuelstranges/chronos/internal/config"
	"github.com/samuelstranges/chronos/internal/database"
	"github.com/samuelstranges/chronos/internal/eventmanager"
	"github.com/samuelstranges/chronos/internal/utils"
	"github.com/samuelstranges/chronos/internal/weather"
	"github.com/jroimartin/gocui"
	"github.com/nsf/termbox-go"
)

var (
	MainViewWidthRatio = 1.0
)

type AppView struct {
	*BaseView

	Database     *database.Database
	EventManager *eventmanager.EventManager
	Calendar     *calendar.Calendar
	Config       *config.Config
	DebugMode    bool
	
	colorPickerEvent  *EventView
	colorPickerActive bool
	copiedEvent       *calendar.Event
	
	// Search functionality
	searchQuery       string
	searchMatches     []*calendar.Event
	currentMatchIndex int
	isSearchActive    bool
	
	// View initialization
	initialViewMode   string
	viewInitialized   bool
	
	// Weather functionality
	weatherCache      *weather.WeatherCache
}


func NewAppView(g *gocui.Gui, db *database.Database, cfg *config.Config) *AppView {
	now := time.Now()
	t := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())

	c := calendar.NewCalendar(calendar.NewDay(t))
	em := eventmanager.NewEventManager(db)

	av := &AppView{
		BaseView:     NewBaseView("app"),
		Database:     db,
		EventManager: em,
		Calendar:     c,
		Config:       cfg,
		DebugMode:    db.DebugMode,
		weatherCache: weather.NewWeatherCache(),
	}
	

	titleView := NewTitleView(c)
	defaultView := config.GetDefaultView(cfg)
	titleView.SetViewMode(defaultView)
	av.AddChild("title", titleView)
	av.AddChild("popup", NewEvenPopup(g, c, db, av.EventManager, cfg))
	av.AddChild("main", NewMainView(c, db, av.EventManager))
	
	// Set up error handler for EventManager after popup is created
	av.setupErrorHandler(g)
	
	// Store the default view mode for later initialization
	// We can't switch views during NewAppView because GUI isn't fully initialized yet
	av.initialViewMode = defaultView
	
	av.AddChild("keybinds", NewKeybindsView())
	
	// Preload weather data if enabled to avoid lag when switching views
	av.preloadWeatherData()

	return av
}

// setupErrorHandler configures the EventManager to show error messages via popup
func (av *AppView) setupErrorHandler(g *gocui.Gui) {
	av.EventManager.SetErrorHandler(func(title, message string) {
		if popup, ok := av.GetChild("popup"); ok {
			if popupView, ok := popup.(*EventPopupView); ok {
				popupView.ShowErrorMessage(g, title, message)
			}
		}
	})
}

func (av *AppView) Layout(g *gocui.Gui) error {
	return av.Update(g)
}

func (av *AppView) Update(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	av.SetProperties(0, 1, maxX-1, maxY-1)

	v, err := g.SetView(
		av.Name,
		av.X,
		av.Y,
		av.X+av.W,
		av.Y+av.H,
	)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
	}

	if err = av.updateEventsFromDatabase(); err != nil {
		return err
	}
	
	// Update weather data if enabled
	if err = av.updateWeatherData(); err != nil {
		// Don't fail the entire update if weather fails - just log it
		// Weather is optional functionality
	}
	
	// Update month view weather if in month view
	if err = av.updateMonthViewWeather(); err != nil {
		// Don't fail the entire update if weather fails - just log it
		// Weather is optional functionality
	}
	
	// Update week view weather if in week view
	if err = av.updateWeekViewWeather(); err != nil {
		// Don't fail the entire update if weather fails - just log it
		// Weather is optional functionality
	}

	// Initialize the view mode on first update
	if !av.viewInitialized {
		av.initializeViewMode(g)
		av.viewInitialized = true
	}

	av.updateChildViewProperties()

	if err = av.UpdateChildren(g); err != nil {
		return err
	}

	if err = av.UpdateCurrentView(g); err != nil {
		return err
	}
	
	if err = av.showSearchStatus(g); err != nil {
		return err
	}

	return nil
}

// initializeViewMode sets up the initial view mode after GUI is initialized
func (av *AppView) initializeViewMode(g *gocui.Gui) {
	switch av.initialViewMode {
	case "month":
		av.SwitchToMonthView(g)
	case "agenda":
		av.SwitchToAgendaView(g)
	default:
		// Week view is already the default, no need to switch
	}
}

func (av *AppView) updateEventsFromDatabase() error {
	for _, v := range av.Calendar.CurrentWeek.Days {
		// Don't use clear() - it affects existing day views pointing to this slice
		// Instead, create a new slice entirely
		var err error
		events, err := av.EventManager.GetEventsByDate(v.Date)
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

		v.Events = localEvents
		v.SortEventsByTime()
	}

	return nil
}


func (av *AppView) JumpToToday() {
	av.Calendar.JumpToToday()
}

func (av *AppView) UpdateToNextWeek() {
	av.Calendar.UpdateToNextWeek()
}

func (av *AppView) UpdateToPrevWeek() {
	av.Calendar.UpdateToPrevWeek()
}

func (av *AppView) UpdateToNextMonth() {
	av.Calendar.UpdateToNextMonth()
	
	// Also update the month view's current month if we're in month mode
	if mainView, ok := av.GetChild("main"); ok {
		if mv, ok := mainView.(*MainView); ok {
			if mv.CalendarView != nil && mv.CalendarView.ViewMode == "month" && mv.CalendarView.MonthView != nil {
				mv.CalendarView.MonthView.UpdateToNextMonth()
			}
		}
	}
}

func (av *AppView) UpdateToPrevMonth() {
	av.Calendar.UpdateToPrevMonth()
	
	// Also update the month view's current month if we're in month mode
	if mainView, ok := av.GetChild("main"); ok {
		if mv, ok := mainView.(*MainView); ok {
			if mv.CalendarView != nil && mv.CalendarView.ViewMode == "month" && mv.CalendarView.MonthView != nil {
				mv.CalendarView.MonthView.UpdateToPrevMonth()
			}
		}
	}
}

func (av *AppView) SwitchToWeekView(g *gocui.Gui) error {
	if mainView, ok := av.GetChild("main"); ok {
		if mv, ok := mainView.(*MainView); ok {
			// Switch the view mode - this will cause the week view to be recreated properly
			err := mv.SwitchToWeekView(g)
			if err != nil {
				return err
			}
			
			// Update title view mode
			if titleView, ok := av.GetChild("title"); ok {
				if tv, ok := titleView.(*TitleView); ok {
					tv.SetViewMode("week")
				}
			}
			
			// Update weather data asynchronously after switching to week view
			go func() {
				av.updateWeekViewWeather()
				// Force a UI refresh after weather data is set
				g.Update(func(g *gocui.Gui) error {
					return nil // Just trigger a refresh
				})
			}()
			
			// Let the main Update() cycle handle event loading and day view updates
			return av.UpdateCurrentView(g)
		}
	}
	return nil
}

func (av *AppView) SwitchToMonthView(g *gocui.Gui) error {
	if mainView, ok := av.GetChild("main"); ok {
		if mv, ok := mainView.(*MainView); ok {
			err := mv.SwitchToMonthView(g)
			if err != nil {
				return err
			}
			
			// Update title view mode
			if titleView, ok := av.GetChild("title"); ok {
				if tv, ok := titleView.(*TitleView); ok {
					tv.SetViewMode("month")
				}
			}
			
			return nil
		}
	}
	return nil
}

func (av *AppView) SwitchToAgendaView(g *gocui.Gui) error {
	if mainView, ok := av.GetChild("main"); ok {
		if mv, ok := mainView.(*MainView); ok {
			err := mv.SwitchToAgendaView(g)
			if err != nil {
				return err
			}
			
			// Update title view mode
			if titleView, ok := av.GetChild("title"); ok {
				if tv, ok := titleView.(*TitleView); ok {
					tv.SetViewMode("agenda")
				}
			}
			
			return nil
		}
	}
	return nil
}

func (av *AppView) ToggleView(g *gocui.Gui) error {
	currentMode := av.GetViewMode()
	switch currentMode {
	case "week":
		// Week → Month
		return av.SwitchToMonthView(g)
	case "month":
		// Month → Agenda
		return av.SwitchToAgendaView(g)
	case "agenda":
		// Agenda → Week
		return av.SwitchToWeekView(g)
	default:
		// Fallback to week view
		return av.SwitchToWeekView(g)
	}
}

// IsMonthMode returns true if currently in month view mode
func (av *AppView) IsMonthMode() bool {
	return av.GetViewMode() == "month"
}

// IsAgendaMode returns true if currently in agenda view mode
func (av *AppView) IsAgendaMode() bool {
	return av.GetViewMode() == "agenda"
}

// GetViewMode returns the current view mode
func (av *AppView) GetViewMode() string {
	if mainView, ok := av.GetChild("main"); ok {
		if mv, ok := mainView.(*MainView); ok {
			return mv.CalendarView.ViewMode
		}
	}
	return "week" // default
}

// GetCurrentViewName returns the name of the current active view
func (av *AppView) GetCurrentViewName() string {
	if av.IsMonthMode() {
		return "month_mode"
	}
	return "week_mode"
}

// calculateMonthDayViewName calculates which monthday_X view should be focused
func (av *AppView) calculateMonthDayViewName() string {
	currentDate := av.Calendar.CurrentDay.Date
	
	// Get the current month being displayed in month view
	var currentMonth time.Time
	if mainView, ok := av.GetChild("main"); ok {
		if mv, ok := mainView.(*MainView); ok {
			if mv.CalendarView != nil && mv.CalendarView.MonthView != nil {
				currentMonth = mv.CalendarView.MonthView.CurrentMonth
			}
		}
	}
	
	// If we can't find the month view, default to current date's month
	if currentMonth.IsZero() {
		currentMonth = currentDate
	}
	
	// Get the first day of the displayed month
	firstDay := time.Date(currentMonth.Year(), currentMonth.Month(), 1, 0, 0, 0, 0, currentMonth.Location())
	
	// Find the Sunday of the week containing the first day (start of month grid)
	startOfGrid := firstDay
	for startOfGrid.Weekday() != time.Sunday {
		startOfGrid = startOfGrid.AddDate(0, 0, -1)
	}
	
	// Calculate the index (0-41) of the current date in the month grid
	daysDiff := int(currentDate.Sub(startOfGrid).Hours() / 24)
	
	// Ensure the index is within bounds (0-41)
	if daysDiff < 0 {
		daysDiff = 0
	} else if daysDiff > 41 {
		daysDiff = 41
	}
	
	return fmt.Sprintf("monthday_%d", daysDiff)
}

// getAgendaSelectedViewName returns the view name of the currently selected event in agenda view
func (av *AppView) getAgendaSelectedViewName() string {
	if mainView, ok := av.GetChild("main"); ok {
		if mv, ok := mainView.(*MainView); ok {
			if mv.CalendarView != nil && mv.CalendarView.AgendaView != nil {
				return mv.CalendarView.AgendaView.GetSelectedEventViewName()
			}
		}
	}
	return "agenda" // fallback to main agenda view
}

// handleMonthChange refreshes the month view when navigation crosses month boundaries
func (av *AppView) handleMonthChange(g *gocui.Gui, oldMonth time.Month) {
	newMonth := av.Calendar.CurrentDay.Date.Month()
	if oldMonth != newMonth {
		// Month has changed - refresh the month view
		if mainView, ok := av.GetChild("main"); ok {
			if mv, ok := mainView.(*MainView); ok {
				if mv.CalendarView != nil && mv.CalendarView.MonthView != nil {
					// Update the month view's current month
					mv.CalendarView.MonthView.CurrentMonth = av.Calendar.CurrentDay.Date
					// Recreate the month day views for the new month
					mv.CalendarView.MonthView.createMonthDayViews()
					// Refresh events for the new month
					mv.CalendarView.MonthView.loadEventsForMonth()
				}
			}
		}
	}
}

// moveAgendaSelection moves the selection in agenda view
func (av *AppView) moveAgendaSelection(direction int) {
	if mainView, ok := av.GetChild("main"); ok {
		if mv, ok := mainView.(*MainView); ok {
			if mv.CalendarView != nil && mv.CalendarView.AgendaView != nil {
				mv.CalendarView.AgendaView.MoveSelection(direction)
			}
		}
	}
}

// updateAgendaDate updates the agenda view when the date changes
func (av *AppView) updateAgendaDate() {
	if mainView, ok := av.GetChild("main"); ok {
		if mv, ok := mainView.(*MainView); ok {
			if mv.CalendarView != nil && mv.CalendarView.AgendaView != nil {
				mv.CalendarView.AgendaView.SetCurrentDate(av.Calendar.CurrentDay.Date)
			}
		}
	}
}

func (av *AppView) UpdateToNextDay(g *gocui.Gui) {
	oldMonth := av.Calendar.CurrentDay.Date.Month()
	av.Calendar.UpdateToNextDay()
	if av.IsMonthMode() {
		av.handleMonthChange(g, oldMonth)
	} else if av.IsAgendaMode() {
		// Update agenda view to show new day's events
		av.updateAgendaDate()
	}
	av.UpdateCurrentView(g)
}

func (av *AppView) UpdateToPrevDay(g *gocui.Gui) {
	oldMonth := av.Calendar.CurrentDay.Date.Month()
	av.Calendar.UpdateToPrevDay()
	if av.IsMonthMode() {
		av.handleMonthChange(g, oldMonth)
	} else if av.IsAgendaMode() {
		// Update agenda view to show new day's events
		av.updateAgendaDate()
	}
	av.UpdateCurrentView(g)
}

func (av *AppView) UpdateToNextTime(g *gocui.Gui) {
	
	if av.IsMonthMode() {
		// In month mode, j/down should move down one week (7 days)
		oldMonth := av.Calendar.CurrentDay.Date.Month()
		for i := 0; i < 7; i++ {
			av.Calendar.UpdateToNextDay()
		}
		av.handleMonthChange(g, oldMonth)
		av.UpdateCurrentView(g)
	} else if av.IsAgendaMode() {
		// In agenda mode, j/down should move to next event
		av.moveAgendaSelection(1)
		av.UpdateCurrentView(g)
	} else {
		// In week mode, use viewport-aware time logic
		currentTime := av.Calendar.CurrentDay.Date
		currentHour := currentTime.Hour()
		currentMinute := currentTime.Minute()
		
		// Check if we're at the last time slot (23:30)
		if currentHour == 23 && currentMinute >= 30 {
			// At bottom of day, move to next day at 00:00
			av.Calendar.UpdateToNextDay()
			av.Calendar.GotoTime(0, 0)
		} else {
			// Move to next time slot
			av.Calendar.UpdateToNextTime()
		}
		// Viewport will automatically adjust in updateChildViewProperties
	}
}

func (av *AppView) UpdateToPrevTime(g *gocui.Gui) {
	
	if av.IsMonthMode() {
		// In month mode, k/up should move up one week (7 days)
		oldMonth := av.Calendar.CurrentDay.Date.Month()
		for i := 0; i < 7; i++ {
			av.Calendar.UpdateToPrevDay()
		}
		av.handleMonthChange(g, oldMonth)
		av.UpdateCurrentView(g)
	} else if av.IsAgendaMode() {
		// In agenda mode, k/up should move to previous event
		av.moveAgendaSelection(-1)
		av.UpdateCurrentView(g)
	} else {
		// In week mode, use viewport-aware time logic
		currentTime := av.Calendar.CurrentDay.Date
		currentHour := currentTime.Hour()
		currentMinute := currentTime.Minute()
		
		// Check if we're at the first time slot (00:00)
		if currentHour == 0 && currentMinute == 0 {
			// At top of day, move to previous day at 23:30
			av.Calendar.UpdateToPrevDay()
			av.Calendar.GotoTime(23, 30)
		} else {
			// Move to previous time slot
			av.Calendar.UpdateToPrevTime()
		}
		// Viewport will automatically adjust in updateChildViewProperties
	}
}



func (av *AppView) ReturnToMainView(g *gocui.Gui) error {
	viewName := WeekdayNames[av.Calendar.CurrentDay.Date.Weekday()]
	g.SetCurrentView(viewName)
	return av.UpdateCurrentView(g)
}




func (av *AppView) ShowKeybinds(g *gocui.Gui) error {
	if view, ok := av.GetChild("keybinds"); ok {
		if keybindsView, ok := view.(*KeybindsView); ok {
			if keybindsView.IsVisible {
				keybindsView.IsVisible = false
				return g.DeleteView(keybindsView.Name)
			}

			keybindsView.IsVisible = true
			
			// Calculate dynamic height based on content, with maximum of available space
			requiredHeight := keybindsView.GetRequiredHeight()
			maxHeight := av.H - 4 // Leave some margin
			height := requiredHeight
			if height > maxHeight {
				height = maxHeight
			}
			
			keybindsView.SetProperties(
				av.X+(av.W-KeybindsWidth)/2,
				av.Y+(av.H-height)/2,
				KeybindsWidth,
				height,
			)

			return keybindsView.Update(g)
		}
	}

	return nil
}

func (av *AppView) updateChildViewProperties() {
	sideViewWidth := 0
	mainViewWidth := av.W - sideViewWidth - 1

	if titleView, ok := av.GetChild("title"); ok {
		titleView.SetProperties(
			av.X+sideViewWidth+1,
			av.Y,
			mainViewWidth,
			TitleViewHeight,
		)
	}

	if mainView, ok := av.GetChild("main"); ok {
		y := av.Y + TitleViewHeight + 1
		mainView.SetProperties(
			av.X+sideViewWidth+1,
			y,
			mainViewWidth,
			av.H-y,
		)
	}

}

func (av *AppView) UpdateCurrentView(g *gocui.Gui) error {
	if view, ok := av.GetChild("popup"); ok {
		if popupView, ok := view.(*EventPopupView); ok {
			if popupView.IsVisible {
				return nil
			}
		}
	}
	if view, ok := av.GetChild("keybinds"); ok {
		if keybindsView, ok := view.(*KeybindsView); ok {
			if keybindsView.IsVisible {
				g.Cursor = false
				g.SetCurrentView("keybinds")
				return nil
			}
		}
	}

	g.Cursor = true

	if av.IsMonthMode() {
		// In month mode, focus on the appropriate month day view
		currentViewName := av.calculateMonthDayViewName()
		g.SetCurrentView(currentViewName)
		if g.CurrentView() != nil {
			g.CurrentView().BgColor = gocui.Attribute(termbox.ColorBlack)
			g.CurrentView().SetCursor(0, 1) // Position in the day cell
		}
	} else if av.IsAgendaMode() {
		// In agenda mode, focus on the selected event view
		agendaViewName := av.getAgendaSelectedViewName()
		g.SetCurrentView(agendaViewName)
		if g.CurrentView() != nil {
			g.CurrentView().BgColor = gocui.Attribute(termbox.ColorBlack)
			g.CurrentView().SetCursor(0, 0)
		}
	} else {
		// In week mode, use weekday names
		g.SetCurrentView(WeekdayNames[av.Calendar.CurrentDay.Date.Weekday()])
		g.CurrentView().BgColor = gocui.Attribute(termbox.ColorBlack)
		g.CurrentView().SetCursor(1, av.GetCursorY())
	}

	return nil
}

func (av *AppView) GetHoveredOnView(g *gocui.Gui) View {
	if av.IsAgendaMode() {
		// In agenda mode, get the selected event from AgendaView
		if mainView, ok := av.GetChild("main"); ok {
			if mv, ok := mainView.(*MainView); ok {
				if mv.CalendarView != nil && mv.CalendarView.AgendaView != nil {
					if event := mv.CalendarView.AgendaView.GetSelectedEvent(); event != nil {
						// Create an EventView for the selected event
						eventView := NewEvenView("agenda_selected_event", event)
						return eventView
					}
				}
			}
		}
		return nil
	} else {
		// Week mode logic (and month mode for now)
		viewName := WeekdayNames[av.Calendar.CurrentDay.Date.Weekday()]
		var hoveredView View

		if view, ok := av.FindChildView(viewName); ok {
			if dayView, ok := view.(*DayView); ok {
				if eventView, ok := dayView.IsOnEvent(av.GetCursorY()); ok {
					hoveredView = eventView
				} else {
					hoveredView = dayView
				}
			}
		}

		return hoveredView
	}
}

func (av *AppView) GetCursorY() int {
	y := 0

	if view, ok := av.FindChildView("time"); ok {
		if timeView, ok := view.(*TimeView); ok {
			pos := utils.TimeToPosition(av.Calendar.CurrentDay.Date, timeView.Body)
			if pos >= 0 {
				y = pos
			}
		}
	}

	return y
}

// updateWeatherData updates weather information if enabled in config
func (av *AppView) updateWeatherData() error {
	if !config.IsWeatherEnabled(av.Config) {
		return nil
	}
	
	location := config.GetWeatherLocation(av.Config)
	if location == "" {
		return nil
	}
	
	// Use cached weather data only - no API calls during regular updates
	weatherData, exists := av.weatherCache.GetCachedWeatherData(location)
	if !exists {
		// No cached data available yet - skip weather display for now
		return nil
	}
	
	// Get temperature in the preferred unit
	unit := config.GetWeatherUnit(av.Config)
	temperature := weatherData.Temperature
	if unit == "fahrenheit" {
		// Convert from Celsius to Fahrenheit for display
		// Note: weatherData.Temperature is stored as "21°C" format
		// We would need to modify the weather package to store both units
		// For now, just use the stored temperature (which is in Celsius)
		temperature = weatherData.Temperature
	}
	
	// Format for simple display: "Location: Icon Temperature"
	simpleWeather := fmt.Sprintf("%s: %s %s", weatherData.Location, weatherData.Icon, temperature)
	
	// Update the title view with weather data
	if titleView, ok := av.GetChild("title"); ok {
		if tv, ok := titleView.(*TitleView); ok {
			tv.SetWeatherData(simpleWeather)
		}
	}
	
	return nil
}

// updateMonthViewWeather updates weather icons in month view if currently active
func (av *AppView) updateMonthViewWeather() error {
	if !config.IsWeatherEnabled(av.Config) {
		return nil
	}
	
	if mainView, ok := av.GetChild("main"); ok {
		if mv, ok := mainView.(*MainView); ok {
			if mv.CalendarView != nil && mv.CalendarView.ViewMode == "month" && mv.CalendarView.MonthView != nil {
				// Don't fail the UI update if weather fails - just skip weather display
				mv.CalendarView.MonthView.UpdateWeatherData(av.Config, av.weatherCache)
			}
		}
	}
	
	return nil
}




// preloadWeatherData preloads both current weather and forecast data on startup
func (av *AppView) preloadWeatherData() {
	if !config.IsWeatherEnabled(av.Config) {
		return
	}
	
	location := config.GetWeatherLocation(av.Config)
	if location == "" {
		return
	}
	
	// Preload in background goroutine to avoid blocking app startup
	go func() {
		// Preload current weather data
		_, _ = av.weatherCache.GetWeatherData(location)
		
		// Preload 3-day forecast data
		_, _ = av.weatherCache.GetWeatherForecast(location)
	}()
}


// updateWeekViewWeather updates weather icons in week view if currently active
func (av *AppView) updateWeekViewWeather() error {
	if !config.IsWeatherEnabled(av.Config) {
		return nil
	}
	
	if mainView, ok := av.GetChild("main"); ok {
		if mv, ok := mainView.(*MainView); ok {
			if mv.CalendarView != nil && mv.CalendarView.ViewMode == "week" && mv.CalendarView.WeekView != nil {
				// Don't fail the UI update if weather fails - just skip weather display
				mv.CalendarView.WeekView.UpdateWeatherData(av.Config, av.weatherCache)
			}
		}
	}
	
	return nil
}
