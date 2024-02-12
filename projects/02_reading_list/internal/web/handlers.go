package web

import (
	"fmt"
	"net/http"
	"strconv"
)

func (app *Application) home(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("home page")
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, "/books", http.StatusMovedPermanently)
}

func (app *Application) books(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("books page")
	books, err := app.service.GetAll()
	if err != nil {
		app.logger.Println("internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, `
<html>
	<head>
		<title>Reading List</title>
	</head>
	<body>
		<h1>Reading List</h1>
		<ul>
	`)
	for _, book := range *books {
		fmt.Fprintf(w, "<li>%s (%d)</li>", book.Title, book.Pages)
	}
	fmt.Fprintf(w, `
		</ul>
	</body>
</html>
`)
}

func (app *Application) addBook(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("add book page")
	fmt.Fprintln(w, "The add book page")
}

func (app *Application) showBook(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("show book page")

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.logger.Println("not found")
		http.NotFound(w, r)
		return
	}

	book, err := app.service.Get(int64(id))
	if err != nil {
		app.logger.Println("internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, `
	<html>
	<head>
		<title>Show Book</title>
	</head>
	<body>
	`)

	fmt.Fprintf(w, "<h1>%s</h1><p>%s</p><p>%d</p>", book.Title, book.Author, book.Published)
}

func (app *Application) updateBook(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("update book page")
	fmt.Fprintln(w, "The update book page")
}

func (app *Application) deleteBook(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "The delete book page")
}
