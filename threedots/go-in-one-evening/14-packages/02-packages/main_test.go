// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"

	"shop/money"
)

func TestPackages(t *testing.T) {
	m := money.New(1000, "EUR")

	if m.Amount != 1000 {
		t.Error("Expected 1000, got ", m.Amount)
	}

	if m.Currency != "EUR" {
		t.Error("Expected EUR, got ", m.Currency)
	}
}
