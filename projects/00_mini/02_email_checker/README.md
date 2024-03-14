# Email Checker

This is a simple email checker tool.

- [Email Checker](#email-checker)
  - [Features](#features)
  - [Technologies](#technologies)
  - [Setup](#setup)

## Features

This tool can be used to:

- check if an email address is valid or not.

## Technologies

The application is built using the following technologies, libraries, frameworks, and tools:

- [Go](https://golang.org/).

## Setup

To run this application:

```bash
go build . && ./02_email_checker
```

If you want to add some extra dependencies to the project, you might need to run the following command (as we are using Go modules and vendoring) afterwards:

```bash
go mod tidy && go mod vendor
```
