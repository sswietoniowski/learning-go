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

	"github.com/joho/godotenv"
	"github.com/sswietoniowski/learning-go/projects/02_reading_list/web_service/internal/data"
)

const version = "1.0.0"

type config struct {
	port     int
	env      string
	dbConfig data.DbConfig
}

type application struct {
	config   config
	logger   *log.Logger
	database data.Databaser
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "set port to run the server on")
	flag.StringVar(&cfg.env, "env", "development", "set environment for the server (development, staging, production)")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	loadEnv(logger)

	cfg.dbConfig = data.NewConfig()

	app := newApplication(cfg, logger)

	addr := fmt.Sprintf(":%d", cfg.port)

	logger.Printf("starting \"%s\" server on %s\n", cfg.env, addr)

	srv := &http.Server{
		Addr:         addr,
		Handler:      app.routes(),
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

func loadEnv(logger *log.Logger) error {
	if err := godotenv.Load(); err != nil {
		logger.Printf("could not load .env file: %v\n", err)
		return err
	}

	_ = godotenv.Load(".env.local")

	return nil
}
