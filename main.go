package main

import (
	"fmt"
	"log/slog"

	"github.com/RyuichiroYoshida/SchedulerDiscordBot/discord"
	"github.com/RyuichiroYoshida/SchedulerDiscordBot/internal/config"
	"github.com/RyuichiroYoshida/SchedulerDiscordBot/notion"
	"github.com/RyuichiroYoshida/SchedulerDiscordBot/scheduler"
	"github.com/RyuichiroYoshida/SchedulerDiscordBot/utils"
)

func main() {
	// 環境変数を読み込み
	envData, err := utils.LoadEnv("env.json")
	if err != nil {
		slog.Error("環境変数の読み込みに失敗しました", slog.Any("error", err))
		return
	}

	// 設定を初期化
	cfg, err := config.NewConfig(envData)
	if err != nil {
		slog.Error("設定の初期化に失敗しました", slog.Any("error", err))
		return
	}

	// 各パッケージのインスタンスを作成
	notionClient := notion.NewClient()
	notionParser := notion.NewParser()
	scheduleFilter := scheduler.NewFilter()
	timeParser := scheduler.NewTimeParser()
	embedBuilder := discord.NewEmbedBuilder()
	discordSender := discord.NewWebhookSender(embedBuilder)

	// プロジェクトごとに処理を実行
	for projectName, projectConfig := range cfg.Projects {
		err := processProject(
			projectName,
			projectConfig,
			notionClient,
			notionParser,
			scheduleFilter,
			timeParser,
			discordSender,
		)
		if err != nil {
			slog.Error("プロジェクトの処理に失敗しました",
				slog.String("projectName", projectName),
				slog.Any("error", err))
			continue
		}
	}
}

func processProject(
	projectName string,
	cfg config.ProjectConfig,
	notionClient notion.Client,
	parser notion.Parser,
	filter scheduler.Filter,
	timeParser scheduler.TimeFormatter,
	discordSender discord.Sender,
) error {
	slog.Info("プロジェクト処理開始",
		slog.String("projectName", projectName),
		slog.String("notionDatabaseId", cfg.NotionDatabaseID))

	// Notionからデータを取得
	data, err := notionClient.GetCalendar(cfg.NotionAPIToken, cfg.NotionDatabaseID)
	if err != nil {
		return fmt.Errorf("Notionカレンダーの取得に失敗しました: %w", err)
	}

	results, ok := data["results"].([]any)
	if !ok {
		return fmt.Errorf("resultsの型が不正です")
	}

	// データを解析
	events, err := parser.Parse(results)
	if err != nil {
		return fmt.Errorf("データの解析に失敗しました: %w", err)
	}

	// 翌日の予定をフィルタリングしてDiscordに送信
	for _, event := range events {
		if !filter.IsScheduleForTomorrow(event.Date) {
			continue
		}

		start, err := timeParser.ParseTimeStamp(event.Date.Start)
		if err != nil {
			slog.Error("開始日時のパースに失敗しました",
				slog.Any("error", err),
				slog.String("date", event.Date.Start))
			continue
		}

		end, err := timeParser.ParseTimeStamp(event.Date.End)
		if err != nil {
			// 終了日時が空の場合はエラーではない
			if event.Date.End != "" {
				slog.Error("終了日時のパースに失敗しました",
					slog.Any("error", err),
					slog.String("date", event.Date.End))
			}
			end = ""
		}

		err = discordSender.SendEmbed(
			cfg.DiscordWebhook,
			event.Title,
			start,
			end,
			event.Location,
			event.Role,
		)
		if err != nil {
			slog.Error("Discord送信に失敗しました",
				slog.Any("error", err),
				slog.String("title", event.Title))
			continue
		}

		slog.Info("スケジュール通知を送信しました",
			slog.String("projectName", projectName),
			slog.String("title", event.Title))
	}

	return nil
}
