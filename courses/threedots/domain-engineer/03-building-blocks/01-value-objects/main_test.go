// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import "testing"

func TestNewMoney_negativeAmount(t *testing.T) {
	_, err := NewMoney(-1, "USD")
	if err == nil {
		t.Error("expected NewMoney to reject a negative amount")
	}
}

func TestNewMoney_emptyCurrency(t *testing.T) {
	_, err := NewMoney(100, "")
	if err == nil {
		t.Error("expected NewMoney to reject an empty currency")
	}
}

func TestNewMoney_zeroAmount(t *testing.T) {
	_, err := NewMoney(0, "USD")
	if err != nil {
		t.Errorf("zero amount should be valid, got error: %v", err)
	}
}

func TestNewMoney_validInputs(t *testing.T) {
	m, err := NewMoney(100, "USD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Amount() != 100 {
		t.Errorf("got Amount() = %d, want 100", m.Amount())
	}
	if m.Currency() != "USD" {
		t.Errorf("got Currency() = %q, want %q", m.Currency(), "USD")
	}
}

func TestAdd_sameCurrency(t *testing.T) {
	a, _ := NewMoney(100, "USD")
	b, _ := NewMoney(50, "USD")

	sum, err := a.Add(b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sum.Amount() != 150 {
		t.Errorf("got Amount() = %d, want 150", sum.Amount())
	}
	if sum.Currency() != "USD" {
		t.Errorf("got Currency() = %q, want %q", sum.Currency(), "USD")
	}
}

func TestAdd_differentCurrencies(t *testing.T) {
	a, _ := NewMoney(100, "USD")
	b, _ := NewMoney(50, "EUR")

	_, err := a.Add(b)
	if err == nil {
		t.Error("expected Add to return an error when the currencies differ")
	}
}

func TestAdd_doesNotMutateOriginal(t *testing.T) {
	a, _ := NewMoney(100, "USD")
	b, _ := NewMoney(50, "USD")

	_, _ = a.Add(b)

	if a.Amount() != 100 {
		t.Errorf("a was mutated after Add: got Amount() = %d, want 100", a.Amount())
	}
}
