# Slack Bot

This is a simple Slack bot.

- [Slack Bot](#slack-bot)
  - [Features](#features)
  - [Technologies](#technologies)
  - [Setup](#setup)
    - [Usage](#usage)

## Features

The bot can be used to automate tasks in a Slack workspace. It can be used to:

- calculate the age of a person based on their year of birth [🎥](https://www.youtube.com/watch?v=jFfo23yIWac&t=9057s),
- answer (philosophical) questions with the help of the [Wolfram Alpha](https://www.wolframalpha.com/) API and the [wit.ai](https://wit.ai/) API [🎥](https://www.youtube.com/watch?v=jFfo23yIWac&t=26935s).

## Technologies

The application is built using the following technologies, libraries, frameworks, and tools:

- [Go](https://golang.org/),
- [godotenv](https://github.com/joho/godotenv),
- [gjson](https://github.com/tidwall/gjson),
- [Slack API in Go](https://github.com/slack-go/slack),
- [slacker](https://github.com/slack-io/slacker),
- [go-wolfram](https://github.com/krognol/go-wolfram),
- [wit-go](https://github.com/wit-ai/wit-go).

## Setup

Before running the application, you need to create a `.env` file in the root directory of the project with the following content:

```env
SLACK_BOT_TOKEN=PUT_YOUR_SLACK_BOT_TOKEN_HERE
SLACK_APP_TOKEN=PUT_YOUR_SLACK_APP_TOKEN_HERE
WOLFRAM_APP_ID=PUT_YOUR_WOLFRAM_APP_ID_HERE
WIT_AI_TOKEN=PUT_YOUR_WIT_AI_TOKEN_HERE
```

or set the environment variables directly in your environment.

To know how to get the `SLACK_BOT_TOKEN` and `SLACK_APP_TOKEN`, please refer to the [official documentation](https://api.slack.com/start/quickstart).

To get the `WOLFRAM_APP_ID`, please refer to the [official documentation](https://products.wolframalpha.com/api/).

To get the `WIT_AI_TOKEN`, please refer to the [official documentation](https://wit.ai/docs/http/20200513).

To run this application:

```bash
go build . && ./01_slack_bot
```

If you want to add some extra dependencies to the project, you might need to run the following command (as we are using Go modules and vendoring) afterwards:

```bash
go mod tidy && go mod vendor
```

Provided that you've configured Slack properly, you should be able to interact with the bot in your Slack workspace.

In my case, I've created a bot called `simple-slack-bot` and I can interact with it by mentioning it in a message.

### Usage

To calculate the age of a person based on their year of birth, you can use the following command:

```text
@simple-slack-bot my yob is 2000
```

The bot will respond with the age of the person.

```text
You are 24 years old
```

To ask a (philosophical) question, you can use the following command:

```text
@simple-slack-bot answer question: "What is the meaning of life?"
```

The bot will respond with an answer to the question (if it can find one):

```text
42
```
