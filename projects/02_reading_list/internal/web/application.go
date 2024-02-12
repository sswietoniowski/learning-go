package web

import (
	"log"

	"github.com/sswietoniowski/learning-go/projects/02_reading_list/internal/web/service"
)

// Application is the main application struct.
type Application struct {
	config  Config
	logger  *log.Logger
	service *service.BooksService
}

// NewApplication creates a new Application instance with the given configuration and logger.
func NewApplication(config Config, logger *log.Logger) *Application {
	service := service.NewBookService(config.BackendEndpoint, logger)

	return &Application{
		config:  config,
		logger:  logger,
		service: service,
	}
}
