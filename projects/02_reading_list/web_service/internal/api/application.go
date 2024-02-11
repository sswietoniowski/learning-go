package api

import (
	"log"

	"github.com/sswietoniowski/learning-go/projects/02_reading_list/web_service/internal/data"
)

// Application is the main application struct.
type Application struct {
	config   Config
	logger   *log.Logger
	database data.Databaser
}

// NewApplication creates a new Application instance with the given configuration and logger.
func NewApplication(config Config, logger *log.Logger) *Application {
	var database data.Databaser
	if config.DatabaseType == "postgresql" {
		dbConfig := data.NewPostgreSQLConfig()
		database = data.NewPostgreSQLDatabase(dbConfig.Dsn(), logger)
	} else {
		database = data.NewInMemoryDatabase(logger)
	}

	return &Application{
		config:   config,
		logger:   logger,
		database: database,
	}
}
