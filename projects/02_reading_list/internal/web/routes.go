package web

import (
	"net/http"
)

// Routes returns the router with all the routes defined.
func (app *Application) Routes() *http.ServeMux {
	r := http.NewServeMux()

	// TODO: add routes here

	return r
}
