package web

import (
	"fmt"
	"net/http"
)

func (app *Application) home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "The home page")
}

func (app *Application) books(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "The books page")
}

func (app *Application) addBook(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "The add book page")
}

func (app *Application) showBook(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "The show book page")
}

func (app *Application) updateBook(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "The update book page")
}

func (app *Application) deleteBook(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "The delete book page")
}
