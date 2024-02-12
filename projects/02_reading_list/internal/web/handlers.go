package web

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/sswietoniowski/learning-go/projects/02_reading_list/internal/web/service"
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

func (app *Application) addBookForm(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("add book form")

	fmt.Fprintf(w, `
	<html>
	<head>
		<title>Add Book</title>
	</head>
	<body>
		<h1>Add Book</h1>
		<form action="/books/add" method="post">
			<label for="title">Title</label>
			<input type="text" name="title" id="title">
			<label for="author">Author</label>
			<input type="text" name="author" id="author">
			<button type="submit">Add</button>
		</form>
	</body>
</html>	
`)
}

func (app *Application) addBookProcess(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("add book process")

	title := strings.TrimSpace(r.PostFormValue("title"))
	if title == "" {
		app.logger.Println("bad request")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	author := strings.TrimSpace(r.PostFormValue("author"))
	if author == "" {
		app.logger.Println("bad request")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// We could add more fields to the form & then process them, but it is just an
	// example of how to handle form submissions so I see no point to do that.

	book := service.Book{
		Title:  title,
		Author: author,
	}

	err := app.service.Add(book)
	if err != nil {
		app.logger.Println("internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/books", http.StatusSeeOther)
}

func (app *Application) addBook(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("add book page")

	switch r.Method {
	case http.MethodGet:
		app.addBookForm(w, r)
	case http.MethodPost:
		app.addBookProcess(w, r)
	default:
		app.logger.Println("internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
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
