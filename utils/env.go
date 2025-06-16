package utils

import (
	"encoding/json"
	"log/slog"
	"os"
)

func LoadEnv(filename string) (map[string]any, error) {
	var envJson map[string]any

	bytes, err := os.ReadFile(filename)
	if err != nil {
		slog.Error("環境変数ファイルの読み込みに失敗しました", slog.Any("error", err))
		return nil, err
	}

	err = json.Unmarshal(bytes, &envJson)
	if err != nil {
		slog.Error("JSONファイルの読み込みに失敗しました", slog.Any("error", err))
		return nil, err
	}
	return envJson, nil
}
