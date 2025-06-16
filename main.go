package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/RyuichiroYoshida/SchedulerDiscordBot/utils"
)

func main() {
	envData := utils.LoadEnv("env.json")

	for n, item := range envData {
		c, _ := item.(map[string]any)
		fmt.Println(n, c["discord_webhook"])
	}

	for projectName, item := range envData {
		params, ok := item.(map[string]any)
		if !ok {
			slog.Error("環境変数の形式が不正です", slog.Any("item", item))
		}

		notionToken := params["notion_api_token"].(string)
		notionDatabaseId := params["notion_database_id"].(string)
		slog.Info("プロジェクト情報", slog.String("projectName", projectName), slog.String("notionToken", notionToken), slog.String("notionDatabaseId", notionDatabaseId))

		callback := make(chan map[string]any)
		go func(token, id string) {
			data := GetNotionCalendar(token, id)
			callback <- data
		}(notionToken, notionDatabaseId)

		slog.Info("Notionカレンダーのデータを取得中...")
		data := <-callback
		if data == nil {
			slog.Error("Notionカレンダーのデータ取得に失敗しました")
			return
		}

		results, ok := data["results"].([]any)
		if !ok {
			slog.Error("resultsの型が不正です")
			return
		}
		parseData := notionParse(results)
		for _, page := range parseData {
			date, ok := page["date"].(map[string]any)
			if !ok {
				slog.Error("日付情報の形式が不正です", slog.Any("page", page))
				continue
			}

			if !isScheduleForTomorrow(date) {
				continue
			}

			start := parseTimeStamp(date["start"].(string))
			end := parseTimeStamp(date["end"].(string))
			if start == "" || end == "" {
				slog.Error("日付のパースに失敗しました", slog.Any("date", date))
				continue
			}

			err := SendDiscordEmbed(params["discord_webhook"].(string), page["title"].(string), start, end, page["location"].(string), page["role"].(string))
			if err != nil {
				slog.Error("Discord Webhookの送信に失敗しました", slog.Any("error", err))
				continue
			}
		}
	}

	// slog.Info("Notionカレンダーのデータ取得に成功", slog.Any("data", *data))
	// chooseGuildIdをここで利用可能
}

func GetNotionCalendar(notionToken, databaseId string) map[string]any {

	if notionToken == "" || databaseId == "" {
		slog.Error("NotionのAPIトークンまたはデータベースIDが設定されていません")
		return nil
	}

	url := fmt.Sprintf("https://api.notion.com/v1/databases/%s/query", databaseId)

	reqBody := map[string]any{}
	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		slog.Error("リクエストの作成に失敗しました", slog.Any("error", err))
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+notionToken)
	req.Header.Set("Notion-Version", "2022-06-28")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("リクエストの送信に失敗しました", slog.Any("error", err))
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("レスポンスの読み取りに失敗しました", slog.Any("error", err))
		return nil
	}

	var responseData map[string]any
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		slog.Error("レスポンスのJSONパースに失敗しました", slog.Any("error", err))
		return nil
	}
	// slog.Info("Notionカレンダーのデータ", slog.Any("data", responseData))
	return responseData
}

