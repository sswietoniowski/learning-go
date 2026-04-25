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
	"github.com/stretchr/testify/require"

	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/common/testutils"
	"eats/backend/orders/adapters/db"
	"eats/backend/orders/api/http"
	"eats/backend/orders/app"
)

func TestReadModel_ListMenuItemsWithRestaurant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)

	restaurantRepo := db.NewRestaurantRepository(dbPool)
	readModel := db.NewReadModel(dbPool)

	// Create test restaurants with menu items
	restaurant1UUID := app.RestaurantUUID{common.NewUUIDv7()}
	restaurant1 := app.OnboardRestaurant{
		Name:        "Pizza Palace",
		Address:     testutils.GenerateRandomAddress(testutils.GenerateRandomCountry()),
		Currency:    shared.MustNewCurrency("USD"),
		Description: gofakeit.LoremIpsumSentence(5),
		MenuItems: []app.MenuItem{
			{
				MenuItemUUID: app.RestaurantMenuItemUUID{common.NewUUIDv7()},
				Name:         "Margherita Pizza",
				GrossPrice:   decimal.NewFromFloat(12.99),
				Ordering:     1,
			},
			{
				MenuItemUUID: app.RestaurantMenuItemUUID{common.NewUUIDv7()},
				Name:         "Pepperoni Pizza",
				GrossPrice:   decimal.NewFromFloat(14.99),
				Ordering:     2,
			},
		},
	}
	err := restaurantRepo.UpsertRestaurant(ctx, restaurant1UUID, restaurant1)
	require.NoError(t, err)

	restaurant2UUID := app.RestaurantUUID{common.NewUUIDv7()}
	restaurant2 := app.OnboardRestaurant{
		Name:        "Burger Barn",
		Address:     testutils.GenerateRandomAddress(testutils.GenerateRandomCountry()),
		Currency:    shared.MustNewCurrency("USD"),
		Description: gofakeit.LoremIpsumSentence(5),
		MenuItems: []app.MenuItem{
			{
				MenuItemUUID: app.RestaurantMenuItemUUID{common.NewUUIDv7()},
				Name:         "Classic Burger",
				GrossPrice:   decimal.NewFromFloat(9.99),
				Ordering:     1,
			},
		},
	}
	err = restaurantRepo.UpsertRestaurant(ctx, restaurant2UUID, restaurant2)
	require.NoError(t, err)

	// Call the read model with empty filter
	items, err := readModel.ListMenuItemsWithRestaurant(ctx, http.ListMenuItemsFilter{})
	require.NoError(t, err)

	// Build expected results
	expected := []http.MenuItemWithRestaurant{
		{
			MenuItemUuid:   restaurant2.MenuItems[0].MenuItemUUID,
			MenuItemName:   "Classic Burger",
			GrossPrice:     decimal.NewFromFloat(9.99),
			Currency:       shared.MustNewCurrency("USD"),
			RestaurantUuid: restaurant2UUID,
			RestaurantName: "Burger Barn",
		},
		{
			MenuItemUuid:   restaurant1.MenuItems[0].MenuItemUUID,
			MenuItemName:   "Margherita Pizza",
			GrossPrice:     decimal.NewFromFloat(12.99),
			Currency:       shared.MustNewCurrency("USD"),
			RestaurantUuid: restaurant1UUID,
			RestaurantName: "Pizza Palace",
		},
		{
			MenuItemUuid:   restaurant1.MenuItems[1].MenuItemUUID,
			MenuItemName:   "Pepperoni Pizza",
			GrossPrice:     decimal.NewFromFloat(14.99),
			Currency:       shared.MustNewCurrency("USD"),
			RestaurantUuid: restaurant1UUID,
			RestaurantName: "Pizza Palace",
		},
	}

	// Filter items to only include ones from our test restaurants
	var testItems []http.MenuItemWithRestaurant
	for _, item := range items {
		if item.RestaurantUuid == restaurant1UUID || item.RestaurantUuid == restaurant2UUID {
			testItems = append(testItems, item)
		}
	}

	// Compare using cmp.Diff with sorting to handle any order differences
	if diff := cmp.Diff(
		expected,
		testItems,
		cmpopts.EquateComparable(shared.SharedTypes...),
		cmpopts.SortSlices(func(a, b http.MenuItemWithRestaurant) bool {
			if a.RestaurantName != b.RestaurantName {
				return a.RestaurantName < b.RestaurantName
			}
			return a.MenuItemName < b.MenuItemName
		}),
	); diff != "" {
		t.Errorf("menu items mismatch (-want +got):\n%s", diff)
	}
}

