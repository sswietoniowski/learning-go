package db

import (
	"context"
	"fmt"
	"time"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/orders/adapters/db/dbmodels"
	"eats/backend/orders/app"
)

type OrdersRepo struct {
	db *pgxpool.Pool
}

func NewOrdersRepository(db *pgxpool.Pool) *OrdersRepo {
	if db == nil {
		panic("db connection pool cannot be nil")
	}

	return &OrdersRepo{db: db}
}

func (r *OrdersRepo) CreateQuote(
	ctx context.Context,
	restaurantID app.RestaurantUUID,
	menuItems app.CreateQuoteItems,
	updateFn func(
		ctx context.Context,
		menuItems map[app.RestaurantMenuItemUUID]app.MenuItem,
		restaurantCurrency shared.Currency,
		restaurantAddress shared.Address,
	) (app.Quote, []app.QuoteMenuItem, error),
) (app.Quote, error) {
	var quote app.Quote

	err := common.UpdateInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		menuItemsUUIDs := make([]common.UUID, 0, len(menuItems))
		for _, item := range menuItems {
			menuItemsUUIDs = append(menuItemsUUIDs, item.MenuItemUUID.UUID)
		}

		appMenuItems, err := r.getMenuItems(ctx, queries, restaurantID, menuItemsUUIDs)
		if err != nil {
			return err
		}

		restaurant, err := queries.GetRestaurant(ctx, restaurantID)
		if err != nil {
			return fmt.Errorf("failed to get restaurant currency for restaurant %s: %w", restaurantID, err)
		}

		var items []app.QuoteMenuItem
		quote, items, err = updateFn(ctx, appMenuItems, restaurant.Currency, restaurant.Address)
		if err != nil {
			return fmt.Errorf("failed to create quote using updateFn: %w", err)
		}

		err = queries.AddQuote(ctx, dbmodels.AddQuoteParams{
			quote.QuoteUUID,
			quote.CustomerUUID,
			quote.RestaurantUUID,
			quote.DeliveryAddress,
			quote.ItemsSubtotalGross,
			quote.ServiceFeeGross,
			quote.DeliveryFeeGross,
			quote.TotalAmountGross,
			quote.TotalTax,
			time.Now(),
			quote.Currency,
		})
		if err != nil {
			return fmt.Errorf("failed to add quote %s: %w", quote.QuoteUUID, err)
		}

		quoteItems := dbQuoteItemsFromApp(items, quote)

		if _, err := queries.AddQuoteItems(ctx, quoteItems); err != nil {
			return fmt.Errorf("failed to add quote items for quote %s: %w", quote.QuoteUUID, err)
		}

		return nil
	})
	if err != nil {
		return app.Quote{}, err
	}

	return quote, nil
}

func dbQuoteItemsFromApp(menuItems []app.QuoteMenuItem, quote app.Quote) []dbmodels.AddQuoteItemsParams {
	quoteItems := make([]dbmodels.AddQuoteItemsParams, 0, len(menuItems))
	for _, position := range menuItems {
		quoteItems = append(quoteItems, dbmodels.AddQuoteItemsParams{
			common.NewUUIDv7(),
			quote.QuoteUUID,
			position.MenuItemUUID,
			position.GrossPrice,
			int32(position.Quantity),
		})
	}
	return quoteItems
}

func (r *OrdersRepo) getMenuItems(ctx context.Context, queries *dbmodels.Queries, restaurantUUID app.RestaurantUUID, menuItemsUUIDs []common.UUID) (map[app.RestaurantMenuItemUUID]app.MenuItem, error) {
	dbMenuItems, err := queries.GetMenuItemsByUUIDs(ctx, dbmodels.GetMenuItemsByUUIDsParams{
		RestaurantUuid: restaurantUUID,
		Column2:        menuItemsUUIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get menu positions: %w", err)
	}

	return appMenuItemsFromDbMenuItems(dbMenuItems), nil
}

func appMenuItemsFromDbMenuItems(dbMenuItems []dbmodels.OrdersRestaurantMenuItem) map[app.RestaurantMenuItemUUID]app.MenuItem {
	appMenuItems := make(map[app.RestaurantMenuItemUUID]app.MenuItem, len(dbMenuItems))

	for _, dbItemPosition := range dbMenuItems {
		appMenuItems[dbItemPosition.RestaurantMenuItemUuid] = app.MenuItem{
			dbItemPosition.RestaurantMenuItemUuid,
			dbItemPosition.Name,
			dbItemPosition.Ordering,
			dbItemPosition.GrossPrice,
			dbItemPosition.IsArchived,
		}
	}

	return appMenuItems
}
