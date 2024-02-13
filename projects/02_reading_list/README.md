# Reading List

This sample application demonstrates how to use web services (REST API) and web applications (HTML, CSS, JS) in Go.

## Features

This application has only one feature: managing a list of books you want to read.

## Technologies

The application is built using the following technologies, libraries, frameworks, and tools:

- [Go](https://golang.org/),
- [PostgreSQL](https://www.postgresql.org/),
- [Docker](https://www.docker.com/),
- [Docker Compose](https://docs.docker.com/compose/),
- [gorilla/mux](https://github.com/gorilla/mux),
- [godotenv](https://github.com/joho/godotenv),
- [pq](github.com/lib/pq),
- [GORM](https://gorm.io/) [:file_folder:](gorm.io/gorm).
- [GORM MySQL Driver](https://gorm.io/docs/connecting_to_the_database.html#MySQL) [:file_folder:](gorm.io/driver/mysql).

## Setup

There are two ways to run this application:

- running the API and web application separately as standalone services,
- using Docker Compose (preferred).

## Standalone Setup

To run this application, you might install Docker on your machine or have the PostgreSQL database already installed.

To start the PostgreSQL database as a Docker container, run the following command:

```bash
docker run --name readinglist -e POSTGRES_PASSWORD=PUT_REAL_PASSWORD_HERE -e POSTGRES_DB=readinglist -p 5433:5432 -d postgres
```

To setup the database, run the following commands in the terminal to copy the `setup.sql` file to the container and execute it with `psql`:

```bash
docker cp ./scripts/setup.sql readinglist:/setup.sql
docker exec -it readinglist psql -U postgres -d readinglist -f /setup.sql
```

To tear down the database, run the following commands in the terminal to copy the `teardown.sql` file to the container and execute it with `psql`:

```bash
docker cp ./scripts/teardown.sql readinglist:/teardown.sql
docker exec -it readinglist psql -U postgres -d readinglist -f /teardown.sql
```

You need to edit the `.env` file (or use `.env.local`) and add an actual password for the PostgreSQL user.

To start the API, run the following command in the terminal:

```bash
go run ./cmd/api/ --port 4000 --env production --db postgresql --frontend http://localhost:8080
```

If you don't have the PostgreSQL database installed, you can use the in-memory database by running the following command:

```bash
go run ./cmd/api/ --port 4000 --env development --db in-memory --frontend http://localhost:8080
```

The API can use MySQL database and **GORM** (Golang ORM) library instead of PostgreSQL and `database/sql` package.

To do so, you must have the MySQL database installed and running on your machine. The easiest way to do so is to use Docker:

```bash
docker run --name readinglist -e MYSQL_ROOT_PASSWORD=PUT_REAL_PASSWORD_HERE -e MYSQL_DATABASE=readinglist -p 3307:3306 -d mysql
```

Then, you need to edit the `.env` file (or use `.env.local`) and add an actual database port, user, and password pointing to your MySQL database.

To start the API, run the following command in the terminal:

```bash
go run ./cmd/api/ --port 4000 --env mysql --db gorm-mysql --frontend http://localhost:8080
```

Regardless of our database of choice, our API will be accessible at `http://localhost:4000/api/v1/books`.

With the API running, we can finally start (in a separate terminal) our web application (HTML, CSS, JS):

```bash
go run ./cmd/web/ --port 8080 --env development --backend "http://localhost:4000/api/v1"
```

Our web application will be accessible at `http://localhost:8080`.

_Voila! Job done; of course, we could use Docker Compose and simplify the whole process ..._

## Docker Compose Setup

To run this application, you must install Docker and Docker Compose on your machine.

Then, you can run the following command from the root directory of the project:

```bash
docker compose up
```

This command will start the web service, web application, and the database.

The database is running on `localhost:5433`.

You can access it using the following credentials:

- username: `postgres`,
- password: `P@ssw0rd`,
- database: `readinglist`.

You can access the web service (REST API) at `http://localhost:4000`.

The web service serves the web application.

You can access the application (HTML, CSS, JS) at `http://localhost:8080`.

To stop the application, you can run the following command:

```bash
docker compose down
```

_A lot simpler, right :-)?_
