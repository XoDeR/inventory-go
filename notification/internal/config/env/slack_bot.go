package env

import (
	"github.com/caarlos0/env/v11"
)

type slackBotEnvConfig struct {
	BotToken  string `env:"SLACK_BOT_TOKEN,required"`
	AppToken  string `env:"SLACK_APP_TOKEN,required"`
	ChannelID string `env:"SLACK_CHANNEL_ID,required"`
}

type slackBotConfig struct {
	raw slackBotEnvConfig
}

func NewSlackBotConfig() (*slackBotConfig, error) {
	var raw slackBotEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}

	return &slackBotConfig{raw: raw}, nil
}

// BotToken is the Bot User OAuth Token (xoxb-...), used to post messages.
func (cfg *slackBotConfig) BotToken() string {
	return cfg.raw.BotToken
}

// AppToken is the app-level token (xapp-...), used by Socket Mode.
func (cfg *slackBotConfig) AppToken() string {
	return cfg.raw.AppToken
}

// ChannelID is the ID (C...) of the channel notifications are posted to.
func (cfg *slackBotConfig) ChannelID() string {
	return cfg.raw.ChannelID
}
