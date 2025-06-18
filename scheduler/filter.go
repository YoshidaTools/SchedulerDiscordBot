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

// ShouldNotifyNow 翌日の予定の場合にtrueを返す
func (f *ScheduleFilter) ShouldNotifyNow(date notion.DateInfo) bool {
	return f.IsScheduleForTomorrow(date)
}

// ShouldNotifyByRemindDate リマインド日時に達している場合にtrueを返す
func (f *ScheduleFilter) ShouldNotifyByRemindDate(remindDate notion.RemindDate) bool {
	if remindDate.NotifyStartTime == "" {
		return false
	}

	remindTime, err := time.Parse(time.RFC3339, remindDate.NotifyStartTime)
	if err != nil {
		// パースエラーの場合は通知しない
		return false
	}

	now := time.Now()
	return !now.Before(remindTime)
}

// ShouldNotifyOnTargetDate 通知したい日が当日の場合にtrueを返す
func (f *ScheduleFilter) ShouldNotifyOnTargetDate(notificationDate notion.NotificationDate) bool {
	if notificationDate.TargetDate == "" {
		return false
	}

	// 日付のみの形式（YYYY-MM-DD）または日時形式（RFC3339）に対応
	var targetTime time.Time
	var err error

	// まずYYYY-MM-DD形式でパースを試行
	targetTime, err = time.Parse("2006-01-02", notificationDate.TargetDate)
	if err != nil {
		// RFC3339形式でパースを試行
		targetTime, err = time.Parse(time.RFC3339, notificationDate.TargetDate)
		if err != nil {
			// パースエラーの場合は通知しない
			return false
		}
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.AddDate(0, 0, 1)

	// targetTimeが今日の範囲内（今日の00:00:00から明日の00:00:00未満）かチェック
	return !targetTime.Before(today) && targetTime.Before(tomorrow)
}
