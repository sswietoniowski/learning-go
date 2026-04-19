// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import "testing"

func TestNumberOfColors(t *testing.T) {
	n := NumberOfColors()
	if n != 5 {
		t.Errorf("Expected NumberOfColors() to return 5, got %v", n)
	}
}

func TestNumberOfSystems(t *testing.T) {
	n := NumberOfSystems()
	if n != 3 {
		t.Errorf("Expected NumberOfSystems() to return 3, got %v", n)
	}
}
