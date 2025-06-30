package views

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/HubertBel/lazyorg/internal/calendar"
	"github.com/HubertBel/lazyorg/internal/database"
	"github.com/HubertBel/lazyorg/internal/eventmanager"
	"github.com/HubertBel/lazyorg/internal/utils"
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

func (epv *EventPopupView) NewEventForm(g *gocui.Gui, title, name, time, location, duration, frequency, occurence, description, color string) *component.Form {
	form := component.NewForm(g, title, epv.X, epv.Y, epv.W, epv.H)

	form.AddInputField("Name", LabelWidth, FieldWidth).SetText(name).AddValidate("Invalid name", utils.ValidateName)
	form.AddInputField("Time", LabelWidth, FieldWidth).SetText(time).AddValidate("Invalid time", utils.ValidateTime)
	form.AddInputField("Location", LabelWidth, FieldWidth).SetText(location)
	form.AddInputField("Duration (eg. 1.5)", LabelWidth, FieldWidth).SetText(duration).AddValidate("Invalid duration", utils.ValidateDuration)
	form.AddInputField("Frequency", LabelWidth, FieldWidth).SetText(frequency).AddValidate("Invalid frequency", utils.ValidateNumber)
	form.AddInputField("Occurence", LabelWidth, FieldWidth).SetText(occurence).AddValidate("Invalid occurence", utils.ValidateNumber)
	form.AddInputField("Color", LabelWidth, FieldWidth).SetText(color)
	form.AddInputField("Description", LabelWidth, FieldWidth).SetText(description)

	return form
}

func (epv *EventPopupView) EditEventForm(g *gocui.Gui, title, name, time, location, duration, description, color string) *component.Form {
	form := component.NewForm(g, title, epv.X, epv.Y, epv.W, epv.H)

	form.AddInputField("Name", LabelWidth, FieldWidth).SetText(name).AddValidate("Invalid name", utils.ValidateName)
	form.AddInputField("Time", LabelWidth, FieldWidth).SetText(time).AddValidate("Invalid time", utils.ValidateTime)
	form.AddInputField("Location", LabelWidth, FieldWidth).SetText(location)
	form.AddInputField("Duration", LabelWidth, FieldWidth).SetText(duration).AddValidate("Invalid duration", utils.ValidateDuration)
	form.AddInputField("Color", LabelWidth, FieldWidth).SetText(color)
	form.AddInputField("Description", LabelWidth, FieldWidth).SetText(description)

	return form
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

func (epv *EventPopupView) CreateEventFromInputs(existingEvent *calendar.Event) *calendar.Event {
	for _, v := range epv.Form.GetInputs() {
		if !v.IsValid() {
			return nil
		}
	}

	name := epv.Form.GetFieldText("Name")
	time, _ := time.Parse(TimeFormat, epv.Form.GetFieldText("Time"))
	location := epv.Form.GetFieldText("Location")

	// Try both field names since NewEventForm and EditEventForm use different labels
	durationText := strings.TrimSpace(epv.Form.GetFieldText("Duration (eg. 1.5)"))
	if durationText == "" {
		durationText = strings.TrimSpace(epv.Form.GetFieldText("Duration"))
	}
	duration := 1.0 // Default to 1 hour for new events
	
	// First check if there's a value in the form field
	if durationText != "" {
		if parsedDuration, err := strconv.ParseFloat(durationText, 64); err == nil && parsedDuration > 0 {
			duration = parsedDuration
		} else if existingEvent != nil {
			// If parsing failed but we're editing, use existing duration
			duration = existingEvent.DurationHour
		}
	} else if existingEvent != nil {
		// If field is empty and we're editing, preserve original duration
		duration = existingEvent.DurationHour
	}
	frequency, _ := strconv.Atoi(epv.Form.GetFieldText("Frequency"))
	occurence, _ := strconv.Atoi(epv.Form.GetFieldText("Occurence"))
	colorName := epv.Form.GetFieldText("Color")
	description := epv.Form.GetFieldText("Description")

	color := calendar.ColorNameToAttribute(colorName)
	if color == gocui.ColorDefault {
		color = calendar.GenerateColorFromName(name)
	}

	return calendar.NewEvent(name, description, location, time, duration, frequency, occurence, color)
}

func (epv *EventPopupView) AddEvent(g *gocui.Gui, v *gocui.View) error {
	if !epv.IsVisible {
		return nil
	}

	var newEvent *calendar.Event
	if newEvent = epv.CreateEventFromInputs(nil); newEvent == nil {
		return nil
	}
	events := newEvent.GetReccuringEvents()

	for _, event := range events {
		if _, success := epv.EventManager.AddEvent(event); !success {
			// Error is handled by EventManager internally
			return nil
		}
	}

	return epv.Close(g, v)
}

func (epv *EventPopupView) EditEvent(g *gocui.Gui, v *gocui.View, event *calendar.Event) error {
	if !epv.IsVisible {
		return nil
	}

	var newEvent *calendar.Event
	if newEvent = epv.CreateEventFromInputs(event); newEvent == nil {
		return nil
	}
	newEvent.Id = event.Id

	if !epv.EventManager.UpdateEvent(event.Id, newEvent) {
		// Error is handled by EventManager internally
		return nil
	}

	return epv.Close(g, v)
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
		v.BgColor = gocui.ColorRed
		v.FgColor = gocui.ColorWhite
		v.Frame = true
		
		// Center the message
		lines := []string{
			"",
			"  " + message,
			"",
			"  Press ENTER or ESC to close",
		}
		
		for _, line := range lines {
			fmt.Fprintln(v, line)
		}
	}
	
	// Set focus to error popup
	g.SetCurrentView("error-popup")
	
	// Set up keybindings for closing the error popup
	g.SetKeybinding("error-popup", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return epv.closeErrorPopup(g)
	})
	g.SetKeybinding("error-popup", gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return epv.closeErrorPopup(g)
	})
	
	return nil
}

