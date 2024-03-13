package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack/socketmode"
	"github.com/slack-io/slacker"
	"github.com/tidwall/gjson"

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
		ctx.Response().Reply(message, slacker.WithInThread(true))
	}

	return &slacker.CommandDefinition{
		Command:     "my yob is <year>",
		Description: "yob calculator",
		Handler:     handler,
	}
}

func answerQuestion(witAiClient *witai.Client, wolframClient *wolfram.Client) *slacker.CommandDefinition {
	handler := func(ctx *slacker.CommandContext) {
		question := ctx.Request().Param("question")

		const errorMessage = "I'm sorry, I didn't understand that."

		msg, err := witAiClient.Parse(&witai.MessageRequest{
			Query: question,
		})
		if err != nil {
			ctx.Response().Reply(errorMessage, slacker.WithInThread(true))
			log.Println(err)
			return
		}

		data, err := json.MarshalIndent(msg, "", "  ")
		if err != nil {
			ctx.Response().Reply(errorMessage, slacker.WithInThread(true))
			log.Println(err)
			return
		}
		log.Printf("Wit.ai response: %s\n", string(data))

		rough := string(data[:])
		value := gjson.Get(rough, "entities.wit$wolfram_search_query:wolfram_search_query.0.value")
		answer := value.String()
		log.Printf("Wit.ai response: %s\n", answer)

		timeout := 1000 // 1 second
		res, err := wolframClient.GetSpokentAnswerQuery(answer, wolfram.Metric, timeout)
		if err != nil {
			ctx.Response().Reply(errorMessage, slacker.WithInThread(true))
			log.Println(err)
			return
		}

		log.Printf("Wolfram response: %s\n", res)
		ctx.Response().Reply(res, slacker.WithInThread(true))
	}

	return &slacker.CommandDefinition{
		Command:     "answer question: <question>",
		Description: "answer a question",
		Handler:     handler,
	}
}

func main() {
	godotenv.Load()

	slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
	slackAppToken := os.Getenv("SLACK_APP_TOKEN")
	slackBotClient := slacker.NewClient(slackBotToken, slackAppToken)

	go printSlackCommandEvents(slackBotClient.SocketModeClient().Events)

	witAiToken := os.Getenv("WIT_AI_TOKEN")
	witAiClient := witai.NewClient(witAiToken)

	wolframAppID := os.Getenv("WOLFRAM_APP_ID")
	wolframClient := &wolfram.Client{AppID: wolframAppID}

	slackBotClient.AddCommand(calculateAge())
	slackBotClient.AddCommand(answerQuestion(witAiClient, wolframClient))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := slackBotClient.Listen(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
