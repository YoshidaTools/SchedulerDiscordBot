package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type WebhookSender struct {
	httpClient   *http.Client
	embedBuilder EmbedBuilder
}

func NewWebhookSender(embedBuilder EmbedBuilder) *WebhookSender {
	return &WebhookSender{
		httpClient:   &http.Client{},
		embedBuilder: embedBuilder,
	}
}

func (s *WebhookSender) SendEmbed(webhookURL, title, start, end, location, role string) error {
	if webhookURL == "" {
		err := fmt.Errorf("DiscordのWebhook URLが設定されていません")
		slog.Error(err.Error())
		return err
	}

	payload := s.embedBuilder.BuildScheduleEmbed(title, start, end, location, role)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Discord EmbedのJSON変換に失敗しました", slog.Any("error", err))
		return err
	}

	resp, err := s.httpClient.Post(webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		slog.Error("Discordへのリクエスト送信に失敗しました", slog.Any("error", err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf("discordへのリクエストが失敗しました: %s", body)
		slog.Error("Discordへのリクエストが失敗しました", 
			slog.Int("statusCode", resp.StatusCode), 
			slog.String("body", string(body)))
		return err
	}

	slog.Info("Discord Embedを送信しました")
	return nil
}