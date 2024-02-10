package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

const bookIDPath = "/api/v1/books/{id:[0-9]+}"

func (app *application) routes() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/api/v1/healthcheck", app.getHealthcheckHandler)
	r.HandleFunc("/api/v1/books", app.getBooksHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/books", app.createBooksHandler).Methods(http.MethodPost)
	r.HandleFunc(bookIDPath, app.getBookByIdHandler).Methods(http.MethodGet)
	r.HandleFunc(bookIDPath, app.updateBookByIdHandler).Methods(http.MethodPut)
	r.HandleFunc(bookIDPath, app.deleteBookByIdHandler).Methods(http.MethodDelete)

	return r
}
