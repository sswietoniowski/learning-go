package web

import (
	"net/http"
)

// Routes returns the router with all the routes defined.
func (app *Application) Routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/books", app.books)
	mux.HandleFunc("/books/add", app.addBook)
	mux.HandleFunc("/books/show", app.showBook)
	mux.HandleFunc("/books/update", app.updateBook)
	mux.HandleFunc("/books/delete", app.deleteBook)

	return mux
}
