// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"sort"
	"testing"
)

var testProducts = map[int]string{
	1: "Book",
	2: "Video Course",
	3: "Lecture",
	4: "Talk",
	5: "Training",
}

func TestKeys(t *testing.T) {
	k := Keys(testProducts)
	assertIntElementsMatch(t, []int{1, 2, 3, 4, 5}, k)
}

func TestValues(t *testing.T) {
	v := Values(testProducts)
	assertStringElementsMatch(t, []string{"Book", "Video Course", "Lecture", "Talk", "Training"}, v)
}

func assertIntElementsMatch(t *testing.T, expected []int, actual []int) {
	if len(expected) != len(actual) {
		t.Errorf("Expected length %v, got %v", len(expected), len(actual))
	}

	sort.Ints(expected)
	sort.Ints(actual)

	for i, e := range expected {
		a := actual[i]

		if e != a {
			t.Errorf("Expected %v at index %v, got %v", e, i, a)
		}
	}
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
