// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"testing"
)

type CustomStorage struct {
	user User
}

func (c *CustomStorage) Store(user User) {
	c.user = user
}

func TestNewUser(t *testing.T) {
	customStorage := &CustomStorage{}

	NewUser("John", customStorage)
	if customStorage.user.Name != "John" {
		t.Errorf("Expected user name to be 'John', got '%s'", customStorage.user.Name)
	}

	NewUser("Anna", customStorage)
	if customStorage.user.Name != "Anna" {
		t.Errorf("Expected user name to be 'Anna', got '%s'", customStorage.user.Name)
	}
}
