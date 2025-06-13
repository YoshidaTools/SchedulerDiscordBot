package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/RyuichiroYoshida/SchedulerDiscordBot/utils"
)

func main() {
	// var chooseGuildId string
	// if len(os.Args) > 1 {
	// 	chooseGuildId = os.Args[1]
	// } else {
	// 	slog.Error("ギルドIDが指定されていません")
	// 	return
	// }
	// slog.Info("選択されたギルドID", slog.String("guildId", chooseGuildId))

	callback := make(chan map[string]any)
	go func() {
		data := GetNotionCalendar()
		callback <- data
	}()
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
		CreateDiscordEmbed(page)
	}

	// slog.Info("Notionカレンダーのデータ取得に成功", slog.Any("data", *data))
	// chooseGuildIdをここで利用可能
}

func GetNotionCalendar() map[string]any {
	loader := utils.DotenvLoader{}
	loader.LoadEnv(".env")
	notionToken := os.Getenv("NOTION_API_TOKEN")
	databaseID := os.Getenv("NOTION_DATABASE_ID")

	if notionToken == "" || databaseID == "" {
		slog.Error("NotionのAPIトークンまたはデータベースIDが設定されていません")
		return nil
	}

	url := fmt.Sprintf("https://api.notion.com/v1/databases/%s/query", databaseID)

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
		// 必要なプロパティを抽出
		date, ok := properties["日付"].(map[string]any)
		if !ok {
			slog.Error("日付の形式が不正です", slog.Any("properties", properties))
			continue
		}
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
		// text{}を抽出
		results = append(results, map[string]any{
			"title": textObj["content"],
			"date":  date,
		})
	}
	slog.Info("Notionカレンダーのデータをパースしました", slog.Any("results", results))
	return results
}

func parseTimeStamp(date string) string {
	// 日付をフォーマット
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		slog.Error("日付のパースに失敗しました", slog.Any("error", err))
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func CreateDiscordEmbed(date map[string]any) string {
	var start, end string
	if dateMap, ok := date["date"].(map[string]any); ok {
		if s, ok := dateMap["start"].(string); ok {
			start = parseTimeStamp(s)
		}
		if e, ok := dateMap["end"].(string); ok && e != "" {
			end = parseTimeStamp(e)
		}
	}
	description := fmt.Sprintf("日付: %s -> %s", start, end)

	embed := fmt.Sprintf(`"title": "%s", "description": "%s"`, date["title"].(string), description)
	return embed
}
