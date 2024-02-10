package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

const contentTypeHeader = "Content-Type"
const jsonContentType = "application/json"

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

	booksJSON, err := json.Marshal(books)
	if err != nil {
		app.logger.Println("get all books: internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set(contentTypeHeader, jsonContentType)
	w.Write(booksJSON)
}

func (app *application) createBooksHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("create a new book")

	if r.Header.Get(contentTypeHeader) != jsonContentType {
		http.Error(w, "Invalid Content-Type, expected 'application/json'", http.StatusUnsupportedMediaType)
		return
	}

	if r.Method != http.MethodPost {
		app.logger.Println("create a new book: method not allowed")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var book Book
	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		app.logger.Println("create a new book: bad request")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	book.Id = int64(len(books) + 1)
	books = append(books, book)

	bookJSON, err := json.Marshal(book)
	if err != nil {
		app.logger.Println("create a new book: internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set(contentTypeHeader, jsonContentType)
	w.Write(bookJSON)
	w.WriteHeader(http.StatusCreated)
}

const booksPath = "/api/v1/books/"

func (app *application) getBookByIdHandler(w http.ResponseWriter, r *http.Request) {
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

	app.logger.Printf("get book by id: %d\n", idInt)

	var book Book
	for _, book = range books {
		if book.Id == idInt {
			break
		}
	}

	bookJSON, err := json.Marshal(book)
	if err != nil {
		app.logger.Println("get book by id: internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set(contentTypeHeader, jsonContentType)
	w.Write(bookJSON)
}

func (app *application) updateBookByIdHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("update book by id")

	if r.Header.Get(contentTypeHeader) != jsonContentType {
		http.Error(w, "Invalid Content-Type, expected 'application/json'", http.StatusUnsupportedMediaType)
		return
	}

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

	app.logger.Printf("update book by id: %d\n", idInt)

	var book Book
	err = json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		app.logger.Println("update book by id: bad request")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	book.Id = idInt
	for i, b := range books {
		if b.Id == idInt {
			books[i] = book
			break
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) deleteBookByIdHandler(w http.ResponseWriter, r *http.Request) {
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

	app.logger.Printf("delete book by id: %d\n", idInt)

	for i, book := range books {
		if book.Id == idInt {
			books = append(books[:i], books[i+1:]...)
			break
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
