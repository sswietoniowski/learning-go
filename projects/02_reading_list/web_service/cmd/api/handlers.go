package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

const contentTypeHeader = "Content-Type"
const jsonContentType = "application/json"

func isValidMethod(w http.ResponseWriter, r *http.Request, expectedMethod string) bool {
	if r.Method != expectedMethod {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return false
	}

	return true
}

func isValidContentType(w http.ResponseWriter, r *http.Request, expectedContentType string) bool {
	if r.Header.Get(contentTypeHeader) != expectedContentType {
		http.Error(w, "Invalid Content-Type, expected 'application/json'", http.StatusUnsupportedMediaType)
		return false
	}

	return true
}

func (app *application) getHealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("get healthcheck")

	if !isValidMethod(w, r, http.MethodGet) {
		app.logger.Println("get healthcheck: method not allowed")
		return
	}

	data := map[string]string{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		app.logger.Println("get healthcheck: internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set(contentTypeHeader, jsonContentType)
	w.Write(dataJSON)
}

func (app *application) getBooksHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("get all books")

	if !isValidMethod(w, r, http.MethodGet) {
		app.logger.Println("get all books: method not allowed")
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

	if !isValidMethod(w, r, http.MethodPost) {
		app.logger.Println("create a new book: method not allowed")
		return
	}

	if !isValidContentType(w, r, jsonContentType) {
		app.logger.Println("create a new book: unsupported media type")
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
	w.WriteHeader(http.StatusCreated)
	w.Write(bookJSON)
}

const booksPath = "/api/v1/books/"

func (app *application) getBookByIdHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("get book by id")

	if !isValidMethod(w, r, http.MethodGet) {
		app.logger.Println("get book by id: method not allowed")
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

	if !isValidMethod(w, r, http.MethodPut) {
		app.logger.Println("update book by id: method not allowed")
		return
	}

	if !isValidContentType(w, r, jsonContentType) {
		app.logger.Println("update book by id: unsupported media type")
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

	if !isValidMethod(w, r, http.MethodDelete) {
		app.logger.Println("delete book by id: method not allowed")
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
