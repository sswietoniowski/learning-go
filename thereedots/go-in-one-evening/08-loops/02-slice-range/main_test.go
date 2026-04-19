// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

func TestSum(t *testing.T) {
	s := Sum(10, 20, 30)
	if s != 60 {
		t.Errorf("Expected Sum(10, 20, 30) to be 60, got %v", s)
	}

	s = Sum(10)
	if s != 10 {
		t.Errorf("Expected Sum(10) to be 10, got %v", s)
	}

	s = Sum()
	if s != 0 {
		t.Errorf("Expected Sum(0) to be 0, got %v", s)
	}

	s = Sum(1, 1, 1, 1, 1)
	if s != 5 {
		t.Errorf("Expected Sum(1, 1, 1, 1, 1) to be 5, got %v", s)
	}
}
