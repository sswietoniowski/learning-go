// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"os"
	"testing"
)

func TestHandlingErrors(t *testing.T) {
	dirName := "directory"

	t.Cleanup(func() {
		_ = os.Remove(dirName)
	})

	ok := CreateDirectory(dirName)
	if !ok {
		t.Error("Expected true when there's no error")
	}

	ok = CreateDirectory(dirName)
	if ok {
		t.Error("Expected false when the directory already exists")
	}
}
