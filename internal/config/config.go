package config

import "fmt"

type ProjectConfig struct {
	NotionAPIToken   string `json:"notion_api_token"`
	NotionDatabaseID string `json:"notion_database_id"`
	DiscordWebhook   string `json:"discord_webhook"`
}

type Config struct {
	Projects map[string]ProjectConfig
}

func (c *ProjectConfig) Validate() error {
	if c.NotionAPIToken == "" {
		return fmt.Errorf("NotionAPITokenが設定されていません")
	}
	if c.NotionDatabaseID == "" {
		return fmt.Errorf("NotionDatabaseIDが設定されていません")
	}
	if c.DiscordWebhook == "" {
		return fmt.Errorf("DiscordWebhookが設定されていません")
	}
	return nil
}

func NewConfig(projectsData map[string]any) (*Config, error) {
	projects := make(map[string]ProjectConfig)

	for projectName, item := range projectsData {
		params, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("プロジェクト %s の設定形式が不正です", projectName)
		}

		projectConfig := ProjectConfig{
			NotionAPIToken:   getStringValue(params, "notion_api_token"),
			NotionDatabaseID: getStringValue(params, "notion_database_id"),
			DiscordWebhook:   getStringValue(params, "discord_webhook"),
		}

		if err := projectConfig.Validate(); err != nil {
			return nil, fmt.Errorf("プロジェクト %s の設定が不正です: %w", projectName, err)
		}

		projects[projectName] = projectConfig
	}

	return &Config{Projects: projects}, nil
}

func getStringValue(params map[string]any, key string) string {
	if value, ok := params[key].(string); ok {
		return value
	}
	return ""
}