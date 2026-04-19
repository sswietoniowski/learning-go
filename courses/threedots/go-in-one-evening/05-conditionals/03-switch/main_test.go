// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"strings"
	"testing"
)

func TestDirection(t *testing.T) {
	testCases := []struct {
		Input  string
		Output string
	}{
		{
			Input:  "N",
			Output: "north",
		},
		{
			Input:  "E",
			Output: "east",
		},
		{
			Input:  "S",
			Output: "south",
		},
		{
			Input:  "W",
			Output: "west",
		},
		{
			Input:  "A",
			Output: "unknown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Input, func(t *testing.T) {
			v := Direction(tc.Input)
			if strings.ToLower(v) != strings.ToLower(tc.Output) {
				t.Error("Expected", tc.Output, "for", tc.Input, "got", v)
			}
		})
	}
}
