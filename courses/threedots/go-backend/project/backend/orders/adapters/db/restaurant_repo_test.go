// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
//go:build integration

package db_test

import (
	"context"
	"testing"

	gofakeit "github.com/brianvoe/gofakeit/v7"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/common/testutils"
	"eats/backend/orders/adapters/db"
	"eats/backend/orders/app"
)

func TestUpsertRestaurant_CreateNew(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)
	repo := db.NewRestaurantRepository(dbPool)

	restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
	onboardRestaurant := newTestOnboardRestaurant()

	err := repo.UpsertRestaurant(ctx, restaurantUUID, onboardRestaurant)
	require.NoError(t, err)

	// Verify menu was created
	menu, err := repo.GetRestaurantMenu(ctx, restaurantUUID)
	require.NoError(t, err)

	// Compare restaurant details using cmp.Diff (ignoring Positions which are compared separately)
	expectedMenu := app.RestaurantMenu{
		RestaurantName: onboardRestaurant.Name,
		Address:        onboardRestaurant.Address,
		Description:    onboardRestaurant.Description,
		Currency:       onboardRestaurant.Currency,
		Positions:      menu.Positions, // Use actual positions for comparison
	}

	if diff := cmp.Diff(
		expectedMenu,
		menu,
		cmpopts.EquateComparable(shared.SharedTypes...),
	); diff != "" {
		t.Errorf("restaurant menu mismatch (-want +got):\n%s", diff)
	}

	// Verify menu positions separately (sorted by ordering field)
	assert.Len(t, menu.Positions, len(onboardRestaurant.MenuItems))
	if diff := cmp.Diff(
		onboardRestaurant.MenuItems,
		menu.Positions,
		cmpopts.SortSlices(func(a, b app.MenuItem) bool { return a.Ordering < b.Ordering }),
	); diff != "" {
		t.Errorf("menu positions mismatch (-want +got):\n%s", diff)
	}
}

func TestUpsertRestaurant_UpdateExisting(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)
	repo := db.NewRestaurantRepository(dbPool)

	restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
	onboardRestaurant := newTestOnboardRestaurant()

	// Create restaurant
	err := repo.UpsertRestaurant(ctx, restaurantUUID, onboardRestaurant)
	require.NoError(t, err)

	// Update restaurant details
	updatedRestaurant := onboardRestaurant
	updatedRestaurant.Name = gofakeit.Company()
	updatedRestaurant.Address = testutils.GenerateRandomAddress(testutils.GenerateRandomCountry())
	updatedRestaurant.Description = gofakeit.LoremIpsumSentence(10)

	err = repo.UpsertRestaurant(ctx, restaurantUUID, updatedRestaurant)
	require.NoError(t, err)

	menu, err := repo.GetRestaurantMenu(ctx, restaurantUUID)
	require.NoError(t, err)

	// Compare restaurant details using cmp.Diff (ignoring Positions which are compared separately)
	expectedMenu := app.RestaurantMenu{
		RestaurantName: updatedRestaurant.Name,
		Address:        updatedRestaurant.Address,
		Description:    updatedRestaurant.Description,
		Currency:       updatedRestaurant.Currency,
		Positions:      menu.Positions, // Use actual positions for comparison
	}
	if diff := cmp.Diff(
		expectedMenu,
		menu,
		cmpopts.EquateComparable(shared.SharedTypes...),
	); diff != "" {
		t.Errorf("restaurant menu mismatch (-want +got):\n%s", diff)
	}
}

