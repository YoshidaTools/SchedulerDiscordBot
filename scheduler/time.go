package scheduler

import (
	"fmt"
	"log/slog"
	"time"
)

type TimeParser struct{}

func NewTimeParser() *TimeParser {
	return &TimeParser{}
}

func (p *TimeParser) ParseTimeStamp(date string) (string, error) {
	if date == "" {
		return "", fmt.Errorf("日付が空です")
	}

	t, err := time.Parse(time.RFC3339Nano, date)
	if err != nil {
		slog.Error("日付のパースに失敗しました", slog.Any("error", err), slog.String("date", date))
		return "", err
	}

	return t.Format(time.DateTime), nil
}