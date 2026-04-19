// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package account

import (
	"testing"
)

func TestAccount(t *testing.T) {
	acc, err := New("joe@example.com", "secret")
	if err != nil {
		t.Fatal(err)
	}

	acc.email = ""
	acc.password = ""
}
