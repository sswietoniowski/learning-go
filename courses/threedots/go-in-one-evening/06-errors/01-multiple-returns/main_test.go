// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

func TestMultipleReturns(t *testing.T) {
	y, x := Swap("x", "y")
	if y != "y" || x != "x" {
		t.Error("Swap(x, y) != (y, x)")
	}

	b, a := Swap("a", "b")
	if b != "b" || a != "a" {
		t.Error("Swap(a, b) != (b, a)")
	}
}
