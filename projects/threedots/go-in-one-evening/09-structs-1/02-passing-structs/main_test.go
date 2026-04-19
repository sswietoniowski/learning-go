// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

func TestArea(t *testing.T) {
	r := Rectangle{
		Width:  150,
		Length: 30,
	}
	area := Area(r)
	expected := 4500

	if area != expected {
		t.Errorf("Expected area %v of Rectangle%+v, got %v", expected, r, area)
	}

	r = Rectangle{
		Width:  20,
		Length: 20,
	}
	area = Area(r)
	expected = 400

	if area != expected {
		t.Errorf("Expected area %v of Rectangle%+v, got %v", expected, r, area)
	}
}
