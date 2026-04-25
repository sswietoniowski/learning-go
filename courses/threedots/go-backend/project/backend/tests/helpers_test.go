// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package tests_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	gofakeit "github.com/brianvoe/gofakeit/v7"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/common/testutils"
	ordersclient "eats/backend/orders/api/http/client"
	"eats/backend/orders/app"
)

func registerCustomer(ctx context.Context, t *testing.T, clients testClients, country shared.CountryCode) ordersclient.CustomerUUID {
	t.Helper()

	customerToCreate := ordersclient.RegisterCustomer{
		Name:        gofakeit.Name(),
		Email:       openapi_types.Email(gofakeit.Email()),
		Address:     testutils.GenerateRandomOpenapiAddress(country),
		PhoneNumber: gofakeit.Phone(),
	}

	resp, err := clients.Orders.RegisterCustomerWithResponse(ctx, customerToCreate)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode())
	require.NotNil(t, resp.JSON201)

	return resp.JSON201.CustomerUuid
}

func registerCustomerInCity(ctx context.Context, t *testing.T, clients testClients, country shared.CountryCode, city string) ordersclient.CustomerUUID {
	t.Helper()

	customerToCreate := ordersclient.RegisterCustomer{
		Name:        gofakeit.Name(),
		Email:       openapi_types.Email(gofakeit.Email()),
		Address:     testutils.GenerateOpenapiAddressInCity(country, city),
		PhoneNumber: gofakeit.Phone(),
	}

	resp, err := clients.Orders.RegisterCustomerWithResponse(ctx, customerToCreate)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode())
	require.NotNil(t, resp.JSON201)

	return resp.JSON201.CustomerUuid
}

func assertJsonReprEqual(t *testing.T, expected, actual any) {
	t.Helper()

	expectedJSON, err := json.Marshal(expected)
	require.NoError(t, err)

	actualJSON, err := json.Marshal(actual)
	require.NoError(t, err)

	require.JSONEq(t, string(expectedJSON), string(actualJSON))
}

func onboardRestaurant(ctx context.Context, t *testing.T, clients testClients, country shared.CountryCode, name string) (ordersclient.RestaurantUUID, []ordersclient.MenuItem) {
	t.Helper()

	restaurantUUID := app.RestaurantUUID{UUID: common.NewUUIDv7()}

	menuItems := []ordersclient.MenuItem{
		{
			Uuid:       app.RestaurantMenuItemUUID{UUID: common.NewUUIDv7()},
			Name:       gofakeit.ProductName(),
			GrossPrice: decimal.NewFromFloat(10.99),
			Ordering:   1,
		},
		{
			Uuid:       app.RestaurantMenuItemUUID{UUID: common.NewUUIDv7()},
			Name:       gofakeit.ProductName(),
			GrossPrice: decimal.NewFromFloat(15.50),
			Ordering:   2,
		},
	}

	currency := shared.MustNewCurrency("USD")

	resp, err := clients.Orders.OnboardRestaurantWithResponse(
		ctx,
		restaurantUUID,
		&ordersclient.OnboardRestaurantParams{
			OperatorUUID: common.NewUUIDv7(),
		},
		ordersclient.OnboardRestaurant{
			Name:        name,
			Address:     testutils.GenerateRandomOpenapiAddress(country),
			Currency:    currency,
			Description: gofakeit.Sentence(10),
			MenuItems:   menuItems,
		},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode())

	return restaurantUUID, menuItems
}
