// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

func TestMerge(t *testing.T) {
	a := []string{"a", "b", "c"}
	b := []string{"d", "e", "f"}
	expected := []string{"a", "b", "c", "d", "e", "f"}

	m := Merge(a, b)

	assertEqual(t, expected, m)
}

func TestRemove(t *testing.T) {
	input := []string{"a", "b", "c", "d", "e", "f"}
	expected := []string{"a", "b", "c", "e", "f"}

	r := Remove(input, 3)

	assertEqual(t, expected, r)
}

func TestRemoveLast(t *testing.T) {
	input := []string{"a", "b", "c", "d", "e", "f"}
	expected := []string{"a", "b", "c", "d", "e"}

	r := RemoveLast(input)

	assertEqual(t, expected, r)
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
