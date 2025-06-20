package calendar

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/HubertBel/lazyorg/internal/utils"
)

type Day struct {
	Date   time.Time
	Events []*Event
}

func NewDay(date time.Time) *Day {
	return &Day{Date: date}
}

func (d *Day) FormatTitle() string {
	return d.Date.Weekday().String() + "-" + strconv.Itoa(d.Date.Day())
}

func (d *Day) FormatTimeAndHour() string {
	return fmt.Sprintf("%s %d | %s", d.Date.Month().String(), d.Date.Day(), utils.FormatHourFromTime(d.Date))
}

func (d *Day) FormatBody() string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("\nEvents :\n")
	sb.WriteString("---------\n")
	for _, v := range d.Events {
		s := "-> " + v.FormatTimeAndName() + "\n"
		sb.WriteString(s)
	}

	return sb.String()
}

func (d *Day) SortEventsByTime() {
	sort.Slice(d.Events, func(i, j int) bool {
		return d.Events[i].Time.Before(d.Events[j].Time)
	})
}
