package web

import (
	"log"
)

// Application is the main application struct.
type Application struct {
	config Config
	logger *log.Logger
}

// NewApplication creates a new Application instance with the given configuration and logger.
func NewApplication(config Config, logger *log.Logger) *Application {
	return &Application{
		config: config,
		logger: logger,
	}
}
