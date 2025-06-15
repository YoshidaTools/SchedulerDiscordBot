package utils

import (
	"encoding/json"
	"log/slog"
	"os"
)

func LoadEnv(filename string) map[string]any {
	var envJson map[string]any

	bytes, err := os.ReadFile(filename)
	if err != nil {
		slog.Error("環境変数ファイルの読み込みに失敗しました", slog.Any("error", err))
		return nil
	}

	err = json.Unmarshal(bytes, &envJson)
	if err != nil {
		slog.Error("環境変数の読み込みに失敗しました", slog.Any("error", err))
		return nil
	}
	slog.Info("環境変数の読み込みに成功", slog.Any("env", envJson))
	return envJson
}
