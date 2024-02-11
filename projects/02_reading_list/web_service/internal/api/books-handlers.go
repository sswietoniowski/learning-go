package api

import (
	"net/http"

	"github.com/sswietoniowski/learning-go/projects/02_reading_list/web_service/internal/data"
)

const booksPath = "/api/v1/books/"

func (app *Application) getHealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("get healthcheck")

	if !isValidMethod(w, r, http.MethodGet) {
		app.logger.Println("get healthcheck: method not allowed")
		return
	}

	data := map[string]string{
		"status":      "available",
		"environment": app.config.EnvironmentName,
		"version":     Version,
	}

	err := sendJsonResponse(w, http.StatusOK, data)
	if err != nil {
		app.logger.Printf("get healthcheck: internal server error: %v\n", err)
	}
}

func (app *Application) getBooksHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("get all books")

	if !isValidMethod(w, r, http.MethodGet) {
		app.logger.Println("get all books: method not allowed")
		return
	}

	books := app.database.GetAll()

	err := sendJsonResponse(w, http.StatusOK, books)
	if err != nil {
		app.logger.Printf("get all books: internal server error: %v\n", err)
	}
}

func (app *Application) createBooksHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("create a new book")

	if !isValidMethod(w, r, http.MethodPost) {
		app.logger.Println("create a new book: method not allowed")
		return
	}

	if !isValidContentType(w, r, jsonContentType) {
		app.logger.Println("create a new book: unsupported media type")
		return
	}

	var book data.Book
	err := parseJsonRequest(w, r, &book)
	if err != nil {
		app.logger.Printf("create a new book: bad request: %v\n", err)
		return
	}

	book = app.database.Add(book)

	err = sendJsonResponse(w, http.StatusCreated, book)
	if err != nil {
		app.logger.Printf("create a new book: internal server error: %v\n", err)
	}
}

func (app *Application) getBookByIdHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("get book by id")

	if !isValidMethod(w, r, http.MethodGet) {
		app.logger.Println("get book by id: method not allowed")
		return
	}

	id, err := extractIdFromRoute(w, r, booksPath)
	if err != nil {
		app.logger.Printf("get book by id: bad request: %v\n", err)
		return
	}

	app.logger.Printf("get book by id: %d\n", id)

	book, found := app.database.GetById(id)
	if !found {
		sendJsonResponse(w, http.StatusNotFound, nil)
		return
	}

	err = sendJsonResponse(w, http.StatusOK, book)
	if err != nil {
		app.logger.Printf("get book by id: internal server error: %v\n", err)
	}
}

func (app *Application) updateBookByIdHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("update book by id")

	if !isValidMethod(w, r, http.MethodPut) {
		app.logger.Println("update book by id: method not allowed")
		return
	}

	if !isValidContentType(w, r, jsonContentType) {
		app.logger.Println("update book by id: unsupported media type")
		return
	}

	id, err := extractIdFromRoute(w, r, booksPath)
	if err != nil {
		app.logger.Printf("update book by id: bad request: %v\n", err)
		return
	}

	app.logger.Printf("update book by id: %d\n", id)

	var book data.Book
	err = parseJsonRequest(w, r, &book)
	if err != nil {
		app.logger.Printf("update book by id: bad request: %v\n", err)
		return
	}

	_, updated := app.database.ModifyById(id, book)
	if !updated {
		sendJsonResponse(w, http.StatusNotFound, nil)
		return
	}

	err = sendJsonResponse(w, http.StatusNoContent, nil)
	if err != nil {
		app.logger.Printf("update book by id: internal server error: %v\n", err)
	}
}

func (app *Application) deleteBookByIdHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("delete book by id")

	if !isValidMethod(w, r, http.MethodDelete) {
		app.logger.Println("delete book by id: method not allowed")
		return
	}

	id, err := extractIdFromRoute(w, r, booksPath)
	if err != nil {
		app.logger.Printf("delete book by id: bad request: %v\n", err)
		return
	}

	app.logger.Printf("delete book by id: %d\n", id)

	_, deleted := app.database.RemoveById(id)
	if !deleted {
		sendJsonResponse(w, http.StatusNotFound, nil)
		return
	}

	err = sendJsonResponse(w, http.StatusNoContent, nil)
	if err != nil {
		app.logger.Printf("delete book by id: internal server error: %v\n", err)
	}
}
