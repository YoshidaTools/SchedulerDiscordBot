package discord

type Embed struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Color       int     `json:"color"`
	Fields      []Field `json:"fields"`
}

type Field struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type WebhookPayload struct {
	Content string  `json:"content"`
	Embeds  []Embed `json:"embeds"`
}

type Sender interface {
	SendEmbed(webhookURL, title, start, end, location, role string) error
}

type EmbedBuilder interface {
	BuildScheduleEmbed(title, start, end, location, role string) WebhookPayload
}