// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

func TestCalculateTax(t *testing.T) {
	tax, err := CalculateTax("120.00", "0.2")
	if err != nil {
		t.Fatal("Expected no error, got", err)
	}

	if tax != "24.00" {
		t.Fatal("Expected tax to be 24.00, got", tax)
	}
}
