// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEvent(t *testing.T) {
	event := Event{
		ID:         "random-id",
		Name:       "ClientSignedUp",
		OccurredAt: time.Date(2000, 1, 31, 12, 30, 40, 500, time.UTC),
	}

	expected := `{"id":"random-id","name":"ClientSignedUp","occurred_at":"2000-01-31T12:30:40.0000005Z"}`

	marshaled, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal Event: %v", err)
	}

	if string(marshaled) != expected {
		t.Errorf("Incorrect JSON\nExpected: %v\nGot: %v", expected, string(marshaled))
	}
}
