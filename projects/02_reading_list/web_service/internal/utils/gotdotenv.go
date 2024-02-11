package utils

import (
	"errors"
	"log"

	"github.com/joho/godotenv"
)

// DotEnvLoad loads the environment variables from the .env files, first from .env.local
// and if not exists from .env if it exists, otherwise returns an error.
func DotEnvLoad(logger *log.Logger) error {
	if err := godotenv.Load(".env.local"); err != nil {
		logger.Printf("could not load .env.local file: %v\n", err)
	} else {
		logger.Println("loaded .env.local file")
		return nil
	}

	if err := godotenv.Load(); err != nil {
		logger.Printf("could not load .env file: %v\n", err)
	} else {
		logger.Println("loaded .env file")
		return nil
	}

	return errors.New("could not load .env.local or .env file")
}
