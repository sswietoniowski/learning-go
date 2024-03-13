package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/slack-io/slacker"
)

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
