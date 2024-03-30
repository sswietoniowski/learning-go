package main

import (
	"errors"
	"os"
)

type discordBotConfig struct {
	Token string
}

const (
	errMissingDiscordBotToken = "missing Discord bot token"
)

func newDiscordBotConfig() (*discordBotConfig, error) {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		return nil, errors.New(errMissingDiscordBotToken)
	}

	return &discordBotConfig{Token: token}, nil
}
