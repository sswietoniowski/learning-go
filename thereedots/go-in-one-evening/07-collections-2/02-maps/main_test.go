// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

func TestStats(t *testing.T) {
	c := Stats["create"]
	if c != 0 {
		t.Error("expected no users created")
	}
	u := Stats["update"]
	if u != 0 {
		t.Error("expected no users updated")
	}

	CreateUser("user")
	CreateUser("user")
	CreateUser("user")

	c = Stats["create"]
	if c != 3 {
		t.Errorf("expected 3 users created, have %d", c)
	}

	UpdateUser("user")
	UpdateUser("user")

	u = Stats["update"]
	if u != 2 {
		t.Errorf("expected 2 users updated, have %d", u)
	}

	PurgeStats()
	c = Stats["create"]
	if c != 0 {
		t.Error("expected no users created")
	}
	u = Stats["update"]
	if u != 0 {
		t.Error("expected no users updated")
	}
}
