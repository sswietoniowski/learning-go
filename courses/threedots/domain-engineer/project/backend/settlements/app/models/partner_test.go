// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eats/backend/common"
	"eats/backend/settlements/app/models"
	"eats/backend/settlements/domain"
)

func TestNewPartner(t *testing.T) {
	uuid, kind, name, taxID, address, iban, currency := validLegalEntityArgs(t)

	legalEntity, err := models.NewLegalEntity(uuid, kind, name, taxID, address, iban, currency)
	require.NoError(t, err)

	platformUUID := models.PlatformEntityUUID{
		LegalEntityUUID: domain.LegalEntityUUID{UUID: common.NewUUIDv7()},
	}

	partner := models.NewPartner(legalEntity, platformUUID)

	assert.Equal(t, legalEntity, partner.LegalEntity)
	assert.Equal(t, platformUUID, partner.PlatformEntityUUID)
}
