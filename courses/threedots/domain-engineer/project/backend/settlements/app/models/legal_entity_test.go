// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/settlements/app/models"
	"eats/backend/settlements/domain"
)

func validLegalEntityArgs(t *testing.T) (
	domain.LegalEntityUUID,
	models.LegalEntityType,
	string,
	shared.TaxID,
	shared.Address,
	domain.IBAN,
	shared.Currency,
) {
	t.Helper()

	uuid := domain.LegalEntityUUID{UUID: common.NewUUIDv7()}

	taxID, err := shared.NewTaxID("PL1234567890")
	require.NoError(t, err)

	address, err := shared.NewAddress("Main Street 1", "", "00-001", "Warsaw", shared.MustNewCountryCode("PL"))
	require.NoError(t, err)

	iban, err := domain.NewIBAN("DE89370400440532013000")
	require.NoError(t, err)

	return uuid, models.LegalEntityPlatform, "Acme Inc.", taxID, address, iban, shared.MustNewCurrency("EUR")
}

func TestNewLegalEntity_Valid(t *testing.T) {
	uuid, kind, name, taxID, address, iban, currency := validLegalEntityArgs(t)

	entity, err := models.NewLegalEntity(uuid, kind, name, taxID, address, iban, currency)
	require.NoError(t, err)

	assert.Equal(t, uuid, entity.UUID)
	assert.Equal(t, kind, entity.Type)
	assert.Equal(t, name, entity.BusinessName)
	assert.Equal(t, taxID, entity.TaxID)
	assert.Equal(t, address, entity.Address)
	assert.Equal(t, iban, entity.BankAccountNumber)
	assert.Equal(t, currency, entity.Currency)
}

func TestNewLegalEntity_RejectsZeroFields(t *testing.T) {
	t.Run("zero_uuid", func(t *testing.T) {
		_, kind, name, taxID, address, iban, currency := validLegalEntityArgs(t)
		_, err := models.NewLegalEntity(domain.LegalEntityUUID{}, kind, name, taxID, address, iban, currency)
		require.Error(t, err)
	})

	t.Run("zero_type", func(t *testing.T) {
		uuid, _, name, taxID, address, iban, currency := validLegalEntityArgs(t)
		_, err := models.NewLegalEntity(uuid, models.LegalEntityType{}, name, taxID, address, iban, currency)
		require.Error(t, err)
	})

	t.Run("empty_business_name", func(t *testing.T) {
		uuid, kind, _, taxID, address, iban, currency := validLegalEntityArgs(t)
		_, err := models.NewLegalEntity(uuid, kind, "", taxID, address, iban, currency)
		require.Error(t, err)
	})

	t.Run("zero_tax_id", func(t *testing.T) {
		uuid, kind, name, _, address, iban, currency := validLegalEntityArgs(t)
		_, err := models.NewLegalEntity(uuid, kind, name, shared.TaxID{}, address, iban, currency)
		require.Error(t, err)
	})

	t.Run("zero_address", func(t *testing.T) {
		uuid, kind, name, taxID, _, iban, currency := validLegalEntityArgs(t)
		_, err := models.NewLegalEntity(uuid, kind, name, taxID, shared.Address{}, iban, currency)
		require.Error(t, err)
	})

	t.Run("zero_iban", func(t *testing.T) {
		uuid, kind, name, taxID, address, _, currency := validLegalEntityArgs(t)
		_, err := models.NewLegalEntity(uuid, kind, name, taxID, address, domain.IBAN{}, currency)
		require.Error(t, err)
	})

	t.Run("zero_currency", func(t *testing.T) {
		uuid, kind, name, taxID, address, iban, _ := validLegalEntityArgs(t)
		_, err := models.NewLegalEntity(uuid, kind, name, taxID, address, iban, shared.Currency{})
		require.Error(t, err)
	})
}
