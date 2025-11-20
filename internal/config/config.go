package config

import (
	"github.com/tkcrm/mx/logger"
	"github.com/tkcrm/mx/ops"
)

type Config struct {
	Log      logger.Config
	Ops      ops.Config
	DataDir  string `yaml:"data_dir" validate:"required" default:"./data"`
	Server   ServerConfig
	Telegram TelegramConfig
	MCP      MCPConfig
	Timezone string `yaml:"timezone" validate:"required" default:"UTC"`
}

type ServerConfig struct {
	IsEnabled bool   `yaml:"is_enabled" default:"false"`
	Addr      string `yaml:"addr" validate:"required" default:":8080"`
}

type TelegramConfig struct {
	BotToken string `yaml:"bot_token" validate:"required"`
}

type MCPConfig struct {
	Groq GroqConfig
}

type GroqConfig struct {
	Model  string `yaml:"model" validate:"required" default:"meta-llama/llama-4-scout-17b-16e-instruct"`
	APIKey string `yaml:"api_key" validate:"required"`
}
