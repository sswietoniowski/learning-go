package main

import (
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-io/slacker"
)

var loremIpsum = "lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua"

func getLoremIpsumWords() []string {
	return strings.Split(loremIpsum, " ")
}

func generateLoremIpsum(n int) string {
	words := getLoremIpsumWords()
	var loremIpsumWords []string
	for i := 0; i < n; i++ {
		loremIpsumWords = append(loremIpsumWords, words[rand.Intn(len(words))])
	}
	return strings.Join(loremIpsumWords, " ")
}

func createLoremIpsumFile(wordCount int) (string, error) {
	fileName := "lorem_ipsum.txt"
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	defer file.Close()

	_, err = file.WriteString(generateLoremIpsum(wordCount))
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	return fileName, nil
}

func uploadLoremIpsumFile(slackBotClient *slack.Client) *slacker.CommandDefinition {
	handler := func(ctx *slacker.CommandContext) {
		wordCount, err := strconv.Atoi(ctx.Request().Param("word_count"))
		if err != nil {
			ctx.Response().Reply("Invalid word count!")
			log.Println(err)
			return
		}

		fileName, err := createLoremIpsumFile(wordCount)
		if err != nil {
			ctx.Response().Reply("Error creating file!")
			log.Println(err)
			return
		}

		file, err := os.Open(fileName)
		if err != nil {
			ctx.Response().Reply("Error opening file!")
			log.Println(err)
			return
		}
		defer file.Close()

		slackFile, err := slackBotClient.UploadFile(slack.FileUploadParameters{
			Reader:   file,
			Filename: "lorem_ipsum.txt",
			Filetype: "text",
			Title:    "lorem_ipsum.txt",
			Channels: []string{ctx.Event().Channel.ID},
		})
		if err != nil {
			ctx.Response().Reply("Error uploading file!")
			log.Println(err)
			return
		}
		log.Printf("File [%s] uploaded: %s\n", slackFile.ID, slackFile.URL)

		ctx.Response().Reply("File uploaded successfully!", slacker.WithInThread(true))

		err = os.Remove(fileName)
		if err != nil {
			log.Println(err)
		}
	}

	return &slacker.CommandDefinition{
		Command:     "upload lorem ipsum file with <word_count> words",
		Description: "upload a lorem ipsum file",
		Handler:     handler,
	}
}
