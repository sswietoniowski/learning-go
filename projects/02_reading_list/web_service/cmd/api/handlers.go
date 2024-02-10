package main

import (
	"net/http"

	. "github.com/sswietoniowski/learning-go/projects/02_reading_list/web_service/internal/data"
	. "github.com/sswietoniowski/learning-go/projects/02_reading_list/web_service/internal/helper"
)

var database = NewDatabase()

const booksPath = "/api/v1/books/"

func (app *application) getHealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("get healthcheck")

	if !IsValidMethod(w, r, http.MethodGet) {
		app.logger.Println("get healthcheck: method not allowed")
		return
	}

	data := map[string]string{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
	}

	err := SendJsonResponse(w, http.StatusOK, data)
	if err != nil {
		app.logger.Printf("get healthcheck: internal server error: %v\n", err)
	}
}

func (app *application) getBooksHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("get all books")

	if !IsValidMethod(w, r, http.MethodGet) {
		app.logger.Println("get all books: method not allowed")
		return
	}

	books := database.GetAll()
	err := SendJsonResponse(w, http.StatusOK, books)
	if err != nil {
		app.logger.Printf("get all books: internal server error: %v\n", err)
	}
}

func (app *application) createBooksHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("create a new book")

	if !IsValidMethod(w, r, http.MethodPost) {
		app.logger.Println("create a new book: method not allowed")
		return
	}

	if !IsValidContentType(w, r, JsonContentType) {
		app.logger.Println("create a new book: unsupported media type")
		return
	}

	var book Book
	err := ParseJsonRequest(w, r, &book)
	if err != nil {
		app.logger.Printf("create a new book: bad request: %v\n", err)
		return
	}

	book = database.Add(book)

	err = SendJsonResponse(w, http.StatusCreated, book)
	if err != nil {
		app.logger.Printf("create a new book: internal server error: %v\n", err)
	}
}

func (app *application) getBookByIdHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("get book by id")

	if !IsValidMethod(w, r, http.MethodGet) {
		app.logger.Println("get book by id: method not allowed")
		return
	}

	id, err := ExtractIdFromRoute(w, r, booksPath)
	if err != nil {
		app.logger.Printf("get book by id: bad request: %v\n", err)
		return
	}

	app.logger.Printf("get book by id: %d\n", id)

	book, found := database.GetById(id)
	if !found {
		SendJsonResponse(w, http.StatusNotFound, nil)
		return
	}

	err = SendJsonResponse(w, http.StatusOK, book)
	if err != nil {
		app.logger.Printf("get book by id: internal server error: %v\n", err)
	}
}

func (app *application) updateBookByIdHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("update book by id")

	if !IsValidMethod(w, r, http.MethodPut) {
		app.logger.Println("update book by id: method not allowed")
		return
	}

	if !IsValidContentType(w, r, JsonContentType) {
		app.logger.Println("update book by id: unsupported media type")
		return
	}

	id, err := ExtractIdFromRoute(w, r, booksPath)
	if err != nil {
		app.logger.Printf("update book by id: bad request: %v\n", err)
		return
	}

	app.logger.Printf("update book by id: %d\n", id)

	var book Book
	err = ParseJsonRequest(w, r, &book)
	if err != nil {
		app.logger.Printf("update book by id: bad request: %v\n", err)
		return
	}

	_, updated := database.ModifyById(id, book)
	if !updated {
		SendJsonResponse(w, http.StatusNotFound, nil)
		return
	}

	err = SendJsonResponse(w, http.StatusNoContent, nil)
	if err != nil {
		app.logger.Printf("update book by id: internal server error: %v\n", err)
	}
}

func (app *application) deleteBookByIdHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("delete book by id")

	if !IsValidMethod(w, r, http.MethodDelete) {
		app.logger.Println("delete book by id: method not allowed")
		return
	}

	id, err := ExtractIdFromRoute(w, r, booksPath)
	if err != nil {
		app.logger.Printf("delete book by id: bad request: %v\n", err)
		return
	}

	app.logger.Printf("delete book by id: %d\n", id)

	_, deleted := database.RemoveById(id)
	if !deleted {
		SendJsonResponse(w, http.StatusNotFound, nil)
		return
	}

	err = SendJsonResponse(w, http.StatusNoContent, nil)
	if err != nil {
		app.logger.Printf("delete book by id: internal server error: %v\n", err)
	}
}
