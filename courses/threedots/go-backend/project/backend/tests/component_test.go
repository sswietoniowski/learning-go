// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package tests_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eats/backend/common/testutils"
)

func TestComponent_CriticalFlow(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()

	country := testutils.GenerateRandomCountry()

	customerUUID := registerCustomerInCity(ctx, t, clients, country, "Some city")
	assert.NotEmpty(t, customerUUID)
}

func TestComponent_ListMenuItems(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()
	country := testutils.GenerateRandomCountry()

	// Onboard a restaurant with menu items
	restaurantUUID, menuItems := onboardRestaurant(ctx, t, clients, country, "Test Restaurant")
	require.NotEmpty(t, restaurantUUID)
	require.NotEmpty(t, menuItems)

	// Call the read model endpoint
	resp, err := clients.Orders.ListMenuItemsWithResponse(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.NotNil(t, resp.JSON200)

	// Verify our menu items are in the response
	items := *resp.JSON200
	found := 0
	for _, item := range items {
		for _, expected := range menuItems {
			if item.MenuItemUuid == expected.Uuid {
				assert.Equal(t, expected.Name, item.MenuItemName)
				assert.Equal(t, "Test Restaurant", item.RestaurantName)
				found++
			}
		}
	}
	assert.Equal(t, len(menuItems), found, "all menu items should be returned by read model")
}
