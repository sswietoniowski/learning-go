// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"sort"
	"testing"
)

func TestDeduplicate(t *testing.T) {
	testAddresses := []string{
		"127.0.0.1",
		"10.0.0.1",
		"10.0.0.1",
		"127.0.0.1",
		"10.0.0.2",
	}

	Deduplicate(&testAddresses)

	assertStringElementsMatch(t, []string{"127.0.0.1", "10.0.0.1", "10.0.0.2"}, testAddresses)
}

func assertStringElementsMatch(t *testing.T, expected []string, actual []string) {
	if len(expected) != len(actual) {
		t.Errorf("Expected length %v, got %v", len(expected), len(actual))
	}

	sort.Strings(expected)
	sort.Strings(actual)

	for i, e := range expected {
		a := actual[i]

		if e != a {
			t.Errorf("Expected %v at index %v, got %v", e, i, a)
		}
	}
}
