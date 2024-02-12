/*
To start the application, run the following command in the terminal:

go run ./cmd/web/ --port 8080 --env development
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

	"github.com/sswietoniowski/learning-go/projects/02_reading_list/internal/web"
)

func main() {
	var config web.Config

	flag.IntVar(&config.ServerPort, "port", 8080, "set port to run the server on")
	flag.StringVar(&config.EnvironmentName, "env", "development", "set environment for the server (development, staging, production)")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	app := web.NewApplication(config, logger)

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
