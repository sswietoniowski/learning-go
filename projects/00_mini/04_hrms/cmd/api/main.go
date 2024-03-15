package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
	"github.com/sswietoniowski/learning-go/projects/00_mini/04_hrms/internal/api"
	"github.com/sswietoniowski/learning-go/projects/00_mini/04_hrms/internal/data"
)

func getServerAddress() (string, error) {
	portString := os.Getenv("PORT")
	if portString == "" {
		return "", fmt.Errorf("PORT is not found in the environment")
	}

	_, err := strconv.ParseInt(portString, 10, 64)
	if err != nil {
		return "", fmt.Errorf("PORT is not a valid number")
	}

	addr := fmt.Sprintf(":%s", portString)

	return addr, nil
}

type MongoParams struct {
	URI string
	DB  string
}

func getMongoParams() (*MongoParams, error) {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		return nil, fmt.Errorf("MONGO_URI is not found in the environment")
	}

	db := os.Getenv("MONGO_DB")
	if db == "" {
		return nil, fmt.Errorf("MONGO_DB is not found in the environment")
	}

	params := &MongoParams{
		URI: uri,
		DB:  db,
	}

	return params, nil
}

func main() {
	godotenv.Load()

	timeout := 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	mongoParams, err := getMongoParams()
	if err != nil {
		log.Fatalf("Failed to get mongo params: %v", err)
	}
	repository, err := data.NewEmployeesMongoDBRepository(ctx, mongoParams.URI, mongoParams.DB)
	if err != nil {
		log.Fatalf("Failed to create MongoDB repository: %v", err)
	}

	f := fiber.New()

	app := api.NewApplication(f, repository)
	defer app.Close()

	app.SetupRoutes()

	addr, err := getServerAddress()
	if err != nil {
		log.Fatalf("Failed to get server address: %v", err)
	}

	f.Listen(addr)
}
