package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
}

type application struct {
	config config
	logger *log.Logger
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "set port to run the server on")
	flag.StringVar(&cfg.env, "env", "development", "set environment for the server (development, staging, production)")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	app := &application{
		config: cfg,
		logger: logger,
	}

	addr := fmt.Sprintf(":%d", cfg.port)

	logger.Printf("starting \"%s\" server on %s", cfg.env, addr)
	err := http.ListenAndServe(addr, app.routes())

	if err != nil {
		log.Fatal(err)
	}
}
