package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v3"
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

func getMongoURI() (string, error) {
	dbURL := os.Getenv("MONGO_URI")
	if dbURL == "" {
		return "", fmt.Errorf("MONGO_URI is not found in the environment")
	}

	return dbURL, nil
}

func main() {
	app := fiber.New()

	mongoURI, err := getMongoURI()
	if err != nil {
		log.Fatalf("Failed to get mongo URI: %v", err)
	}

	addr, err := getServerAddress()
	if err != nil {
		log.Fatalf("Failed to get server address: %v", err)
	}

	app.Listen(addr)
}
