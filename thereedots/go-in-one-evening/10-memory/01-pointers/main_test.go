// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

func TestUser(t *testing.T) {
	firstOrder := Order{
		Products: []int{545, 490},
	}

	secondOrder := Order{
		Products: []int{98, 829, 245},
	}

	_ = User{
		Name:   "Alice",
		Orders: []*Order{&firstOrder, &secondOrder},
	}
}
