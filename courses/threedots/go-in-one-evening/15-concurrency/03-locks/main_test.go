// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"fmt"
	"testing"
	"time"
)

func TestAddUser(t *testing.T) {
	storage := &Storage{users: make(map[string]User)}

	emails := 1000

	for i := 0; i < emails; i++ {
		go storage.AddUser(fmt.Sprintf("joe-%v@example.com", i))
	}

	time.Sleep(time.Millisecond * 1500)

	if len(storage.users) != emails {
		t.Error("Expected", emails, "users, got", len(storage.users))
	}
}
