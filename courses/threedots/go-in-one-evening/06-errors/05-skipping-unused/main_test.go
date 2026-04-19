// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"os"
	"testing"
)

func TestHandlingErrors(t *testing.T) {
	ok := CheckFile("unknown-file")
	if ok {
		t.Error("Expected false when the file doesn't exist")
	}

	err := os.WriteFile("valid-file", []byte{}, 0644)
	if err != nil {
		t.Error("Failed to create a test file")
	}

	t.Cleanup(func() {
		_ = os.Remove("valid-file")
	})

	ok = CheckFile("valid-file")
	if !ok {
		t.Error("Expected true when the file exists")
	}
}
