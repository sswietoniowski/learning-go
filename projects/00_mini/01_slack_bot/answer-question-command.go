package main

import (
	"encoding/json"
	"log"

	"github.com/krognol/go-wolfram"
	"github.com/slack-io/slacker"
	"github.com/tidwall/gjson"
	witai "github.com/wit-ai/wit-go/v2"
)

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
