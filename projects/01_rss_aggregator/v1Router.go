package main

import (
	"github.com/go-chi/chi/v5"
)

func v1Router() chi.Router {
	r := chi.NewRouter()

	r.HandleFunc("/readiness", readinessHandler)
	r.HandleFunc("/err", errHandler)

	return r
}
