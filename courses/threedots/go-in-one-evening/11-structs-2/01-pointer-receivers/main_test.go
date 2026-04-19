// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

func TestPosition(t *testing.T) {
	p := Position{X: 50, Y: 75}
	p.Move(25, -25)

	if p.X == 50 && p.Y == 75 {
		t.Fatalf("Position didn't change. Did you use a pointer receiver? (*Position)")
	}

	if p.X != 75 && p.Y != 50 {
		t.Error("Position{50, 75} after Move(25, -25) should be Position{75, 50}")
	}
}
