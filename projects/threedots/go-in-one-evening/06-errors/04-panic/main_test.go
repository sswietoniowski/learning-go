// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

func TestPanic(t *testing.T) {
	dp := didPanic(func() {
		MustStoreMessage("message")
	})
	if dp {
		t.Error("Expected the function to not panic on nil error")
	}
	if message != "message" {
		t.Errorf(`MustStoreMessage("message") did not store the message, stored message: "%s"`, message)
	}

	dp = didPanic(func() {
		MustStoreMessage("")
	})
	if !dp {
		t.Error("Expected the function to panic on error")
	}
}

func TestStoreMessage(t *testing.T) {
	err := StoreMessage("foo")
	if err != nil {
		t.Error(`StoreMessage("foo") should not return error, but returned`, err)
	}

	if message != "foo" {
		t.Errorf(`StoreMessage("foo") did not store the message, stored message: "%s"`, message)
	}
}

func didPanic(f func()) (didPanic bool) {
	defer func() {
		r := recover()
		if r != nil {
			didPanic = true
		}
	}()

	f()
	return
}
