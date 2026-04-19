// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
	"time"
)

func TestDateRange(t *testing.T) {
	lifetime, err := NewDateRange(
		time.Date(1815, 12, 10, 0, 0, 0, 0, time.UTC),
		time.Date(1852, 11, 27, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := 324048.0

	if lifetime.Hours() != expected {
		t.Errorf("Expected hours to be %v, got %v instead", expected, lifetime.Hours())
	}

	_, err = NewDateRange(
		time.Date(1852, 11, 27, 0, 0, 0, 0, time.UTC),
		time.Date(1815, 12, 10, 0, 0, 0, 0, time.UTC),
	)
	if err == nil {
		t.Fatalf("Expected error on end before start, got nil")
	}

	_, err = NewDateRange(
		time.Time{},
		time.Date(1852, 11, 27, 0, 0, 0, 0, time.UTC),
	)
	if err == nil {
		t.Fatalf("Expected error on empty start, got nil")
	}

	_, err = NewDateRange(
		time.Date(1815, 12, 10, 0, 0, 0, 0, time.UTC),
		time.Time{},
	)
	if err == nil {
		t.Fatalf("Expected error on empty end, got nil")
	}

	_, err = NewDateRange(
		time.Time{},
		time.Time{},
	)
	if err == nil {
		t.Fatalf("Expected error on empty start and end, got nil")
	}
}
