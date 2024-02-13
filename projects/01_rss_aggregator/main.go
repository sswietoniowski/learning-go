package main

import (
	"log"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	_, err := strconv.ParseInt(os.Getenv("PORT"), 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{}))
}
