// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

func TestDescribeNumber(t *testing.T) {
	if DescribeNumber(-1) != "negative" {
		t.Error("-1 should be negative")
	}

	if DescribeNumber(0) != "zero" {
		t.Error("0 should be zero")
	}

	if DescribeNumber(1) != "positive" {
		t.Error("1 should be positive")
	}
}
