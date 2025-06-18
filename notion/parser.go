package notion

import (
	"fmt"
	"log/slog"
	"time"
)

type NotionParser struct{}

func NewParser() *NotionParser {
	return &NotionParser{}
}

func (p *NotionParser) Parse(data []any) ([]Event, error) {
	var events []Event

	for _, item := range data {
		page, ok := item.(map[string]any)
		if !ok {
			slog.Error("データの形式が不正です", slog.Any("item", item))
			continue
		}

		properties, ok := page["properties"].(map[string]any)
		if !ok {
			slog.Error("プロパティの形式が不正です", slog.Any("page", page))
			continue
		}

		event, err := p.parseEvent(properties)
		if err != nil {
			slog.Error("イベントの解析に失敗しました", slog.Any("error", err))
			continue
		}

		// 過去の予定をスキップ
		if p.isPastEvent(event.Date) {
			continue
		}

		events = append(events, event)
	}

	slog.Info("Notionカレンダーのデータをパースしました", slog.Int("eventCount", len(events)))
	return events, nil
}

func (p *NotionParser) parseEvent(properties map[string]any) (Event, error) {
	var event Event

	// 日付情報を取得
	dateInfo, err := p.parseDateInfo(properties)
	if err != nil {
		return event, err
	}
	event.Date = dateInfo

	// タイトルを取得
	title, err := p.parseTitle(properties)
	if err != nil {
		return event, err
	}
	event.Title = title

	// 開催場所を取得
	location, err := p.parseLocation(properties)
	if err != nil {
		return event, err
	}
	event.Location = location

	// 対象ロールを取得
	role, err := p.parseRole(properties)
	if err != nil {
		return event, err
	}
	event.Role = role

	// リマインド日時を取得
	remindDate := p.parseRemindDate(properties)
	event.RemindDate = remindDate

	return event, nil
}

func (p *NotionParser) parseDateInfo(properties map[string]any) (DateInfo, error) {
	dateAll, ok := properties["日付"].(map[string]any)
	if !ok {
		return DateInfo{}, fmt.Errorf("日付情報の形式が不正です")
	}

	date, ok := dateAll["date"].(map[string]any)
	if !ok {
		return DateInfo{}, fmt.Errorf("日付の形式が不正です")
	}

	start, _ := date["start"].(string)
	end, _ := date["end"].(string)

	return DateInfo{
		Start: start,
		End:   end,
	}, nil
}


func (p *NotionParser) parseRemindDate(properties map[string]any) RemindDate {
	remindAll, ok := properties["リマインド日時"].(map[string]any)
	if !ok {
		// リマインド日時プロパティが存在しない場合は空のRemindDateを返す
		return RemindDate{}
	}

	remindDate, ok := remindAll["date"].(map[string]any)
	if !ok {
		return RemindDate{}
	}

	notifyStartTime, _ := remindDate["start"].(string)
	return RemindDate{NotifyStartTime: notifyStartTime}
}

func (p *NotionParser) parseTitle(properties map[string]any) (string, error) {
	name, ok := properties["名前"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("名前の形式が不正です")
	}

	titleValue, ok := name["title"].([]any)
	if !ok || len(titleValue) == 0 {
		return "", fmt.Errorf("タイトル配列の形式が不正です")
	}

	textObj, ok := titleValue[0].(map[string]any)["text"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("textオブジェクトの形式が不正です")
	}

	title, ok := textObj["content"].(string)
	if !ok {
		return "", fmt.Errorf("titleコンテンツの形式が不正です")
	}

	return title, nil
}

func (p *NotionParser) parseLocation(properties map[string]any) (string, error) {
	location, ok := properties["開催場所"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("開催場所の形式が不正です")
	}

	locationValue, ok := location["rich_text"].([]any)
	if !ok || len(locationValue) == 0 {
		return "", fmt.Errorf("開催場所のリッチテキスト配列の形式が不正です")
	}

	locationText, ok := locationValue[0].(map[string]any)["plain_text"].(string)
	if !ok {
		return "", fmt.Errorf("開催場所のplain_textの形式が不正です")
	}

	return locationText, nil
}

func (p *NotionParser) parseRole(properties map[string]any) (string, error) {
	role, ok := properties["ロール"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("対象ロールの形式が不正です")
	}

	roleValue, ok := role["rich_text"].([]any)
	if !ok || len(roleValue) == 0 {
		return "", fmt.Errorf("対象ロールのリッチテキスト配列の形式が不正です")
	}

	roleText, ok := roleValue[0].(map[string]any)["plain_text"].(string)
	if !ok {
		return "", fmt.Errorf("対象ロールのplain_textの形式が不正です")
	}

	return roleText, nil
}

func (p *NotionParser) isPastEvent(date DateInfo) bool {
	if date.Start == "" {
		return false
	}

	t, err := time.Parse(time.RFC3339, date.Start)
	if err != nil {
		return false
	}

	return t.Before(time.Now())
}