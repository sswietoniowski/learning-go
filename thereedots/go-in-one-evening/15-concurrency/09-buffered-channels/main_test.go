// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import "testing"

func TestCalculatePayments(t *testing.T) {
	payments := []int{100, 200, 50, 150, 200, 400}
	r, err := CalculatePayments(payments)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if r != 1100 {
		t.Errorf("Expected 1100, got %v", r)
	}

	payments = []int{100, 200, 50, 150, 200, 400, 100, 100, 100, 100}
	r, err = CalculatePayments(payments)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if r != 1500 {
		t.Errorf("Expected 1500, got %v", r)
	}
}
