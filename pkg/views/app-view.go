package views

import (
	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/samuelstranges/chronos/internal/config"
	"github.com/samuelstranges/chronos/internal/database"
	"github.com/samuelstranges/chronos/internal/eventmanager"
	"github.com/samuelstranges/chronos/internal/utils"
	"github.com/jroimartin/gocui"
	"github.com/nsf/termbox-go"
	"time"
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
	}
	

	av.AddChild("title", NewTitleView(c))
	av.AddChild("popup", NewEvenPopup(g, c, db, av.EventManager))
	av.AddChild("main", NewMainView(c))
	
	// Set up error handler for EventManager after popup is created
	av.setupErrorHandler(g)
	
	
	av.AddChild("keybinds", NewKeybindsView())

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

func (av *AppView) updateEventsFromDatabase() error {
	for _, v := range av.Calendar.CurrentWeek.Days {
		clear(v.Events)

		var err error
		events, err := av.Database.GetEventsByDate(v.Date)
		if err != nil {
			return err
		}

		v.Events = events
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

func (av *AppView) UpdateToNextDay(g *gocui.Gui) {
	av.Calendar.UpdateToNextDay()
	av.UpdateCurrentView(g)
}

func (av *AppView) UpdateToPrevDay(g *gocui.Gui) {
	av.Calendar.UpdateToPrevDay()
	av.UpdateCurrentView(g)
}

func (av *AppView) UpdateToNextTime(g *gocui.Gui) {
	_, height := g.CurrentView().Size()
	if _, y := g.CurrentView().Cursor(); y < height-1 {
		av.Calendar.UpdateToNextTime()
	}
}

func (av *AppView) UpdateToPrevTime(g *gocui.Gui) {
	if _, y := g.CurrentView().Cursor(); y > 0 {
		av.Calendar.UpdateToPrevTime()
	}
}



func (av *AppView) ReturnToMainView(g *gocui.Gui) error {
	viewName := WeekdayNames[av.Calendar.CurrentDay.Date.Weekday()]
	g.SetCurrentView(viewName)
	return av.UpdateCurrentView(g)
}


// refreshCurrentTimeHighlighting updates the current time highlighting for today's column
func (av *AppView) refreshCurrentTimeHighlighting(g *gocui.Gui) {
	// Iterate through all day views and update their current time highlighting
	for _, day := range av.Calendar.CurrentWeek.Days {
		if view, ok := av.FindChildView(WeekdayNames[day.Date.Weekday()]); ok {
			if dayView, ok := view.(*DayView); ok {
				// Call the new update method on each day view
				dayView.updateCurrentTimeHighlight(g)
			}
		}
	}
}

// Override UpdateChildren to automatically refresh current time highlighting
func (av *AppView) UpdateChildren(g *gocui.Gui) error {
	// Call the base UpdateChildren implementation
	err := av.BaseView.UpdateChildren(g)
	if err != nil {
		return err
	}
	
	// Automatically refresh current time highlighting after any child update
	av.refreshCurrentTimeHighlighting(g)
	
	return nil
}

func (av *AppView) ShowKeybinds(g *gocui.Gui) error {
	if view, ok := av.GetChild("keybinds"); ok {
		if keybindsView, ok := view.(*KeybindsView); ok {
			if keybindsView.IsVisible {
				keybindsView.IsVisible = false
				return g.DeleteView(keybindsView.Name)
			}

			keybindsView.IsVisible = true
			keybindsView.SetProperties(
				av.X+(av.W-KeybindsWidth)/2,
				av.Y+(av.H-KeybindsHeight)/2,
				KeybindsWidth,
				KeybindsHeight,
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

	g.SetCurrentView(WeekdayNames[av.Calendar.CurrentDay.Date.Weekday()])
	g.CurrentView().BgColor = gocui.Attribute(termbox.ColorBlack)
	g.CurrentView().SetCursor(1, av.GetCursorY())

	return nil
}

func (av *AppView) GetHoveredOnView(g *gocui.Gui) View {
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