func TestUpsertRestaurant_UpdateMenuItems(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)
	repo := db.NewRestaurantRepository(dbPool)

	restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
	onboardRestaurant := newTestOnboardRestaurant()

	// Create restaurant
	err := repo.UpsertRestaurant(ctx, restaurantUUID, onboardRestaurant)
	require.NoError(t, err)

	// Update menu position prices and names
	updatedRestaurant := onboardRestaurant
	updatedRestaurant.MenuItems[0].Name = gofakeit.Dessert()
	updatedRestaurant.MenuItems[0].GrossPrice = newTestPrice()
	updatedRestaurant.MenuItems[1].Name = gofakeit.Lunch()
	updatedRestaurant.MenuItems[1].GrossPrice = newTestPrice()

	err = repo.UpsertRestaurant(ctx, restaurantUUID, updatedRestaurant)
	require.NoError(t, err)

	// Verify menu positions were updated
	menu, err := repo.GetRestaurantMenu(ctx, restaurantUUID)
	require.NoError(t, err)
	assert.Len(t, menu.Positions, len(updatedRestaurant.MenuItems))

	if diff := cmp.Diff(
		updatedRestaurant.MenuItems,
		menu.Positions,
		cmpopts.SortSlices(func(a, b app.MenuItem) bool { return a.Ordering < b.Ordering }),
	); diff != "" {
		t.Errorf("menu positions mismatch (-want +got):\n%s", diff)
	}
}

func TestUpsertRestaurant_ArchiveRemovedMenuItems(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)
	repo := db.NewRestaurantRepository(dbPool)

	restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
	onboardRestaurant := newTestOnboardRestaurant()

	// Ensure we have at least 3 menu positions for this test
	for len(onboardRestaurant.MenuItems) < 3 {
		onboardRestaurant.MenuItems = append(onboardRestaurant.MenuItems, newTestMenuItem())
	}

	// Create restaurant with menu
	err := repo.UpsertRestaurant(ctx, restaurantUUID, onboardRestaurant)
	require.NoError(t, err)

	// Get initial menu
	initialMenu, err := repo.GetRestaurantMenu(ctx, restaurantUUID)
	require.NoError(t, err)
	initialCount := len(initialMenu.Positions)
	assert.GreaterOrEqual(t, initialCount, 3)

	// Remove one menu position
	updatedRestaurant := onboardRestaurant
	removedPosition := updatedRestaurant.MenuItems[1]
	updatedRestaurant.MenuItems = append(
		updatedRestaurant.MenuItems[:1],
		updatedRestaurant.MenuItems[2:]...,
	)

	err = repo.UpsertRestaurant(ctx, restaurantUUID, updatedRestaurant)
	require.NoError(t, err)

	// Verify menu position was archived (not returned in menu)
	menu, err := repo.GetRestaurantMenu(ctx, restaurantUUID)
	require.NoError(t, err)
	assert.Len(t, menu.Positions, initialCount-1)

	// Verify removed position is not in the menu
	for _, pos := range menu.Positions {
		assert.NotEqual(t, removedPosition.MenuItemUUID, pos.MenuItemUUID)
	}
}

func TestUpsertRestaurant_PreventCurrencyChange(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)
	repo := db.NewRestaurantRepository(dbPool)

	restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
	onboardRestaurant := newTestOnboardRestaurant()
	onboardRestaurant.Currency = shared.MustNewCurrency("USD")

	// Create restaurant with USD currency
	err := repo.UpsertRestaurant(ctx, restaurantUUID, onboardRestaurant)
	require.NoError(t, err)

	// Try to change currency to EUR
	updatedRestaurant := onboardRestaurant
	updatedRestaurant.Currency = shared.MustNewCurrency("EUR")

	err = repo.UpsertRestaurant(ctx, restaurantUUID, updatedRestaurant)
	require.Error(t, err)

	// Verify error is about currency change
	var invalidInputErr common.Error
	require.ErrorAs(t, err, &invalidInputErr)
	assert.Equal(t, "cannot-change-currency", invalidInputErr.ErrorSlug)
}

func TestUpsertRestaurant_Idempotency(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)
	repo := db.NewRestaurantRepository(dbPool)

	restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
	onboardRestaurant := newTestOnboardRestaurant()

	// Upsert same restaurant twice
	err := repo.UpsertRestaurant(ctx, restaurantUUID, onboardRestaurant)
	require.NoError(t, err)

	err = repo.UpsertRestaurant(ctx, restaurantUUID, onboardRestaurant)
	require.NoError(t, err)

	// Verify restaurant exists with correct data
	menu, err := repo.GetRestaurantMenu(ctx, restaurantUUID)
	require.NoError(t, err)

	// Compare restaurant details using cmp.Diff (ignoring Positions which are compared separately)
	expectedMenu := app.RestaurantMenu{
		RestaurantName: onboardRestaurant.Name,
		Address:        onboardRestaurant.Address,
		Description:    onboardRestaurant.Description,
		Currency:       onboardRestaurant.Currency,
		Positions:      menu.Positions, // Use actual positions for comparison
	}
	if diff := cmp.Diff(
		expectedMenu,
		menu,
		cmpopts.EquateComparable(shared.SharedTypes...),
	); diff != "" {
		t.Errorf("restaurant menu mismatch (-want +got):\n%s", diff)
	}

	assert.Len(t, menu.Positions, len(onboardRestaurant.MenuItems))
}

