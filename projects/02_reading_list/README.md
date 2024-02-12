# Reading List

This sample application demonstrates how to use web services (REST API) and web applications (HTML, CSS, JS) in Go.

## Features

This application has only one feature: managing a list of books you want to read.

## Technologies

The application is built using the following technologies, libraries, frameworks, and tools:

- [Go](https://golang.org/),
- [PostgreSQL](https://www.postgresql.org/),
- [Docker](https://www.docker.com/),
- [gorilla/mux](https://github.com/gorilla/mux),
- [godotenv](https://github.com/joho/godotenv),
- [pq](github.com/lib/pq).

## Setup

To run this application, you must install Docker on your machine or have the PostgreSQL database already installed.

To start the PostgreSQL database as a Docker container, run the following command:

```bash
docker run --name readinglist -e POSTGRES_PASSWORD=PUT_REAL_PASSWORD_HERE -e POSTGRES_DB=readinglist -p 5433:5432 -d postgres
```

To setup the database, run the following commands in the terminal to copy the setup.sql file to the container and execute it with psql:

```bash
docker cp ./scripts/setup.sql readinglist:/setup.sql
docker exec -it readinglist psql -U postgres -d readinglist -f /setup.sql
```

To tear down the database, run the following commands in the terminal to copy the teardown.sql file to the container and execute it with psql:

```bash
docker cp ./scripts/teardown.sql readinglist:/teardown.sql
docker exec -it readinglist psql -U postgres -d readinglist -f /teardown.sql
```

You need to edit the .env file (or use .env.local) and add an actual password for the PostgreSQL user.

Having the database started, we can run our web service (REST API):

```bash
go run ./cmd/api/ --port 4000 --env development --db in-memory --frontend http://localhost:8080
```

Our API will be accessible at:

```text
http://localhost:4000/api/v1/books
```

With the API running, we can finally start (in a separate terminal) our web application (HTML, CSS, JS):

```bash
go run ./cmd/web/ --port 8080 --env development --backend "http://localhost:4000/api/v1"
```

Our web application will be accessible at:

```text
http://localhost:8080
```

Voila! Job done; of course, we could use Docker Compose and simplify the whole process ... do so if you like :-).
