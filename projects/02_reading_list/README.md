# Reading List

This is a sample application, that demonstrates how to use web services (REST API) and web applications (HTML, CSS, JS) in Go.

## Features

There is only one feature in this application, which is to manage a list of books you want to read.

## Technologies

The application is built using the following technologies, libraries, frameworks, and tools:

- [Go](https://golang.org/),
- [PostgreSQL](https://www.postgresql.org/),
- [Docker](https://www.docker.com/),
- [Docker Compose](https://docs.docker.com/compose/),
- [gorilla/mux](https://github.com/gorilla/mux),
- [godotenv](https://github.com/joho/godotenv),
- [pq](github.com/lib/pq).

## Setup

To run this application, you need to have Docker and Docker Compose installed on your machine.

Then, you can run the following command from the root directory of the project:

```bash
docker compose up
```

This command will start the web service, web application and the database.

The database is running on `localhost:5433`.

You can access it using the following credentials:

- username: `postgres`,
- password: `P@ssw0rd`,
- database: `readinglist`.

You can access the web service (REST API) at `http://localhost:4000`.

The web application is served by the web service.

You can access the application at `http://localhost:8080`.

To stop the application, you can run the following command:

```bash
docker compose down
```
