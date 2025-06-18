package discord

import "fmt"

type ScheduleEmbedBuilder struct{}

func NewEmbedBuilder() *ScheduleEmbedBuilder {
	return &ScheduleEmbedBuilder{}
}

func (b *ScheduleEmbedBuilder) BuildScheduleEmbed(title, start, end, location, role string) WebhookPayload {
	if end == "" {
		end = "未定"
	}

	const color = 2859167 // DiscordのEmbedの色

	embed := Embed{
		Title:       "スケジュール通知です!",
		Description: "明日のスケジュールをお知らせします。\n",
		Color:       color,
		Fields: []Field{
			{
				Name:  "タイトル",
				Value: title,
			},
			{
				Name:  "対象者",
				Value: role,
			},
			{
				Name:  "日時",
				Value: fmt.Sprintf("開始%s -> 終了%s", start, end),
			},
			{
				Name:  "開催場所",
				Value: location,
			},
		},
	}

	return WebhookPayload{
		Content: fmt.Sprintf("@%s", role),
		Embeds:  []Embed{embed},
	}
}