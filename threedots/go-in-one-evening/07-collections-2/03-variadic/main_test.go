// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

func TestDebugLog(t *testing.T) {
	expected := []string{"[DEBUG]", "Querying database", "INSERT INTO users"}
	actual := DebugLog("Querying database", "INSERT INTO users")

	assertEqual(t, expected, actual)
}

func TestInfoLog(t *testing.T) {
	expected := []string{"[INFO]", "User created:", "42"}
	actual := InfoLog("User created:", "42")

	assertEqual(t, expected, actual)
}

func TestErrorLog(t *testing.T) {
	expected := []string{"[ERROR]", "Could not create user", "unknown error"}
	actual := ErrorLog("Could not create user", "unknown error")

	assertEqual(t, expected, actual)
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
