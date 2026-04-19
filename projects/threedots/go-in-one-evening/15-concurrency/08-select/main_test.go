// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"errors"
	"testing"
	"time"
)

func TestRunWithTimeout(t *testing.T) {
	err := RunWithTimeout(func(errChan chan error) {
		errChan <- nil
	})
	if err != nil {
		t.Error("Expected nil error, got", err)
	}

	err = RunWithTimeout(func(errChan chan error) {
		errChan <- errors.New("some error")
	})
	if err == nil || err.Error() != "some error" {
		t.Error("Expected 'some error', got", err)
	}

	err = RunWithTimeout(func(errChan chan error) {
		time.Sleep(time.Second * 2)
		errChan <- nil
	})
	if err == nil || err.Error() != "timeout" {
		t.Error("Expected 'timeout', got", err)
	}
}
