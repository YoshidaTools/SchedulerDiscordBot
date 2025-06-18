package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type NotionClient struct {
	httpClient *http.Client
}

func NewClient() *NotionClient {
	return &NotionClient{
		httpClient: &http.Client{},
	}
}

func (c *NotionClient) GetCalendar(notionToken, databaseID string) (map[string]any, error) {
	if notionToken == "" || databaseID == "" {
		err := fmt.Errorf("NotionのAPIトークンまたはデータベースIDが設定されていません")
		slog.Error(err.Error())
		return nil, err
	}

	url := fmt.Sprintf("https://api.notion.com/v1/databases/%s/query", databaseID)

	reqBody := map[string]any{}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		slog.Error("リクエストボディのJSON変換に失敗しました", slog.Any("error", err))
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		slog.Error("リクエストの作成に失敗しました", slog.Any("error", err))
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+notionToken)
	req.Header.Set("Notion-Version", "2022-06-28")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error("リクエストの送信に失敗しました", slog.Any("error", err))
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("レスポンスの読み取りに失敗しました", slog.Any("error", err))
		return nil, err
	}

	var responseData map[string]any
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		slog.Error("レスポンスのJSONパースに失敗しました", slog.Any("error", err))
		return nil, err
	}

	return responseData, nil
}
