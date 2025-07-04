package views

import (
	"strconv"
	"strings"
	"time"

	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/samuelstranges/chronos/internal/database"
	"github.com/jroimartin/gocui"
)

// CreateEventFromInputs creates an event from form inputs
func (epv *EventPopupView) CreateEventFromInputs(existingEvent *calendar.Event) *calendar.Event {
	for _, v := range epv.Form.GetInputs() {
		if !v.IsValid() {
			return nil
		}
	}

	name := epv.Form.GetFieldText("Name")
	dateStr := epv.Form.GetFieldText("Date")
	timeStr := epv.Form.GetFieldText("Time")
	location := epv.Form.GetFieldText("Location")
	
	// Parse date and time separately then combine
	dateTime, _ := time.ParseInLocation("2006-01-02 15:04", dateStr+" "+timeStr, epv.Calendar.CurrentDay.Date.Location())

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

	return calendar.NewEvent(name, description, location, dateTime, duration, frequency, occurence, color)
}

// AddEvent handler for adding new events
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

// EditEvent handler for editing existing events
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

// Goto handler for navigating to specific date/time
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

// SelectColor handler for color selection
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

// ExecuteSearch handler for executing search
func (epv *EventPopupView) ExecuteSearch(g *gocui.Gui, v *gocui.View) error {
	if !epv.IsVisible {
		return nil
	}

	// Collect all search criteria
	criteria := database.SearchCriteria{
		Query:     strings.TrimSpace(epv.Form.GetFieldText("Query")),
		StartDate: strings.TrimSpace(epv.Form.GetFieldText("From Date")),
		StartTime: "", // No separate time fields in simplified form
		EndDate:   strings.TrimSpace(epv.Form.GetFieldText("To Date")),
		EndTime:   "", // No separate time fields in simplified form
	}

	// At least one search parameter must be provided
	if criteria.Query == "" && criteria.StartDate == "" && criteria.EndDate == "" {
		return epv.Close(g, v)
	}

	// Call the search callback if it exists
	if epv.SearchCallback != nil {
		if err := epv.SearchCallback(criteria); err != nil {
			return err
		}
	}
	
	return epv.Close(g, v)
}

// addKeybind adds a keybinding to all form items
func (epv *EventPopupView) addKeybind(key interface{}, handler func(g *gocui.Gui, v *gocui.View) error) {
	for _, item := range epv.Form.GetItems() {
		item.AddHandlerOnly(key, handler)
	}
}