func TestReadModel_ListMenuItemsWithRestaurant_ExcludesArchived(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)

	restaurantRepo := db.NewRestaurantRepository(dbPool)
	readModel := db.NewReadModel(dbPool)

	// Create a restaurant with multiple menu items
	restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
	menuItem1 := app.MenuItem{
		MenuItemUUID: app.RestaurantMenuItemUUID{common.NewUUIDv7()},
		Name:         "Active Item",
		GrossPrice:   decimal.NewFromFloat(10.00),
		Ordering:     1,
	}
	menuItem2 := app.MenuItem{
		MenuItemUUID: app.RestaurantMenuItemUUID{common.NewUUIDv7()},
		Name:         "To Be Archived",
		GrossPrice:   decimal.NewFromFloat(15.00),
		Ordering:     2,
	}

	restaurant := app.OnboardRestaurant{
		Name:        "Test Restaurant " + gofakeit.UUID(),
		Address:     testutils.GenerateRandomAddress(testutils.GenerateRandomCountry()),
		Currency:    shared.MustNewCurrency("USD"),
		Description: gofakeit.LoremIpsumSentence(5),
		MenuItems:   []app.MenuItem{menuItem1, menuItem2},
	}
	err := restaurantRepo.UpsertRestaurant(ctx, restaurantUUID, restaurant)
	require.NoError(t, err)

	// Verify both items are returned initially
	items, err := readModel.ListMenuItemsWithRestaurant(ctx, http.ListMenuItemsFilter{})
	require.NoError(t, err)

	countBefore := 0
	for _, item := range items {
		if item.RestaurantUuid == restaurantUUID {
			countBefore++
		}
	}
	require.Equal(t, 2, countBefore, "should have 2 items before archiving")

	// Update restaurant to remove one menu item (archives it)
	restaurant.MenuItems = []app.MenuItem{menuItem1}
	err = restaurantRepo.UpsertRestaurant(ctx, restaurantUUID, restaurant)
	require.NoError(t, err)

	// Verify only active item is returned
	items, err = readModel.ListMenuItemsWithRestaurant(ctx, http.ListMenuItemsFilter{})
	require.NoError(t, err)

	countAfter := 0
	for _, item := range items {
		if item.RestaurantUuid == restaurantUUID {
			countAfter++
			require.Equal(t, "Active Item", item.MenuItemName, "only active item should be returned")
		}
	}
	require.Equal(t, 1, countAfter, "should have 1 item after archiving")
}

