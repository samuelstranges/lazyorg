package calendar

import "time"

type Week struct {
	StartDate time.Time
	EndDate   time.Time
	Days      []*Day
}

func NewWeek() *Week {
	return &Week{
		Days: []*Day{
			NewDay(time.Time{}),
			NewDay(time.Time{}),
			NewDay(time.Time{}),
			NewDay(time.Time{}),
			NewDay(time.Time{}),
			NewDay(time.Time{}),
			NewDay(time.Time{}),
		},
	}
}
