package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("PORT is not found in the environment")
	}

	_, err := strconv.ParseInt(portString, 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Port: %s\n", portString)

	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{}))
}
