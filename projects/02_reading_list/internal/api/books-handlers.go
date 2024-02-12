package api

import (
	"net/http"
	"strconv"

	"github.com/sswietoniowski/learning-go/projects/02_reading_list/internal/api/data"
)

const booksPath = "/api/v1/books/"

func (app *Application) getBooksHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("get all books")

	if !IsValidMethod(w, r, http.MethodGet) {
		app.logger.Println("method not allowed")
		SendMethodNotAllowed(w)
		return
	}

	books, err := app.database.GetAll()
	if err != nil {
		app.logger.Printf("internal server error: %v\n", err)
		SendInternalServerError(w)
		return
	}

	err = SendOk(w, books)
	if err != nil {
		app.logger.Printf("internal server error: %v\n", err)
	}
}

func (app *Application) createBooksHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("create a new book")

	if !IsValidMethod(w, r, http.MethodPost) {
		app.logger.Println("method not allowed")
		SendMethodNotAllowed(w)
		return
	}

	if !IsValidContentType(w, r, jsonContentType) {
		app.logger.Println("unsupported media type")
		SendUnsupportedMediaType(w)
		return
	}

	var book data.Book
	err := ParseJsonRequest(w, r, &book)
	if err != nil {
		app.logger.Printf("bad request: %v\n", err)
		SendBadRequest(w)
		return
	}

	createdBook, err := app.database.Add(book)
	if err != nil {
		app.logger.Printf("internal server error: %v\n", err)
		SendInternalServerError(w)
		return
	}

	location := booksPath + strconv.FormatInt(createdBook.Id, 10)

	err = SendCreated(w, createdBook, location)
	if err != nil {
		app.logger.Printf("internal server error: %v\n", err)
	}
}

func (app *Application) getBookByIdHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("get book by id")

	if !IsValidMethod(w, r, http.MethodGet) {
		app.logger.Println("method not allowed")
		SendMethodNotAllowed(w)
		return
	}

	id, err := ExtractIdFromRoute(w, r, booksPath)
	if err != nil {
		app.logger.Printf("bad request: %v\n", err)
		SendBadRequest(w)
		return
	}

	book, err := app.database.GetById(id)
	if err != nil {
		switch err.(type) {
		case *data.NotFoundError:
			app.logger.Printf("not found: %v\n", err)
			SendNotFound(w)
		default:
			app.logger.Printf("internal server error: %v\n", err)
			SendInternalServerError(w)
		}
		return
	}

	err = SendOk(w, book)
	if err != nil {
		app.logger.Printf("internal server error: %v\n", err)
	}
}

func (app *Application) updateBookByIdHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("update book by id")

	if !IsValidMethod(w, r, http.MethodPut) {
		app.logger.Println("method not allowed")
		SendMethodNotAllowed(w)
		return
	}

	if !IsValidContentType(w, r, jsonContentType) {
		app.logger.Println("unsupported media type")
		SendUnsupportedMediaType(w)
		return
	}

	id, err := ExtractIdFromRoute(w, r, booksPath)
	if err != nil {
		app.logger.Printf("bad request: %v\n", err)
		SendBadRequest(w)
		return
	}

	var book data.Book
	err = ParseJsonRequest(w, r, &book)
	if err != nil {
		app.logger.Printf("bad request: %v\n", err)
		SendBadRequest(w)
		return
	}

	_, err = app.database.ModifyById(id, book)
	if err != nil {
		switch err.(type) {
		case *data.NotFoundError:
			app.logger.Printf("not found: %v\n", err)
			SendNotFound(w)
		default:
			app.logger.Printf("internal server error: %v\n", err)
			SendInternalServerError(w)
		}
		return
	}

	err = SendNoContent(w)
	if err != nil {
		app.logger.Printf("internal server error: %v\n", err)
	}
}

func (app *Application) deleteBookByIdHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("delete book by id")

	if !IsValidMethod(w, r, http.MethodDelete) {
		app.logger.Println("method not allowed")
		SendMethodNotAllowed(w)
		return
	}

	id, err := ExtractIdFromRoute(w, r, booksPath)
	if err != nil {
		app.logger.Printf("bad request: %v\n", err)
		SendBadRequest(w)
		return
	}

	_, err = app.database.RemoveById(id)
	if err != nil {
		switch err.(type) {
		case *data.NotFoundError:
			app.logger.Printf("not found: %v\n", err)
			SendNotFound(w)
		default:
			app.logger.Printf("internal server error: %v\n", err)
			SendInternalServerError(w)
		}
		return
	}

	err = SendNoContent(w)
	if err != nil {
		app.logger.Printf("internal server error: %v\n", err)
	}
}
