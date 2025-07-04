package calendar

import (
	"fmt"
	"hash/fnv"
	"strings"
	"time"

	"github.com/samuelstranges/chronos/internal/utils"
	"github.com/jroimartin/gocui"
)

type Event struct {
	Id           int
	Name         string
	Description  string
	Location     string
	Time         time.Time
	DurationHour float64
	FrequencyDay int
	Occurence    int
	Color        gocui.Attribute
}

func NewEvent(name, description, location string, time time.Time, duration float64, frequency, occurence int, color gocui.Attribute) *Event {
	return &Event{Name: name, Description: description, Location: location, Time: time, DurationHour: duration, FrequencyDay: frequency, Occurence: occurence, Color: color}
}

func NewEventWithAutoColor(name, description, location string, time time.Time, duration float64, frequency, occurence int) *Event {
	color := GenerateColorFromName(name)
	return NewEvent(name, description, location, time, duration, frequency, occurence, color)
}

func GenerateColorFromName(name string) gocui.Attribute {
	colors := GetAvailableColors()

	h := fnv.New32a()
	h.Write([]byte(name))
	return colors[h.Sum32()%uint32(len(colors))]
}

// Custom purple color - using 256-color palette index for a nice purple
// Color 93 is a bright/vibrant purple that stands out from event colors
var ColorCustomPurple = gocui.Attribute(93)

func GetAvailableColors() []gocui.Attribute {
	return []gocui.Attribute{
		gocui.ColorRed,
		gocui.ColorGreen,
		gocui.ColorYellow,
		gocui.ColorBlue,
		gocui.ColorMagenta,
		gocui.ColorCyan,
		gocui.ColorWhite,
	}
}

func GetColorNames() []string {
	return []string{
		"Red",
		"Green",
		"Yellow",
		"Blue",
		"Magenta",
		"Cyan",
		"White",
	}
}

func ColorNameToAttribute(colorName string) gocui.Attribute {
	colors := GetAvailableColors()
	names := GetColorNames()

	for i, name := range names {
		if name == colorName {
			return colors[i]
		}
	}
	return gocui.ColorDefault
}

func ColorAttributeToName(color gocui.Attribute) string {
	colors := GetAvailableColors()
	names := GetColorNames()

	for i, c := range colors {
		if c == color {
			return names[i]
		}
	}
	return "Default"
}

func (e *Event) FormatTimeAndName() string {
	return fmt.Sprintf("%s | %s", e.FormatDurationTime(), e.Name)
}

func (e *Event) FormatDurationTime() string {
	startTimeString := utils.FormatHourFromTime(e.Time)

	duration := time.Duration(e.DurationHour * float64(time.Hour))
	endTime := e.Time.Add(duration)
	endTimeString := utils.FormatHourFromTime(endTime)

	return fmt.Sprintf("%s-%s", startTimeString, endTimeString)
}

func (e *Event) FormatBody() string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("\n%s | %s\n", e.FormatDurationTime(), e.Location))
	sb.WriteString("\nDescription :\n")
	sb.WriteString("--------------\n")
	sb.WriteString(e.Description)

	return sb.String()
}

func (e Event) GetReccuringEvents() []Event {

	var events []Event
	f := e.FrequencyDay
	initTime := e.Time

	for i := range e.Occurence {
		e.Time = initTime.AddDate(0, 0, i*f)
		events = append(events, e)
	}

	return events
}
