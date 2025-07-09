package views

import (
	"fmt"

	"github.com/samuelstranges/chronos/internal/utils"
	component "github.com/j-04/gocui-component"
	"github.com/jroimartin/gocui"
)

// NewEventForm creates a form for adding new events
func (epv *EventPopupView) NewEventForm(g *gocui.Gui, title, name, date, time, location, duration, frequency, occurence, description, color string) *component.Form {
	form := component.NewForm(g, title, epv.X, epv.Y, epv.W, epv.H)

	form.AddInputField("Name", LabelWidth, FieldWidth).SetText(name).AddValidate("Invalid name", utils.ValidateName)
	form.AddInputField("Date", LabelWidth, FieldWidth).SetText(date).AddValidate("Invalid date (YYYYMMDD)", utils.ValidateDate)
	form.AddInputField("Time", LabelWidth, FieldWidth).SetText(time).AddValidate("Invalid time (HH:MM)", utils.ValidateEventTime)
	form.AddInputField("Location", LabelWidth, FieldWidth).SetText(location)
	form.AddInputField("Duration (eg. 1.5)", LabelWidth, FieldWidth).SetText(duration).AddValidate("Invalid duration", utils.ValidateDuration)
	form.AddInputField("Frequency", LabelWidth, FieldWidth).SetText(frequency).AddValidate("Invalid frequency", utils.ValidateNumber)
	form.AddInputField("Occurence", LabelWidth, FieldWidth).SetText(occurence).AddValidate("Invalid occurence", utils.ValidateNumber)
	form.AddInputField("Color", LabelWidth, FieldWidth).SetText(color)
	form.AddInputField("Description", LabelWidth, FieldWidth).SetText(description)

	return form
}

// EditEventForm creates a form for editing existing events
func (epv *EventPopupView) EditEventForm(g *gocui.Gui, title, name, date, time, location, duration, description, color string) *component.Form {
	form := component.NewForm(g, title, epv.X, epv.Y, epv.W, epv.H)

	form.AddInputField("Name", LabelWidth, FieldWidth).SetText(name).AddValidate("Invalid name", utils.ValidateName)
	form.AddInputField("Date", LabelWidth, FieldWidth).SetText(date).AddValidate("Invalid date (YYYYMMDD)", utils.ValidateDate)
	form.AddInputField("Time", LabelWidth, FieldWidth).SetText(time).AddValidate("Invalid time (HH:MM)", utils.ValidateEventTime)
	form.AddInputField("Location", LabelWidth, FieldWidth).SetText(location)
	form.AddInputField("Duration", LabelWidth, FieldWidth).SetText(duration).AddValidate("Invalid duration", utils.ValidateDuration)
	form.AddInputField("Color", LabelWidth, FieldWidth).SetText(color)
	form.AddInputField("Description", LabelWidth, FieldWidth).SetText(description)

	return form
}

// GotoForm creates a form for navigating to a specific time on the same day
func (epv *EventPopupView) GotoForm(g *gocui.Gui, title string) *component.Form {
	form := component.NewForm(g, title, epv.X, epv.Y, epv.W, epv.H)

	currentTime := epv.Calendar.CurrentDay.Date
	defaultHour := fmt.Sprintf("%02d", currentTime.Hour())
	
	form.AddInputField("Hour", LabelWidth, FieldWidth).SetText(defaultHour).AddValidate("Invalid hour (00-23)", utils.ValidateHourMinute)

	return form
}

// DateForm creates a form for navigating to a specific date
func (epv *EventPopupView) DateForm(g *gocui.Gui, title string) *component.Form {
	form := component.NewForm(g, title, epv.X, epv.Y, epv.W, epv.H)

	currentTime := epv.Calendar.CurrentDay.Date
	defaultDate := fmt.Sprintf("%04d%02d%02d", currentTime.Year(), currentTime.Month(), currentTime.Day())
	
	form.AddInputField("Date", LabelWidth, FieldWidth).SetText(defaultDate).AddValidate("Invalid date (YYYYMMDD)", utils.ValidateDate)

	return form
}

// ColorPickerForm creates a form for selecting event colors
func (epv *EventPopupView) ColorPickerForm(g *gocui.Gui, title string) *component.Form {
	form := component.NewForm(g, title, epv.X, epv.Y, epv.W, epv.H)

	// Start with empty field so user can type single letters directly
	form.AddInputField("Color", LabelWidth, FieldWidth).SetText("")

	return form
}

// DurationForm creates a form for changing an event's duration
func (epv *EventPopupView) DurationForm(g *gocui.Gui, title string) *component.Form {
	form := component.NewForm(g, title, epv.X, epv.Y, epv.W, epv.H)

	form.AddInputField("Duration", LabelWidth, FieldWidth).SetText("").AddValidate("Invalid duration", utils.ValidateDuration)

	return form
}

// SearchForm creates a form for searching events with optional date filters
func (epv *EventPopupView) SearchForm(g *gocui.Gui, title string) *component.Form {
	form := component.NewForm(g, title, epv.X, epv.Y, epv.W, epv.H)
	
	form.AddInputField("Query", LabelWidth, FieldWidth).SetText("")
	form.AddInputField("From Date", LabelWidth, FieldWidth).SetText("").AddValidate("Invalid date (YYYYMMDD, 't' for today, or empty)", utils.ValidateOptionalDate)
	form.AddInputField("To Date", LabelWidth, FieldWidth).SetText("").AddValidate("Invalid date (YYYYMMDD, 't' for today, or empty)", utils.ValidateOptionalDate)
	
	return form
}

// positionCursorsAtEnd positions cursors at the end of field text
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