// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

func TestMessages(t *testing.T) {
	testFn := func(t *testing.T, n int) {
		messagesChan := make(chan string)

		go SendMessages(n, messagesChan)

		messages := ReadMessages(messagesChan)
		if len(messages) != n {
			t.Errorf("Expected %v messages, got %v instead", n, len(messages))
		}
	}

	t.Run("4", func(t *testing.T) {
		testFn(t, 4)
	})

	t.Run("42", func(t *testing.T) {
		testFn(t, 42)
	})

	t.Run("120", func(t *testing.T) {
		testFn(t, 120)
	})
}
