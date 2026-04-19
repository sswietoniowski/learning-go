// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"encoding/json"
	"testing"
)

func TestAccount(t *testing.T) {
	account := Account{
		Name:     "John",
		password: "top-secret",
	}

	expected := `{"name":"John"}`

	marshaled, err := json.Marshal(account)
	if err != nil {
		t.Fatalf("Failed to marshal Account: %v", err)
	}

	if string(marshaled) != expected {
		t.Errorf("Incorrect JSON\nExpected: %v\nGot: %v", expected, string(marshaled))
	}
}
