package web

import (
	"fmt"
	"html/template"
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

	books, err := app.service.GetAll()
	if err != nil {
		app.logger.Println("internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	files := []string{
		"./ui/html/base.html",
		"./ui/html/partials/nav.html",
		"./ui/html/pages/home.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = ts.ExecuteTemplate(w, "base", books)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

/*
func (app *Application) books(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("books page")

	// For others endpoints we will use templates, but for this one we will use
	// plain HTML to keep it simple and to show that we can use both.

	books, err := app.service.GetAll()
	if err != nil {
		app.logger.Println(err)

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
*/

func (app *Application) addBookForm(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("add book form")

	files := []string{
		"./ui/html/base.html",
		"./ui/html/partials/nav.html",
		"./ui/html/pages/add.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = ts.ExecuteTemplate(w, "base", nil)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (app *Application) addBookProcess(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("add book process")

	err := r.ParseForm()
	if err != nil {
		app.logger.Println("bad request")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

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

	published, err := strconv.Atoi(r.PostFormValue("published"))
	if err != nil || published < 0 {
		app.logger.Println("bad request")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	pages, err := strconv.Atoi(r.PostFormValue("pages"))
	if err != nil || pages < 0 {
		app.logger.Println("bad request")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	genres := strings.Split(r.PostFormValue("genres"), ",")

	rating, err := strconv.ParseFloat(r.PostFormValue("rating"), 32)
	if err != nil || rating < 0 {
		app.logger.Println("bad request")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	book := service.Book{
		Title:     title,
		Author:    author,
		Published: published,
		Pages:     pages,
		Genres:    genres,
		Rating:    float32(rating),
	}

	err = app.service.Add(book)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
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
		app.logger.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	files := []string{
		"./ui/html/base.html",
		"./ui/html/partials/nav.html",
		"./ui/html/pages/show.html",
	}

	// Used to convert comma-separated genres to a slice within the template.
	funcs := template.FuncMap{"join": strings.Join}

	ts, err := template.New("show").Funcs(funcs).ParseFiles(files...)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = ts.ExecuteTemplate(w, "base", book)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (app *Application) updateBook(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("update book page")

	// We could implement this endpoint in the same way as the addBook endpoint,
	// so for now I will just leave it as a placeholder.

	fmt.Fprintln(w, "The update book page")
}

func (app *Application) deleteBook(w http.ResponseWriter, r *http.Request) {
	app.logger.Println("delete book page")

	// We could implement this endpoint in the same way as the addBook endpoint,
	// so for now I will just leave it as a placeholder.

	fmt.Fprintln(w, "The delete book page")
}
