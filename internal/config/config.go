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
	Agent    AgentConfig `yaml:"agent"`
	STT      STTConfig  `yaml:"stt"`
	Timezone string    `yaml:"timezone" validate:"required" default:"UTC"`
}

type AuthConfig struct {
	AccessTokenSecretKey  string `json:"access_token_secret_key"`
	RefreshTokenSecretKey string `json:"refresh_token_secret_key"`
}

type ServerConfig struct {
	Addr           string     `yaml:"addr" validate:"required" default:":9000"`
	ReflectEnabled bool       `yaml:"reflect_enabled" default:"false"`
	Auth           AuthConfig `yaml:"auth"`
}

type TelegramConfig struct {
	BotToken    string `yaml:"bot_token" validate:"required"`
	BotUsername string `yaml:"bot_username"`
}

type AgentConfig struct {
	Groq GroqConfig `yaml:"groq"`
}

type STTConfig struct {
	Enabled     bool   `yaml:"enabled" default:"false"`
	ModelPath   string `yaml:"model_path"`
	MaxDuration int    `yaml:"max_duration" default:"30"`
}

type GroqConfig struct {
	Model  string `yaml:"model" validate:"required" default:"meta-llama/llama-4-scout-17b-16e-instruct"`
	APIKey string `yaml:"api_key" validate:"required"`
}