// closeErrorPopup removes the error popup and restores focus
func (epv *EventPopupView) closeErrorPopup(g *gocui.Gui) error {
	// Delete the error popup view
	if err := g.DeleteView("error-popup"); err != nil {
		return err
	}
	
	// Remove keybindings
	g.DeleteKeybinding("error-popup", gocui.KeyEnter, gocui.ModNone)
	g.DeleteKeybinding("error-popup", gocui.KeyEsc, gocui.ModNone)
	
	// Return focus to the main popup
	if epv.IsVisible {
		_, err := g.SetCurrentView("popup")
		return err
	}
	
	// Return focus to the main calendar view
	_, err := g.SetCurrentView("Monday-30") // This should be dynamic based on current day
	return err
}

func (epv *EventPopupView) positionCursorsAtEnd(g *gocui.Gui) {
	for _, input := range epv.Form.GetInputs() {
		fieldName := input.GetLabel()
		fieldText := input.GetFieldText()
		if fieldText != "" {
			if v, err := g.View(fieldName); err == nil {
				v.MoveCursor(len(fieldText), 0, true)
			}
		}
	}
}

func (epv *EventPopupView) GotoForm(g *gocui.Gui, title string) *component.Form {
	form := component.NewForm(g, title, epv.X, epv.Y, epv.W, epv.H)

	currentTime := epv.Calendar.CurrentDay.Date
	defaultDate := fmt.Sprintf("%04d%02d%02d", currentTime.Year(), currentTime.Month(), currentTime.Day())
	defaultHour := fmt.Sprintf("%02d", currentTime.Hour())
	
	form.AddInputField("Date", LabelWidth, FieldWidth).SetText(defaultDate).AddValidate("Invalid date (YYYYMMDD)", utils.ValidateDate)
	form.AddInputField("Hour", LabelWidth, FieldWidth).SetText(defaultHour).AddValidate("Invalid hour (00-23)", utils.ValidateHourMinute)

	return form
}

func (epv *EventPopupView) ColorPickerForm(g *gocui.Gui, title string) *component.Form {
	form := component.NewForm(g, title, epv.X, epv.Y, epv.W, epv.H)

	// Start with empty field so user can type single letters directly
	form.AddInputField("Color", LabelWidth, FieldWidth).SetText("")

	return form
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

func (epv *EventPopupView) Goto(g *gocui.Gui, v *gocui.View) error {
	if !epv.IsVisible {
		return nil
	}

	for _, v := range epv.Form.GetInputs() {
		if !v.IsValid() {
			return nil
		}
	}

	dateStr := epv.Form.GetFieldText("Date")
	hourStr := epv.Form.GetFieldText("Hour")
	
	year, _ := strconv.Atoi(dateStr[:4])
	month, _ := strconv.Atoi(dateStr[4:6])
	day, _ := strconv.Atoi(dateStr[6:8])
	hour, _ := strconv.Atoi(hourStr)

	// Set both date and time together
	currentDate := epv.Calendar.CurrentDay.Date
	newDate := time.Date(year, time.Month(month), day, hour, 0, 0, 0, currentDate.Location())
	epv.Calendar.CurrentDay.Date = newDate
	epv.Calendar.UpdateWeek()

	return epv.Close(g, v)
}

// expandColorShorthand expands single letter shortcuts to full color names
func (epv *EventPopupView) expandColorShorthand(input string) string {
	input = strings.ToLower(strings.TrimSpace(input))
	
	switch input {
	case "r":
		return "Red"
	case "g":
		return "Green"
	case "y":
		return "Yellow"
	case "b":
		return "Blue"
	case "m":
		return "Magenta"
	case "c":
		return "Cyan"
	case "w":
		return "White"
	default:
		// Return the input as-is if it's not a single letter shortcut
		return input
	}
}

func (epv *EventPopupView) SelectColor(g *gocui.Gui, v *gocui.View) error {
	if !epv.IsVisible {
		return nil
	}

	colorInput := epv.Form.GetFieldText("Color")
	colorName := epv.expandColorShorthand(colorInput)
	
	// Use callback to handle color selection
	if epv.ColorPickerCallback != nil {
		if err := epv.ColorPickerCallback(colorName); err != nil {
			return err
		}
	}

	return epv.Close(g, v)
}

func (epv *EventPopupView) SearchForm(g *gocui.Gui, title string) *component.Form {
	form := component.NewForm(g, title, epv.X, epv.Y, epv.W, epv.H)
	
	form.AddInputField("Search", LabelWidth, FieldWidth).SetText("")
	
	return form
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

func (epv *EventPopupView) ExecuteSearch(g *gocui.Gui, v *gocui.View) error {
	if !epv.IsVisible {
		return nil
	}

	query := strings.TrimSpace(epv.Form.GetFieldText("Search"))
	if query == "" {
		return epv.Close(g, v)
	}

	// Call the search callback if it exists
	if epv.SearchCallback != nil {
		if err := epv.SearchCallback(query); err != nil {
			return err
		}
	}
	
	return epv.Close(g, v)
}

func (epv *EventPopupView) addKeybind(key interface{}, handler func(g *gocui.Gui, v *gocui.View) error) {
	for _, item := range epv.Form.GetItems() {
		item.AddHandlerOnly(key, handler)
	}
}
