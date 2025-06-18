package notion

type Event struct {
	Title            string
	Role             string
	Date             DateInfo
	Location         string
	RemindDate       RemindDate
	NotificationDate NotificationDate
}

type DateInfo struct {
	Start string
	End   string
}

type RemindDate struct {
	NotifyStartTime string
}

type NotificationDate struct {
	TargetDate string
}

type NotionResponse struct {
	Results []map[string]any `json:"results"`
}

type Client interface {
	GetCalendar(token, databaseID string) (map[string]any, error)
}

type Parser interface {
	Parse(data []any) ([]Event, error)
}

type TimeValidator interface {
	IsScheduleForTomorrow(date DateInfo) bool
	ParseTimeStamp(date string) (string, error)
}
