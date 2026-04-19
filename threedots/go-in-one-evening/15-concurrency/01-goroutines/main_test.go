// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
	"time"
)

func TestSignUp(t *testing.T) {
	done := make(chan bool)
	go func() {
		SignUp("joe@example.com")
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(time.Millisecond * 1500):
		t.Error("SignUp timed out")
	}
}
