package tests_test

import (
	"eats/backend/common/testutils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComponent_CriticalFlow(t *testing.T) {
	t.Parallel()

	clients := newTestClients(t)
	ctx := t.Context()
	country := testutils.GenerateRandomCountry()
	customerUUID := registerCustomerInCity(ctx, t, clients, country, "Some city")
	assert.NotEmpty(t, customerUUID)
}
