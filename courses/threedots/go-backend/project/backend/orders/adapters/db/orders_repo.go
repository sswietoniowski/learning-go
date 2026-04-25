package db

import (
	"context"
	"fmt"

	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/orders/adapters/db/dbmodels"
	"eats/backend/orders/app"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrdersRepository struct {
	db *pgxpool.Pool
}

func NewOrdersRepository(db *pgxpool.Pool) *OrdersRepository {
	if db == nil {
		panic("db connection pool cannot be nil")
	}

	return &OrdersRepository{
		db: db,
	}
}

func (r *OrdersRepository) CreateQuote(
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

		dbRestaurant, err := queries.GetRestaurant(ctx, restaurantID)
		if err != nil {
			return fmt.Errorf("get restaurant failed: %w", err)
		}

		dbMenuItems, err := queries.GetRestaurantMenu(ctx, restaurantID)
		if err != nil {
			return fmt.Errorf("get restaurant menu failed: %w", err)
		}

		menuItemsMap := make(map[app.RestaurantMenuItemUUID]app.MenuItem, len(dbMenuItems))
		for _, dbMenuItem := range dbMenuItems {
			menuItemsMap[dbMenuItem.OrdersRestaurantMenuItem.RestaurantMenuItemUuid] = app.MenuItem{
				Name:       dbMenuItem.OrdersRestaurantMenuItem.Name,
				Ordering:   dbMenuItem.OrdersRestaurantMenuItem.Ordering,
				GrossPrice: dbMenuItem.OrdersRestaurantMenuItem.GrossPrice,
			}
		}

		quoteMenuItems := make([]app.QuoteMenuItem, 0, len(menuItems))
		quote, quoteMenuItems, err = updateFn(ctx, menuItemsMap, dbRestaurant.Currency, dbRestaurant.Address)
		if err != nil {
			return fmt.Errorf("update quote failed: %w", err)
		}

		err = queries.AddQuote(ctx, dbmodels.AddQuoteParams{
			QuoteUuid:          quote.QuoteUUID,
			CustomerUuid:       quote.CustomerUUID,
			RestaurantUuid:     restaurantID,
			DeliveryAddress:    quote.DeliveryAddress,
			ItemsSubtotalGross: quote.ItemsSubtotalGross,
			ServiceFeeGross:    quote.ServiceFeeGross,
			DeliveryFeeGross:   quote.DeliveryFeeGross,
			TotalAmountGross:   quote.TotalAmountGross,
			TotalTax:           quote.TotalTax,
			CreatedAt:          quote.CreatedAt,
			Currency:           dbRestaurant.Currency,
		})
		if err != nil {
			return fmt.Errorf("add quote failed: %w", err)
		}

		addQuoteItemsParams := make([]dbmodels.AddQuoteItemsParams, len(quoteMenuItems))
		for i, quoteMenuItem := range quoteMenuItems {
			addQuoteItemsParams[i] = dbmodels.AddQuoteItemsParams{
				QuoteItemUuid: common.NewUUIDv7(),
				QuoteUuid:     quote.QuoteUUID,
				MenuItemUuid:  quoteMenuItem.MenuItemUUID,
				GrossPrice:    quoteMenuItem.GrossPrice,
				Quantity:      int32(quoteMenuItem.Quantity),
			}
		}

		_, err = queries.AddQuoteItems(ctx, addQuoteItemsParams)
		if err != nil {
			return fmt.Errorf("add quote menu items failed: %w", err)
		}

		return nil
	})

	return quote, err
}
