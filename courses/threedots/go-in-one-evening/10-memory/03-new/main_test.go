// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

func TestAllocateBuffers(t *testing.T) {
	for i := 0; i < 3; i++ {
		b := AllocateBuffer()

		if b == nil || *b != "" {
			t.Errorf("Expected a pointer to an empty string, got %#v instead", b)
		}
	}

	b := AllocateBuffer()
	if b != nil {
		t.Errorf("Expected nil on the 4th call, got %#v instead", b)
	}
}
