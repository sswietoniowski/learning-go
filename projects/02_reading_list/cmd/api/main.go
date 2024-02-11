/*

To start the application, run the following command in the terminal:

go run ./cmd/api/ --port 4000 --env development --db in-memory

or

go run ./cmd/api/ --port 4000 --env production --db postgresql

*/

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/sswietoniowski/learning-go/projects/02_reading_list/internal/api"
	"github.com/sswietoniowski/learning-go/projects/02_reading_list/internal/utils"
)

func main() {
	var config api.Config

	// Part of the configuration is defined as flags, the rest (for the database) is loaded from the environment.
	// Note, that it is not a good practice to mix flags and environment variables in a single application,
	// but it is done here for the sake of the example (to show how to use both).
	flag.IntVar(&config.ServerPort, "port", 4000, "set port to run the server on")
	flag.StringVar(&config.EnvironmentName, "env", "development", "set environment for the server (development, staging, production)")
	flag.StringVar(&config.DatabaseType, "db", "in-memory", "set database to use (in-memory, postgresql)")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// Load the rest of the configuration from the environment.
	utils.DotEnvLoad(logger)

	app := api.NewApplication(config, logger)

	addr := fmt.Sprintf(":%d", config.ServerPort)

	logger.Printf("starting \"%s\" server on %s\n", config.EnvironmentName, addr)

	srv := &http.Server{
		Addr:         addr,
		Handler:      app.Routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt)

	var hasError bool

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Printf("could not listen on %s: %v\n", srv.Addr, err)
			hasError = true
			close(done)
		}
	}()

	<-quit
	logger.Printf("server is shutting down...\n")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Printf("could not gracefully shutdown the server: %v\n", err)
		hasError = true
	}

	close(done)
	<-done

	if hasError {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
