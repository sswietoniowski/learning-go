// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
	"time"
)

func TestSendNewsletterToEmail(t *testing.T) {
	done := make(chan bool)
	go SendNewsletterToEmail("email@example.com", done)

	select {
	case d := <-done:
		if !d {
			t.Error("SendNewsletterToEmail() failed")
		}
	case <-time.After(time.Second * 2):
		t.Error("SendNewsletterToEmail() timed out")
	}
}
