// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package tests_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/common/testutils"
	ordersclient "eats/backend/orders/api/http/client"
	"eats/backend/orders/app"
)

func TestComponent_IssueReceipt(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()

	documentUUID := issueReceipt(ctx, t, clients)
	require.NotEmpty(t, documentUUID, "document UUID should not be empty")
}

func TestComponent_CriticalFlow(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()

	country := testutils.GenerateRandomCountry()

	platform := createPlatformEntity(ctx, t, clients)
	restaurant := onboardRestaurant(ctx, t, clients, platform.UUID, country)
	assertRestaurantMenuPublished(ctx, t, clients, restaurant.UUID, restaurant.Data)

	courier := registerCourierInCity(ctx, t, clients, platform.UUID, country, restaurant.Data.Address.City)

	customerUUID := registerCustomerInCity(ctx, t, clients, country, restaurant.Data.Address.City)

	_, cardNumber := createBankAccountWithBalance(ctx, t, decimal.NewFromInt(1000), common.NewUUIDv7().String())

	orderUUID := placeOrder(ctx, t, clients, customerUUID, restaurant.UUID, restaurant.Data, country, cardNumber)

	assertOrderVisibleToRestaurant(ctx, t, clients, restaurant.UUID, orderUUID)

	restaurantAcceptOrder(ctx, t, clients, restaurant.UUID, orderUUID)

	assertOrderVisibleToCourier(ctx, t, clients, courier.UUID, orderUUID)

	courierAcceptDelivery(ctx, t, clients, courier.UUID, orderUUID)

	restaurantMarkOrderReady(ctx, t, clients, restaurant.UUID, orderUUID)

	assertOrderReadyForCourier(ctx, t, clients, courier.UUID, orderUUID)

	courierReportPickup(ctx, t, clients, courier.UUID, orderUUID)

	assertPickupReported(ctx, t, clients, courier.UUID, orderUUID)

	courierReportDelivered(ctx, t, clients, courier.UUID, orderUUID)

	assertDeliveryReported(ctx, t, clients, courier.UUID, orderUUID)
}

