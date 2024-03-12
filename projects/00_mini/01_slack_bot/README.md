# Slack Bot

This is a simple Slack bot.

- [Slack Bot](#slack-bot)
  - [Features](#features)
  - [Technologies](#technologies)
  - [Setup](#setup)

## Features

The bot can be used to automate tasks in a Slack workspace. It can be used to:

- TODO: Add features

## Technologies

The application is built using the following technologies, libraries, frameworks, and tools:

- [Go](https://golang.org/),
- [godotenv](https://github.com/joho/godotenv),
- [slacker](https://github.com/slack-io/slacker).

## Setup

Before running the application, you need to create a `.env` file in the root directory of the project with the following content:

```env
SLACK_BOT_TOKEN=PUT_YOUR_SLACK_BOT_TOKEN_HERE
SLACK_APP_TOKEN=PUT_YOUR_SLACK_APP_TOKEN_HERE
```

To know how to get the `SLACK_BOT_TOKEN` and `SLACK_APP_TOKEN`, please refer to the [official documentation](https://api.slack.com/start/quickstart).

To run this application:

```bash
go build . && ./01_slack_bot
```

If you want to add some extra dependencies to the project, you might need to run the following command (as we are using Go modules and vendoring) afterwards:

```bash
go mod tidy && go mod vendor
```
