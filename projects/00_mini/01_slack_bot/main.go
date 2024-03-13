package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"github.com/slack-io/slacker"

	"github.com/krognol/go-wolfram"
	witai "github.com/wit-ai/wit-go/v2"
)

func printSlackCommandEvents(eventsCh <-chan socketmode.Event) {
	for event := range eventsCh {
		eventJSON, err := json.Marshal(event)
		if err != nil {
			fmt.Printf("Error marshalling event: %v\n", err)
		}
		fmt.Printf("Event received:\n%s\n", eventJSON)
	}
}

func main() {
	godotenv.Load()

	slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
	slackAppToken := os.Getenv("SLACK_APP_TOKEN")
	slackerBotClient := slacker.NewClient(slackBotToken, slackAppToken)
	slackBotClient := slack.New(slackBotToken)

	go printSlackCommandEvents(slackerBotClient.SocketModeClient().Events)

	witAiToken := os.Getenv("WIT_AI_TOKEN")
	witAiClient := witai.NewClient(witAiToken)

	wolframAppID := os.Getenv("WOLFRAM_APP_ID")
	wolframClient := &wolfram.Client{AppID: wolframAppID}

	slackerBotClient.AddCommand(calculateAge())
	slackerBotClient.AddCommand(uploadLoremIpsumFile(slackBotClient))
	slackerBotClient.AddCommand(answerQuestion(witAiClient, wolframClient))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := slackerBotClient.Listen(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
