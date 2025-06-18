# SchedulerDiscordBot

NotionカレンダーとDiscordを連携させた、スケジュール通知ボットです。Notionデータベースから翌日の予定を取得し、Discord WebhookでEmbedメッセージとして通知します。

## 機能

- **Notion API連携**: 指定されたNotionデータベースから予定情報を取得
- **スケジュールフィルタリング**: 翌日の予定のみを抽出
- **Discord通知**: Webhook経由でリッチなEmbedメッセージを送信
- **マルチプロジェクト対応**: 複数のプロジェクト・チームを同時管理
- **自動実行**: GitHub Actionsによる定期実行（毎日15:00 UTC）

## アーキテクチャ

### システム構成図
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   GitHub        │    │   Notion        │    │   Discord       │
│   Actions       │───▶│   Database      │───▶│   Webhook       │
│   (Scheduler)   │    │   (Calendar)    │    │   (Notification)│
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### コード構成（リファクタリング後）
```
SchedulerDiscordBot/
├── main.go                    # エントリーポイント（134行）
├── notion/                    # Notion API連携パッケージ
│   ├── client.go             # Notion APIクライアント
│   ├── parser.go             # データ解析処理
│   └── types.go              # 型定義・インターフェース
├── discord/                   # Discord通知パッケージ
│   ├── webhook.go            # Webhook送信処理
│   ├── embed.go              # Embed構築処理
│   └── types.go              # 型定義・インターフェース
├── scheduler/                 # スケジュール処理パッケージ
│   ├── filter.go             # 日付フィルタリング
│   ├── time.go               # 時刻処理
│   └── types.go              # 型定義・インターフェース
├── internal/config/          # 設定管理パッケージ
│   └── config.go             # 設定構造体・バリデーション
├── utils/                    # ユーティリティパッケージ
│   └── env.go                # 環境変数ローダー
├── env.json                  # プロジェクト設定ファイル
├── go.mod                    # Go モジュール設定
├── go.sum                    # 依存関係チェックサム
└── .github/workflows/
    └── main.yml              # GitHub Actions設定
```

### パッケージ設計原則
- **単一責任原則**: 各パッケージは特定の機能に特化
- **依存関係注入**: インターフェースによる疎結合設計
- **エラーハンドリング**: 構造化ログによる詳細なエラー追跡
- **テスタビリティ**: モックしやすいインターフェース設計

## セットアップ

### 前提条件

- Go 1.24.4以上
- Notion APIアクセス権限
- Discord Webhook URL

### インストール

1. リポジトリをクローン:
```bash
git clone https://github.com/RyuichiroYoshida/SchedulerDiscordBot.git
cd SchedulerDiscordBot
```

2. 依存関係をインストール:
```bash
go mod tidy
```

3. 設定ファイルを作成:
```bash
cp env.json.example env.json
```

### 設定

`env.json`に各プロジェクトの設定を追加:

```json
{
  "project_name": {
    "notion_api_token": "ntn_xxxxxxxxxxxxxxx",
    "notion_database_id": "xxxxxxxxxxxxxxx",
    "discord_webhook": "https://discord.com/api/webhooks/xxxxx/xxxxx"
  }
}
```

#### Notionデータベース要件

データベースには以下のプロパティが必要です:

| プロパティ名 | タイプ | 説明 |
|------------|--------|------|
| 名前 | タイトル | 予定のタイトル |
| 日付 | 日付 | 開始日時・終了日時 |
| 開催場所 | リッチテキスト | イベントの場所 |
| ロール | リッチテキスト | 通知対象のDiscordロール |
| 通知開始 | 日付 | 通知を開始する日時（オプション） |
| リマインド日時 | 日付 | 指定された日時に通知送信（オプション） |
| 通知したい日 | 日付 | 指定された日が当日になったら通知（オプション） |

**通知タイミングの条件**:
1. **翌日の予定**: 予定開始日時が翌日の場合に通知
2. **リマインド日時**: 指定された日時に達したら通知送信
3. **通知したい日**: 指定された日が当日になったら通知送信

**注意**: 複数の条件が同時に満たされた場合でも、通知は1回のみ送信されます。

## 使用方法

### ローカル実行

```bash
go run main.go
```

### テスト実行

```bash
go test ./...
```

### ビルド

```bash
go build -o scheduler-bot main.go
```

## GitHub Actions設定

`.github/workflows/main.yml`で定期実行を設定:

- **実行時間**: 毎日15:00 UTC（日本時間24:00）
- **実行環境**: セルフホストWindowsランナー
- **手動実行**: `workflow_dispatch`で任意実行可能

## API仕様

### パッケージ別API

#### Notion パッケージ
- **`NotionClient.GetCalendar(token, databaseID string) (map[string]any, error)`**
  - Notion APIからデータベース情報を取得
  - POSTリクエストでデータベースクエリを実行
  - エラーハンドリングとログ出力を含む

