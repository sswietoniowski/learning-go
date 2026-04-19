// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

func TestRunSafely(t *testing.T) {
	err := RunSafely(func() {
		panic("test panic")
	})
	if err == nil {
		t.Error("Expected error, got nil")
	}

	err = RunSafely(func() {})
	if err != nil {
		t.Error("Expected nil, got error")
	}
}
