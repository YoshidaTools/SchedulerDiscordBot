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

// IsNotificationTime 現在時刻が通知開始時刻に達しているかを判定
func (f *ScheduleFilter) IsNotificationTime(date notion.DateInfo) bool {
	// 通知開始時刻が設定されていない場合は常にtrue（従来の動作）
	if date.NotifyStartTime == "" {
		return true
	}

	notifyTime, err := time.Parse(time.RFC3339, date.NotifyStartTime)
	if err != nil {
		// パースエラーの場合は従来の動作（通知する）
		return true
	}

	now := time.Now()
	return !now.Before(notifyTime)
}

// ShouldNotifyNow 翌日の予定かつ通知開始時刻に達している場合にtrueを返す
func (f *ScheduleFilter) ShouldNotifyNow(date notion.DateInfo) bool {
	return f.IsScheduleForTomorrow(date) || f.IsNotificationTime(date)
}
