// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

func TestSum(t *testing.T) {
	if Sum(1, 2, 3, 4, 5) != 15 {
		t.Error("Sum(1, 2, 3, 4, 5) != 15")
	}

	if Sum(9, 8, 7, 6, 5) != 35 {
		t.Error("Sum(9, 8, 7, 6, 5) != 35")
	}

	if Sum(0, 0, 0, 0, 0) != 0 {
		t.Error("Sum(0, 0, 0, 0, 0) != 0")
	}
}
