package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Event struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	OccurredAt time.Time `json:"occurred_at"`
}

func main() {
	event := Event{
		ID:         "random-id",
		Name:       "ClientSignedUp",
		OccurredAt: time.Now(),
	}

	marshaled, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(marshaled))
}
