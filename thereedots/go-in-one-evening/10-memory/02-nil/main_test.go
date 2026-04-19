// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

func TestUpdatePassword(t *testing.T) {
	assertEqual(t, "Alice", Name, "Incorrect Name")
	assertEqual(t, "top-secret", Password, "Incorrect Password")

	UpdateUser(nil, nil)

	assertEqual(t, "Alice", Name, "Expected unchanged Name")
	assertEqual(t, "top-secret", Password, "Expected unchanged Password")

	newPassword := "much-more-secure"
	UpdateUser(nil, &newPassword)

	assertEqual(t, "Alice", Name, "Expected unchanged Name")
	assertEqual(t, newPassword, Password, "Incorrect Password after updating")

	newName := "Bob"
	UpdateUser(&newName, &newPassword)

	assertEqual(t, newName, Name, "Incorrect Name after updating")
	assertEqual(t, newPassword, Password, "Incorrect Password after updating")
}

func assertEqual(t *testing.T, expected string, actual string, message string) {
	if expected != actual {
		t.Errorf("%v: expected %v, got %v instead", message, expected, actual)
	}
}