func TestComponent_ListMenuItems(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()
	country := testutils.GenerateRandomCountry()

	// Onboard a restaurant with menu items
	restaurantUUID, restaurant := onboardRestaurantWithName(ctx, t, clients, country, "Test Restaurant")
	require.NotEmpty(t, restaurantUUID)
	require.NotEmpty(t, restaurant.MenuItems)

	// Call the read model endpoint (no filters)
	resp, err := clients.Orders.ListMenuItemsWithResponse(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.NotNil(t, resp.JSON200)

	// Verify our menu items are in the response
	items := *resp.JSON200
	found := 0
	for _, item := range items {
		for _, expected := range restaurant.MenuItems {
			if item.MenuItemUuid == expected.Uuid {
				assert.Equal(t, expected.Name, item.MenuItemName)
				assert.Equal(t, "Test Restaurant", item.RestaurantName)
				found++
			}
		}
	}
	assert.Equal(t, len(restaurant.MenuItems), found, "all menu items should be returned by read model")
}

func TestComponent_ListMenuItems_WithFiltering(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()
	country := testutils.GenerateRandomCountry()

	// Onboard two restaurants
	_, _ = onboardRestaurantWithName(ctx, t, clients, country, "Pizza Palace")
	_, _ = onboardRestaurantWithName(ctx, t, clients, country, "Burger Barn")

	// Filter by restaurant name
	restaurantName := "Pizza"
	resp, err := clients.Orders.ListMenuItemsWithResponse(ctx, &ordersclient.ListMenuItemsParams{
		RestaurantName: &restaurantName,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.NotNil(t, resp.JSON200)

	// All items should be from Pizza Palace
	items := *resp.JSON200
	for _, item := range items {
		assert.Contains(t, item.RestaurantName, "Pizza", "all items should be from filtered restaurant")
	}
}

func TestComponent_ListMenuItems_WithOrdering(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()
	country := testutils.GenerateRandomCountry()

	// Onboard a restaurant
	_, _ = onboardRestaurantWithName(ctx, t, clients, country, "Test Restaurant")

	// Order by price ascending
	orderBy := ordersclient.PriceAsc
	resp, err := clients.Orders.ListMenuItemsWithResponse(ctx, &ordersclient.ListMenuItemsParams{
		OrderBy: &orderBy,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.NotNil(t, resp.JSON200)

	// Verify items are ordered by price
	items := *resp.JSON200
	if len(items) > 1 {
		for i := 1; i < len(items); i++ {
			assert.True(t, items[i-1].GrossPrice.LessThanOrEqual(items[i].GrossPrice),
				"items should be ordered by price ascending")
		}
	}
}

func TestComponent_ListMenuItems_WithFullTextSearch(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()
	country := testutils.GenerateRandomCountry()

	// Onboard restaurants with specific menu items for search testing
	_, _ = onboardRestaurantWithItems(ctx, t, clients, country, "Italian Trattoria", []string{
		"Spaghetti Carbonara",
		"Margherita Pizza",
		"Tiramisu Dessert",
	})
	_, _ = onboardRestaurantWithItems(ctx, t, clients, country, "Burger Joint", []string{
		"Classic Cheeseburger",
		"Bacon Burger",
		"Veggie Burger",
	})

	// Search for "pizza" - should find items/restaurants mentioning pizza
	search := "pizza"
	resp, err := clients.Orders.ListMenuItemsWithResponse(ctx, &ordersclient.ListMenuItemsParams{
		Search: &search,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.NotNil(t, resp.JSON200)

	items := *resp.JSON200
	require.NotEmpty(t, items, "should find items matching 'pizza'")

	// All results should contain pizza in either item name or restaurant name
	for _, item := range items {
		found := strings.Contains(strings.ToLower(item.MenuItemName), "pizza") ||
			strings.Contains(strings.ToLower(item.RestaurantName), "pizza")
		assert.True(t, found, "item should match search term: %s at %s", item.MenuItemName, item.RestaurantName)
	}
}

func TestComponent_ListMenuItems_WithSearchAndRelevanceOrdering(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()
	country := testutils.GenerateRandomCountry()

	// Create restaurants where one has "burger" in its name
	_, _ = onboardRestaurantWithItems(ctx, t, clients, country, "Best Burger Place", []string{
		"Simple Salad", // doesn't contain burger
	})
	_, _ = onboardRestaurantWithItems(ctx, t, clients, country, "Random Diner", []string{
		"Deluxe Burger", // contains burger in item name
	})

	// Search for "burger" with relevance ordering
	search := "burger"
	orderBy := ordersclient.Relevance
	resp, err := clients.Orders.ListMenuItemsWithResponse(ctx, &ordersclient.ListMenuItemsParams{
		Search:  &search,
		OrderBy: &orderBy,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.NotNil(t, resp.JSON200)

	items := *resp.JSON200
	require.NotEmpty(t, items, "should find items matching 'burger'")
}

func TestComponent_RegisterCourier(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()

	country := testutils.GenerateRandomCountry()
	city := testutils.GenerateRandomOpenapiAddress(country).City

	// Register a courier
	platform := createPlatformEntity(ctx, t, clients)
	courier := registerCourierInCity(ctx, t, clients, platform.UUID, country, city)
	require.NotEmpty(t, courier.UUID)
}

func TestComponent_RegisterCourier_Validation(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()

	t.Run("empty_name", func(t *testing.T) {
		resp, err := clients.Orders.RegisterCourierWithResponse(ctx, ordersclient.RegisterCourier{
			Name:        "",
			PhoneNumber: "123456789",
			City:        "Warsaw",
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode())
	})

	t.Run("empty_phone_number", func(t *testing.T) {
		resp, err := clients.Orders.RegisterCourierWithResponse(ctx, ordersclient.RegisterCourier{
			Name:        "John Doe",
			PhoneNumber: "",
			City:        "Warsaw",
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode())
	})

	t.Run("empty_city", func(t *testing.T) {
		resp, err := clients.Orders.RegisterCourierWithResponse(ctx, ordersclient.RegisterCourier{
			Name:        "John Doe",
			PhoneNumber: "123456789",
			City:        "",
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode())
	})
}

func TestComponent_CourierDeliveryFlow(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()

	country := testutils.GenerateRandomCountry()

	platform := createPlatformEntity(ctx, t, clients)
	restaurant := onboardRestaurant(ctx, t, clients, platform.UUID, country)
	assertRestaurantMenuPublished(ctx, t, clients, restaurant.UUID, restaurant.Data)

	courier := registerCourierInCity(ctx, t, clients, platform.UUID, country, restaurant.Data.Address.City)

	customerUUID := registerCustomerInCity(ctx, t, clients, country, restaurant.Data.Address.City)

	// Create quote and place order
	orderItems := []ordersclient.OrderItem{
		{
			MenuItemUuid: restaurant.Data.MenuItems[0].Uuid,
			Quantity:     2,
		},
	}
	deliveryAddress := testutils.GenerateOpenapiAddressInCity(country, restaurant.Data.Address.City)
	quote := createQuote(ctx, t, clients, customerUUID, restaurant.UUID, orderItems, deliveryAddress)

	_, cardNumber := createBankAccountWithBalance(ctx, t, decimal.NewFromInt(1000), common.NewUUIDv7().String())
	order := placeOrderFromQuote(ctx, t, clients, customerUUID, quote, cardNumber)
	require.NotEmpty(t, order.OrderUuid)
	assertOrderMatchesQuote(t, order, quote)

	// Restaurant accepts order
	restaurantAcceptOrder(ctx, t, clients, restaurant.UUID, order.OrderUuid)

	// Courier accepts delivery
	courierAcceptDelivery(ctx, t, clients, courier.UUID, order.OrderUuid)

	// Restaurant marks order as ready
	restaurantMarkOrderReady(ctx, t, clients, restaurant.UUID, order.OrderUuid)

	// Courier reports pickup
	courierReportPickup(ctx, t, clients, courier.UUID, order.OrderUuid)

	// Courier reports delivery
	courierReportDelivered(ctx, t, clients, courier.UUID, order.OrderUuid)

	t.Run("report_delivery_idempotent", func(t *testing.T) {
		courierReportDelivered(ctx, t, clients, courier.UUID, order.OrderUuid)

		assertDeliveryReported(ctx, t, clients, courier.UUID, order.OrderUuid)
	})
}

func TestComponent_CourierCityFiltering(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()
	country := testutils.GenerateRandomCountry()

	platform := createPlatformEntity(ctx, t, clients)
	restaurant := onboardRestaurant(ctx, t, clients, platform.UUID, country)
	customerUUID := registerCustomerInCity(ctx, t, clients, country, restaurant.Data.Address.City)

	courierInSameCity := registerCourierInCity(ctx, t, clients, platform.UUID, country, restaurant.Data.Address.City)

	differentCity := restaurant.Data.Address.City + "_Different"
	courierInDifferentCity := registerCourierInCity(ctx, t, clients, platform.UUID, country, differentCity)

	_, cardNumber := createBankAccountWithBalance(ctx, t, decimal.NewFromInt(1000), common.NewUUIDv7().String())

	orderUUID := placeOrder(ctx, t, clients, customerUUID, restaurant.UUID, restaurant.Data, country, cardNumber)

	restaurantAcceptOrder(ctx, t, clients, restaurant.UUID, orderUUID)

	t.Run("courier_in_same_city_sees_available_order", func(t *testing.T) {
		resp, err := clients.Orders.CourierGetAvailableOrdersWithResponse(ctx, &ordersclient.CourierGetAvailableOrdersParams{
			CourierUUID: courierInSameCity.UUID,
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
		require.NotNil(t, resp.JSON200)

		// Should see at least the order we created
		found := false
		for _, order := range resp.JSON200.Orders {
			if order.OrderUuid == orderUUID {
				found = true
				break
			}
		}
		require.True(t, found, "Courier in same city should see the available order")
	})

	t.Run("courier_in_different_city_does_not_see_order", func(t *testing.T) {
		resp, err := clients.Orders.CourierGetAvailableOrdersWithResponse(ctx, &ordersclient.CourierGetAvailableOrdersParams{
			CourierUUID: courierInDifferentCity.UUID,
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
		require.NotNil(t, resp.JSON200)

		// Should not see the order in different city
		for _, order := range resp.JSON200.Orders {
			require.NotEqual(t, orderUUID, order.OrderUuid, "Courier in different city should not see orders from other cities")
		}
	})

	t.Run("courier_from_different_city_cannot_accept_order", func(t *testing.T) {
		resp, err := clients.Orders.CourierAcceptDeliveryWithResponse(
			ctx,
			&ordersclient.CourierAcceptDeliveryParams{
				CourierUUID: courierInDifferentCity.UUID,
			},
			ordersclient.CourierAcceptDeliveryJSONRequestBody{
				OrderUuid: orderUUID,
			},
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode())
		require.NotNil(t, resp.JSON400)

		expectedError := ordersclient.ErrorResponse{
			Details: []ordersclient.ErrorDetails{
				{
					Message:    fmt.Sprintf("courier operates in %s only", differentCity),
					ErrorSlug:  "courier-out-of-delivery-zone",
					EntityType: common.ToPtr("order"),
				},
			},
			Message: "courier cannot accept orders outside their delivery zone",
			Slug:    "courier-out-of-delivery-zone",
		}

		assertJsonReprEqual(t, expectedError, resp.JSON400)
	})

	t.Run("courier_from_same_city_can_accept_order", func(t *testing.T) {
		resp, err := clients.Orders.CourierAcceptDeliveryWithResponse(
			ctx,
			&ordersclient.CourierAcceptDeliveryParams{
				CourierUUID: courierInSameCity.UUID,
			},
			ordersclient.CourierAcceptDeliveryJSONRequestBody{
				OrderUuid: orderUUID,
			},
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, resp.StatusCode())
	})
}

func TestComponent_WrongRestaurantCannotManageOrder(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()

	country := testutils.GenerateRandomCountry()

	platform := createPlatformEntity(ctx, t, clients)
	restaurant := onboardRestaurant(ctx, t, clients, platform.UUID, country)
	assertRestaurantMenuPublished(ctx, t, clients, restaurant.UUID, restaurant.Data)

	// Onboard a second restaurant (the "wrong" one)
	wrongRestaurant := onboardRestaurant(ctx, t, clients, platform.UUID, country)

	customerUUID := registerCustomerInCity(ctx, t, clients, country, restaurant.Data.Address.City)

	// Create quote and place order for the first restaurant
	orderItems := []ordersclient.OrderItem{
		{
			MenuItemUuid: restaurant.Data.MenuItems[0].Uuid,
			Quantity:     1,
		},
	}
	deliveryAddress := testutils.GenerateOpenapiAddressInCity(country, restaurant.Data.Address.City)
	quote := createQuote(ctx, t, clients, customerUUID, restaurant.UUID, orderItems, deliveryAddress)

	_, cardNumber := createBankAccountWithBalance(ctx, t, decimal.NewFromInt(1000), common.NewUUIDv7().String())
	order := placeOrderFromQuote(ctx, t, clients, customerUUID, quote, cardNumber)

	t.Run("wrong_restaurant_cannot_accept_order", func(t *testing.T) {
		resp, err := clients.Orders.RestaurantAcceptOrderWithResponse(
			ctx,
			&ordersclient.RestaurantAcceptOrderParams{
				RestaurantUUID: wrongRestaurant.UUID,
			},
			ordersclient.AcceptOrder{
				OrderUuid: order.OrderUuid,
			},
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, resp.StatusCode(),
			"wrong restaurant should not be able to accept the order")
	})

	t.Run("wrong_restaurant_cannot_mark_order_ready", func(t *testing.T) {
		// First let the correct restaurant accept the order
		restaurantAcceptOrder(ctx, t, clients, restaurant.UUID, order.OrderUuid)

		resp, err := clients.Orders.RestaurantMarkOrderReadyForPickupWithResponse(
			ctx,
			&ordersclient.RestaurantMarkOrderReadyForPickupParams{
				RestaurantUUID: wrongRestaurant.UUID,
			},
			ordersclient.MarkOrderReady{
				OrderUuid: order.OrderUuid,
			},
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, resp.StatusCode(),
			"wrong restaurant should not be able to mark the order as ready")
	})
}

func TestComponent_SecondCourierCannotAcceptSameOrder(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()

	country := testutils.GenerateRandomCountry()

	platform := createPlatformEntity(ctx, t, clients)
	restaurant := onboardRestaurant(ctx, t, clients, platform.UUID, country)
	assertRestaurantMenuPublished(ctx, t, clients, restaurant.UUID, restaurant.Data)

	courierA := registerCourierInCity(ctx, t, clients, platform.UUID, country, restaurant.Data.Address.City)
	courierB := registerCourierInCity(ctx, t, clients, platform.UUID, country, restaurant.Data.Address.City)
	customerUUID := registerCustomerInCity(ctx, t, clients, country, restaurant.Data.Address.City)

	// Create quote and place order
	orderItems := []ordersclient.OrderItem{
		{
			MenuItemUuid: restaurant.Data.MenuItems[0].Uuid,
			Quantity:     1,
		},
	}
	deliveryAddress := testutils.GenerateOpenapiAddressInCity(country, restaurant.Data.Address.City)
	quote := createQuote(ctx, t, clients, customerUUID, restaurant.UUID, orderItems, deliveryAddress)

	_, cardNumber := createBankAccountWithBalance(ctx, t, decimal.NewFromInt(1000), common.NewUUIDv7().String())
	order := placeOrderFromQuote(ctx, t, clients, customerUUID, quote, cardNumber)

	// Restaurant accepts order
	restaurantAcceptOrder(ctx, t, clients, restaurant.UUID, order.OrderUuid)

	// Courier A accepts delivery
	courierAcceptDelivery(ctx, t, clients, courierA.UUID, order.OrderUuid)

	// Courier B tries to accept the same order - should fail
	resp, err := clients.Orders.CourierAcceptDeliveryWithResponse(
		ctx,
		&ordersclient.CourierAcceptDeliveryParams{
			CourierUUID: courierB.UUID,
		},
		ordersclient.CourierAcceptDeliveryJSONRequestBody{
			OrderUuid: order.OrderUuid,
		},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusConflict, resp.StatusCode(),
		"second courier should not be able to accept an already-accepted order")
	require.NotNil(t, resp.JSON409)
	assert.Equal(t, "already-accepted", resp.JSON409.Slug)
}

func TestComponent_ListRestaurants(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()

	country := testutils.GenerateRandomCountry()

	platform := createPlatformEntity(ctx, t, clients)
	restaurant1 := onboardRestaurant(ctx, t, clients, platform.UUID, country)
	restaurant2 := onboardRestaurant(ctx, t, clients, platform.UUID, country)
	restaurant3 := onboardRestaurant(ctx, t, clients, platform.UUID, country)

	customerUUID := registerCustomer(ctx, t, clients, country)

	resp, err := clients.Orders.CustomerListRestaurantsWithResponse(ctx, &ordersclient.CustomerListRestaurantsParams{
		CustomerUUID: customerUUID,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.NotNil(t, resp.JSON200)
	require.GreaterOrEqual(t, len(resp.JSON200.Restaurants), 3, "Expected at least 3 restaurants")

	expectedRestaurants := []ordersclient.Restaurant{
		{
			Uuid:        restaurant1.UUID,
			Name:        restaurant1.Data.Name,
			Description: restaurant1.Data.Description,
			Address:     restaurant1.Data.Address,
		},
		{
			Uuid:        restaurant2.UUID,
			Name:        restaurant2.Data.Name,
			Description: restaurant2.Data.Description,
			Address:     restaurant2.Data.Address,
		},
		{
			Uuid:        restaurant3.UUID,
			Name:        restaurant3.Data.Name,
			Description: restaurant3.Data.Description,
			Address:     restaurant3.Data.Address,
		},
	}

	// Filter actual restaurants to only include the ones we created
	var actualRestaurants []ordersclient.Restaurant
	for _, r := range resp.JSON200.Restaurants {
		if r.Uuid == restaurant1.UUID || r.Uuid == restaurant2.UUID || r.Uuid == restaurant3.UUID {
			actualRestaurants = append(actualRestaurants, r)
		}
	}

	require.Empty(
		t,
		cmp.Diff(
			expectedRestaurants,
			actualRestaurants,
			cmpopts.SortSlices(func(a, b ordersclient.Restaurant) bool {
				return a.Uuid.String() < b.Uuid.String()
			}),
			cmpopts.EquateComparable(shared.SharedTypes...),
		),
	)
}

func TestComponent_MenuItemArchival(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()

	country := testutils.GenerateRandomCountry()

	// Onboard restaurant with 5 menu items
	platform := createPlatformEntity(ctx, t, clients)
	originalRestaurant := onboardRestaurant(ctx, t, clients, platform.UUID, country)

	// Verify original menu has all items
	require.GreaterOrEqual(t, len(originalRestaurant.Data.MenuItems), 3, "Need at least 3 menu items for this test")

	assertRestaurantMenuPublished(ctx, t, clients, originalRestaurant.UUID, originalRestaurant.Data)

	itemsToArchive := originalRestaurant.Data.MenuItems[2:]
	itemsToKeep := originalRestaurant.Data.MenuItems[:2]

	// Re-onboard the same restaurant with only first 2 menu items (excluding the rest)
	updatedRestaurant := ordersclient.OnboardRestaurant{
		Name:        originalRestaurant.Data.Name,
		Description: originalRestaurant.Data.Description,
		Address:     originalRestaurant.Data.Address,
		MenuItems:   itemsToKeep,
		Currency:    originalRestaurant.Data.Currency,
	}

	updateRestaurantMenu(ctx, t, clients, originalRestaurant.UUID, updatedRestaurant)

	t.Run("archived_menu_items_not_in_menu", func(t *testing.T) {
		t.Parallel()
		assertMenuItemsEquals(ctx, t, clients, originalRestaurant.UUID, updatedRestaurant.MenuItems)
	})

	t.Run("archived_menu_items_not_in_menu", func(t *testing.T) {
		t.Parallel()
		assertArchivedMenuItemsNotInMenu(ctx, t, clients, originalRestaurant.UUID, itemsToArchive)
	})

	t.Run("initialize_order_with_archived_item_fails", func(t *testing.T) {
		t.Parallel()

		customerUUID := registerCustomerInCity(ctx, t, clients, country, originalRestaurant.Data.Address.City)
		assertOrderWithArchivedItemFails(ctx, t, clients, customerUUID, originalRestaurant.UUID, itemsToArchive[0], country, originalRestaurant.Data.Address.City)
	})
}

func TestComponent_PlaceOrder(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()

	country := testutils.GenerateRandomCountry()

	platform := createPlatformEntity(ctx, t, clients)
	restaurant := onboardRestaurant(ctx, t, clients, platform.UUID, country)
	assertRestaurantMenuPublished(ctx, t, clients, restaurant.UUID, restaurant.Data)

	customerUUID := registerCustomerInCity(ctx, t, clients, country, restaurant.Data.Address.City)

	// Create quote
	orderItems := []ordersclient.OrderItem{
		{
			MenuItemUuid: restaurant.Data.MenuItems[0].Uuid,
			Quantity:     2,
		},
	}
	deliveryAddress := testutils.GenerateOpenapiAddressInCity(country, restaurant.Data.Address.City)
	quote := createQuote(ctx, t, clients, customerUUID, restaurant.UUID, orderItems, deliveryAddress)

	// Place order
	_, cardNumber := createBankAccountWithBalance(ctx, t, decimal.NewFromInt(1000), common.NewUUIDv7().String())
	order := placeOrderFromQuote(ctx, t, clients, customerUUID, quote, cardNumber)
	require.NotEmpty(t, order.OrderUuid)
	assertOrderMatchesQuote(t, order, quote)
}

func TestComponent_CreateQuoteOutOfDeliveryZone(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()

	country := testutils.GenerateRandomCountry()

	// Onboard restaurant in one city
	platform := createPlatformEntity(ctx, t, clients)
	restaurant := onboardRestaurant(ctx, t, clients, platform.UUID, country)
	customerUUID := registerCustomer(ctx, t, clients, country)

	// Try to create quote with delivery address in a different city
	deliveryAddress := testutils.GenerateRandomOpenapiAddress(country)
	// Ensure the delivery city is different from restaurant city
	for deliveryAddress.City == restaurant.Data.Address.City {
		deliveryAddress = testutils.GenerateRandomOpenapiAddress(country)
	}

	createQuoteRequest := ordersclient.CreateQuoteRequest{
		RestaurantUuid: restaurant.UUID,
		Items: []ordersclient.OrderItem{
			{
				MenuItemUuid: restaurant.Data.MenuItems[0].Uuid,
				Quantity:     1,
			},
		},
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
	require.Equal(t, http.StatusBadRequest, resp.StatusCode())
	require.NotNil(t, resp.JSON400)

	expectedError := ordersclient.ErrorResponse{
		Details: []ordersclient.ErrorDetails{
			{
				Message:    fmt.Sprintf("restaurant delivers to %s only", restaurant.Data.Address.City),
				ErrorSlug:  "address-out-of-delivery-zone",
				EntityType: common.ToPtr("quote"),
			},
		},
		Message: "restaurant does not deliver to the provided address",
		Slug:    "address-out-of-delivery-zone",
	}

	assertJsonReprEqual(t, expectedError, resp.JSON400)
}

func TestComponent_PlaceOrderWithArchivedItemFromQuote(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()

	country := testutils.GenerateRandomCountry()

	// Onboard restaurant with menu items
	platform := createPlatformEntity(ctx, t, clients)
	restaurant := onboardRestaurant(ctx, t, clients, platform.UUID, country)
	require.GreaterOrEqual(t, len(restaurant.Data.MenuItems), 3, "Need at least 3 menu items for this test")

	assertRestaurantMenuPublished(ctx, t, clients, restaurant.UUID, restaurant.Data)

	// Create customer and quote with an item that will be archived
	customerUUID := registerCustomerInCity(ctx, t, clients, country, restaurant.Data.Address.City)
	itemToArchive := restaurant.Data.MenuItems[2]

	orderItems := []ordersclient.OrderItem{
		{
			MenuItemUuid: itemToArchive.Uuid,
			Quantity:     1,
		},
	}

	quote := createQuote(ctx, t, clients, customerUUID, restaurant.UUID, orderItems, testutils.GenerateOpenapiAddressInCity(country, restaurant.Data.Address.City))

	// Archive the menu item after quote creation
	updatedRestaurant := ordersclient.OnboardRestaurant{
		Name:        restaurant.Data.Name,
		Description: restaurant.Data.Description,
		Address:     restaurant.Data.Address,
		MenuItems:   restaurant.Data.MenuItems[:2], // Only keep first 2 items
		Currency:    restaurant.Data.Currency,
	}
	updateRestaurantMenu(ctx, t, clients, restaurant.UUID, updatedRestaurant)

	_, cardNumber := createBankAccountWithBalance(ctx, t, decimal.NewFromInt(1000), common.NewUUIDv7().String())

	nonce := preauthPayment(
		ctx,
		t,
		cardNumber,
		decimal.NewFromInt(100),
		"USD",
		quote.QuoteUuid.String(),
	)

	// Try to place order with quote containing archived item
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
	require.Equal(
		t,
		http.StatusGone,
		resp.StatusCode(),
	)
	require.NotNil(t, resp.JSON410)

	expectedErrors := []ordersclient.ErrorDetails{
		{
			Message:    fmt.Sprintf("menu position '%s' is archived", itemToArchive.Name),
			ErrorSlug:  "archived-menu-position",
			EntityType: common.ToPtr("menu_item"),
			EntityId:   common.ToPtr(itemToArchive.Uuid.String()),
		},
	}

	assertJsonReprEqual(t, expectedErrors, resp.JSON410.Details)
}

func TestComponent_AmountsRegression(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()

	// Create restaurant with fixed prices for regression testing
	menuItems := []ordersclient.MenuItem{
		{
			Uuid:       app.RestaurantMenuItemUUID{common.NewUUIDv7()},
			Name:       "Classic Burger",
			Category:   app.ItemCategoryFood,
			GrossPrice: decimal.RequireFromString("12.50"),
			Ordering:   0.1,
		},
		{
			Uuid:       app.RestaurantMenuItemUUID{common.NewUUIDv7()},
			Name:       "Caesar Salad",
			Category:   app.ItemCategoryFood,
			GrossPrice: decimal.RequireFromString("8.75"),
			Ordering:   0.2,
		},
		{
			Uuid:       app.RestaurantMenuItemUUID{common.NewUUIDv7()},
			Name:       "Margherita Pizza",
			Category:   app.ItemCategoryFood,
			GrossPrice: decimal.RequireFromString("15.00"),
			Ordering:   0.3,
		},
	}

	country := shared.MustNewCountryCode("US")

	restaurant := ordersclient.OnboardRestaurant{
		Name:        "Test Restaurant",
		Description: "A test restaurant for regression testing",
		Address:     testutils.GenerateRandomOpenapiAddress(country),
		MenuItems:   menuItems,
		Currency:    shared.MustNewCurrency("USD"),
	}

	restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
	onboardRestaurantWithData(ctx, t, clients, restaurantUUID, restaurant)

	platform := createPlatformEntity(ctx, t, clients)

	// Onboard partner with US address for consistent tax calculation
	_ = onboardPartnerForRestaurant(ctx, t, clients, platform.UUID, restaurantUUID, "Test Restaurant Inc.")

	assertRestaurantMenuPublished(ctx, t, clients, restaurantUUID, restaurant)

	customerUUID := registerCustomerInCity(ctx, t, clients, country, restaurant.Address.City)
	courier := registerCourierInCity(ctx, t, clients, platform.UUID, country, restaurant.Address.City)

	// Place order with specific items: 2x Classic Burger + 1x Margherita Pizza
	orderItems := []ordersclient.OrderItem{
		{MenuItemUuid: menuItems[0].Uuid, Quantity: 2}, // Classic Burger
		{MenuItemUuid: menuItems[2].Uuid, Quantity: 1}, // Margherita Pizza
	}

	quote := createQuote(
		ctx,
		t,
		clients,
		customerUUID,
		restaurantUUID,
		orderItems,
		testutils.GenerateOpenapiAddressInCity(country, restaurant.Address.City),
	)

	// just to be sure 3d place is always 0
	const places = 3

	expectedItemsSubtotalGross := decimal.RequireFromString("40")
	expectedServiceFeeGross := decimal.RequireFromString("2.40")
	expectedDeliveryFeeGross := decimal.RequireFromString("10.00")

	expectedTotalGross := decimal.RequireFromString("52.40")
	expectedTotalTax := decimal.RequireFromString("9.80")

	// sanity check
	require.Equal(
		t,
		expectedItemsSubtotalGross.Add(expectedServiceFeeGross).Add(expectedDeliveryFeeGross).StringFixed(places),
		quote.TotalGross.StringFixed(places),
	)

	assert.Equal(
		t,
		expectedItemsSubtotalGross.StringFixed(places),
		quote.ItemsSubtotalGross.StringFixed(places),
		"items subtotal should match expected value",
	)
	assert.Equal(
		t,
		expectedServiceFeeGross.StringFixed(places),
		quote.ServiceFeeGross.StringFixed(places),
		"service fee should match expected value",
	)
	assert.Equal(
		t,
		expectedDeliveryFeeGross.StringFixed(places),
		quote.DeliveryFeeGross.StringFixed(places),
		"delivery fee should match expected value",
	)
	assert.Equal(
		t,
		expectedTotalGross.StringFixed(places),
		quote.TotalGross.StringFixed(places),
		"total gross should match expected value",
	)
	assert.Equal(
		t,
		expectedTotalTax.StringFixed(places),
		quote.TotalTax.StringFixed(places),
		"total gross should match expected value",
	)

	_, cardNumber := createBankAccountWithBalance(ctx, t, expectedTotalGross, common.NewUUIDv7().String())

	order := placeOrderFromQuote(ctx, t, clients, customerUUID, quote, cardNumber)

	assertOrderMatchesQuote(t, order, quote)

	assertOrderVisibleToRestaurant(ctx, t, clients, restaurantUUID, order.OrderUuid)

	restaurantAcceptOrder(ctx, t, clients, restaurantUUID, order.OrderUuid)

	assertOrderVisibleToCourier(ctx, t, clients, courier.UUID, order.OrderUuid)

	courierAcceptDelivery(ctx, t, clients, courier.UUID, order.OrderUuid)

	restaurantMarkOrderReady(ctx, t, clients, restaurantUUID, order.OrderUuid)

	assertOrderReadyForCourier(ctx, t, clients, courier.UUID, order.OrderUuid)

	courierReportPickup(ctx, t, clients, courier.UUID, order.OrderUuid)

	assertPickupReported(ctx, t, clients, courier.UUID, order.OrderUuid)

	courierReportDelivered(ctx, t, clients, courier.UUID, order.OrderUuid)

	assertDeliveryReported(ctx, t, clients, courier.UUID, order.OrderUuid)
}

func TestComponent_CreateQuoteValidationErrors(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()

	country := testutils.GenerateRandomCountry()

	platform := createPlatformEntity(ctx, t, clients)
	restaurant := onboardRestaurant(ctx, t, clients, platform.UUID, country)
	customerUUID := registerCustomerInCity(ctx, t, clients, country, restaurant.Data.Address.City)

	t.Run("multiple_validation_errors_with_details", func(t *testing.T) {
		t.Parallel()

		// Create a request that triggers multiple validation errors:
		// 1. Invalid quantity (with details)
		// 2. Invalid delivery address
		createQuoteRequest := ordersclient.CreateQuoteRequest{
			RestaurantUuid: restaurant.UUID,
			Items: []ordersclient.OrderItem{
				{
					MenuItemUuid: restaurant.Data.MenuItems[0].Uuid,
					Quantity:     0, // Invalid: quantity must be > 0
				},
				{
					MenuItemUuid: restaurant.Data.MenuItems[1].Uuid,
					Quantity:     -5, // Invalid: quantity must be > 0
				},
			},
			DeliveryAddress: testutils.GenerateOpenapiAddressInCity(country, restaurant.Data.Address.City),
		}

		resp, err := clients.Orders.CustomerCreateQuoteWithResponse(
			ctx,
			&ordersclient.CustomerCreateQuoteParams{
				CustomerUUID: customerUUID,
			},
			createQuoteRequest,
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode())
		require.NotNil(t, resp.JSON400)

		// Verify we have multiple errors in the response
		errorResponse := resp.JSON400
		require.NotNil(t, errorResponse.Details)
		require.Len(t, errorResponse.Details, 2)

		// Define expected errors
		expectedErrors := []ordersclient.ErrorDetails{
			{
				Message:    "menu position quantity must be greater than zero",
				ErrorSlug:  "invalid-quantity",
				EntityId:   common.ToPtr(restaurant.Data.MenuItems[0].Uuid.String()),
				EntityType: common.ToPtr("menu_item"),
			},
			{
				Message:    "menu position quantity must be greater than zero",
				ErrorSlug:  "invalid-quantity",
				EntityId:   common.ToPtr(restaurant.Data.MenuItems[1].Uuid.String()),
				EntityType: common.ToPtr("menu_item"),
			},
		}

		assertJsonReprEqual(t, expectedErrors, errorResponse.Details)
	})

	t.Run("empty_order_validation_error", func(t *testing.T) {
		t.Parallel()

		// Create a request with empty items
		createQuoteRequest := ordersclient.CreateQuoteRequest{
			RestaurantUuid:  restaurant.UUID,
			Items:           []ordersclient.OrderItem{}, // Empty order
			DeliveryAddress: testutils.GenerateOpenapiAddressInCity(country, restaurant.Data.Address.City),
		}

		resp, err := clients.Orders.CustomerCreateQuoteWithResponse(
			ctx,
			&ordersclient.CustomerCreateQuoteParams{
				CustomerUUID: customerUUID,
			},
			createQuoteRequest,
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode())
		require.NotNil(t, resp.JSON400)

		errorResponse := resp.JSON400
		require.NotNil(t, errorResponse.Details)
		require.Len(t, errorResponse.Details, 1, "should have 1 validation error")

		assert.Equal(t, "empty-order", errorResponse.Details[0].ErrorSlug)
		assert.Contains(t, errorResponse.Details[0].Message, "at least one menu position")
	})
}

// TestComponent_JapaneseYenNoDecimals verifies that Japanese Yen amounts
// are handled correctly without decimal places.
func TestComponent_JapaneseYenNoDecimals(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()

	// Create menu items with whole number JPY prices (no decimals)
	menuItems := []ordersclient.MenuItem{
		{
			Uuid:       app.RestaurantMenuItemUUID{common.NewUUIDv7()},
			Name:       "Ramen",
			Category:   app.ItemCategoryFood,
			GrossPrice: decimal.RequireFromString("1200"), // 1,200
			Ordering:   0.1,
		},
		{
			Uuid:       app.RestaurantMenuItemUUID{common.NewUUIDv7()},
			Name:       "Gyoza",
			Category:   app.ItemCategoryFood,
			GrossPrice: decimal.RequireFromString("600"), // 600
			Ordering:   0.2,
		},
		{
			Uuid:       app.RestaurantMenuItemUUID{common.NewUUIDv7()},
			Name:       "Matcha Ice Cream",
			Category:   app.ItemCategoryFood,
			GrossPrice: decimal.RequireFromString("450"), // 450
			Ordering:   0.3,
		},
	}

	country := shared.MustNewCountryCode("JP")
	jpyCurrency := shared.MustNewCurrency("JPY")

	restaurant := ordersclient.OnboardRestaurant{
		Name:        "Tokyo Ramen House",
		Description: "Authentic Japanese ramen",
		Address:     testutils.GenerateRandomOpenapiAddress(country),
		MenuItems:   menuItems,
		Currency:    jpyCurrency,
	}

	restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
	onboardRestaurantWithData(ctx, t, clients, restaurantUUID, restaurant)

	platform := createPlatformEntityWithCurrency(ctx, t, clients, jpyCurrency)

	_ = onboardPartnerForRestaurantWithCurrency(
		ctx, t, clients, platform.UUID, restaurantUUID, "Tokyo Ramen Inc.", jpyCurrency,
	)

	assertRestaurantMenuPublished(ctx, t, clients, restaurantUUID, restaurant)

	customerUUID := registerCustomerInCity(ctx, t, clients, country, restaurant.Address.City)
	courier := registerCourierInCityWithCurrency(ctx, t, clients, platform.UUID, country, restaurant.Address.City, jpyCurrency)

	// Place order: 2x Ramen + 1x Gyoza = 3,000
	orderItems := []ordersclient.OrderItem{
		{MenuItemUuid: menuItems[0].Uuid, Quantity: 2}, // 2x Ramen = 2,400
		{MenuItemUuid: menuItems[1].Uuid, Quantity: 1}, // 1x Gyoza = 600
	}

	quote := createQuote(
		ctx,
		t,
		clients,
		customerUUID,
		restaurantUUID,
		orderItems,
		testutils.GenerateOpenapiAddressInCity(country, restaurant.Address.City),
	)

	// Verify JPY amounts have no decimal places
	assert.True(t, quote.ItemsSubtotalGross.Equal(quote.ItemsSubtotalGross.Truncate(0)),
		"JPY items subtotal should have no decimal places")
	assert.True(t, quote.TotalGross.Equal(quote.TotalGross.Truncate(0)),
		"JPY total should have no decimal places")
	assert.True(t, quote.DeliveryFeeGross.Equal(quote.DeliveryFeeGross.Truncate(0)),
		"JPY delivery fee should have no decimal places")
	assert.True(t, quote.ServiceFeeGross.Equal(quote.ServiceFeeGross.Truncate(0)),
		"JPY service fee should have no decimal places")

	// Expected items subtotal: 2x1200 + 1x600 = 3,000
	expectedItemsSubtotal := decimal.RequireFromString("3000")
	assert.Equal(t, expectedItemsSubtotal.String(), quote.ItemsSubtotalGross.String(),
		"items subtotal should be 3,000")

	_, cardNumber := createBankAccountWithBalance(ctx, t, quote.TotalGross, common.NewUUIDv7().String())

	order := placeOrderFromQuote(ctx, t, clients, customerUUID, quote, cardNumber)

	assertOrderMatchesQuote(t, order, quote)

	// Complete the order flow
	assertOrderVisibleToRestaurant(ctx, t, clients, restaurantUUID, order.OrderUuid)
	restaurantAcceptOrder(ctx, t, clients, restaurantUUID, order.OrderUuid)
	assertOrderVisibleToCourier(ctx, t, clients, courier.UUID, order.OrderUuid)
	courierAcceptDelivery(ctx, t, clients, courier.UUID, order.OrderUuid)
	restaurantMarkOrderReady(ctx, t, clients, restaurantUUID, order.OrderUuid)
	assertOrderReadyForCourier(ctx, t, clients, courier.UUID, order.OrderUuid)
	courierReportPickup(ctx, t, clients, courier.UUID, order.OrderUuid)
	assertPickupReported(ctx, t, clients, courier.UUID, order.OrderUuid)

	courierReportDelivered(ctx, t, clients, courier.UUID, order.OrderUuid)
	assertDeliveryReported(ctx, t, clients, courier.UUID, order.OrderUuid)
}

func TestComponent_OnboardingValidation(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()

	t.Run("invalid_country_code", func(t *testing.T) {
		t.Parallel()

		// Using an invalid country code that won't pass validation
		invalidCountry := shared.CountryCode{} // Zero value is invalid
		restaurant := ordersclient.OnboardRestaurant{
			Name:        "Test Restaurant",
			Description: "A test restaurant",
			Address: ordersclient.Address{
				Line1:       "123 Test St",
				Line2:       "",
				City:        "Test City",
				PostalCode:  "12345",
				CountryCode: invalidCountry,
			},
			MenuItems: []ordersclient.MenuItem{
				{
					Uuid:       app.RestaurantMenuItemUUID{common.NewUUIDv7()},
					Name:       "Test Item",
					Category:   app.ItemCategoryFood,
					GrossPrice: decimal.RequireFromString("10.00"),
					Ordering:   0.1,
				},
			},
			Currency: shared.MustNewCurrency("USD"),
		}

		restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
		resp, err := clients.Orders.OnboardRestaurantWithResponse(
			ctx,
			restaurantUUID,
			&ordersclient.OnboardRestaurantParams{
				OperatorUUID: common.NewUUIDv7(),
			},
			restaurant,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode(),
			"should reject invalid country code")
	})

	t.Run("negative_menu_item_price", func(t *testing.T) {
		t.Parallel()

		country := testutils.GenerateRandomCountry()
		restaurant := ordersclient.OnboardRestaurant{
			Name:        "Test Restaurant",
			Description: "A test restaurant",
			Address:     testutils.GenerateRandomOpenapiAddress(country),
			MenuItems: []ordersclient.MenuItem{
				{
					Uuid:       app.RestaurantMenuItemUUID{common.NewUUIDv7()},
					Name:       "Negative Price Item",
					Category:   app.ItemCategoryFood,
					GrossPrice: decimal.RequireFromString("-10.00"), // Invalid: negative price
					Ordering:   0.1,
				},
			},
			Currency: shared.MustNewCurrency("USD"),
		}

		restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
		resp, err := clients.Orders.OnboardRestaurantWithResponse(
			ctx,
			restaurantUUID,
			&ordersclient.OnboardRestaurantParams{
				OperatorUUID: common.NewUUIDv7(),
			},
			restaurant,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode(),
			"should reject negative menu item price")
	})

	t.Run("empty_restaurant_name", func(t *testing.T) {
		t.Parallel()

		country := testutils.GenerateRandomCountry()
		restaurant := ordersclient.OnboardRestaurant{
			Name:        "", // Invalid: empty name
			Description: "A test restaurant",
			Address:     testutils.GenerateRandomOpenapiAddress(country),
			MenuItems: []ordersclient.MenuItem{
				{
					Uuid:       app.RestaurantMenuItemUUID{common.NewUUIDv7()},
					Name:       "Test Item",
					Category:   app.ItemCategoryFood,
					GrossPrice: decimal.RequireFromString("10.00"),
					Ordering:   0.1,
				},
			},
			Currency: shared.MustNewCurrency("USD"),
		}

		restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
		resp, err := clients.Orders.OnboardRestaurantWithResponse(
			ctx,
			restaurantUUID,
			&ordersclient.OnboardRestaurantParams{
				OperatorUUID: common.NewUUIDv7(),
			},
			restaurant,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode(),
			"should reject empty restaurant name")
	})

	t.Run("empty_menu", func(t *testing.T) {
		t.Parallel()

		country := testutils.GenerateRandomCountry()
		restaurant := ordersclient.OnboardRestaurant{
			Name:        "Test Restaurant",
			Description: "A test restaurant",
			Address:     testutils.GenerateRandomOpenapiAddress(country),
			MenuItems:   []ordersclient.MenuItem{}, // Invalid: empty menu
			Currency:    shared.MustNewCurrency("USD"),
		}

		restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
		resp, err := clients.Orders.OnboardRestaurantWithResponse(
			ctx,
			restaurantUUID,
			&ordersclient.OnboardRestaurantParams{
				OperatorUUID: common.NewUUIDv7(),
			},
			restaurant,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode(),
			"should reject empty menu")
	})
}
