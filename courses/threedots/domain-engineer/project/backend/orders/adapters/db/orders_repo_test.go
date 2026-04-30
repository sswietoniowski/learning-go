// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
//go:build integration

package db_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/common/testutils"
	"eats/backend/orders/adapters/db"
	"eats/backend/orders/adapters/db/dbmodels"
	"eats/backend/orders/app"
)

func TestCreateQuote(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)
	ordersRepo := db.NewOrdersRepository(dbPool)
	restaurantRepo := db.NewRestaurantRepository(dbPool)
	customerRepo := db.NewCustomerRepository(dbPool)

	// Setup: Create a restaurant with menu positions
	restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
	restaurant := newTestOnboardRestaurant()
	err := restaurantRepo.UpsertRestaurant(ctx, restaurantUUID, restaurant)
	require.NoError(t, err)

	// Create a customer
	customerUUID := app.CustomerUUID{common.NewUUIDv7()}
	customer := newTestCustomer(customerUUID)
	err = customerRepo.RegisterCustomer(ctx, customer)
	require.NoError(t, err)

	deliveryAddress := testutils.GenerateRandomAddress(testutils.GenerateRandomCountry())

	quoteMenuItems := []app.CreateQuoteItem{
		{MenuItemUUID: restaurant.MenuItems[0].MenuItemUUID, Quantity: 2},
		{MenuItemUUID: restaurant.MenuItems[1].MenuItemUUID, Quantity: 1},
	}

	var createdQuote app.Quote
	var positions []app.QuoteMenuItem
	quote, err := ordersRepo.CreateQuote(ctx, restaurantUUID, quoteMenuItems, func(
		ctx context.Context,
		menuItems map[app.RestaurantMenuItemUUID]app.MenuItem,
		r app.Restaurant,
	) (app.Quote, []app.QuoteMenuItem, error) {
		// Verify menu positions were fetched correctly
		require.Len(t, menuItems, 2)
		if diff := cmp.Diff(
			app.Restaurant{
				RestaurantUUID: restaurantUUID,
				Name:           restaurant.Name,
				Description:    restaurant.Description,
				Address:        restaurant.Address,
				Currency:       restaurant.Currency,
			},
			r,
			cmpopts.EquateComparable(shared.SharedTypes...),
		); diff != "" {
			t.Errorf("restaurant mismatch (-want +got):\n%s", diff)
		}

		// Calculate totals
		itemsSubtotal := decimal.Zero
		positions = make([]app.QuoteMenuItem, 0, len(quoteMenuItems))

		for _, qmp := range quoteMenuItems {
			mp, ok := menuItems[qmp.MenuItemUUID]
			require.True(t, ok)
			itemsSubtotal = itemsSubtotal.Add(mp.GrossPrice.Mul(decimal.NewFromInt(int64(qmp.Quantity))))

			positions = append(positions, app.QuoteMenuItem{
				MenuItemUUID: qmp.MenuItemUUID,
				GrossPrice:   mp.GrossPrice,
				Quantity:     qmp.Quantity,
			})
		}

		serviceFee := decimal.NewFromFloat(5.00)
		deliveryFee := decimal.NewFromFloat(3.00)
		totalAmount := itemsSubtotal.Add(serviceFee).Add(deliveryFee)
		totalTax := totalAmount.Mul(decimal.NewFromFloat(0.1)).RoundBank(2)

		createdQuote = app.Quote{
			QuoteUUID:          app.QuoteUUID{common.NewUUIDv7()},
			CustomerUUID:       customerUUID,
			RestaurantUUID:     restaurantUUID,
			DeliveryAddress:    deliveryAddress,
			ItemsSubtotalGross: itemsSubtotal,
			ServiceFeeGross:    serviceFee,
			DeliveryFeeGross:   deliveryFee,
			TotalAmountGross:   totalAmount,
			TotalTax:           totalTax,
			Currency:           r.Currency,
			CreatedAt:          time.Now(),
		}

		return createdQuote, positions, nil
	})

	require.NoError(t, err)

	cmpOpts := cmp.Options{
		cmpopts.EquateComparable(shared.SharedTypes...),
		cmp.Comparer(func(a, b time.Time) bool {
			return a.Truncate(time.Second).Equal(b.Truncate(time.Second))
		}),
		cmp.Comparer(func(a, b decimal.Decimal) bool {
			return a.Equal(b)
		}),
	}

	// Verify the returned quote matches what we created
	if diff := cmp.Diff(createdQuote, quote, cmpOpts); diff != "" {
		t.Errorf("quote mismatch (-want +got):\n%s", diff)
	}

	// Verify the quote was persisted correctly
	queries := dbmodels.New(dbPool)
	persistedQuote, err := queries.GetQuote(ctx, quote.QuoteUUID)
	require.NoError(t, err)

	persistedAsAppQuote := app.Quote{
		QuoteUUID:          persistedQuote.QuoteUuid,
		CustomerUUID:       persistedQuote.CustomerUuid,
		RestaurantUUID:     persistedQuote.RestaurantUuid,
		DeliveryAddress:    persistedQuote.DeliveryAddress,
		ItemsSubtotalGross: persistedQuote.ItemsSubtotalGross,
		ServiceFeeGross:    persistedQuote.ServiceFeeGross,
		DeliveryFeeGross:   persistedQuote.DeliveryFeeGross,
		TotalAmountGross:   persistedQuote.TotalAmountGross,
		TotalTax:           persistedQuote.TotalTax,
		Currency:           persistedQuote.Currency,
		CreatedAt:          persistedQuote.CreatedAt,
	}
	if diff := cmp.Diff(createdQuote, persistedAsAppQuote, cmpOpts); diff != "" {
		t.Errorf("persisted quote mismatch (-want +got):\n%s", diff)
	}

	// Verify quote items were persisted correctly
	quoteItems, err := queries.GetQuoteItems(ctx, quote.QuoteUUID)
	require.NoError(t, err)
	require.Len(t, quoteItems, len(positions))

	persistedPositions := make([]app.QuoteMenuItem, 0, len(quoteItems))
	for _, item := range quoteItems {
		persistedPositions = append(persistedPositions, app.QuoteMenuItem{
			MenuItemUUID: item.MenuItemUuid,
			GrossPrice:   item.GrossPrice,
			Quantity:     int(item.Quantity),
		})
	}
	if diff := cmp.Diff(positions, persistedPositions, cmpOpts, cmpopts.SortSlices(func(a, b app.QuoteMenuItem) bool {
		return a.MenuItemUUID.String() < b.MenuItemUUID.String()
	})); diff != "" {
		t.Errorf("persisted quote items mismatch (-want +got):\n%s", diff)
	}
}

