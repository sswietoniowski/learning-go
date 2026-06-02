// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eats/backend/common"
	"eats/backend/settlements/domain"
)

func TestNewInitialBillingCycle_Valid(t *testing.T) {
	partnerUUID := domain.LegalEntityUUID{UUID: common.NewUUIDv7()}
	partnerType := domain.PartnerTypeRestaurant

	cycle, err := domain.NewInitialBillingCycle(partnerUUID, partnerType)
	require.NoError(t, err)
	require.NotNil(t, cycle)

	assert.Equal(t, partnerUUID, cycle.PartnerUUID())
	assert.Equal(t, partnerType, cycle.PartnerType())
	assert.Equal(t, 1, cycle.Number())
	assert.False(t, cycle.Closed())
	assert.False(t, cycle.Settled())
	assert.WithinDuration(t, time.Now().UTC(), cycle.StartDate(), 5*time.Second)
}

func TestNewInitialBillingCycle_RejectsZeroPartnerUUID(t *testing.T) {
	_, err := domain.NewInitialBillingCycle(domain.LegalEntityUUID{}, domain.PartnerTypeRestaurant)
	require.Error(t, err)
}

func TestNewInitialBillingCycle_RejectsZeroPartnerType(t *testing.T) {
	partnerUUID := domain.LegalEntityUUID{UUID: common.NewUUIDv7()}
	_, err := domain.NewInitialBillingCycle(partnerUUID, domain.PartnerType{})
	require.Error(t, err)
}

func TestNewNextBillingCycle_RejectsNilPrevious(t *testing.T) {
	_, err := domain.NewNextBillingCycle(nil)
	require.Error(t, err)
}

func TestNewNextBillingCycle_RejectsOpenPrevious(t *testing.T) {
	partnerUUID := domain.LegalEntityUUID{UUID: common.NewUUIDv7()}
	open, err := domain.NewInitialBillingCycle(partnerUUID, domain.PartnerTypeRestaurant)
	require.NoError(t, err)

	_, err = domain.NewNextBillingCycle(open)
	require.Error(t, err)
}

func TestNewNextBillingCycle_ValidRollover(t *testing.T) {
	partnerUUID := domain.LegalEntityUUID{UUID: common.NewUUIDv7()}
	partnerType := domain.PartnerTypeCourier
	endDate := time.Now().UTC()

	closed := domain.UnmarshalBillingCycle(
		domain.BillingCycleUUID{UUID: common.NewUUIDv7()},
		partnerUUID,
		partnerType,
		1,
		true,
		false,
		endDate.AddDate(0, 0, -7),
		&endDate,
	)

	next, err := domain.NewNextBillingCycle(closed)
	require.NoError(t, err)
	require.NotNil(t, next)

	assert.Equal(t, partnerUUID, next.PartnerUUID())
	assert.Equal(t, partnerType, next.PartnerType())
	assert.Equal(t, 2, next.Number())
	assert.False(t, next.Closed())
	assert.False(t, next.Settled())
	assert.Equal(t, endDate.AddDate(0, 0, 1), next.StartDate())
}

func TestClose_OpenCycle_Succeeds(t *testing.T) {
	partnerUUID := domain.LegalEntityUUID{UUID: common.NewUUIDv7()}
	cycle, err := domain.NewInitialBillingCycle(partnerUUID, domain.PartnerTypeRestaurant)
	require.NoError(t, err)

	require.False(t, cycle.Closed())
	require.Nil(t, cycle.EndDate())

	err = cycle.Close()
	require.NoError(t, err)

	assert.True(t, cycle.Closed())
	require.NotNil(t, cycle.EndDate())
	assert.WithinDuration(t, time.Now().UTC(), *cycle.EndDate(), 5*time.Second)
}

func TestClose_AlreadyClosed_Fails(t *testing.T) {
	partnerUUID := domain.LegalEntityUUID{UUID: common.NewUUIDv7()}
	cycle, err := domain.NewInitialBillingCycle(partnerUUID, domain.PartnerTypeRestaurant)
	require.NoError(t, err)

	require.NoError(t, cycle.Close())

	err = cycle.Close()
	require.Error(t, err)
}

func TestSettle_ClosedCycle_Succeeds(t *testing.T) {
	partnerUUID := domain.LegalEntityUUID{UUID: common.NewUUIDv7()}
	cycle, err := domain.NewInitialBillingCycle(partnerUUID, domain.PartnerTypeCourier)
	require.NoError(t, err)

	require.NoError(t, cycle.Close())
	require.False(t, cycle.Settled())

	err = cycle.Settle()
	require.NoError(t, err)
	assert.True(t, cycle.Settled())
}

func TestSettle_OpenCycle_Fails(t *testing.T) {
	partnerUUID := domain.LegalEntityUUID{UUID: common.NewUUIDv7()}
	cycle, err := domain.NewInitialBillingCycle(partnerUUID, domain.PartnerTypeRestaurant)
	require.NoError(t, err)

	err = cycle.Settle()
	require.Error(t, err)
}

func TestSettle_AlreadySettled_Fails(t *testing.T) {
	partnerUUID := domain.LegalEntityUUID{UUID: common.NewUUIDv7()}
	cycle, err := domain.NewInitialBillingCycle(partnerUUID, domain.PartnerTypeCourier)
	require.NoError(t, err)

	require.NoError(t, cycle.Close())
	require.NoError(t, cycle.Settle())

	err = cycle.Settle()
	require.Error(t, err)
}

func TestUnmarshalBillingCycle_RoundTrips(t *testing.T) {
	uuid := domain.BillingCycleUUID{UUID: common.NewUUIDv7()}
	partnerUUID := domain.LegalEntityUUID{UUID: common.NewUUIDv7()}
	partnerType := domain.PartnerTypeRestaurant
	startDate := time.Now().UTC().Add(-7 * 24 * time.Hour)
	endDate := time.Now().UTC()

	cycle := domain.UnmarshalBillingCycle(uuid, partnerUUID, partnerType, 3, true, true, startDate, &endDate)

	assert.Equal(t, uuid, cycle.UUID())
	assert.Equal(t, partnerUUID, cycle.PartnerUUID())
	assert.Equal(t, partnerType, cycle.PartnerType())
	assert.Equal(t, 3, cycle.Number())
	assert.True(t, cycle.Closed())
	assert.True(t, cycle.Settled())
	assert.Equal(t, startDate, cycle.StartDate())
	assert.Equal(t, &endDate, cycle.EndDate())
}
