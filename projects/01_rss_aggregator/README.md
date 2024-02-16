# RSS Aggregator

This is a simple [RSS](https://en.wikipedia.org/wiki/RSS) aggregator.

## Features

It is a web server that allows clients to:

- add RSS feeds to be collected,
- follow and un-follow RSS feeds that other users have added,
- fetch all of the latest posts from the RSS feeds they follow.

It is based on [this](https://github.com/bootdotdev/fcc-learn-golang-assets/tree/main/project) project.

## Technologies

The application is built using the following technologies, libraries, frameworks, and tools:

- [Go](https://golang.org/),
- [PostgreSQL](https://www.postgresql.org/),
- [Docker](https://www.docker.com/),
- [Docker Compose](https://docs.docker.com/compose/),
- [chi](https://github.com/go-chi/chi),
- [cors](https://github.com/go-chi/cors),
- [godotenv](https://github.com/joho/godotenv).

## Setup

To run this application, you might install Docker on your machine or have the PostgreSQL database already installed.

To start the PostgreSQL database as a Docker container, run the following command:

```bash
docker run --name readinglist -e POSTGRES_PASSWORD=PUT_REAL_PASSWORD_HERE -e POSTGRES_DB=readinglist -p 5433:5432 -d postgres
```

To run this the application (at this stage), just run the following command:

```bash
go build . && ./01_rss_aggregator
```