func TestGetRestaurantMenu(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)
	repo := db.NewRestaurantRepository(dbPool)

	restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
	onboardRestaurant := newTestOnboardRestaurant()

	// Create restaurant
	err := repo.UpsertRestaurant(ctx, restaurantUUID, onboardRestaurant)
	require.NoError(t, err)

	// Test GetRestaurantMenu method
	menu, err := repo.GetRestaurantMenu(ctx, restaurantUUID)
	require.NoError(t, err)

	// Compare restaurant details using cmp.Diff (ignoring Positions which are compared separately)
	expectedMenu := app.RestaurantMenu{
		RestaurantName: onboardRestaurant.Name,
		Address:        onboardRestaurant.Address,
		Description:    onboardRestaurant.Description,
		Currency:       onboardRestaurant.Currency,
		Positions:      menu.Positions, // Use actual positions for comparison
	}
	if diff := cmp.Diff(
		expectedMenu,
		menu,
		cmpopts.EquateComparable(shared.SharedTypes...),
	); diff != "" {
		t.Errorf("restaurant menu mismatch (-want +got):\n%s", diff)
	}

	// Verify menu positions separately (sorted by ordering field)
	assert.Len(t, menu.Positions, len(onboardRestaurant.MenuItems))
	if diff := cmp.Diff(
		onboardRestaurant.MenuItems,
		menu.Positions,
		cmpopts.SortSlices(func(a, b app.MenuItem) bool { return a.Ordering < b.Ordering }),
	); diff != "" {
		t.Errorf("menu positions mismatch (-want +got):\n%s", diff)
	}
}

func TestGetRestaurantMenu_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)
	repo := db.NewRestaurantRepository(dbPool)

	// Try to get menu of non-existent restaurant
	restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
	_, err := repo.GetRestaurantMenu(ctx, restaurantUUID)
	require.Error(t, err)

	// Verify error is NotFoundError
	var notFoundErr common.Error
	require.ErrorAs(t, err, &notFoundErr)
	assert.Equal(t, "restaurant-not-found", notFoundErr.ErrorSlug)
}

func newTestOnboardRestaurant() app.OnboardRestaurant {
	return app.OnboardRestaurant{
		Name:        gofakeit.Company(),
		Address:     testutils.GenerateRandomAddress(testutils.GenerateRandomCountry()),
		Currency:    shared.MustNewCurrency("USD"),
		Description: gofakeit.LoremIpsumSentence(10),
		MenuItems: []app.MenuItem{
			newTestMenuItem(),
			newTestMenuItem(),
		},
	}
}

func newTestMenuItem() app.MenuItem {
	return app.MenuItem{
		MenuItemUUID: app.RestaurantMenuItemUUID{common.NewUUIDv7()},
		Name:         gofakeit.Dinner(),
		Ordering:     gofakeit.Float64Range(1, 100),
		GrossPrice:   newTestPrice(),
	}
}

func newTestPrice() decimal.Decimal {
	// Generate price with 2 decimal places (standard for USD)
	return decimal.New(int64(gofakeit.IntRange(1000, 10000)), -2)
}

func newTestCustomer(uuid app.CustomerUUID) app.Customer {
	return app.Customer{
		CustomerUUID: uuid,
		Name:         gofakeit.Name(),
		Email:        gofakeit.Email(),
		Address:      testutils.GenerateRandomAddress(testutils.GenerateRandomCountry()),
		PhoneNumber:  gofakeit.Phone(),
	}
}
