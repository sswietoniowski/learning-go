# RSS Aggregator

This is a simple [RSS](https://en.wikipedia.org/wiki/RSS) aggregator.

- [RSS Aggregator](#rss-aggregator)
  - [Features](#features)
  - [Technologies](#technologies)
  - [Setup](#setup)
    - [SQLC](#sqlc)
    - [Goose](#goose)
  - [Further Improvements](#further-improvements)

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
- [chi](https://github.com/go-chi/chi),
- [cors](https://github.com/go-chi/cors),
- [godotenv](https://github.com/joho/godotenv),
- [uuid](https://github.com/google/uuid),
- [pq](https://github.com/lib/pq),
- [sqlc](https://sqlc.dev/),
- [goose](https://pressly.github.io/goose/).

## Setup

To run this application, you might install Docker on your machine or have the PostgreSQL database already installed.

To start the PostgreSQL database as a Docker container, run the following command:

```bash
docker run --name rssaggregator -e POSTGRES_PASSWORD=PUT_REAL_PASSWORD_HERE -e POSTGRES_DB=rssaggregator -p 5433:5432 -d postgres
```

Before running the application, you need to create a `.env` file in the root directory of the project with the following content:

```env
PORT=8080
DATABASE_URL=postgres://user:password@host:port/database?sslmode=disable
```

or set the environment variables directly in your environment.

To run this application (at this stage), just run the following command:

```bash
go build . && ./01_rss_aggregator
```

If you want to add some extra dependencies to the project, you might need to run the following command (as we are using Go modules and vendoring) afterwards:

```bash
go mod tidy && go mod vendor
```

### SQLC

This project is also using [`sqlc`](https://github.com/sqlc-dev/sqlc) to generate Go code from SQL.

> SQLC is an _amazing_ Go program that generates Go code from SQL queries. It's not exactly an ORM, but rather a tool that makes working with raw SQL almost as easy as using an ORM.

To install `sqlc`, run the following command:

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

This project is using the following SQLC configuration file `sqlc.yaml`:

```yaml
version: '2'
sql:
  - schema: 'sql/schema'
    queries: 'sql/queries'
    engine: 'postgresql'
    gen:
      go:
        out: 'internal/database'
```

> As you can see, the SQLC configuration file is pointing to the `sql/schema` directory to find the SQL files and to the `sql/queries` directory to find the query files. The Go code generated by SQLC will be placed in the `internal/database` directory. It is also using the PostgreSQL database engine.

To generate the Go code from SQL, run the following command:

```bash
sqlc generate
```

### Goose

This project is also using [`goose`](https://github.com/pressly/goose) to manage the database migrations.

> Goose is a database migration tool written in Go. It runs migrations from the same SQL files that SQLC uses, making the pair of tools a perfect fit.

To install `goose`, run the following command:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

To run the migrations, run the following command:

```bash
goose -dir sql/schema postgres "postgres://postgres:PUT_REAL_PASSWORD_HERE@localhost:5433/rssaggregator" up
```

To revert the migrations, run the following command:

```bash
goose -dir sql/schema postgres "postgres://postgres:PUT_REAL_PASSWORD_HERE@localhost:5433/rssaggregator" down
```

## Further Improvements

The following are some ideas for further improvements:

- support pagination of the endpoints that can return many items,
- support different options for sorting and filtering posts using query parameters,
- classify different types of feeds and posts (e.g. blog, podcast, video, etc.),
- add a CLI client that uses the API to fetch and display posts, maybe it even allows you to read them in your terminal,
- scrape lists of feeds themselves from a third-party site that aggregates feed URLs,
- add support for other types of feeds (e.g. Atom, JSON, etc.),
- add integration tests that use the API to create, read, update, and delete feeds and posts,
- add bookmarking or "liking" to posts,
- create a simple web UI that uses your backend API.
