package views

import (
	"strconv"
	"strings"
	"time"

	"github.com/HubertBel/lazyorg/internal/calendar"
	"github.com/HubertBel/lazyorg/internal/database"
	"github.com/HubertBel/lazyorg/internal/utils"
	component "github.com/j-04/gocui-component"
	"github.com/jroimartin/gocui"
)

type EventPopupView struct {
	*BaseView
	Form     *component.Form
	Calendar *calendar.Calendar
	Database *database.Database

	IsVisible bool
	SearchCallback func(query string) error
}

func NewEvenPopup(g *gocui.Gui, c *calendar.Calendar, db *database.Database) *EventPopupView {

	epv := &EventPopupView{
		BaseView:  NewBaseView("popup"),
		Form:      nil,
		Calendar:  c,
		Database:  db,
		IsVisible: false,
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

func (epv *EventPopupView) CreateEventFromInputs() *calendar.Event {
	for _, v := range epv.Form.GetInputs() {
		if !v.IsValid() {
			return nil
		}
	}

	name := epv.Form.GetFieldText("Name")
	time, _ := time.Parse(TimeFormat, epv.Form.GetFieldText("Time"))
	location := epv.Form.GetFieldText("Location")

	durationText := strings.TrimSpace(epv.Form.GetFieldText("Duration (eg. 1.5)"))
	duration := 1.0 // Default to 1 hour
	if durationText != "" {
		if parsedDuration, err := strconv.ParseFloat(durationText, 64); err == nil && parsedDuration > 0 {
			duration = parsedDuration
		}
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
	if newEvent = epv.CreateEventFromInputs(); newEvent == nil {
		return nil
	}
	events := newEvent.GetReccuringEvents()

	for _, v := range events {
		if _, err := epv.Database.AddEvent(v); err != nil {
			return err
		}
	}

	return epv.Close(g, v)
}

func (epv *EventPopupView) EditEvent(g *gocui.Gui, v *gocui.View, event *calendar.Event) error {
	if !epv.IsVisible {
		return nil
	}

	var newEvent *calendar.Event
	if newEvent = epv.CreateEventFromInputs(); newEvent == nil {
		return nil
	}
	newEvent.Id = event.Id

	if err := epv.Database.UpdateEventById(event.Id, newEvent); err != nil {
		return err
	}

	return epv.Close(g, v)
}

func (epv *EventPopupView) Close(g *gocui.Gui, v *gocui.View) error {
	epv.IsVisible = false
	return epv.Form.Close(g, v)
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

func (epv *EventPopupView) GotoTimeForm(g *gocui.Gui, title string) *component.Form {
	form := component.NewForm(g, title, epv.X, epv.Y, epv.W, epv.H)
	
	form.AddInputField("Hour (eg. 14)", LabelWidth, FieldWidth).SetText("").AddValidate("Invalid hour (00-23)", utils.ValidateHourMinute)

	return form
}

func (epv *EventPopupView) GotoDateForm(g *gocui.Gui, title string) *component.Form {
	form := component.NewForm(g, title, epv.X, epv.Y, epv.W, epv.H)
	
	form.AddInputField("Date (eg. 20250622)", LabelWidth, FieldWidth).SetText("").AddValidate("Invalid date (YYYYMMDD)", utils.ValidateDate)

	return form
}

func (epv *EventPopupView) ShowGotoTimePopup(g *gocui.Gui) error {
	if epv.IsVisible {
		return nil
	}

	epv.Form = epv.GotoTimeForm(g, "Goto Time")

	epv.addKeybind(gocui.KeyEsc, epv.Close)
	epv.addKeybind(gocui.KeyEnter, epv.GotoTime)

	epv.Form.AddButton("Goto", epv.GotoTime)
	epv.Form.AddButton("Cancel", epv.Close)

	epv.Form.SetCurrentItem(0)
	epv.IsVisible = true
	epv.Form.Draw()

	epv.positionCursorsAtEnd(g)

	return nil
}

func (epv *EventPopupView) ShowGotoDatePopup(g *gocui.Gui) error {
	if epv.IsVisible {
		return nil
	}

	epv.Form = epv.GotoDateForm(g, "Goto Date")

	epv.addKeybind(gocui.KeyEsc, epv.Close)
	epv.addKeybind(gocui.KeyEnter, epv.GotoDate)

	epv.Form.AddButton("Goto", epv.GotoDate)
	epv.Form.AddButton("Cancel", epv.Close)

	epv.Form.SetCurrentItem(0)
	epv.IsVisible = true
	epv.Form.Draw()

	epv.positionCursorsAtEnd(g)

	return nil
}

func (epv *EventPopupView) GotoTime(g *gocui.Gui, v *gocui.View) error {
	if !epv.IsVisible {
		return nil
	}

	for _, v := range epv.Form.GetInputs() {
		if !v.IsValid() {
			return nil
		}
	}

	hourStr := epv.Form.GetFieldText("Hour (eg. 14)")
	hour, _ := strconv.Atoi(hourStr)

	epv.Calendar.GotoTime(hour, 0)

	return epv.Close(g, v)
}

func (epv *EventPopupView) GotoDate(g *gocui.Gui, v *gocui.View) error {
	if !epv.IsVisible {
		return nil
	}

	for _, v := range epv.Form.GetInputs() {
		if !v.IsValid() {
			return nil
		}
	}

	dateStr := epv.Form.GetFieldText("Date (eg. 20250622)")
	year, _ := strconv.Atoi(dateStr[:4])
	month, _ := strconv.Atoi(dateStr[4:6])
	day, _ := strconv.Atoi(dateStr[6:8])

	epv.Calendar.GotoDate(year, month, day)

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
