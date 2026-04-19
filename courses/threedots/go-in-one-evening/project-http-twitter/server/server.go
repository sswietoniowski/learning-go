package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type tweet struct {
	Message  string `json:"message"`
	Location string `json:"location"`
}

type id struct {
	ID int `json:"ID"`
}

type tweetsList struct {
	Tweets []tweet `json:"tweets"`
}

type TweetsRepository interface {
	AddTweet(t tweet) (int, error)
	ListTweets() ([]tweet, error)
}

type TweetsMemoryRepository struct {
	lock   sync.RWMutex
	tweets []tweet
}

func (t *TweetsMemoryRepository) AddTweet(tw tweet) (int, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.tweets = append(t.tweets, tw)
	return len(t.tweets), nil
}

func (t *TweetsMemoryRepository) ListTweets() ([]tweet, error) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.tweets, nil
}

type Server struct {
	TweetsRepository TweetsRepository
}

func (s *Server) Tweets(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		end := time.Now()
		duration := end.Sub(start)
		fmt.Printf("%s %s %s\n", r.Method, r.URL, duration)
	}()

	if r.Method == http.MethodPost {
		s.AddTweet(w, r)
	} else if r.Method == http.MethodGet {
		s.ListTweets(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) AddTweet(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Failed to read body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	t := tweet{}

	if err := json.Unmarshal(body, &t); err != nil {
		log.Println("Failed to unmarshal payload:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Printf("Tweet: `%s` from %s\n", t.Message, t.Location)

	id, err := s.TweetsRepository.AddTweet(t)
	if err != nil {
		log.Println("Failed to add tweet:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	idAsString := fmt.Sprintf(`{"ID":%d}`, id)
	payload := []byte(idAsString)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(payload)
}

func (s *Server) ListTweets(w http.ResponseWriter, r *http.Request) {
	tweets, err := s.TweetsRepository.ListTweets()
	if err != nil {
		log.Println("Failed to list tweets:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tweetsList := tweetsList{
		Tweets: tweets,
	}

	payload, err := json.Marshal(tweetsList)
	if err != nil {
		log.Println("Failed to marshal tweets:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(payload)
}
