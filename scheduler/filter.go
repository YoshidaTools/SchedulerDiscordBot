package scheduler

import (
	"time"

	"github.com/RyuichiroYoshida/SchedulerDiscordBot/notion"
)

type ScheduleFilter struct{}

func NewFilter() *ScheduleFilter {
	return &ScheduleFilter{}
}

func (f *ScheduleFilter) IsScheduleForTomorrow(date notion.DateInfo) bool {
	if date.Start == "" {
		return false
	}

	t, err := time.Parse(time.RFC3339, date.Start)
	if err != nil {
		return false
	}

	now := time.Now()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	dayAfterTomorrow := tomorrow.AddDate(0, 0, 1)

	return !t.Before(tomorrow) && t.Before(dayAfterTomorrow)
}