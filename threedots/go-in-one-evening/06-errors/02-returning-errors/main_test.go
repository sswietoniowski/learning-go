// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

func TestDivide(t *testing.T) {
	r, err := Divide(200, 50)
	if err != nil {
		t.Error("Divide returned unexpected error:", err)
	}

	expected := 4.0

	if r != expected {
		t.Errorf("Divide(200, 50) returned %v, expected %v", r, expected)
	}

	_, err = Divide(200, 0)
	if err == nil {
		t.Error("Divide(200, 0) should return an error")
	}
}
