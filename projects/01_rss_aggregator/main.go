package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/sswietoniowski/learning-go/projects/01_rss_aggregator/internal/database"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {
	godotenv.Load()

	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("PORT is not found in the environment")
	}

	_, err := strconv.ParseInt(portString, 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not found in the environment")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	apiCfg := &apiConfig{
		DB: dbQueries,
	}

	router := chi.NewRouter()

	corsOptions := cors.Options{
		AllowedOrigins:   []string{"http://*", "https://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}
	router.Use(cors.Handler(corsOptions))

	v1Router := chi.NewRouter()

	v1Router.Post("/users", apiCfg.handlerUsersCreate)
	v1Router.Get("/users", apiCfg.middlewareAuth(apiCfg.handlerUsersGet))

	v1Router.Post("/feeds", apiCfg.middlewareAuth(apiCfg.handlerFeedsCreate))

	v1Router.Get("/healthz", handlerReadiness)
	v1Router.Get("/err", handlerErr)

	router.Mount("/api/v1", v1Router)

	addr := fmt.Sprintf(":%s", portString)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	fmt.Printf("Server is running on port %s\n", portString)

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
