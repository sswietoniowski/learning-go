// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

func TestDescribeNumber(t *testing.T) {
	if DescribeNumber(-1) != "negative" {
		t.Error("-1 is negative")
	}

	if DescribeNumber(0) != "zero or more" {
		t.Error("0 is zero or more")
	}

	if DescribeNumber(1) != "zero or more" {
		t.Error("1 is zero or more")
	}
}
