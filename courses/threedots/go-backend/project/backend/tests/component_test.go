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
)

func TestComponent_CriticalFlow(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()

	country := testutils.GenerateRandomCountry()

	restaurant := onboardRestaurant(ctx, t, clients, country)
	assertRestaurantMenuPublished(ctx, t, clients, restaurant.UUID, restaurant.Data)

	courier := registerCourierInCity(ctx, t, clients, country, restaurant.Data.Address.City)

	customerUUID := registerCustomerInCity(ctx, t, clients, country, restaurant.Data.Address.City)

	orderUUID := placeOrder(ctx, t, clients, customerUUID, restaurant.UUID, restaurant.Data, country)

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
	city := testutils.GenerateRandomAddress(country).City

	// Register a courier
	courier := registerCourierInCity(ctx, t, clients, country, city)
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

	restaurant := onboardRestaurant(ctx, t, clients, country)
	assertRestaurantMenuPublished(ctx, t, clients, restaurant.UUID, restaurant.Data)

	courier := registerCourierInCity(ctx, t, clients, country, restaurant.Data.Address.City)
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

	order := placeOrderFromQuote(ctx, t, clients, customerUUID, restaurant.UUID, quote)
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
}

func TestComponent_CourierCityFiltering(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()
	country := testutils.GenerateRandomCountry()

	restaurant := onboardRestaurant(ctx, t, clients, country)
	customerUUID := registerCustomerInCity(ctx, t, clients, country, restaurant.Data.Address.City)

	courierInSameCity := registerCourierInCity(ctx, t, clients, country, restaurant.Data.Address.City)

	differentCity := restaurant.Data.Address.City + "_Different"
	courierInDifferentCity := registerCourierInCity(ctx, t, clients, country, differentCity)

	orderUUID := placeOrder(ctx, t, clients, customerUUID, restaurant.UUID, restaurant.Data, country)

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

	restaurant := onboardRestaurant(ctx, t, clients, country)
	assertRestaurantMenuPublished(ctx, t, clients, restaurant.UUID, restaurant.Data)

	// Onboard a second restaurant (the "wrong" one)
	wrongRestaurant := onboardRestaurant(ctx, t, clients, country)

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
	order := placeOrderFromQuote(ctx, t, clients, customerUUID, restaurant.UUID, quote)

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

	restaurant := onboardRestaurant(ctx, t, clients, country)
	assertRestaurantMenuPublished(ctx, t, clients, restaurant.UUID, restaurant.Data)

	courierA := registerCourierInCity(ctx, t, clients, country, restaurant.Data.Address.City)
	courierB := registerCourierInCity(ctx, t, clients, country, restaurant.Data.Address.City)
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
	order := placeOrderFromQuote(ctx, t, clients, customerUUID, restaurant.UUID, quote)

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

	restaurant1 := onboardRestaurant(ctx, t, clients, country)
	restaurant2 := onboardRestaurant(ctx, t, clients, country)
	restaurant3 := onboardRestaurant(ctx, t, clients, country)

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
	originalRestaurant := onboardRestaurant(ctx, t, clients, country)

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

	restaurant := onboardRestaurant(ctx, t, clients, country)
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
	order := placeOrderFromQuote(ctx, t, clients, customerUUID, restaurant.UUID, quote)
	require.NotEmpty(t, order.OrderUuid)
	assertOrderMatchesQuote(t, order, quote)
}

func TestComponent_CreateQuoteOutOfDeliveryZone(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()

	country := testutils.GenerateRandomCountry()

	// Onboard restaurant in one city
	restaurant := onboardRestaurant(ctx, t, clients, country)
	customerUUID := registerCustomer(ctx, t, clients, country)

	// Ensure the delivery city is different from restaurant city
	deliveryAddress := testutils.GenerateRandomOpenapiAddress(country)
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
	restaurant := onboardRestaurant(ctx, t, clients, country)
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

	_, cardNumber := createBankAccountWithBalance(ctx, t, clients, decimal.NewFromInt(1000), common.NewUUIDv7().String())
	createBankAccount(ctx, t, clients, restaurant.UUID.String())
	nonce := preauthPayment(
		ctx,
		t,
		clients,
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

func TestComponent_CreateQuoteValidationErrors(t *testing.T) {
	t.Parallel()
	clients := newTestClients(t)

	ctx := t.Context()

	country := testutils.GenerateRandomCountry()

	restaurant := onboardRestaurant(ctx, t, clients, country)
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
