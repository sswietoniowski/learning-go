// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
//go:build integration

package db_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/common/testutils"
	"eats/backend/settlements/adapters/db"
	"eats/backend/settlements/app/models"
	"eats/backend/settlements/domain"
)

func TestLegalEntityRepository_SaveAndLoadPlatformEntity(t *testing.T) {
	ctx := context.Background()
	repo := db.NewLegalEntityRepository(testutils.NewDB(t))

	platform := newPlatformLegalEntity(t)

	t.Run("create", func(t *testing.T) {
		err := repo.SavePlatformEntity(ctx, platform)
		require.NoError(t, err)

		saved, err := repo.LegalEntityByUUID(ctx, platform.UUID)
		require.NoError(t, err)

		assert.Equal(t, platform.UUID, saved.UUID)
		assert.Equal(t, models.LegalEntityPlatform, saved.Type)
		assert.Equal(t, platform.BusinessName, saved.BusinessName)
		assert.Equal(t, platform.TaxID, saved.TaxID)
		assert.Equal(t, platform.Address, saved.Address)
		assert.Equal(t, platform.BankAccountNumber.String(), saved.BankAccountNumber.String())
		assert.Equal(t, platform.Currency, saved.Currency)

	})

	t.Run("update", func(t *testing.T) {
		updatedTaxID, err := shared.NewTaxID("PL987654321")
		require.NoError(t, err)

		updatedAddress, err := shared.NewAddress("Updated 1", "", "00-002", "Cracow", shared.MustNewCountryCode("PL"))
		require.NoError(t, err)

		updatedIBAN, err := domain.NewIBAN("DE77666555444333222111")
		require.NoError(t, err)

		platform.Type = models.LegalEntityPartner // This should NOT be updated
		platform.BusinessName = "Updated name"
		platform.TaxID = updatedTaxID
		platform.Address = updatedAddress
		platform.BankAccountNumber = updatedIBAN
		platform.Currency = shared.MustNewCurrency("USD")

		err = repo.SavePlatformEntity(ctx, platform)
		require.NoError(t, err)

		updated, err := repo.LegalEntityByUUID(ctx, platform.UUID)
		require.NoError(t, err)

		assert.Equal(t, platform.UUID, updated.UUID)
		assert.Equal(t, models.LegalEntityPlatform, models.LegalEntityPlatform)
		assert.Equal(t, platform.BusinessName, updated.BusinessName)
		assert.Equal(t, platform.TaxID, updated.TaxID)
		assert.Equal(t, platform.Address, updated.Address)
		assert.Equal(t, platform.BankAccountNumber.String(), updated.BankAccountNumber.String())
		assert.Equal(t, platform.Currency, updated.Currency)
	})
}

func TestLegalEntityRepository_LegalEntityByUUID_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := db.NewLegalEntityRepository(testutils.NewDB(t))

	missing := domain.LegalEntityUUID{UUID: common.NewUUIDv7()}

	_, err := repo.LegalEntityByUUID(ctx, missing)
	require.Error(t, err)
}

func TestLegalEntityRepository_SavePartner_AtomicSaveAndPartnerByUUID(t *testing.T) {
	ctx := context.Background()
	repo := db.NewLegalEntityRepository(testutils.NewDB(t))

	platform := newPlatformLegalEntity(t)
	require.NoError(t, repo.SavePlatformEntity(ctx, platform))

	partner := newPartner(t, models.PlatformEntityUUID{LegalEntityUUID: platform.UUID})

	err := repo.SavePartner(ctx, partner)
	require.NoError(t, err)

	loaded, err := repo.PartnerByUUID(ctx, partner.LegalEntity.UUID)
	require.NoError(t, err)
	assert.Equal(t, partner.LegalEntity.UUID, loaded.LegalEntity.UUID)
	assert.Equal(t, models.LegalEntityPartner, loaded.LegalEntity.Type)
	assert.Equal(t, partner.PlatformEntityUUID, loaded.PlatformEntityUUID)
}

func TestLegalEntityRepository_SavePartner_RejectsNonPlatformReference(t *testing.T) {
	ctx := context.Background()
	repo := db.NewLegalEntityRepository(testutils.NewDB(t))

	// Reference is a partner-type entity, not a platform.
	platform := newPlatformLegalEntity(t)
	require.NoError(t, repo.SavePlatformEntity(ctx, platform))

	wrongRef := newPartner(t, models.PlatformEntityUUID{LegalEntityUUID: platform.UUID})
	require.NoError(t, repo.SavePartner(ctx, wrongRef))

	// Try to onboard another partner referencing the partner UUID instead of the platform UUID.
	bogus := newPartner(t, models.PlatformEntityUUID{LegalEntityUUID: wrongRef.LegalEntity.UUID})

	err := repo.SavePartner(ctx, bogus)
	require.Error(t, err)

	// Tx must roll back: bogus partner's legal entity should not be persisted.
	_, err = repo.LegalEntityByUUID(ctx, bogus.LegalEntity.UUID)
	require.Error(t, err)
}

func TestLegalEntityRepository_SavePartner_RejectsMissingPlatform(t *testing.T) {
	ctx := context.Background()
	repo := db.NewLegalEntityRepository(testutils.NewDB(t))

	missingPlatform := models.PlatformEntityUUID{LegalEntityUUID: domain.LegalEntityUUID{UUID: common.NewUUIDv7()}}
	partner := newPartner(t, missingPlatform)

	err := repo.SavePartner(ctx, partner)
	require.Error(t, err)

	// Tx must roll back: partner legal entity should not be persisted.
	_, err = repo.LegalEntityByUUID(ctx, partner.LegalEntity.UUID)
	require.Error(t, err)
}

func newPlatformLegalEntity(t *testing.T) models.LegalEntity {
	t.Helper()

	taxID, err := shared.NewTaxID("PL1234567890")
	require.NoError(t, err)

	address, err := shared.NewAddress("Main 1", "", "00-001", "Warsaw", shared.MustNewCountryCode("PL"))
	require.NoError(t, err)

	iban, err := domain.NewIBAN("DE89370400440532013000")
	require.NoError(t, err)

	entity, err := models.NewLegalEntity(
		domain.LegalEntityUUID{UUID: common.NewUUIDv7()},
		models.LegalEntityPlatform,
		"Three Dots Eats Platform",
		taxID,
		address,
		iban,
		shared.MustNewCurrency("EUR"),
	)
	require.NoError(t, err)

	return entity
}

func newPartner(t *testing.T, platformUUID models.PlatformEntityUUID) models.Partner {
	t.Helper()

	taxID, err := shared.NewTaxID("PL9876543210")
	require.NoError(t, err)

	address, err := shared.NewAddress("Side 5", "", "00-002", "Warsaw", shared.MustNewCountryCode("PL"))
	require.NoError(t, err)

	iban, err := domain.NewIBAN("DE89370400440532013111")
	require.NoError(t, err)

	legalEntity, err := models.NewLegalEntity(
		domain.LegalEntityUUID{UUID: common.NewUUIDv7()},
		models.LegalEntityPartner,
		"Mama's Pizzeria",
		taxID,
		address,
		iban,
		shared.MustNewCurrency("EUR"),
	)
	require.NoError(t, err)

	return models.NewPartner(legalEntity, platformUUID)
}
