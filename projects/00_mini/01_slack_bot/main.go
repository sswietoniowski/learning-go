package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack/socketmode"
	"github.com/slack-io/slacker"
)

func printCommandEvents(eventsCh <-chan socketmode.Event) {
	for event := range eventsCh {
		fmt.Printf("Event received: %v\n", event)
	}
}

func calculateAge() *slacker.CommandDefinition {
	handler := func(ctx *slacker.CommandContext) {
		year := ctx.Request().Param("year")
		yob, err := strconv.Atoi(year)
		if err != nil {
			ctx.Response().Reply("Invalid year of birth!")
			log.Println(err)
			return
		}

		age := time.Now().Year() - yob

		message := fmt.Sprintf("You are %d years old", age)
		ctx.Response().Reply(message)
	}

	return &slacker.CommandDefinition{
		Command:     "my yob is <year>",
		Description: "yob calculator",
		Handler:     handler,
	}
}

func main() {
	godotenv.Load()

	slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
	slackAppToken := os.Getenv("SLACK_APP_TOKEN")

	bot := slacker.NewClient(slackBotToken, slackAppToken)

	go printCommandEvents(bot.SocketModeClient().Events)

	bot.AddCommand(calculateAge())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := bot.Listen(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
