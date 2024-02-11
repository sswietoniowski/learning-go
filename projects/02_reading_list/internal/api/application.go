package api

import (
	"log"
	"net/http"

	"github.com/sswietoniowski/learning-go/projects/02_reading_list/internal/data"
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
	// Create a new database based on the configuration.
	// If the database type is not supported, use the in-memory database.
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

func (app *Application) getHealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("get healthcheck")

	if !IsValidMethod(w, r, http.MethodGet) {
		app.logger.Println("method not allowed")
		SendMethodNotAllowed(w)
		return
	}

	data := map[string]string{
		"status":      "available",
		"environment": app.config.EnvironmentName,
		"version":     Version,
	}

	err := SendOk(w, data)
	if err != nil {
		app.logger.Printf("internal server error: %v\n", err)
	}
}
