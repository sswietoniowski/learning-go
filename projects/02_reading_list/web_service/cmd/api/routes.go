package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (app *application) routes() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/api/v1/healthcheck", app.getHealthcheckHandler)
	r.HandleFunc("/api/v1/books", app.getBooksHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/books", app.postBooksHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/books/{id:[0-9]+}", app.getBookHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/books/{id:[0-9]+}", app.updateBookHandler).Methods(http.MethodPut)
	r.HandleFunc("/api/v1/books/{id:[0-9]+}", app.deleteBookHandler).Methods(http.MethodDelete)

	return r
}
