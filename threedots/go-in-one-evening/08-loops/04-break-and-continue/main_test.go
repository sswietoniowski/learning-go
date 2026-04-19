// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

func TestCountCreatedEvents(t *testing.T) {
	events := []string{
		"product_created",
		"product_updated",
		"product_created",
		"product_updated",
		"product_assigned",
		"product_created",
		"order_created",
		"order_created",
		"order_updated",
		"client_created",
		"client_updated",
		"client_refreshed",
		"client_deleted",
		"order_updated",
		"order_created",
		"order_created",
		"order_updated",
		"client_created",
		"client_updated",
	}

	created := CountCreatedEvents(events)
	if created == 9 {
		t.Errorf("Expected 6 created events, got 9. Did you forget to break after a deleted event?")
	} else if created != 6 {
		t.Errorf("Expected 6 created events, got %v", created)
	}
}
