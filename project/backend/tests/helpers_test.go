// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package tests_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	bank2 "github.com/ThreeDotsLabs/the-domain-engineer/clients/bank"
	gofakeit "github.com/brianvoe/gofakeit/v7"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/common/testutils"
	ordersclient "eats/backend/orders/api/http/client"
	"eats/backend/orders/app"
)

func assertRestaurantMenuPublished(ctx context.Context, t *testing.T, clients testClients, restaurantUUID app.RestaurantUUID, restaurant ordersclient.OnboardRestaurant) {
	t.Helper()

	resp, err := clients.Orders.CustomerGetRestaurantMenuWithResponse(ctx, restaurantUUID, &ordersclient.CustomerGetRestaurantMenuParams{
		CustomerUUID: app.CustomerUUID{common.NewUUIDv7()},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	require.Empty(
		t,
		cmp.Diff(
			restaurant.MenuItems,
			resp.JSON200.Items,
			cmpopts.SortSlices(func(a, b ordersclient.MenuItem) bool {
				return a.Uuid.String() < b.Uuid.String()
			}),
		),
	)
	require.Equal(t, resp.JSON200.RestaurantName, restaurant.Name)
	require.Equal(t, resp.JSON200.Address, restaurant.Address)
	require.Equal(t, resp.JSON200.Description, restaurant.Description)
	require.Equal(t, resp.JSON200.Currency, restaurant.Currency)
}

type testRestaurant struct {
	UUID app.RestaurantUUID
	Data ordersclient.OnboardRestaurant
}

func onboardRestaurant(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	country shared.CountryCode,
) testRestaurant {
	t.Helper()

	var menuItems []ordersclient.MenuItem
	for i := 0; i < 5; i++ {
		menuItems = append(menuItems, ordersclient.MenuItem{
			Uuid:       app.RestaurantMenuItemUUID{common.NewUUIDv7()},
			Name:       gofakeit.Lunch(),
			GrossPrice: randomPrice(),
			Ordering:   rand.Float32(),
		})
	}

	name := ""
	if rand.Intn(2) == 0 {
		name += gofakeit.FirstName() + "'s "
	}
	name += gofakeit.HipsterWord()

	address := testutils.GenerateRandomOpenapiAddress(country)

	restaurantToCreate := ordersclient.OnboardRestaurant{
		Address:     address,
		Description: gofakeit.HipsterSentence(),
		MenuItems:   menuItems,
		Name:        cases.Title(language.Und).String(name),
		Currency:    currencyForCountry(t, country),
	}

	restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
	resp, err := clients.Orders.OnboardRestaurantWithResponse(
		ctx,
		restaurantUUID,
		&ordersclient.OnboardRestaurantParams{
			OperatorUUID: common.NewUUIDv7(),
		},
		restaurantToCreate,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode())

	return testRestaurant{
		UUID: restaurantUUID,
		Data: restaurantToCreate,
	}
}

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

type testCourier struct {
	UUID ordersclient.CourierUUID
}

func registerCourierInCity(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	country shared.CountryCode,
	city string,
) testCourier {
	t.Helper()

	courierToCreate := ordersclient.RegisterCourier{
		Name:        gofakeit.Name(),
		PhoneNumber: gofakeit.Phone(),
		City:        city,
	}

	resp, err := clients.Orders.RegisterCourierWithResponse(ctx, courierToCreate)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode())
	require.NotNil(t, resp.JSON201)

	return testCourier{
		UUID: resp.JSON201.CourierUuid,
	}
}

func placeOrder(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	customerUUID app.CustomerUUID,
	restaurantUUID app.RestaurantUUID,
	restaurant ordersclient.OnboardRestaurant,
	country shared.CountryCode,
) app.OrderUUID {
	t.Helper()

	// Select 1-3 random items from the menu
	numItems := rand.Intn(3) + 1
	if numItems > len(restaurant.MenuItems) {
		numItems = len(restaurant.MenuItems)
	}

	var orderItems []ordersclient.OrderItem
	for i := 0; i < numItems; i++ {
		orderItems = append(orderItems, ordersclient.OrderItem{
			MenuItemUuid: restaurant.MenuItems[i].Uuid,
			Quantity:     rand.Intn(3) + 1,
		})
	}

	createQuoteRequest := ordersclient.CreateQuoteRequest{
		RestaurantUuid:  restaurantUUID,
		Items:           orderItems,
		DeliveryAddress: testutils.GenerateOpenapiAddressInCity(country, restaurant.Address.City),
	}

	quoteResp, err := clients.Orders.CustomerCreateQuoteWithResponse(
		ctx,
		&ordersclient.CustomerCreateQuoteParams{
			CustomerUUID: customerUUID,
		},
		createQuoteRequest,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, quoteResp.StatusCode())
	require.Equal(t, quoteResp.JSON201.Currency, restaurant.Currency)
	require.NotNil(t, quoteResp.JSON201)
	require.False(t, quoteResp.JSON201.DeliveryFeeGross.IsZero())
	require.False(t, quoteResp.JSON201.ServiceFeeGross.IsZero())
	require.False(t, quoteResp.JSON201.TotalGross.IsZero())

	_, cardNumber := createBankAccountWithBalance(ctx, t, clients, decimal.NewFromInt(1000), common.NewUUIDv7().String())
	createBankAccount(ctx, t, clients, restaurantUUID.String())
	nonce := preauthPayment(
		ctx,
		t,
		clients,
		cardNumber,
		quoteResp.JSON201.TotalGross,
		quoteResp.JSON201.Currency.String(),
		quoteResp.JSON201.QuoteUuid.String(),
	)

	placeOrderRequest := ordersclient.PlaceOrder{
		QuoteUuid:    quoteResp.JSON201.QuoteUuid,
		PaymentNonce: nonce,
	}

	orderResp, err := clients.Orders.CustomerPlaceOrderWithResponse(
		ctx,
		&ordersclient.CustomerPlaceOrderParams{
			CustomerUUID: customerUUID,
		},
		placeOrderRequest,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, orderResp.StatusCode())

	assertOrderMatchesQuote(t, orderResp.JSON201, quoteResp.JSON201)

	require.Empty(
		t,
		cmp.Diff(
			&ordersclient.CustomerOrder{
				CourierAcceptedAt:     nil,
				CourierUuid:           nil,
				DeliveredAt:           nil,
				DeliveryAddress:       createQuoteRequest.DeliveryAddress,
				DeliveryFeeGross:      quoteResp.JSON201.DeliveryFeeGross,
				ItemsSubtotalGross:    quoteResp.JSON201.ItemsSubtotalGross,
				OrderUuid:             orderResp.JSON201.OrderUuid,
				OrderedAt:             time.Now(),
				PickedUpAt:            nil,
				RestaurantConfirmedAt: nil,
				RestaurantName:        restaurant.Name,
				RestaurantPreparedAt:  nil,
				RestaurantUuid:        restaurantUUID,
				ServiceFeeGross:       quoteResp.JSON201.ServiceFeeGross,
				TotalGross:            quoteResp.JSON201.TotalGross,
				TotalTax:              quoteResp.JSON201.TotalTax,
				Currency:              quoteResp.JSON201.Currency,
			},
			orderResp.JSON201,
			cmpopts.EquateApproxTime(time.Minute),
			cmpopts.EquateComparable(shared.SharedTypes...),
		),
	)

	customerOrdersResp, err := clients.Orders.CustomerGetOrdersWithResponse(ctx, &ordersclient.CustomerGetOrdersParams{
		CustomerUUID: customerUUID,
	})
	require.NoError(t, err)
	require.NotEmpty(t, customerOrdersResp.JSON200.Orders, "Expected at least one order for the customer")
	require.Empty(
		t,
		cmp.Diff(
			*orderResp.JSON201,
			customerOrdersResp.JSON200.Orders[0],
			cmpopts.EquateComparable(shared.SharedTypes...),
			cmpopts.EquateApproxTime(time.Second),
		),
	)

	restaurantOrdersResp, err := clients.Orders.RestaurantGetOrdersWithResponse(ctx, &ordersclient.RestaurantGetOrdersParams{
		RestaurantUUID: restaurantUUID,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, restaurantOrdersResp.StatusCode())
	require.NotEmpty(t, restaurantOrdersResp.JSON200.Orders, "Expected at least one order")

	orderUUID := app.OrderUUID{common.NewUUIDv7()}
	for _, order := range restaurantOrdersResp.JSON200.Orders {
		if order.CustomerUuid == customerUUID {
			orderUUID = order.OrderUuid
			break
		}
	}
	require.NotEqual(t, openapi_types.UUID{}, orderUUID, "Placed order not found in restaurant's order list")

	return orderUUID
}

func assertOrderVisibleToRestaurant(ctx context.Context, t *testing.T, clients testClients, restaurantUUID app.RestaurantUUID, orderUUID app.OrderUUID) {
	t.Helper()

	resp, err := clients.Orders.RestaurantGetOrdersWithResponse(ctx, &ordersclient.RestaurantGetOrdersParams{
		RestaurantUUID: restaurantUUID,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	// Verify the specific order is in the list
	found := false
	for _, order := range resp.JSON200.Orders {
		if order.OrderUuid == orderUUID {
			found = true
			break
		}
	}
	require.True(t, found, "Order %s should be visible to restaurant", orderUUID)
}

func restaurantAcceptOrder(ctx context.Context, t *testing.T, clients testClients, restaurantUUID app.RestaurantUUID, orderUUID app.OrderUUID) {
	t.Helper()

	resp, err := clients.Orders.RestaurantAcceptOrderWithResponse(
		ctx,
		&ordersclient.RestaurantAcceptOrderParams{
			RestaurantUUID: restaurantUUID,
		},
		ordersclient.AcceptOrder{
			OrderUuid: orderUUID,
		},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, resp.StatusCode())
}

func assertOrderVisibleToCourier(ctx context.Context, t *testing.T, clients testClients, courierUUID app.CourierUUID, orderUUID app.OrderUUID) {
	t.Helper()

	resp, err := clients.Orders.CourierGetAvailableOrdersWithResponse(ctx, &ordersclient.CourierGetAvailableOrdersParams{
		CourierUUID: courierUUID,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	// Verify the specific order is in the list
	found := false
	for _, order := range resp.JSON200.Orders {
		if order.OrderUuid == orderUUID {
			found = true
			break
		}
	}
	require.True(t, found, "Order %s should be visible to courier", orderUUID)
}

func courierAcceptDelivery(ctx context.Context, t *testing.T, clients testClients, courierUUID app.CourierUUID, orderUUID app.OrderUUID) {
	t.Helper()

	resp, err := clients.Orders.CourierAcceptDeliveryWithResponse(
		ctx,
		&ordersclient.CourierAcceptDeliveryParams{
			CourierUUID: courierUUID,
		},
		ordersclient.AcceptDelivery{
			OrderUuid: orderUUID,
		},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, resp.StatusCode())
}

func restaurantMarkOrderReady(ctx context.Context, t *testing.T, clients testClients, restaurantUUID app.RestaurantUUID, orderUUID app.OrderUUID) {
	t.Helper()

	resp, err := clients.Orders.RestaurantMarkOrderReadyForPickupWithResponse(
		ctx,
		&ordersclient.RestaurantMarkOrderReadyForPickupParams{
			RestaurantUUID: restaurantUUID,
		},
		ordersclient.MarkOrderReady{
			OrderUuid: orderUUID,
		},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, resp.StatusCode())
}

func assertOrderReadyForCourier(ctx context.Context, t *testing.T, clients testClients, courierUUID app.CourierUUID, orderUUID app.OrderUUID) {
	t.Helper()

	resp, err := clients.Orders.CourierGetCurrentOrdersWithResponse(ctx, &ordersclient.CourierGetCurrentOrdersParams{
		CourierUUID: courierUUID,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	// Find the specific order and verify it's marked as ready
	found := false
	for _, order := range resp.JSON200.Orders {
		if order.OrderUuid == orderUUID {
			require.NotNil(t, order.RestaurantPreparedAt, "Order should be marked as ready")
			found = true
			break
		}
	}

	require.True(t, found, "Order %s not found in courier's order list", orderUUID)
}

func courierReportPickup(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	courierUUID app.CourierUUID,
	orderUUID app.OrderUUID,
) {
	t.Helper()

	resp, err := clients.Orders.CourierReportPickupWithResponse(
		ctx,
		&ordersclient.CourierReportPickupParams{
			CourierUUID: courierUUID,
		},
		ordersclient.ReportPickup{
			OrderUuid: orderUUID,
		},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, resp.StatusCode())
}

func assertPickupReported(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	courierUUID app.CourierUUID,
	orderUUID app.OrderUUID,
) {
	t.Helper()

	resp, err := clients.Orders.CourierGetCurrentOrdersWithResponse(ctx, &ordersclient.CourierGetCurrentOrdersParams{
		CourierUUID: courierUUID,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	// Find the specific order and verify it's marked as picked up
	found := false
	for _, order := range resp.JSON200.Orders {
		if order.OrderUuid == orderUUID {
			require.NotNil(t, order.PickedUpAt, "Order should be marked as picked up")
			found = true
			break
		}
	}

	require.True(t, found, "Order %s not found in courier's order list", orderUUID)
}

func courierReportDelivered(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	courierUUID app.CourierUUID,
	orderUUID app.OrderUUID,
) {
	t.Helper()

	resp, err := clients.Orders.CourierReportDeliveryWithResponse(
		ctx,
		&ordersclient.CourierReportDeliveryParams{
			CourierUUID: courierUUID,
		},
		ordersclient.ReportDelivery{
			OrderUuid: orderUUID,
		},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, resp.StatusCode())
}

func assertDeliveryReported(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	courierUUID app.CourierUUID,
	orderUUID app.OrderUUID,
) {
	t.Helper()

	resp, err := clients.Orders.CourierGetCurrentOrdersWithResponse(ctx, &ordersclient.CourierGetCurrentOrdersParams{
		CourierUUID: courierUUID,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	// Find the specific order and verify it's marked as delivered
	found := false
	for _, order := range resp.JSON200.Orders {
		if order.OrderUuid == orderUUID {
			require.NotNil(t, order.DeliveredAt, "Order should be marked as delivered")
			found = true
			break
		}
	}

	require.True(t, found, "Order %s not found in courier's order list", orderUUID)
}

func updateRestaurantMenu(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	restaurantUUID app.RestaurantUUID,
	restaurant ordersclient.OnboardRestaurant,
) {
	t.Helper()

	resp, err := clients.Orders.OnboardRestaurantWithResponse(
		ctx,
		restaurantUUID,
		&ordersclient.OnboardRestaurantParams{
			OperatorUUID: common.NewUUIDv7(),
		},
		restaurant,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode())
}

func assertMenuItemsEquals(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	restaurantUUID app.RestaurantUUID,
	expectedMenuItems []ordersclient.MenuItem,
) {
	t.Helper()

	menuResp, err := clients.Orders.CustomerGetRestaurantMenuWithResponse(ctx, restaurantUUID, &ordersclient.CustomerGetRestaurantMenuParams{
		CustomerUUID: app.CustomerUUID{common.NewUUIDv7()},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, menuResp.StatusCode())

	require.Len(t, menuResp.JSON200.Items, len(expectedMenuItems), "Menu should contain expected number of items")

	expectedUUIDs := make([]app.RestaurantMenuItemUUID, len(expectedMenuItems))
	for i, item := range expectedMenuItems {
		expectedUUIDs[i] = item.Uuid
	}

	actualUUIDs := make([]app.RestaurantMenuItemUUID, len(menuResp.JSON200.Items))
	for i, item := range menuResp.JSON200.Items {
		actualUUIDs[i] = item.Uuid
	}

	require.ElementsMatch(t, expectedUUIDs, actualUUIDs, "Menu should contain only expected items")
}

func assertArchivedMenuItemsNotInMenu(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	restaurantUUID app.RestaurantUUID,
	archivedItems []ordersclient.MenuItem,
) {
	t.Helper()

	menuResp, err := clients.Orders.CustomerGetRestaurantMenuWithResponse(ctx, restaurantUUID, &ordersclient.CustomerGetRestaurantMenuParams{
		CustomerUUID: app.CustomerUUID{common.NewUUIDv7()},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, menuResp.StatusCode())

	for _, archivedItem := range archivedItems {
		for _, actualItem := range menuResp.JSON200.Items {
			require.NotEqual(
				t,
				archivedItem.Uuid,
				actualItem.Uuid,
				"Archived item %s should not be in the menu",
				archivedItem.Uuid,
			)
		}
	}
}

func assertOrderWithArchivedItemFails(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	customerUUID app.CustomerUUID,
	restaurantUUID app.RestaurantUUID,
	archivedItems ordersclient.MenuItem,
	country shared.CountryCode,
	city string,
) {
	t.Helper()

	// Try to create an order offer with an archived item
	createOfferRequest := ordersclient.CreateQuoteRequest{
		RestaurantUuid: restaurantUUID,
		Items: []ordersclient.OrderItem{
			{
				MenuItemUuid: archivedItems.Uuid,
				Quantity:     1,
			},
		},
		DeliveryAddress: testutils.GenerateOpenapiAddressInCity(country, city),
	}

	offerResp, err := clients.Orders.CustomerCreateQuoteWithResponse(
		ctx,
		&ordersclient.CustomerCreateQuoteParams{
			CustomerUUID: customerUUID,
		},
		createOfferRequest,
	)
	require.NoError(t, err)
	require.Equal(
		t,
		http.StatusGone,
		offerResp.StatusCode(),
		"Should return 400 when creating offer with archived items",
	)

	require.NotNil(t, offerResp.JSON410, "Should have error response body")
	require.NotEmpty(t, offerResp.JSON410.Details, "Should have errors in response")
	require.Equal(
		t,
		"archived-menu-position",
		offerResp.JSON410.Details[0].ErrorSlug,
		"Error slug should indicate unavailable menu items",
	)
	require.Equal(
		t,
		// it's part of the public contract - we want to assert it
		fmt.Sprintf("menu position '%s' is archived", archivedItems.Name),
		offerResp.JSON410.Details[0].Message,
		"Error message should indicate items not available",
	)
}

func randomPrice() decimal.Decimal {
	return decimal.New(int64(rand.Intn(200)+50), -1)
}

func onboardRestaurantWithData(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	restaurantUUID app.RestaurantUUID,
	restaurant ordersclient.OnboardRestaurant,
) {
	t.Helper()

	resp, err := clients.Orders.OnboardRestaurantWithResponse(
		ctx,
		restaurantUUID,
		&ordersclient.OnboardRestaurantParams{
			OperatorUUID: common.NewUUIDv7(),
		},
		restaurant,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode())
}

func createQuote(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	customerUUID app.CustomerUUID,
	restaurantUUID app.RestaurantUUID,
	orderItems []ordersclient.OrderItem,
	deliveryAddress ordersclient.Address,
) *ordersclient.CreateQuoteResponse {
	t.Helper()

	createQuoteRequest := ordersclient.CreateQuoteRequest{
		RestaurantUuid:  restaurantUUID,
		Items:           orderItems,
		DeliveryAddress: deliveryAddress,
	}

	resp, err := clients.Orders.CustomerCreateQuoteWithResponse(
		ctx,
		&ordersclient.CustomerCreateQuoteParams{
			CustomerUUID: customerUUID,
		},
		createQuoteRequest,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode())
	require.NotNil(t, resp.JSON201)

	return resp.JSON201
}

func placeOrderFromQuote(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	customerUUID app.CustomerUUID,
	restaurantUUID app.RestaurantUUID,
	quote *ordersclient.CreateQuoteResponse,
) *ordersclient.CustomerOrder {
	t.Helper()

	_, cardNumber := createBankAccountWithBalance(ctx, t, clients, decimal.NewFromInt(1000), common.NewUUIDv7().String())
	createBankAccount(ctx, t, clients, restaurantUUID.String())
	nonce := preauthPayment(ctx, t, clients, cardNumber, quote.TotalGross, quote.Currency.String(), quote.QuoteUuid.String())

	placeOrderRequest := ordersclient.PlaceOrder{
		QuoteUuid:    quote.QuoteUuid,
		PaymentNonce: nonce,
	}

	resp, err := clients.Orders.CustomerPlaceOrderWithResponse(
		ctx,
		&ordersclient.CustomerPlaceOrderParams{
			CustomerUUID: customerUUID,
		},
		placeOrderRequest,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode())
	require.NotNil(t, resp.JSON201)

	return resp.JSON201
}

func assertOrderMatchesQuote(t *testing.T, order *ordersclient.CustomerOrder, quote *ordersclient.CreateQuoteResponse) {
	t.Helper()

	assert.Equal(
		t,
		quote.ItemsSubtotalGross.String(),
		order.ItemsSubtotalGross.String(),
		"order items subtotal should match quote",
	)
	assert.Equal(
		t,
		quote.ServiceFeeGross.String(),
		order.ServiceFeeGross.String(),
		"order service fee should match quote",
	)
	assert.Equal(
		t,
		quote.DeliveryFeeGross.String(),
		order.DeliveryFeeGross.String(),
		"order delivery fee should match quote",
	)
	assert.Equal(
		t,
		quote.TotalGross.String(),
		order.TotalGross.String(),
		"order total gross should match quote",
	)
	assert.Equal(
		t,
		quote.TotalTax.String(),
		order.TotalTax.String(),
		"order total tax should match quote",
	)
}

func currencyForCountry(t *testing.T, country shared.CountryCode) shared.Currency {
	t.Helper()
	switch country.Code() {
	case "US":
		return shared.MustNewCurrency("USD")
	case "DE":
		return shared.MustNewCurrency("EUR")
	case "GB":
		return shared.MustNewCurrency("GBP")
	case "JP":
		return shared.MustNewCurrency("JPY")
	case "PL":
		return shared.MustNewCurrency("PLN")
	default:
		t.Fatalf("unsupported country for currency mapping: %s", country.Code())
		return shared.Currency{} // unreachable
	}
}

func assertJsonReprEqual(t *testing.T, expected, actual any) {
	t.Helper()

	expectedJSON, err := json.Marshal(expected)
	require.NoError(t, err)

	actualJSON, err := json.Marshal(actual)
	require.NoError(t, err)

	require.JSONEq(t, string(expectedJSON), string(actualJSON))
}

func createBankAccount(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	merchantID string,
) (string, string) {
	t.Helper()
	return createBankAccountWithBalance(ctx, t, clients, decimal.Zero, merchantID)
}

func createBankAccountWithBalance(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	balance decimal.Decimal,
	merchantID string,
) (string, string) {
	t.Helper()
	resp, err := clients.CommonClients.Bank.CreateAccountWithResponse(ctx, bank2.CreateAccountJSONRequestBody{
		InitialBalance: balance,
		MerchantId:     merchantID,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode())
	require.NotNil(t, resp.JSON201)
	require.NotNil(t, resp.JSON201.Card)
	return resp.JSON201.AccountNumber, resp.JSON201.Card.CardNumber
}

func preauthPayment(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	cardNumber string,
	amount decimal.Decimal,
	currency string,
	idempotencyKey string,
) string {
	paymentResp, err := clients.CommonClients.Bank.PreauthorizePaymentWithResponse(ctx, bank2.PreauthorizePaymentJSONRequestBody{
		Amount:     amount,
		CardNumber: cardNumber,
		Currency:   currency,
		Cvv:        "123",
		ExpiryDate: openapi_types.Date{
			Time: time.Date(2030, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		IdempotencyKey: idempotencyKey,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, paymentResp.StatusCode())

	return paymentResp.JSON200.PaymentNonce
}

func onboardRestaurantWithName(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	country shared.CountryCode,
	name string,
) (app.RestaurantUUID, ordersclient.OnboardRestaurant) {
	t.Helper()

	var menuItems []ordersclient.MenuItem
	for i := 0; i < 5; i++ {
		menuItems = append(menuItems, ordersclient.MenuItem{
			Uuid:       app.RestaurantMenuItemUUID{common.NewUUIDv7()},
			Name:       gofakeit.Lunch(),
			GrossPrice: randomPrice(),
			Ordering:   rand.Float32(),
		})
	}

	restaurantToCreate := ordersclient.OnboardRestaurant{
		Address:     testutils.GenerateRandomOpenapiAddress(country),
		Description: gofakeit.HipsterSentence(),
		MenuItems:   menuItems,
		Name:        name,
		Currency:    currencyForCountry(t, country),
	}

	restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
	resp, err := clients.Orders.OnboardRestaurantWithResponse(
		ctx,
		restaurantUUID,
		&ordersclient.OnboardRestaurantParams{
			OperatorUUID: common.NewUUIDv7(),
		},
		restaurantToCreate,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode())

	return restaurantUUID, restaurantToCreate
}

func onboardRestaurantWithItems(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	country shared.CountryCode,
	name string,
	itemNames []string,
) (app.RestaurantUUID, ordersclient.OnboardRestaurant) {
	t.Helper()

	menuItems := make([]ordersclient.MenuItem, 0, len(itemNames))
	for i, itemName := range itemNames {
		menuItems = append(menuItems, ordersclient.MenuItem{
			Uuid:       app.RestaurantMenuItemUUID{common.NewUUIDv7()},
			Name:       itemName,
			GrossPrice: decimal.NewFromFloat(10.00 + float64(i)),
			Ordering:   float32(i + 1),
		})
	}

	restaurantToCreate := ordersclient.OnboardRestaurant{
		Address:     testutils.GenerateRandomOpenapiAddress(country),
		Description: gofakeit.HipsterSentence(),
		MenuItems:   menuItems,
		Name:        name,
		Currency:    currencyForCountry(t, country),
	}

	restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
	resp, err := clients.Orders.OnboardRestaurantWithResponse(
		ctx,
		restaurantUUID,
		&ordersclient.OnboardRestaurantParams{
			OperatorUUID: common.NewUUIDv7(),
		},
		restaurantToCreate,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode())

	return restaurantUUID, restaurantToCreate
}
