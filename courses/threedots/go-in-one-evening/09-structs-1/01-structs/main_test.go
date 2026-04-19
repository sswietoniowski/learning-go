// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

func TestPoint(t *testing.T) {
	p := Point{100, 200}

	if p.X != 100 && p.Y != 20 {
		t.Error("Point{100, 200} has unexpected fields")
	}
}
