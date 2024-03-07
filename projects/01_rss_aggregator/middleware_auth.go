package main

import (
	"net/http"

	"github.com/sswietoniowski/learning-go/projects/01_rss_aggregator/internal/auth"
	"github.com/sswietoniowski/learning-go/projects/01_rss_aggregator/internal/database"
)

type authHandler func(http.ResponseWriter, *http.Request, database.User)

func (cfg *apiConfig) middlewareAuth(handler authHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey, err := auth.GetApiKey(r.Header)
		if err != nil {
			switch err {
			case auth.ErrNoAuthHeaderIncluded:
				respondWithError(w, http.StatusUnauthorized, "No API key included")
			case auth.ErrMalformedAuthHeader:
				respondWithError(w, http.StatusBadRequest, "Malformed API key")
			default:
				respondWithError(w, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		user, err := cfg.DB.GetUserByApiKey(r.Context(), apiKey)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Invalid API key")
			return
		}

		handler(w, r, user)
	}
}