func TestPlaceOrder(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)
	ordersRepo := db.NewOrdersRepository(dbPool)
	restaurantRepo := db.NewRestaurantRepository(dbPool)
	customerRepo := db.NewCustomerRepository(dbPool)

	// Setup: Create restaurant and quote
	restaurantUUID, quote, quoteMenuItems := setupRestaurantAndQuote(t, ctx, restaurantRepo, ordersRepo, customerRepo)

	// Fetch quote with menu items
	fetchedQuote, menuItems, err := ordersRepo.QuoteWithMenuItems(ctx, quote.QuoteUUID)
	require.NoError(t, err)
	assert.Equal(t, quote.QuoteUUID, fetchedQuote.QuoteUUID)
	assert.Equal(t, quote.CustomerUUID, fetchedQuote.CustomerUUID)
	assert.Equal(t, restaurantUUID, fetchedQuote.RestaurantUUID)
	require.Len(t, menuItems, len(quoteMenuItems))

	// Create and save order from quote
	order, err := app.NewOrderFromQuote(fetchedQuote)
	require.NoError(t, err)

	err = ordersRepo.SaveOrder(ctx, order)
	require.NoError(t, err)

	// Verify order was created by fetching it
	fetchedOrder, err := ordersRepo.OrderByID(ctx, order.OrderUUID)
	require.NoError(t, err)

	if diff := cmp.Diff(
		order,
		fetchedOrder,
		cmpopts.EquateComparable(shared.SharedTypes...),
		cmp.Comparer(func(a, b time.Time) bool {
			return a.Truncate(time.Second).Equal(b.Truncate(time.Second))
		}),
	); diff != "" {
		t.Errorf("order mismatch (-want +got):\n%s", diff)
	}
}

