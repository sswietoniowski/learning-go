package api

import (
	"net/http"

	"github.com/sswietoniowski/learning-go/projects/02_reading_list/web_service/internal/data"
)

const booksPath = "/api/v1/books/"

func (app *Application) getHealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("get healthcheck")

	if !isValidMethod(w, r, http.MethodGet) {
		app.logger.Println("method not allowed")
		return
	}

	data := map[string]string{
		"status":      "available",
		"environment": app.config.EnvironmentName,
		"version":     Version,
	}

	err := sendJsonResponse(w, http.StatusOK, data)
	if err != nil {
		app.logger.Printf("internal server error: %v\n", err)
	}
}

func (app *Application) getBooksHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("get all books")

	if !isValidMethod(w, r, http.MethodGet) {
		app.logger.Println("method not allowed")
		return
	}

	books, err := app.database.GetAll()
	if err != nil {
		app.logger.Printf("internal server error: %v\n", err)
		sendJsonResponse(w, http.StatusInternalServerError, nil)
		return
	}

	err = sendJsonResponse(w, http.StatusOK, books)
	if err != nil {
		app.logger.Printf("internal server error: %v\n", err)
	}
}

func (app *Application) createBooksHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("create a new book")

	if !isValidMethod(w, r, http.MethodPost) {
		app.logger.Println("method not allowed")
		return
	}

	if !isValidContentType(w, r, jsonContentType) {
		app.logger.Println("unsupported media type")
		return
	}

	var book data.Book
	err := parseJsonRequest(w, r, &book)
	if err != nil {
		app.logger.Printf("bad request: %v\n", err)
		return
	}

	createdBook, err := app.database.Add(book)
	if err != nil {
		app.logger.Printf("internal server error: %v\n", err)
		sendJsonResponse(w, http.StatusInternalServerError, nil)
		return
	}

	err = sendJsonResponse(w, http.StatusCreated, createdBook)
	if err != nil {
		app.logger.Printf("internal server error: %v\n", err)
	}
}

func (app *Application) getBookByIdHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("get book by id")

	if !isValidMethod(w, r, http.MethodGet) {
		app.logger.Println("method not allowed")
		return
	}

	id, err := extractIdFromRoute(w, r, booksPath)
	if err != nil {
		app.logger.Printf("bad request: %v\n", err)
		return
	}

	book, err := app.database.GetById(id)
	if err != nil {
		switch err.(type) {
		case *data.NotFoundError:
			app.logger.Printf("not found: %v\n", err)
			sendJsonResponse(w, http.StatusNotFound, nil)
		default:
			app.logger.Printf("internal server error: %v\n", err)
			sendJsonResponse(w, http.StatusInternalServerError, nil)
		}
		return
	}

	err = sendJsonResponse(w, http.StatusOK, book)
	if err != nil {
		app.logger.Printf("internal server error: %v\n", err)
	}
}

func (app *Application) updateBookByIdHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("update book by id")

	if !isValidMethod(w, r, http.MethodPut) {
		app.logger.Println("method not allowed")
		return
	}

	if !isValidContentType(w, r, jsonContentType) {
		app.logger.Println("unsupported media type")
		return
	}

	id, err := extractIdFromRoute(w, r, booksPath)
	if err != nil {
		app.logger.Printf("bad request: %v\n", err)
		return
	}

	var book data.Book
	err = parseJsonRequest(w, r, &book)
	if err != nil {
		app.logger.Printf("bad request: %v\n", err)
		return
	}

	_, err = app.database.ModifyById(id, book)
	if err != nil {
		switch err.(type) {
		case *data.NotFoundError:
			app.logger.Printf("not found: %v\n", err)
			sendJsonResponse(w, http.StatusNotFound, nil)
		default:
			app.logger.Printf("internal server error: %v\n", err)
			sendJsonResponse(w, http.StatusInternalServerError, nil)
		}
		return
	}

	err = sendJsonResponse(w, http.StatusNoContent, nil)
	if err != nil {
		app.logger.Printf("internal server error: %v\n", err)
	}
}

func (app *Application) deleteBookByIdHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("delete book by id")

	if !isValidMethod(w, r, http.MethodDelete) {
		app.logger.Println("method not allowed")
		return
	}

	id, err := extractIdFromRoute(w, r, booksPath)
	if err != nil {
		app.logger.Printf("bad request: %v\n", err)
		return
	}

	_, err = app.database.RemoveById(id)
	if err != nil {
		switch err.(type) {
		case *data.NotFoundError:
			app.logger.Printf("not found: %v\n", err)
			sendJsonResponse(w, http.StatusNotFound, nil)
		default:
			app.logger.Printf("internal server error: %v\n", err)
			sendJsonResponse(w, http.StatusInternalServerError, nil)
		}
		return
	}

	err = sendJsonResponse(w, http.StatusNoContent, nil)
	if err != nil {
		app.logger.Printf("internal server error: %v\n", err)
	}
}
