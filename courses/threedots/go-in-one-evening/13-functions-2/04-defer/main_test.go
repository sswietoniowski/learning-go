// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"errors"
	"testing"
)

func TestExecute(t *testing.T) {
	metrics := &Metrics{}

	err := Execute(func() error {
		return nil
	}, metrics)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(metrics.execution) != 1 {
		t.Errorf("Expected 1 execution metric, got %v", len(metrics.execution))
	}
	if len(metrics.success) != 1 {
		t.Errorf("Expected 1 success metric, got %v", len(metrics.success))
	}
	if len(metrics.failure) != 0 {
		t.Errorf("Expected no failure metrics, got %v", len(metrics.failure))
	}

	err = Execute(func() error {
		return errors.New("error")
	}, metrics)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	if len(metrics.execution) != 2 {
		t.Errorf("Expected 2 execution metrics, got %v", len(metrics.execution))
	}
	if len(metrics.success) != 1 {
		t.Errorf("Expected 1 success metric, got %v", len(metrics.success))
	}
	if len(metrics.failure) != 1 {
		t.Errorf("Expected 1 failure metric, got %v", len(metrics.failure))
	}
}
