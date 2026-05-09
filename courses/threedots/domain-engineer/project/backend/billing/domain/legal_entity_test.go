// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eats/backend/billing/domain"
	"eats/backend/common/shared"
)

func TestNewLegalEntity_BusinessEntityWithTaxID(t *testing.T) {
	taxID, err := shared.NewTaxID("1234567890")
	require.NoError(t, err)

	entity, err := domain.NewLegalEntity("Food Delivery Inc.", newTestAddress(t), &taxID)

	require.NoError(t, err)
	assert.Equal(t, "Food Delivery Inc.", entity.Name())
	require.NotNil(t, entity.TaxID())
	assert.Equal(t, "1234567890", entity.TaxID().String())
}

func TestNewLegalEntity_IndividualWithoutTaxID(t *testing.T) {
	entity, err := domain.NewLegalEntity("John Doe", newTestAddress(t), nil)

	require.NoError(t, err)
	assert.Equal(t, "John Doe", entity.Name())
	assert.Nil(t, entity.TaxID())
}

func TestNewLegalEntity_EmptyNameRejected(t *testing.T) {
	_, err := domain.NewLegalEntity("", newTestAddress(t), nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "name can't be empty")
}

func TestNewLegalEntity_EmptyAddressRejected(t *testing.T) {
	_, err := domain.NewLegalEntity("Jane Doe", shared.Address{}, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "address can't be empty")
}

func TestLegalEntity_IsZero(t *testing.T) {
	assert.True(t, domain.LegalEntity{}.IsZero())

	entity, err := domain.NewLegalEntity("John Doe", newTestAddress(t), nil)
	require.NoError(t, err)

	assert.False(t, entity.IsZero())
}

func newTestAddress(t *testing.T) shared.Address {
	t.Helper()

	addr, err := shared.NewAddress("123 Main St", "", "12345", "New York", shared.MustNewCountryCode("US"))
	require.NoError(t, err)

	return addr
}
