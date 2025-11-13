package config

import (
	"github.com/tkcrm/mx/logger"
	"github.com/tkcrm/mx/ops"
)

type Config struct {
	Log              logger.Config
	Ops              ops.Config
	DataDir          string `yaml:"data_dir" validate:"required" default:"./data"`
	TelegramBotToken string `yaml:"telegram_bot_token" validate:"required"`
}
