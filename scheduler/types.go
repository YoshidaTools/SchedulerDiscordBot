package scheduler

import "github.com/RyuichiroYoshida/SchedulerDiscordBot/notion"

type Filter interface {
	IsScheduleForTomorrow(date notion.DateInfo) bool
	ShouldNotifyNow(date notion.DateInfo) bool
	ShouldNotifyByRemindDate(remindDate notion.RemindDate) bool
}

type TimeFormatter interface {
	ParseTimeStamp(date string) (string, error)
}