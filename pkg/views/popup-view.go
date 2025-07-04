package views

import (
	"fmt"
	"strconv"
	"time"

	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/samuelstranges/chronos/internal/database"
	"github.com/samuelstranges/chronos/internal/eventmanager"
	component "github.com/j-04/gocui-component"
	"github.com/jroimartin/gocui"
)

type EventPopupView struct {
	*BaseView
	Form         *component.Form
	Calendar     *calendar.Calendar
	Database     *database.Database
	EventManager *eventmanager.EventManager

	IsVisible bool
	SearchCallback func(query string) error
	ColorPickerCallback func(colorName string) error
}

func NewEvenPopup(g *gocui.Gui, c *calendar.Calendar, db *database.Database, em *eventmanager.EventManager) *EventPopupView {

	epv := &EventPopupView{
		BaseView:     NewBaseView("popup"),
		Form:         nil,
		Calendar:     c,
		Database:     db,
		EventManager: em,
		IsVisible:    false,
	}

	return epv
}

func (epv *EventPopupView) Update(g *gocui.Gui) error {
	return nil
}


func (epv *EventPopupView) ShowNewEventPopup(g *gocui.Gui) error {
	if epv.IsVisible {
		return nil
	}

	epv.Form = epv.NewEventForm(g, "New Event", "", epv.Calendar.CurrentDay.Date.Format(TimeFormat), "", "", "7", "1", "", "Red")

	epv.addKeybind(gocui.KeyEsc, epv.Close)
	epv.addKeybind(gocui.KeyEnter, epv.AddEvent)

	epv.Form.AddButton("Add", epv.AddEvent)
	epv.Form.AddButton("Cancel", epv.Close)

	epv.Form.SetCurrentItem(0)
	epv.IsVisible = true
	epv.Form.Draw()
	
	epv.positionCursorsAtEnd(g)

	return nil
}

func (epv *EventPopupView) ShowEditEventPopup(g *gocui.Gui, eventView *EventView) error {
	if epv.IsVisible {
		return nil
	}

	event := eventView.Event

	epv.Form = epv.EditEventForm(g,
		"Edit Event",
		event.Name,
		event.Time.Format(TimeFormat),
		event.Location,
		strconv.FormatFloat(event.DurationHour, 'f', -1, 64),
		event.Description,
		calendar.ColorAttributeToName(event.Color),
	)

	editHandler := func(g *gocui.Gui, v *gocui.View) error {
		return epv.EditEvent(g, v, event)
	}
	epv.addKeybind(gocui.KeyEsc, epv.Close)
	epv.addKeybind(gocui.KeyEnter, editHandler)

	epv.Form.AddButton("Edit", editHandler)
	epv.Form.AddButton("Cancel", epv.Close)

	epv.Form.SetCurrentItem(0)
	epv.IsVisible = true
	epv.Form.Draw()
	
	epv.positionCursorsAtEnd(g)

	return nil
}



func (epv *EventPopupView) Close(g *gocui.Gui, v *gocui.View) error {
	epv.IsVisible = false
	return epv.Form.Close(g, v)
}

// ShowErrorMessage displays an error popup to the user
func (epv *EventPopupView) ShowErrorMessage(g *gocui.Gui, title, message string) error {
	// Calculate popup size and position
	maxX, maxY := g.Size()
	width := 60
	height := 8
	x := (maxX - width) / 2
	y := (maxY - height) / 2
	
	// Create error popup view
	v, err := g.SetView("error-popup", x, y, x+width, y+height)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " " + title + " "
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorRed
		v.Frame = true
		
		// Center the message
		lines := []string{
			"",
			"  " + message,
			"",
			"  Auto-dismissing in 3 seconds...",
		}
		
		for _, line := range lines {
			fmt.Fprintln(v, line)
		}
	}
	
	// Auto-dismiss after 3 seconds
	go func() {
		time.Sleep(3 * time.Second)
		g.Update(func(g *gocui.Gui) error {
			// Check if the error popup still exists before trying to close it
			if _, err := g.View("error-popup"); err == nil {
				return epv.closeErrorPopup(g)
			}
			return nil
		})
	}()
	
	return nil
}

// closeErrorPopup removes the error popup and restores focus
func (epv *EventPopupView) closeErrorPopup(g *gocui.Gui) error {
	// Delete the error popup view
	if err := g.DeleteView("error-popup"); err != nil {
		return err
	}
	
	// Return focus to the main popup if it's visible
	if epv.IsVisible {
		if _, err := g.View("popup"); err == nil {
			g.SetCurrentView("popup")
		}
		return nil
	}
	
	// Try to find any valid view to focus on
	views := g.Views()
	if len(views) > 0 {
		for _, view := range views {
			if view.Name() != "error-popup" {
				g.SetCurrentView(view.Name())
				break
			}
		}
	}
	
	return nil
}


func (epv *EventPopupView) ShowGotoPopup(g *gocui.Gui) error {
	if epv.IsVisible {
		return nil
	}

	epv.Form = epv.GotoForm(g, "Goto Date/Time")

	epv.addKeybind(gocui.KeyEsc, epv.Close)
	epv.addKeybind(gocui.KeyEnter, epv.Goto)

	epv.Form.AddButton("Goto", epv.Goto)
	epv.Form.AddButton("Cancel", epv.Close)

	epv.Form.SetCurrentItem(0)
	epv.IsVisible = true
	epv.Form.Draw()

	epv.positionCursorsAtEnd(g)

	return nil
}

func (epv *EventPopupView) ShowColorPickerPopup(g *gocui.Gui) error {
	if epv.IsVisible {
		return nil
	}

	epv.Form = epv.ColorPickerForm(g, "Select Color (r/g/y/b/m/c/w or full name)")

	epv.addKeybind(gocui.KeyEsc, epv.Close)
	epv.addKeybind(gocui.KeyEnter, epv.SelectColor)

	epv.Form.AddButton("Select", epv.SelectColor)
	epv.Form.AddButton("Cancel", epv.Close)

	epv.Form.SetCurrentItem(0)
	epv.IsVisible = true
	epv.Form.Draw()

	epv.positionCursorsAtEnd(g)

	return nil
}


func (epv *EventPopupView) ShowSearchPopup(g *gocui.Gui) error {
	if epv.IsVisible {
		return nil
	}

	epv.Form = epv.SearchForm(g, "Search Events")

	epv.addKeybind(gocui.KeyEsc, epv.Close)
	epv.addKeybind(gocui.KeyEnter, epv.ExecuteSearch)

	epv.Form.AddButton("Search", epv.ExecuteSearch)
	epv.Form.AddButton("Cancel", epv.Close)

	epv.Form.SetCurrentItem(0)
	epv.IsVisible = true
	epv.Form.Draw()

	return nil
}