func notionParse(data []any) []map[string]any {
	results := make([]map[string]any, 0)
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

		// 日付情報を取得 date map[string]any
		dateAll, ok := properties["日付"].(map[string]any)
		if !ok {
			slog.Error("日付情報の形式が不正です", slog.Any("properties", properties))
			continue
		}
		date, ok := dateAll["date"].(map[string]any)
		if !ok {
			slog.Error("日付の形式が不正です", slog.Any("dateAll", dateAll))
			continue
		}
		// startが現在時刻より過去ならスキップ
		if s, okStart := date["start"].(string); okStart && s != "" {
			t, err := time.Parse(time.RFC3339, s)
			if err == nil && t.Before(time.Now()) {
				continue
			}
		}

		// 予定のタイトルtextを取得
		name, ok := properties["名前"].(map[string]any)
		if !ok {
			slog.Error("名前の形式が不正です", slog.Any("properties", properties))
			continue
		}
		titleValue, ok := name["title"].([]any)
		if !ok || len(titleValue) == 0 {
			slog.Error("タイトル配列の形式が不正です", slog.Any("name", name))
			continue
		}
		textObj, ok := titleValue[0].(map[string]any)["text"].(map[string]any)
		if !ok {
			slog.Error("textオブジェクトの形式が不正です", slog.Any("titleValue", titleValue[0]))
			continue
		}

		// 開催場所textを取得
		location, ok := properties["開催場所"].(map[string]any)
		if !ok {
			slog.Error("開催場所の形式が不正です", slog.Any("properties", properties))
			continue
		}
		locationValue, ok := location["rich_text"].([]any)
		if !ok || len(locationValue) == 0 {
			slog.Error("開催場所のリッチテキスト配列の形式が不正です", slog.Any("location", location))
			continue
		}
		locationText, ok := locationValue[0].(map[string]any)["plain_text"].(string)
		if !ok {
			slog.Error("開催場所のplain_textの形式が不正です", slog.Any("locationValue", locationValue[0]))
			continue
		}

		// 対象ロールtextを取得
		role, ok := properties["ロール"].(map[string]any)
		if !ok {
			slog.Error("対象ロールの形式が不正です", slog.Any("properties", properties))
			continue
		}
		roleValue, ok := role["rich_text"].([]any)
		if !ok || len(roleValue) == 0 {
			slog.Error("対象ロールのリッチテキスト配列の形式が不正です", slog.Any("role", role))
			continue
		}
		roleText, ok := roleValue[0].(map[string]any)["plain_text"].(string)
		if !ok {
			slog.Error("対象ロールのplain_textの形式が不正です", slog.Any("roleValue", roleValue[0]))
			continue
		}

		// 抽出したデータをresultsに追加
		results = append(results, map[string]any{
			"title":    textObj["content"],
			"role":     roleText,
			"date":     date,
			"location": locationText,
		})
	}
	slog.Info("Notionカレンダーのデータをパースしました", slog.Any("results", results))
	return results
}

func parseTimeStamp(date string) string {
	// 日付をフォーマット
	t, err := time.Parse(time.RFC3339Nano, date)
	if err != nil {
		slog.Error("日付のパースに失敗しました", slog.Any("error", err))
		return ""
	}
	return t.Format(time.DateTime)
}

func isScheduleForTomorrow(date map[string]any) bool {
	if s, ok := date["start"].(string); ok && s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return false
		}
		now := time.Now()
		tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		dayAfterTomorrow := tomorrow.AddDate(0, 0, 1)
		return !t.Before(tomorrow) && t.Before(dayAfterTomorrow)
	}
	return false
}

func SendDiscordEmbed(webhookURL, title, start, end, location, role string) error {
	if webhookURL == "" {
		slog.Error("DiscordのWebhook URLが設定されていません")
		return fmt.Errorf("DiscordのWebhook URLが設定されていません")
	}

	if end == "" {
		end = start // 終了時間がない場合は開始時間を使用
	}

	payload := map[string]any{
		"embeds": []map[string]any{
			{
				"title": "スケジュール通知です！",
				"description": fmt.Sprintf(`
				タイトル: %s
				ロール: %s
				日時: %s -> %s
				開催場所: %s`, title, role, start, end, location),
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Discord EmbedのJSON変換に失敗しました", slog.Any("error", err))
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		slog.Error("Discordへのリクエスト送信に失敗しました", slog.Any("error", err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		slog.Error("Discordへのリクエストが失敗しました", slog.Int("statusCode", resp.StatusCode), slog.String("body", string(body)))
		return fmt.Errorf("discordへのリクエストが失敗しました: %s", body)
	}

	slog.Info("Discord Embedを送信しました")
	return nil
}
