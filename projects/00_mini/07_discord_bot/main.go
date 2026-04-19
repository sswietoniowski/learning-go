package main

import (
	"log"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	bot, err := NewDiscordBot()
	if err != nil {
		log.Fatalf("Error creating Discord bot: %v\n", err)
	}
	defer bot.Close()

	err = bot.Run()
	if err != nil {
		log.Fatalf("Error opening connection: %v\n", err)
	}

	select {} // Block forever
}
