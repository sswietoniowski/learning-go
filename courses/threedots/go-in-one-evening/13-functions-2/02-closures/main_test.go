// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

func TestWordGenerator_Protocols(t *testing.T) {
	protocols := []string{"HTTP", "FTP", "SSH"}
	expected := []string{"HTTP", "FTP", "SSH", "HTTP", "FTP", "SSH", "HTTP", "FTP"}
	g := WordGenerator(protocols)

	for i := 0; i < len(expected); i++ {
		if g() != expected[i] {
			t.Errorf("Expected %s at call %v, got %s", expected[i], i, g())
		}
	}
}

func TestWordGenerator_Words(t *testing.T) {
	words := []string{"foo", "bar", "baz"}
	expected := []string{"foo", "bar"}
	g := WordGenerator(words)

	for i := 0; i < len(expected); i++ {
		if g() != expected[i] {
			t.Errorf("Expected %s at call %v, got %s", expected[i], i, g())
		}
	}
}