- **`NotionParser.Parse(data []any) ([]Event, error)`**
  - Notion APIレスポンスを解析
  - 必要なフィールド（タイトル、日付、場所、ロール）を抽出
  - 過去の予定を自動フィルタリング

#### Discord パッケージ
- **`WebhookSender.SendEmbed(webhookURL, title, start, end, location, role string) error`**
  - Discord WebhookでEmbedメッセージを送信
  - ロールメンション機能付き
  - エラーハンドリングとステータスコード確認

- **`ScheduleEmbedBuilder.BuildScheduleEmbed(title, start, end, location, role string) WebhookPayload`**
  - スケジュール通知用Embedを構築
  - カスタマイズ可能なEmbed色設定
  - フィールドの動的生成

#### Scheduler パッケージ
- **`ScheduleFilter.IsScheduleForTomorrow(date DateInfo) bool`**
  - 予定が翌日かどうかを判定
  - タイムゾーンを考慮した日付比較

- **`TimeParser.ParseTimeStamp(date string) (string, error)`**
  - RFC3339形式の日付を人間可読形式に変換
  - エラーハンドリングとログ出力

#### Config パッケージ
- **`NewConfig(projectsData map[string]any) (*Config, error)`**
  - 環境変数からプロジェクト設定を構築
  - 設定値のバリデーション
  - 型安全な設定管理

### Discord Embed形式

```json
{
  "content": "@role_name",
  "embeds": [{
    "title": "スケジュール通知です!",
    "description": "明日のスケジュールをお知らせします。",
    "color": 2859167,
    "fields": [
      {"name": "タイトル", "value": "会議名"},
      {"name": "対象者", "value": "開発チーム"},
      {"name": "日時", "value": "開始2024-01-01 10:00:00 -> 終了2024-01-01 11:00:00"},
      {"name": "開催場所", "value": "会議室A"}
    ]
  }]
}
```

## ログ出力

構造化ログ（`log/slog`）を使用:

```go
slog.Info("プロジェクト情報", 
    slog.String("projectName", projectName),
    slog.String("notionToken", notionToken),
    slog.String("notionDatabaseId", notionDatabaseId))
```

## エラーハンドリング

- **API通信エラー**: リトライ機能なし、ログ出力のみ
- **データ形式エラー**: 不正データをスキップして処理継続
- **設定エラー**: アプリケーション終了

## セキュリティ

- APIトークンは`env.json`で管理（リポジトリには含まない）
- HTTPS通信のみ使用
- 入力データのバリデーション実装

## 制限事項

- 翌日の予定のみ通知（当日・他日程は対象外）
- Notion APIレート制限に依存
- Discord Webhook制限に準拠

## トラブルシューティング

### よくある問題

1. **Notion APIエラー**
   - APIトークンの有効性を確認
   - データベースIDが正しいか確認
   - データベースへのアクセス権限を確認

2. **Discord通知が届かない**
   - Webhook URLの有効性を確認
   - Discord側の権限設定を確認
   - ロール名が正しいか確認

3. **日付フィルタリングが動作しない**
   - タイムゾーン設定を確認
   - Notionの日付形式を確認

## 開発・貢献

### 開発環境

```bash
# アプリケーション実行
go run main.go

# 全パッケージのテスト実行
go test ./...

# テストカバレッジ確認
go test -cover ./...

# ベンチマークテスト
go test -bench=. ./...

# ビルド
go build -o scheduler-bot main.go

# 依存関係更新
go mod tidy
```

### コーディング規約

- **Go標準フォーマット**: `gofmt`による自動整形
- **構造化ログ**: `log/slog`パッケージの使用
- **エラーハンドリング**: 適切なエラーラッピングと詳細ログ
- **インターフェース設計**: テスタビリティを重視した抽象化
- **パッケージ分離**: 単一責任原則に基づくモジュール化

### 拡張方法

#### 新しい通知チャンネルの追加
1. `discord/`パッケージを参考に新しいパッケージを作成
2. 共通のインターフェースを実装
3. `main.go`で依存関係注入

#### 新しいデータソースの追加
1. `notion/`パッケージを参考に新しいパッケージを作成
2. 共通のデータ形式（`Event`構造体）を返すよう実装
3. `main.go`で切り替え可能に設計

## ライセンス

このプロジェクトは個人利用・学習目的で作成されています。

## 更新履歴

- **v2.0.0**: リファクタリング（モジュール化）
  - コードを機能別パッケージに分離
  - インターフェースベースの設計に変更
  - テスタビリティとメンテナンス性を向上
  - main.goを309行から134行に削減

- **v1.0.0**: 初期リリース
  - Notion-Discord連携機能
  - GitHub Actions自動実行
  - マルチプロジェクト対応