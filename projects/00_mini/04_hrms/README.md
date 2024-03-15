# HRMS (Human Resource Management System)

This is a simple HR management system that can be used to manage employees' data.

- [HRMS (Human Resource Management System)](#hrms-human-resource-management-system)
  - [Features](#features)
  - [Technologies](#technologies)
  - [Setup](#setup)

## Features

This application has the following features:

- basic CRUD operations for employees using document-based database (MongoDB).

## Technologies

The application is built using the following technologies, libraries, frameworks, and tools:

- [Go](https://golang.org/),
- [godotenv](https://github.com/joho/godotenv),
- [fiber](https://github.com/gofiber/fiber),
- [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver).

## Setup

To run this application:

```bash
go build . && ./04_hrms
```

If you want to add some extra dependencies to the project, you might need to run the following command (as we are using Go modules and vendoring) afterwards:

```bash
go mod tidy && go mod vendor
```
