// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

func TestGlobals(t *testing.T) {
	if taxRate != 0.1 {
		t.Error("expected taxRate to be equal 0.1")
	}
}
