// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package shared_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"eats/backend/common/shared"
)

func TestNewNetAmount(t *testing.T) {
	amount := shared.NewNetAmount(decimal.NewFromInt(100))

	assert.True(t, amount.Amount().Equal(decimal.NewFromInt(100)))
	assert.True(t, amount.IsNet())
	assert.False(t, amount.IsGross())
}

func TestNewGrossAmount(t *testing.T) {
	amount := shared.NewGrossAmount(decimal.NewFromInt(123))

	assert.True(t, amount.Amount().Equal(decimal.NewFromInt(123)))
	assert.True(t, amount.IsGross())
	assert.False(t, amount.IsNet())
}
