package config

import (
	"github.com/tkcrm/mx/logger"
	"github.com/tkcrm/mx/ops"
)

type Config struct {
	Log      logger.Config
	Ops      ops.Config
	DataDir  string `yaml:"data_dir" validate:"required" default:"./data"`
	Telegram TelegramConfig
	Groq     GroqConfig
}

type TelegramConfig struct {
	BotToken string `yaml:"bot_token" validate:"required"`
}

type GroqConfig struct {
	APIKey string `yaml:"api_key" validate:"required"`
}
