// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package conditionals

import "testing"

func TestGetSecret(t *testing.T) {
	if Password != "current-password" {
		t.Error("Expected no changes in Password")
	}

	ResetPassword(0)
	if Password != "current-password" {
		t.Error("Expected no changes in Password")
	}

	ResetPassword(1000)
	if Password != "current-password" {
		t.Error("Expected no changes in Password")
	}

	ResetPassword(2022)
	if Password != "new-password" {
		t.Error("Expected a new Password")
	}
}
