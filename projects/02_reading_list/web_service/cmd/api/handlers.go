package main

import (
	"fmt"
	"net/http"
	"strconv"
)

func (app *application) getHealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("get healthcheck")

	if r.Method != http.MethodGet {
		app.logger.Println("get healthcheck: method not allowed")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	fmt.Fprintln(w, "status: available")
	fmt.Fprintf(w, "environment: %s\n", app.config.env)
	fmt.Fprintf(w, "version: %s\n", version)
}

func (app *application) getBooksHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("get all books")

	if r.Method != http.MethodGet {
		app.logger.Println("get all books: method not allowed")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	fmt.Fprintln(w, "get all books")
}

func (app *application) createBooksHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("create a new book")

	if r.Method != http.MethodPost {
		app.logger.Println("create a new book: method not allowed")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	fmt.Fprintln(w, "create a new book")
}

const booksPath = "/api/v1/books/"

func (app *application) getBookHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("get book by id")

	if r.Method != http.MethodGet {
		app.logger.Println("get book by id: method not allowed")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Path[len(booksPath):]
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		app.logger.Println("get book by id: bad request")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "get book by id: %d\n", idInt)
}

func (app *application) updateBookHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("update book by id")

	if r.Method != http.MethodPut {
		app.logger.Println("update book by id: method not allowed")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Path[len(booksPath):]
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		app.logger.Println("update book by id: bad request")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "update book by id: %d\n", idInt)
}

func (app *application) deleteBookHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("delete book by id")

	if r.Method != http.MethodDelete {
		app.logger.Println("delete book by id: method not allowed")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Path[len(booksPath):]
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		app.logger.Println("delete book by id: bad request")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "delete book by id: %d\n", idInt)
}
