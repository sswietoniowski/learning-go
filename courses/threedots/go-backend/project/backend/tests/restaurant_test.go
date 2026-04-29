package tests_test

import (
	"net/http"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eats/backend/common"
	"eats/backend/common/testutils"
	ordersclient "eats/backend/orders/api/http/client"
	"eats/backend/orders/app"
)

func TestComponent_OnboardRestaurant(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)
	ctx := t.Context()
	country := testutils.GenerateRandomCountry()

	restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
	resp, err := clients.Orders.OnboardRestaurantWithResponse(
		ctx,
		restaurantUUID,
		&ordersclient.OnboardRestaurantParams{
			OperatorUUID: common.NewUUIDv7(),
		},
		ordersclient.OnboardRestaurant{
			Name:        "Test Restaurant",
			Description: "A test restaurant",
			Address:     testutils.GenerateRandomOpenapiAddress(country),
			MenuItems: []ordersclient.MenuItem{
				{
					Uuid:       app.RestaurantMenuItemUUID{common.NewUUIDv7()},
					Name:       "Test Item",
					GrossPrice: decimal.NewFromFloat(10.0),
					Ordering:   1.0,
				},
			},
			Currency: currencyForCountry(t, country),
		},
	)
	require.NoError(t, err)
	// Restaurant UUID is client-generated and passed as a path parameter,
	// so there is no response body — just a 204.
	require.Equal(t, http.StatusNoContent, resp.StatusCode())
}

func TestComponent_CreateQuote(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)
	ctx := t.Context()
	country := testutils.GenerateRandomCountry()

	restaurant := onboardRestaurant(ctx, t, clients, country)
	customerUUID := registerCustomerInCity(ctx, t, clients, country, restaurant.Data.Address.City)

	orderItems := []ordersclient.OrderItem{
		{
			MenuItemUuid: restaurant.Data.MenuItems[0].Uuid,
			Quantity:     2,
		},
		{
			MenuItemUuid: restaurant.Data.MenuItems[1].Uuid,
			Quantity:     1,
		},
	}
	deliveryAddress := testutils.GenerateOpenapiAddressInCity(country, restaurant.Data.Address.City)

	quote := createQuote(ctx, t, clients, customerUUID, restaurant.UUID, orderItems, deliveryAddress)

	assert.Equal(t, restaurant.Data.Currency, quote.Currency, "quote currency should match restaurant currency")
	assert.False(t, quote.ItemsSubtotalGross.IsZero(), "items subtotal should be non-zero")
	assert.False(t, quote.DeliveryFeeGross.IsZero(), "delivery fee should be non-zero")
	assert.False(t, quote.ServiceFeeGross.IsZero(), "service fee should be non-zero")
	assert.False(t, quote.TotalGross.IsZero(), "total gross should be non-zero")

	expectedTotal := quote.ItemsSubtotalGross.Add(quote.DeliveryFeeGross).Add(quote.ServiceFeeGross)
	assert.True(t, quote.TotalGross.Equal(expectedTotal), "total gross should equal items + delivery + service fees")
}

func TestComponent_IdempotentRestaurantUpsert(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)
	ctx := t.Context()
	country := testutils.GenerateRandomCountry()

	restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
	currency := currencyForCountry(t, country)
	menuItems := []ordersclient.MenuItem{
		{
			Uuid:       app.RestaurantMenuItemUUID{common.NewUUIDv7()},
			Name:       "Burger",
			GrossPrice: decimal.NewFromFloat(12.50),
			Ordering:   1.0,
		},
	}

	first := ordersclient.OnboardRestaurant{
		Name:        "Original Name",
		Description: "Original description",
		Address:     testutils.GenerateRandomOpenapiAddress(country),
		MenuItems:   menuItems,
		Currency:    currency,
	}
	onboardRestaurantWithData(ctx, t, clients, restaurantUUID, first)

	second := ordersclient.OnboardRestaurant{
		Name:        "Updated Name",
		Description: "Updated description",
		Address:     first.Address,
		MenuItems:   menuItems,
		Currency:    currency,
	}
	onboardRestaurantWithData(ctx, t, clients, restaurantUUID, second)

	// Verify the restaurant reflects the updated data.
	assertRestaurantMenuPublished(ctx, t, clients, restaurantUUID, second)
}
