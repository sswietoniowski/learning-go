package main

import (
	"errors"
	"log"

	"github.com/bwmarrin/discordgo"
)

type DiscordBot struct {
	botID   string
	discord *discordgo.Session
}

const (
	errCreatingDiscordSession = "error creating Discord session"
	errRetrievingAccount      = "error retrieving account"
	errOpeningConnection      = "error opening connection"
)

func NewDiscordBot() (*DiscordBot, error) {
	config, err := newDiscordBotConfig()
	if err != nil {
		return nil, errors.New(errMissingDiscordBotToken)
	}

	bot := &DiscordBot{}

	discord, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		log.Printf("Error creating Discord session: %v\n", err)
		return nil, errors.New(errCreatingDiscordSession)
	}

	bot.discord = discord

	user, err := discord.User("@me")
	if err != nil {
		log.Printf("Error retrieving account: %v\n", err)
		return nil, errors.New(errRetrievingAccount)
	}

	bot.botID = user.ID

	bot.discord.AddHandler(messageHandler)

	return bot, nil
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "!ping" {
		_, err := s.ChannelMessageSend(m.ChannelID, "pong!")
		if err != nil {
			log.Printf("Error sending message: %v\n", err)
		}
	}
}

func (bot *DiscordBot) Run() error {
	err := bot.discord.Open()
	if err != nil {
		log.Printf("Error opening connection: %v\n", err)
		return errors.New(errOpeningConnection)
	}

	log.Printf("Bot is running with ID %s\n", bot.botID)

	return nil
}

func (bot *DiscordBot) Close() {
	bot.discord.Close()
}
