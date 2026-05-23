// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"errors"
	"testing"
)

type mockPaymentsClient struct {
	details PaymentDetails
	err     error
}

func (m *mockPaymentsClient) GetPaymentDetails(nonce string) (PaymentDetails, error) {
	return m.details, m.err
}

func validFactory(t *testing.T) OrderFactory {
	t.Helper()
	client := &mockPaymentsClient{
		details: PaymentDetails{Amount: 1000, Currency: "USD"},
	}
	return NewOrderFactory(client)
}

func TestNewOrderFactory(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		client := &mockPaymentsClient{}
		factory := NewOrderFactory(client)
		_ = factory
	})

	t.Run("nil_client_panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected NewOrderFactory to panic with nil client")
			}
		}()
		NewOrderFactory(nil)
	})
}

func TestNewOrder(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		factory := validFactory(t)

		order, err := factory.NewOrder("abc123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.Nonce() != "abc123" {
			t.Errorf("got Nonce() = %q, want %q", order.Nonce(), "abc123")
		}
		if order.Amount() != 1000 {
			t.Errorf("got Amount() = %d, want %d", order.Amount(), 1000)
		}
		if order.Currency() != "USD" {
			t.Errorf("got Currency() = %q, want %q", order.Currency(), "USD")
		}
	})

	t.Run("empty_nonce", func(t *testing.T) {
		factory := validFactory(t)

		_, err := factory.NewOrder("")
		if err == nil {
			t.Error("expected NewOrder to reject an empty nonce")
		}
	})

	t.Run("client_error", func(t *testing.T) {
		client := &mockPaymentsClient{
			err: errors.New("service unavailable"),
		}
		factory := NewOrderFactory(client)

		_, err := factory.NewOrder("abc123")
		if err == nil {
			t.Error("expected NewOrder to return error when client fails")
		}
	})

	t.Run("non_positive_amount", func(t *testing.T) {
		client := &mockPaymentsClient{
			details: PaymentDetails{Amount: 0, Currency: "USD"},
		}
		factory := NewOrderFactory(client)

		_, err := factory.NewOrder("abc123")
		if err == nil {
			t.Error("expected NewOrder to reject a non-positive amount")
		}
	})
}
