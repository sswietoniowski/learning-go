package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
	slackAppToken := os.Getenv("SLACK_APP_TOKEN")

	fmt.Printf("Bot Token: %s\n", slackBotToken)
	fmt.Printf("App Token: %s\n", slackAppToken)
}
