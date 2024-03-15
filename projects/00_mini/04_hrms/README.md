# HRMS (Human Resource Management System)

This is a simple HR management system that can be used to manage employees' data.

- [HRMS (Human Resource Management System)](#hrms-human-resource-management-system)
  - [Features](#features)
  - [Technologies](#technologies)
  - [Setup](#setup)

## Features

This application has the following features:

- basic CRUD operations for employees using document-based database (MongoDB) and RESTful API.

## Technologies

The application is built using the following technologies, libraries, frameworks, and tools:

- [Go](https://golang.org/),
- [godotenv](https://github.com/joho/godotenv),
- [fiber](https://github.com/gofiber/fiber),
- [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver).

## Setup

To run this application, you might install Docker on your machine or have the MongoDB database already installed.

To start the MongoDB database as a Docker container, run the following command:

```bash
docker run --name hrms -d -p 27018:27017 mongo
```

Before running the application, you need to create a `.env` file in the root directory of the project with the following content:

```env
PORT=PUT_REAL_PORT_NUMBER_HERE
MONGO_URI=PUT_REAL_MONGO_URI_HERE
MONGO_DB=PUT_REAL_MONGO_DB_NAME_HERE
```

or set the environment variables directly in your environment.

Sample `.env` file:

```env
PORT=8080
MONGO_URI=mongodb://localhost:27018/?compressors=snappy,zlib,zstd
MONGO_DB=hrms
```

To run this application, run the following command in the terminal:

```bash
go build . && ./04_hrms
```

If you want to add some extra dependencies to the project, you might need to run the following command (as we are using Go modules and vendoring) afterwards:

```bash
go mod tidy && go mod vendor
```