func TestReadModel_ListMenuItemsWithRestaurant_FilterByRestaurantName(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)

	restaurantRepo := db.NewRestaurantRepository(dbPool)
	readModel := db.NewReadModel(dbPool)

	// Create two restaurants with distinct names
	pizzaUUID := app.RestaurantUUID{common.NewUUIDv7()}
	pizzaRestaurant := app.OnboardRestaurant{
		Name:        "Pizza Palace",
		Address:     testutils.GenerateRandomAddress(testutils.GenerateRandomCountry()),
		Currency:    shared.MustNewCurrency("USD"),
		Description: gofakeit.LoremIpsumSentence(5),
		MenuItems: []app.MenuItem{
			{
				MenuItemUUID: app.RestaurantMenuItemUUID{common.NewUUIDv7()},
				Name:         "Margherita",
				GrossPrice:   decimal.NewFromFloat(12.99),
				Ordering:     1,
			},
		},
	}
	err := restaurantRepo.UpsertRestaurant(ctx, pizzaUUID, pizzaRestaurant)
	require.NoError(t, err)

	burgerUUID := app.RestaurantUUID{common.NewUUIDv7()}
	burgerRestaurant := app.OnboardRestaurant{
		Name:        "Burger Barn",
		Address:     testutils.GenerateRandomAddress(testutils.GenerateRandomCountry()),
		Currency:    shared.MustNewCurrency("USD"),
		Description: gofakeit.LoremIpsumSentence(5),
		MenuItems: []app.MenuItem{
			{
				MenuItemUUID: app.RestaurantMenuItemUUID{common.NewUUIDv7()},
				Name:         "Classic Burger",
				GrossPrice:   decimal.NewFromFloat(9.99),
				Ordering:     1,
			},
		},
	}
	err = restaurantRepo.UpsertRestaurant(ctx, burgerUUID, burgerRestaurant)
	require.NoError(t, err)

	// Filter by "Pizza" - should only return Pizza Palace items
	pizzaFilter := "Pizza"
	items, err := readModel.ListMenuItemsWithRestaurant(ctx, http.ListMenuItemsFilter{
		RestaurantName: &pizzaFilter,
	})
	require.NoError(t, err)

	// Check that all returned items belong to Pizza Palace
	for _, item := range items {
		if item.RestaurantUuid == pizzaUUID || item.RestaurantUuid == burgerUUID {
			require.Equal(t, pizzaUUID, item.RestaurantUuid, "filter should only return Pizza Palace items")
			require.Contains(t, item.RestaurantName, "Pizza")
		}
	}

	// Filter by "burger" (lowercase) - should still match due to case-insensitive search
	burgerFilter := "burger"
	items, err = readModel.ListMenuItemsWithRestaurant(ctx, http.ListMenuItemsFilter{
		RestaurantName: &burgerFilter,
	})
	require.NoError(t, err)

	for _, item := range items {
		if item.RestaurantUuid == pizzaUUID || item.RestaurantUuid == burgerUUID {
			require.Equal(t, burgerUUID, item.RestaurantUuid, "case-insensitive filter should match Burger Barn")
			require.Contains(t, item.RestaurantName, "Burger")
		}
	}
}

func TestReadModel_ListMenuItemsWithRestaurant_OrderByPrice(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)

	restaurantRepo := db.NewRestaurantRepository(dbPool)
	readModel := db.NewReadModel(dbPool)

	// Create a restaurant with menu items at various prices
	restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
	restaurantName := "Price Test Restaurant " + gofakeit.UUID()
	restaurant := app.OnboardRestaurant{
		Name:        restaurantName,
		Address:     testutils.GenerateRandomAddress(testutils.GenerateRandomCountry()),
		Currency:    shared.MustNewCurrency("USD"),
		Description: gofakeit.LoremIpsumSentence(5),
		MenuItems: []app.MenuItem{
			{
				MenuItemUUID: app.RestaurantMenuItemUUID{common.NewUUIDv7()},
				Name:         "Expensive",
				GrossPrice:   decimal.NewFromFloat(50.00),
				Ordering:     1,
			},
			{
				MenuItemUUID: app.RestaurantMenuItemUUID{common.NewUUIDv7()},
				Name:         "Cheap",
				GrossPrice:   decimal.NewFromFloat(5.00),
				Ordering:     2,
			},
			{
				MenuItemUUID: app.RestaurantMenuItemUUID{common.NewUUIDv7()},
				Name:         "Medium",
				GrossPrice:   decimal.NewFromFloat(25.00),
				Ordering:     3,
			},
		},
	}
	err := restaurantRepo.UpsertRestaurant(ctx, restaurantUUID, restaurant)
	require.NoError(t, err)

	// Order by price ascending
	orderByPriceAsc := "price_asc"
	nameFilter := restaurantName
	items, err := readModel.ListMenuItemsWithRestaurant(ctx, http.ListMenuItemsFilter{
		RestaurantName: &nameFilter,
		OrderBy:        &orderByPriceAsc,
	})
	require.NoError(t, err)

	// Filter to our test restaurant's items
	var testItems []http.MenuItemWithRestaurant
	for _, item := range items {
		if item.RestaurantUuid == restaurantUUID {
			testItems = append(testItems, item)
		}
	}
	require.Len(t, testItems, 3, "should have all 3 menu items")

	// Verify ascending price order
	require.Equal(t, "Cheap", testItems[0].MenuItemName)
	require.Equal(t, "Medium", testItems[1].MenuItemName)
	require.Equal(t, "Expensive", testItems[2].MenuItemName)

	// Order by price descending
	orderByPriceDesc := "price_desc"
	items, err = readModel.ListMenuItemsWithRestaurant(ctx, http.ListMenuItemsFilter{
		RestaurantName: &nameFilter,
		OrderBy:        &orderByPriceDesc,
	})
	require.NoError(t, err)

	testItems = nil
	for _, item := range items {
		if item.RestaurantUuid == restaurantUUID {
			testItems = append(testItems, item)
		}
	}
	require.Len(t, testItems, 3)

	// Verify descending price order
	require.Equal(t, "Expensive", testItems[0].MenuItemName)
	require.Equal(t, "Medium", testItems[1].MenuItemName)
	require.Equal(t, "Cheap", testItems[2].MenuItemName)
}