func TestOrderByID_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)
	ordersRepo := db.NewOrdersRepository(dbPool)

	// Try to get non-existent order
	nonExistentUUID := app.OrderUUID{common.NewUUIDv7()}
	_, err := ordersRepo.OrderByID(ctx, nonExistentUUID)
	require.Error(t, err)

	// Verify it's a NotFoundError from common
	var commonErr common.Error
	require.True(t, errors.As(err, &commonErr))
	assert.Equal(t, http.StatusNotFound, commonErr.HttpErrorCode)
	assert.Equal(t, "order_not_found", commonErr.ErrorSlug)
}

func TestUpdateOrder(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)
	ordersRepo := db.NewOrdersRepository(dbPool)
	restaurantRepo := db.NewRestaurantRepository(dbPool)
	customerRepo := db.NewCustomerRepository(dbPool)
	courierRepo := db.NewCourierRepository(dbPool)

	// Setup: Create restaurant, quote, and order
	_, _, originalOrder := setupRestaurantQuoteAndOrder(t, ctx, restaurantRepo, ordersRepo, customerRepo)

	// Create a courier
	courierUUID := app.CourierUUID{common.NewUUIDv7()}
	courier := newTestCourier()
	err := courierRepo.RegisterCourier(ctx, courierUUID, courier)
	require.NoError(t, err)

	// Update order - add courier and all timestamps
	confirmedAt := time.Now()
	acceptedAt := time.Now().Add(5 * time.Minute)
	preparedAt := time.Now().Add(10 * time.Minute)
	pickedUpAt := time.Now().Add(15 * time.Minute)
	deliveredAt := time.Now().Add(30 * time.Minute)

	var expectedOrder app.Order
	err = ordersRepo.UpdateOrder(ctx, originalOrder.OrderUUID, func(ctx context.Context, order app.Order) (app.Order, error) {
		// Verify order state
		assert.Equal(t, originalOrder.OrderUUID, order.OrderUUID)
		assert.Nil(t, order.RestaurantConfirmedAt)
		assert.Nil(t, order.CourierAcceptedAt)

		// Update order using struct without named fields to ensure that when new fields are added,
		// the compiler will force us to update this test.
		expectedOrder = app.Order{
			order.OrderUUID,
			order.QuoteUUID,
			order.CustomerUUID,
			order.RestaurantUUID,
			&courierUUID,
			order.DeliveryAddress,
			order.OrderedAt,
			&confirmedAt,
			&acceptedAt,
			&preparedAt,
			&pickedUpAt,
			&deliveredAt,
			order.ItemsSubtotal,
			order.ServiceFeeGross,
			order.DeliveryFeeGross,
			order.TotalAmountGross,
			order.TotalTax,
			order.Currency,
		}
		return expectedOrder, nil
	})
	require.NoError(t, err)

	// Verify all fields were updated correctly
	updatedOrder, err := ordersRepo.OrderByID(ctx, originalOrder.OrderUUID)
	require.NoError(t, err)

	if diff := cmp.Diff(
		expectedOrder,
		updatedOrder,
		cmpopts.EquateComparable(shared.SharedTypes...),
		cmp.Comparer(func(a, b time.Time) bool {
			return a.Truncate(time.Second).Equal(b.Truncate(time.Second))
		}),
	); diff != "" {
		t.Errorf("order mismatch (-want +got):\n%s", diff)
	}
}

func TestUpdateOrder_Idempotency(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)
	ordersRepo := db.NewOrdersRepository(dbPool)
	restaurantRepo := db.NewRestaurantRepository(dbPool)
	customerRepo := db.NewCustomerRepository(dbPool)

	// Setup: Create restaurant, quote, and order
	_, _, order := setupRestaurantQuoteAndOrder(t, ctx, restaurantRepo, ordersRepo, customerRepo)

	// Update order multiple times with same data
	confirmedAt := time.Now()

	for i := 0; i < 3; i++ {
		err := ordersRepo.UpdateOrder(ctx, order.OrderUUID, func(ctx context.Context, order app.Order) (app.Order, error) {
			order.RestaurantConfirmedAt = &confirmedAt
			return order, nil
		})
		require.NoError(t, err)
	}

	// Verify order state is consistent
	fetchedOrder, err := ordersRepo.OrderByID(ctx, order.OrderUUID)
	require.NoError(t, err)

	assert.NotNil(t, fetchedOrder.RestaurantConfirmedAt)
	assert.True(t, fetchedOrder.RestaurantConfirmedAt.Truncate(time.Second).Equal(confirmedAt.Truncate(time.Second)))
}

