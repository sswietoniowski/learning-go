// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

func TestAlphabet(t *testing.T) {
	assertEqual(t, []string{"a"}, Alphabet(1))
	assertEqual(t, []string{"a", "b"}, Alphabet(2))
	assertEqual(t, []string{"a", "b", "c", "d", "e", "f", "g", "h"}, Alphabet(8))
}

func assertEqual(t *testing.T, expected []string, actual []string) {
	if len(expected) != len(actual) {
		t.Errorf("Expected length %v, got %v", len(expected), len(actual))
	}

	for i, e := range expected {
		a := actual[i]

		if e != a {
			t.Errorf("Expected %v at index %v, got %v", e, i, a)
		}
	}
}