func TestReadModel_ListMenuItemsWithRestaurant_OrderByName(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)

	restaurantRepo := db.NewRestaurantRepository(dbPool)
	readModel := db.NewReadModel(dbPool)

	// Create a restaurant with menu items with various names
	restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
	restaurantName := "Name Test Restaurant " + gofakeit.UUID()
	restaurant := app.OnboardRestaurant{
		Name:        restaurantName,
		Address:     testutils.GenerateRandomAddress(testutils.GenerateRandomCountry()),
		Currency:    shared.MustNewCurrency("USD"),
		Description: gofakeit.LoremIpsumSentence(5),
		MenuItems: []app.MenuItem{
			{
				MenuItemUUID: app.RestaurantMenuItemUUID{common.NewUUIDv7()},
				Name:         "Zebra Cake",
				GrossPrice:   decimal.NewFromFloat(10.00),
				Ordering:     1,
			},
			{
				MenuItemUUID: app.RestaurantMenuItemUUID{common.NewUUIDv7()},
				Name:         "Apple Pie",
				GrossPrice:   decimal.NewFromFloat(10.00),
				Ordering:     2,
			},
			{
				MenuItemUUID: app.RestaurantMenuItemUUID{common.NewUUIDv7()},
				Name:         "Mango Smoothie",
				GrossPrice:   decimal.NewFromFloat(10.00),
				Ordering:     3,
			},
		},
	}
	err := restaurantRepo.UpsertRestaurant(ctx, restaurantUUID, restaurant)
	require.NoError(t, err)

	// Order by name ascending
	orderByNameAsc := "name_asc"
	nameFilter := restaurantName
	items, err := readModel.ListMenuItemsWithRestaurant(ctx, http.ListMenuItemsFilter{
		RestaurantName: &nameFilter,
		OrderBy:        &orderByNameAsc,
	})
	require.NoError(t, err)

	var testItems []http.MenuItemWithRestaurant
	for _, item := range items {
		if item.RestaurantUuid == restaurantUUID {
			testItems = append(testItems, item)
		}
	}
	require.Len(t, testItems, 3, "should have all 3 menu items")

	// Verify ascending name order
	require.Equal(t, "Apple Pie", testItems[0].MenuItemName)
	require.Equal(t, "Mango Smoothie", testItems[1].MenuItemName)
	require.Equal(t, "Zebra Cake", testItems[2].MenuItemName)

	// Order by name descending
	orderByNameDesc := "name_desc"
	items, err = readModel.ListMenuItemsWithRestaurant(ctx, http.ListMenuItemsFilter{
		RestaurantName: &nameFilter,
		OrderBy:        &orderByNameDesc,
	})
	require.NoError(t, err)

	testItems = nil
	for _, item := range items {
		if item.RestaurantUuid == restaurantUUID {
			testItems = append(testItems, item)
		}
	}
	require.Len(t, testItems, 3)

	// Verify descending name order
	require.Equal(t, "Zebra Cake", testItems[0].MenuItemName)
	require.Equal(t, "Mango Smoothie", testItems[1].MenuItemName)
	require.Equal(t, "Apple Pie", testItems[2].MenuItemName)
}