func setupRestaurantAndQuote(t *testing.T, ctx context.Context, restaurantRepo *db.RestaurantRepository, ordersRepo *db.OrdersRepo, customerRepo *db.CustomerRepository) (app.RestaurantUUID, app.Quote, []app.CreateQuoteItem) {
	// Create restaurant
	restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
	restaurant := newTestOnboardRestaurant()
	err := restaurantRepo.UpsertRestaurant(ctx, restaurantUUID, restaurant)
	require.NoError(t, err)

	// Create customer
	customerUUID := app.CustomerUUID{common.NewUUIDv7()}
	customer := newTestCustomer(customerUUID)
	err = customerRepo.RegisterCustomer(ctx, customer)
	require.NoError(t, err)

	deliveryAddress := testutils.GenerateRandomAddress(testutils.GenerateRandomCountry())
	quoteMenuItems := []app.CreateQuoteItem{
		{MenuItemUUID: restaurant.MenuItems[0].MenuItemUUID, Quantity: 1},
	}

	quote, err := ordersRepo.CreateQuote(ctx, restaurantUUID, quoteMenuItems, func(
		ctx context.Context,
		menuItems map[app.RestaurantMenuItemUUID]app.MenuItem,
		r app.Restaurant,
	) (app.Quote, []app.QuoteMenuItem, error) {
		itemsSubtotal := decimal.Zero
		quoteItems := make([]app.QuoteMenuItem, 0, len(quoteMenuItems))

		for _, qmp := range quoteMenuItems {
			mp := menuItems[qmp.MenuItemUUID]
			itemsSubtotal = itemsSubtotal.Add(mp.GrossPrice.Mul(decimal.NewFromInt(int64(qmp.Quantity))))

			quoteItems = append(quoteItems, app.QuoteMenuItem{
				MenuItemUUID: qmp.MenuItemUUID,
				Quantity:     qmp.Quantity,
				GrossPrice:   mp.GrossPrice,
			})
		}

		serviceFee := decimal.NewFromFloat(5.00)
		deliveryFee := decimal.NewFromFloat(3.00)
		totalAmount := itemsSubtotal.Add(serviceFee).Add(deliveryFee)
		totalTax := totalAmount.Mul(decimal.NewFromFloat(0.1)).RoundBank(2)

		return app.Quote{
			QuoteUUID:          app.QuoteUUID{common.NewUUIDv7()},
			CustomerUUID:       customerUUID,
			RestaurantUUID:     restaurantUUID,
			DeliveryAddress:    deliveryAddress,
			ItemsSubtotalGross: itemsSubtotal,
			ServiceFeeGross:    serviceFee,
			DeliveryFeeGross:   deliveryFee,
			TotalAmountGross:   totalAmount,
			TotalTax:           totalTax,
			Currency:           r.Currency,
			CreatedAt:          time.Now(),
		}, quoteItems, nil
	})
	require.NoError(t, err)

	return restaurantUUID, quote, quoteMenuItems
}

func setupRestaurantQuoteAndOrder(t *testing.T, ctx context.Context, restaurantRepo *db.RestaurantRepository, ordersRepo *db.OrdersRepo, customerRepo *db.CustomerRepository) (app.RestaurantUUID, app.Quote, app.Order) {
	restaurantUUID, quote, _ := setupRestaurantAndQuote(t, ctx, restaurantRepo, ordersRepo, customerRepo)

	order, err := app.NewOrderFromQuote(quote)
	require.NoError(t, err)

	err = ordersRepo.SaveOrder(ctx, order)
	require.NoError(t, err)

	return restaurantUUID, quote, order
}
