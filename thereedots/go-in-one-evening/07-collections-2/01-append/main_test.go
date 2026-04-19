// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

func TestAddUser(t *testing.T) {
	assertEqual(t, []string{}, users)
	AddUser("Alice")
	assertEqual(t, []string{"Alice"}, users)
	AddUser("Bob")
	assertEqual(t, []string{"Alice", "Bob"}, users)
	AddUser("Charlie")
	AddUser("Dan")
	assertEqual(t, []string{"Alice", "Bob", "Charlie", "Dan"}, users)
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
