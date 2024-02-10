package main

import (
	"encoding/json"
	"errors"
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

func parseJsonRequest(w http.ResponseWriter, r *http.Request, data interface{}) error {
	err := json.NewDecoder(r.Body).Decode(data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return errors.New("could not decode request data from JSON")
	}

	return nil
}

const booksPath = "/api/v1/books/"

func extractBookId(w http.ResponseWriter, r *http.Request) (int64, error) {
	id := r.URL.Path[len(booksPath):]
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return 0, errors.New("bad request, invalid book id")
	}

	return idInt, nil
}

func sendJsonResponse(w http.ResponseWriter, statusCode int, data interface{}) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return errors.New("could not encode response data to JSON")
	}

	w.Header().Set(contentTypeHeader, jsonContentType)
	if statusCode > 0 {
		w.WriteHeader(statusCode)
	}
	if data != nil {
		w.Write(dataJSON)
	}

	return nil
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

	err := sendJsonResponse(w, http.StatusOK, data)
	if err != nil {
		app.logger.Printf("get healthcheck: internal server error: %v\n", err)
	}
}

func (app *application) getBooksHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("get all books")

	if !isValidMethod(w, r, http.MethodGet) {
		app.logger.Println("get all books: method not allowed")
		return
	}

	err := sendJsonResponse(w, http.StatusOK, books)
	if err != nil {
		app.logger.Printf("get all books: internal server error: %v\n", err)
	}
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
	err := parseJsonRequest(w, r, &book)
	if err != nil {
		app.logger.Printf("create a new book: bad request: %v\n", err)
		return
	}

	book.Id = int64(len(books) + 1)
	books = append(books, book)

	err = sendJsonResponse(w, http.StatusCreated, book)
	if err != nil {
		app.logger.Printf("create a new book: internal server error: %v\n", err)
	}
}

func (app *application) getBookByIdHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("get book by id")

	if !isValidMethod(w, r, http.MethodGet) {
		app.logger.Println("get book by id: method not allowed")
		return
	}

	id, err := extractBookId(w, r)
	if err != nil {
		app.logger.Printf("get book by id: bad request: %v\n", err)
		return
	}

	app.logger.Printf("get book by id: %d\n", id)

	var book Book
	for _, book = range books {
		if book.Id == id {
			break
		}
	}

	err = sendJsonResponse(w, http.StatusOK, book)
	if err != nil {
		app.logger.Printf("get book by id: internal server error: %v\n", err)
	}
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

	id, err := extractBookId(w, r)
	if err != nil {
		app.logger.Printf("update book by id: bad request: %v\n", err)
		return
	}

	app.logger.Printf("update book by id: %d\n", id)

	var book Book
	err = parseJsonRequest(w, r, &book)
	if err != nil {
		app.logger.Printf("update book by id: bad request: %v\n", err)
		return
	}

	book.Id = id
	for i, b := range books {
		if b.Id == id {
			books[i] = book
			break
		}
	}

	err = sendJsonResponse(w, http.StatusNoContent, nil)
	if err != nil {
		app.logger.Printf("update book by id: internal server error: %v\n", err)
	}
}

func (app *application) deleteBookByIdHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("delete book by id")

	if !isValidMethod(w, r, http.MethodDelete) {
		app.logger.Println("delete book by id: method not allowed")
		return
	}

	id, err := extractBookId(w, r)
	if err != nil {
		app.logger.Printf("delete book by id: bad request: %v\n", err)
		return
	}

	app.logger.Printf("delete book by id: %d\n", id)

	for i, book := range books {
		if book.Id == id {
			books = append(books[:i], books[i+1:]...)
			break
		}
	}

	err = sendJsonResponse(w, http.StatusNoContent, nil)
	if err != nil {
		app.logger.Printf("delete book by id: internal server error: %v\n", err)
	}
}
