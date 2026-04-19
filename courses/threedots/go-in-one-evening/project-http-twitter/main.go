package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
	"twitter/server"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

func main() {
	s := &server.Server{
		TweetsRepository: &server.TweetsMemoryRepository{},
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/tweets", s.ListTweets)
	r.With(httprate.LimitByIP(10, time.Minute)).Post("/tweets", s.AddTweet)

	go spamTweets()

	log.Fatal(http.ListenAndServe(":8080", r))
}

func spamTweets() {
	messages := []string{
		"Hello from the spam bot!",
		"This is an automated tweet",
		"Go programming is awesome!",
		"Testing the Twitter API",
		"Another spam message here",
		"Concurrent programming rocks!",
		"Building APIs with Go",
		"Chi router is fantastic",
		"HTTP requests are flying!",
		"JSON marshaling in action",
	}

	locations := []string{
		"New York", "London", "Tokyo", "Paris", "Sydney",
		"Berlin", "Toronto", "Mumbai", "San Francisco", "Moscow",
	}

	for {
		// Prepare payload
		tweet := struct {
			Message  string `json:"message"`
			Location string `json:"location"`
		}{
			Message:  messages[rand.Intn(len(messages))],
			Location: locations[rand.Intn(len(locations))],
		}

		// Marshal payload
		payload, err := json.Marshal(tweet)
		if err != nil {
			log.Printf("Failed to marshal tweet: %v", err)
			time.Sleep(time.Second)
			continue
		}

		// Send HTTP POST request
		resp, err := http.Post("http://localhost:8080/tweets", "application/json", bytes.NewBuffer(payload))
		if err != nil {
			log.Printf("Failed to send POST request: %v", err)
			time.Sleep(time.Second)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Failed to read response body: %v", err)
		} else {
			fmt.Printf("Response: %s\n", string(body))
		}

		// Close response body
		resp.Body.Close()

		// (Optionally read and print the response)
		fmt.Printf("Response: %s %s\n", resp.Status, string(body))
	}
}
