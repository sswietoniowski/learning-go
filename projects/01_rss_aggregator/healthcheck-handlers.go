package main

import "net/http"

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func errHandler(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
}
