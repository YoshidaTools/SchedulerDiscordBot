# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

NotionカレンダーデータベースとDiscordを連携させたスケジュール通知ボットです。Goで実装されており、GitHub Actionsで毎日15:00 UTC（日本時間24:00）に自動実行され、翌日の予定をDiscord WebhookでEmbedメッセージとして通知します。v2.0.0でモジュール化リファクタリングを実施し、保守性とテスタビリティが大幅に向上しました。

## アーキテクチャ

### パッケージ構成（リファクタリング後）
- **main.go**: エントリーポイント（134行）- 依存関係注入とプロジェクト処理のオーケストレーション
- **notion/**: Notion API連携パッケージ
  - `client.go`: Notion APIクライアント実装
  - `parser.go`: APIレスポンスの解析とEvent構造体への変換
  - `types.go`: 型定義とインターフェース（Client, Parser）
- **discord/**: Discord通知パッケージ
  - `webhook.go`: Webhook送信処理とエラーハンドリング
  - `embed.go`: スケジュール通知用Embed構築
  - `types.go`: 型定義とインターフェース（Sender, EmbedBuilder）
- **scheduler/**: スケジュール処理パッケージ
  - `filter.go`: 翌日判定ロジック
  - `time.go`: 時刻フォーマット処理
  - `types.go`: 型定義とインターフェース（Filter, TimeFormatter）
- **internal/config/**: 設定管理パッケージ
  - `config.go`: プロジェクト設定の構造体とバリデーション
- **utils/env.go**: 環境変数ファイル（env.json）ローダー
- **env.json**: プロジェクト別設定ファイル

### 設計原則
- **インターフェースベース設計**: 各パッケージはインターフェースで抽象化されテスタブル
- **依存関係注入**: main.goで全依存関係を構成し、各関数に注入
- **単一責任原則**: 各パッケージは特定の機能に特化
- **構造化ログ**: log/slogによる詳細なエラー追跡

アプリケーションは複数プロジェクトを並行処理し、各プロジェクトが独自のNotionデータベースとDiscord Webhook設定を持ちます。

## 一般的なコマンド

### 開発環境
```bash
# アプリケーションをローカル実行
go run main.go

# 全パッケージのテスト実行
go test ./...

# 依存関係の更新
go mod tidy

# アプリケーションのビルド
go build -o scheduler-bot main.go
```

### テスト
```bash
# 特定の関数のテスト
go test -run TestFunctionName ./...

# 詳細出力付きテスト実行
go test -v ./...

# テストカバレッジ確認
go test -cover ./...

# ベンチマークテスト
go test -bench=. ./...
```

## 設定

アプリケーションは`env.json`から設定を読み込みます。設定構造は以下の通りです：
```json
{
  "project_name": {
    "notion_api_token": "ntn_...",
    "notion_database_id": "database_id",
    "discord_webhook": "https://discord.com/api/webhooks/..."
  }
}
```

### 設定管理の仕組み
- `utils.LoadEnv()`: env.jsonファイルを読み込み
- `config.NewConfig()`: プロジェクト設定を構造体に変換しバリデーション実行
- `config.ProjectConfig`: 型安全な設定管理
- エラー発生時は詳細なログ出力でデバッグ支援

## Notionデータベーススキーマ

ボットは以下のプロパティを持つNotionデータベースを想定しています：
- **日付** (Date): スケジュールの日時（開始・終了時刻）
- **名前** (Name/Title): イベントのタイトル
- **開催場所** (Location): イベントの開催場所
- **ロール** (Role): 通知対象のDiscordロール名
- **通知開始** (Date): 通知を開始する日時（オプション）
- **リマインド日時** (Date): 指定された日時に通知送信（オプション）

### データ処理フロー
1. `notion.Client.GetCalendar()`: Notion APIからデータベース全体を取得
2. `notion.Parser.Parse()`: APIレスポンスをEvent構造体に変換（通知開始時刻・リマインド日時を含む）
3. `scheduler.Filter.ShouldNotifyNow()`: 翌日の予定かつ通知開始時刻に達したものを抽出
4. `scheduler.Filter.ShouldNotifyByRemindDate()`: リマインド日時に達したものを抽出
5. `scheduler.TimeParser.ParseTimeStamp()`: 日時を人間可読形式に変換

### 通知タイミングの動作
#### リマインド日時
- **設定されている場合**: 指定された日時に通知送信（予定の日程に関係なく）
- **設定されていない場合**: リマインド通知は送信されない

#### 通知開始時刻
- **設定されている場合**: 翌日の予定で指定時刻に通知開始
- **設定されていない場合**: 従来通り翌日の予定として通知

#### エラーハンドリング
- **パースエラーの場合**: リマインド日時は通知しない（安全側）、通知開始時刻は通知する（安全側）

## 主要関数

### Notionパッケージ
- `notion.NewClient()`: NotionクライアントのファクトリFunction
- `notion.Client.GetCalendar(token, databaseID string)`: Notion APIからカレンダーデータを取得
- `notion.NewParser()`: Notionパーサーのファクトリ関数
- `notion.Parser.Parse(data []any)`: Notion APIレスポンスをEvent配列に変換

### Discordパッケージ
- `discord.NewWebhookSender(embedBuilder)`: Webhook送信者のファクトリ関数
- `discord.Sender.SendEmbed(webhookURL, title, start, end, location, role)`: Discord Embedメッセージ送信
- `discord.NewEmbedBuilder()`: Embed構築者のファクトリ関数
- `discord.EmbedBuilder.BuildScheduleEmbed(...)`: スケジュール通知用Embed構築

### Schedulerパッケージ
- `scheduler.NewFilter()`: スケジュールフィルターのファクトリ関数
- `scheduler.Filter.IsScheduleForTomorrow(date)`: 翌日の予定かどうかを判定
- `scheduler.Filter.IsNotificationTime(date)`: 通知開始時刻に達しているかを判定
- `scheduler.Filter.ShouldNotifyNow(date)`: 翌日の予定かつ通知開始時刻に達している場合にtrueを返す
- `scheduler.Filter.ShouldNotifyByRemindDate(remindDate)`: リマインド日時に達している場合にtrueを返す
- `scheduler.NewTimeParser()`: 時刻パーサーのファクトリ関数
- `scheduler.TimeFormatter.ParseTimeStamp(date)`: タイムスタンプを読みやすい形式に変換

### Configパッケージ
- `config.NewConfig(projectsData)`: 環境変数からプロジェクト設定を構築
- `config.ProjectConfig.Validate()`: 設定値のバリデーション実行

## デプロイメント

アプリケーションはGitHub ActionsでセルフホストされたWindowsランナー上で実行されます。ワークフローは以下の通りです：
1. リポジトリから最新コードをpull
2. `go mod tidy`で依存関係を更新
3. `go run main.go`でアプリケーションを実行

### GitHub Actions設定（.github/workflows/main.yml）
- **実行スケジュール**: 毎日15:00 UTC（日本時間24:00）
- **手動実行**: `workflow_dispatch`で任意のタイミングで実行可能
- **実行環境**: セルフホストWindowsランナー

## エラーハンドリング

アプリケーションは包括的なエラー追跡とデバッグのため、`log/slog`による構造化ログを使用しています。すべてのAPI障害とデータ解析エラーはコンテキスト付きでログ出力されます。

### ログ戦略
- **Info**: 正常な処理フロー（プロジェクト開始、データ解析完了、Discord送信成功）
- **Error**: エラー詳細（API障害、データ形式エラー、設定不備）
- **構造化フィールド**: projectName, error, statusCode等の詳細情報
- **継続処理**: 一つのプロジェクトでエラーが発生しても他のプロジェクトの処理は継続

### 典型的なエラーパターン
- Notion API認証エラー: トークンまたはデータベースIDが無効
- Discord Webhook送信エラー: URL無効またはサーバーエラー
- データ形式エラー: Notionデータベースのスキーマ不一致