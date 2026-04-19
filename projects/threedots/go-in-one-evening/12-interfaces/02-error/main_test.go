// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"errors"
	"testing"
)

func TestError(t *testing.T) {
	storage := MemoryStorage{
		[]User{
			{
				ID:   100,
				Name: "Alice",
			},
		},
	}

	_, err := storage.FindUser(100)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	_, err = storage.FindUser(200)
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}

	expected := "user not found: 200"
	if err.Error() != expected {
		t.Errorf("Incorrect error message\nExpected: %v\nGot: %v", expected, err.Error())
	}

	if !errors.As(err, &UserNotFoundError{}) {
		t.Error("Expected error to be UserNotFoundError")
	}
}
