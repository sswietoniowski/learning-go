# Go Backend Masterclass

A food delivery backend service built as part of the [Three Dots Labs](https://threedots.tech/) Go Backend Masterclass training (`backend-masterclass-beta`).

* [Go Backend Masterclass](#go-backend-masterclass)
    * [Features](#features)
    * [Technologies](#technologies)
    * [Setup](#setup)
        * [Running Tests](#running-tests)
        * [SQLC](#sqlc)
        * [oapi-codegen](#oapi-codegen)
    * [Further Improvements](#further-improvements)

## Features

The application is a food delivery platform (similar to Uber Eats) that supports:

* registering customers and couriers,
* onboarding restaurants with menus,
* creating price quotes and placing orders,
* tracking the full order lifecycle: restaurant confirmation → courier acceptance → pickup → delivery,
* browsing restaurants and menu items with full-text search and price/name filtering,
* inter-module communication between the `orders` and `delivery` modules,
* automatic delivery fee calculation based on order details.

## Technologies

The application is built using the following technologies, libraries, frameworks, and tools:

* [Go](https://golang.org/) 1.25,
* [PostgreSQL](https://www.postgresql.org/) 17.6,
* [Docker](https://www.docker.com/) / [Docker Compose](https://docs.docker.com/compose/),
* [Echo](https://echo.labstack.com/) — HTTP framework,
* [pgx](https://github.com/jackc/pgx) — PostgreSQL driver,
* [sqlc](https://sqlc.dev/) — SQL-to-Go code generator,
* [golang-migrate](https://github.com/golang-migrate/migrate) — database migrations,
* [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) — OpenAPI 3.0 code generator,
* [shopspring/decimal](https://github.com/shopspring/decimal) — precise decimal arithmetic for monetary values,
* [gofakeit](https://github.com/brianvoe/gofakeit) — fake data generation for tests,
* [Task](https://taskfile.dev/) — task runner (Taskfile),
* [golangci-lint](https://golangci-lint.run/) — linter,
* [gofumpt](https://github.com/mvdan/gofumpt) — stricter Go formatter.

## Setup

To run this application, you need [Docker](https://www.docker.com/) and [Task](https://taskfile.dev/) installed on your machine.

All commands should be run from the `project` directory:

```bash
cd project
```

To start all services (gateway, backend, PostgreSQL):

```bash
task up
```

To stop all services:

```bash
task down
```

To stop all services and remove volumes:

```bash
task down-volumes
```

The backend listens on port `8080`. The training gateway (which drives test scenarios) listens on port `8888`.

The following environment variables are used by the backend:

| Variable       | Description                     | Default (Docker Compose)                                      |
|----------------|---------------------------------|---------------------------------------------------------------|
| `POSTGRES_URL` | PostgreSQL connection string    | `postgres://user:password@postgres:5432/eats?sslmode=disable` |
| `GATEWAY_ADDR` | Address of the training gateway | `gateway:8080`                                                |

Database migrations run automatically on startup.

### Running Tests

```bash
# Unit tests
task test-unit

# Integration tests (require a running PostgreSQL)
task test-integration

# Component tests (require the full stack)
task test-component

# All tests
task test
```

### SQLC

This project uses [`sqlc`](https://sqlc.dev/) to generate type-safe Go code from SQL queries.

> sqlc reads your SQL schema and queries, then generates Go code with the correct types — the ergonomics of an ORM without giving up raw SQL.

The configuration lives at `project/backend/orders/adapters/db/sqlc.yaml`. Generated code is placed in the `dbmodels` package alongside the queries.

To regenerate after modifying SQL queries or schema:

```bash
task gen
```

### oapi-codegen

This project uses [`oapi-codegen`](https://github.com/oapi-codegen/oapi-codegen) to generate server stubs and HTTP client code from the OpenAPI 3.0 specification.

The spec lives at `project/backend/orders/api/http/openapi.yaml`. Generated server code and the HTTP client are updated by the same command:

```bash
task gen
```

## Further Improvements

The following are some ideas for further improvements:

* add authentication and authorization (JWT or OAuth2) instead of passing UUIDs in headers,
* add pagination to list endpoints (orders, restaurants, menu items),
* introduce an event-driven approach for order state transitions using an event bus,
* add metrics and distributed tracing (e.g. with OpenTelemetry),
* implement push notifications for order status updates,
* add a rate limiter to the HTTP layer,
* support soft-deletes and archiving for restaurants and menu items,
* add a backoffice API for operators to manage the platform,
* write end-to-end tests covering the complete order flow.
