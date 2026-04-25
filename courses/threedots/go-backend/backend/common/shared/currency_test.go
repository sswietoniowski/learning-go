// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package shared_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eats/backend/common/shared"
)

func TestMustNewCurrency(t *testing.T) {
	validCurrencies := []string{"USD", "EUR", "GBP", "JPY", "PLN"}

	for _, code := range validCurrencies {
		t.Run(code, func(t *testing.T) {
			currency := shared.MustNewCurrency(code)
			assert.Equal(t, code, currency.Code())
			assert.False(t, currency.IsZero())
		})
	}
}

func TestMustNewCurrency_InvalidValue(t *testing.T) {
	assert.Panics(t, func() {
		shared.MustNewCurrency("INVALID")
	})
}

func TestCurrency_ZeroValue(t *testing.T) {
	var c shared.Currency
	assert.True(t, c.IsZero())
	assert.Equal(t, "", c.Code())
}

func TestCurrency_InSharedTypes(t *testing.T) {
	found := false
	for _, st := range shared.SharedTypes {
		if _, ok := st.(shared.Currency); ok {
			found = true
			break
		}
	}
	require.True(t, found, "Currency{} should be in SharedTypes slice")
}
