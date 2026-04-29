package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eats/backend/common"
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

func (r *OrdersRepo) GetRestaurant(
	ctx context.Context,
	restaurantID app.RestaurantUUID,
) (app.Restaurant, error) {
	queries := dbmodels.New(r.db)
	dbRestaurant, err := queries.GetRestaurant(ctx, restaurantID)
	if err != nil {
		return app.Restaurant{}, fmt.Errorf("failed to get restaurant %s: %w", restaurantID, err)
	}
	return appRestaurantFromDB(dbRestaurant), nil
}

func appRestaurantFromDB(dbRestaurant dbmodels.OrdersRestaurant) app.Restaurant {
	return app.Restaurant{
		RestaurantUUID: dbRestaurant.RestaurantUuid,
		Name:           dbRestaurant.Name,
		Description:    dbRestaurant.Description,
		Currency:       dbRestaurant.Currency,
		Address:        dbRestaurant.Address,
	}
}

func (r *OrdersRepo) CreateQuote(
	ctx context.Context,
	restaurantID app.RestaurantUUID,
	menuItems app.CreateQuoteItems,
	updateFn func(
		ctx context.Context,
		menuItems map[app.RestaurantMenuItemUUID]app.MenuItem,
		r app.Restaurant,
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
		quote, items, err = updateFn(ctx, appMenuItems, appRestaurantFromDB(restaurant))
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

func (r *OrdersRepo) GetQuote(ctx context.Context, quoteUUID app.QuoteUUID) (app.Quote, error) {
	queries := dbmodels.New(r.db)
	dbQuote, err := queries.GetQuote(ctx, quoteUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return app.Quote{}, common.NewNotFoundError("quote-not-found", "quote not found")
		}
		return app.Quote{}, fmt.Errorf("failed to get quote %s: %w", quoteUUID, err)
	}
	return app.Quote{
		QuoteUUID:          dbQuote.QuoteUuid,
		CustomerUUID:       dbQuote.CustomerUuid,
		RestaurantUUID:     dbQuote.RestaurantUuid,
		DeliveryAddress:    dbQuote.DeliveryAddress,
		ItemsSubtotalGross: dbQuote.ItemsSubtotalGross,
		ServiceFeeGross:    dbQuote.ServiceFeeGross,
		DeliveryFeeGross:   dbQuote.DeliveryFeeGross,
		TotalAmountGross:   dbQuote.TotalAmountGross,
		TotalTax:           dbQuote.TotalTax,
		Currency:           dbQuote.Currency,
		CreatedAt:          dbQuote.CreatedAt,
	}, nil
}

func (r *OrdersRepo) GetMenuItemsForQuote(ctx context.Context, quoteUUID app.QuoteUUID, restaurantUUID app.RestaurantUUID) (map[app.RestaurantMenuItemUUID]app.MenuItem, error) {
	queries := dbmodels.New(r.db)

	quoteItems, err := queries.GetQuoteItems(ctx, quoteUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote items for %s: %w", quoteUUID, err)
	}

	menuItemUUIDs := make([]common.UUID, 0, len(quoteItems))
	for _, item := range quoteItems {
		menuItemUUIDs = append(menuItemUUIDs, item.MenuItemUuid.UUID)
	}

	return r.getMenuItems(ctx, queries, restaurantUUID, menuItemUUIDs)
}

func dbOrderToAppOrder(dbOrder dbmodels.OrdersOrder) app.Order {
	return app.Order{
		OrderUUID:             dbOrder.OrderUuid,
		QuoteUUID:             dbOrder.QuoteUuid,
		CustomerUUID:          dbOrder.CustomerUuid,
		RestaurantUUID:        dbOrder.RestaurantUuid,
		RestaurantName:        "",
		CourierUUID:           dbOrder.CourierUuid,
		DeliveryAddress:       dbOrder.DeliveryAddress,
		OrderedAt:             dbOrder.OrderedAt,
		RestaurantConfirmedAt: dbOrder.RestaurantConfirmedAt,
		CourierAcceptedAt:     dbOrder.CourierAcceptedAt,
		RestaurantPreparedAt:  dbOrder.RestaurantPreparedAt,
		PickedUpAt:            dbOrder.PickedUpAt,
		DeliveredAt:           dbOrder.DeliveredAt,
		ItemsSubtotalGross:    dbOrder.ItemsSubtotalGross,
		ServiceFeeGross:       dbOrder.ServiceFeeGross,
		DeliveryFeeGross:      dbOrder.DeliveryFeeGross,
		TotalAmountGross:      dbOrder.TotalAmountGross,
		TotalTax:              dbOrder.TotalTax,
		Currency:              dbOrder.Currency,
	}
}

func (r *OrdersRepo) QuoteWithMenuItems(ctx context.Context, quoteUUID app.QuoteUUID) (app.Quote, map[app.RestaurantMenuItemUUID]app.MenuItem, error) {
	quote, err := r.GetQuote(ctx, quoteUUID)
	if err != nil {
		return app.Quote{}, nil, err
	}

	menuItems, err := r.GetMenuItemsForQuote(ctx, quoteUUID, quote.RestaurantUUID)
	if err != nil {
		return app.Quote{}, nil, err
	}

	return quote, menuItems, nil
}

func (r *OrdersRepo) OrderByID(ctx context.Context, orderUUID app.OrderUUID) (app.Order, error) {
	queries := dbmodels.New(r.db)
	dbOrder, err := queries.GetOrder(ctx, orderUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return app.Order{}, common.NewNotFoundError("order_not_found", "order not found")
		}
		return app.Order{}, fmt.Errorf("failed to get order %s: %w", orderUUID, err)
	}
	return dbOrderToAppOrder(dbOrder), nil
}

func (r *OrdersRepo) UpdateOrder(
	ctx context.Context,
	orderUUID app.OrderUUID,
	updateFn func(ctx context.Context, order app.Order) (app.Order, error),
) error {
	return common.UpdateInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		dbOrder, err := queries.GetOrder(ctx, orderUUID)
		if err != nil {
			return fmt.Errorf("failed to get order: %w", err)
		}

		updatedOrder, err := updateFn(ctx, dbOrderToAppOrder(dbOrder))
		if err != nil {
			return fmt.Errorf("failed to update order: %w", err)
		}

		return queries.UpdateOrder(ctx, dbmodels.UpdateOrderParams{
			orderUUID,
			updatedOrder.CourierUUID,
			&updatedOrder.OrderedAt,
			updatedOrder.RestaurantConfirmedAt,
			updatedOrder.CourierAcceptedAt,
			updatedOrder.RestaurantPreparedAt,
			updatedOrder.PickedUpAt,
			updatedOrder.DeliveredAt,
		})
	})
}

func (r *OrdersRepo) SaveOrder(ctx context.Context, order app.Order) error {
	return r.PlaceOrder(ctx, order)
}

func (r *OrdersRepo) PlaceOrder(ctx context.Context, order app.Order) error {
	return common.UpdateInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		err := queries.PlaceOrder(ctx, dbmodels.PlaceOrderParams{
			order.OrderUUID,
			order.QuoteUUID,
			order.CustomerUUID,
			order.RestaurantUUID,
			order.DeliveryAddress,
			order.OrderedAt,
				order.ItemsSubtotalGross,
			order.ServiceFeeGross,
			order.DeliveryFeeGross,
			order.TotalAmountGross,
			order.TotalTax,
			order.Currency,
		})
		if err != nil {
			return fmt.Errorf("failed to place order %s: %w", order.OrderUUID, err)
		}

		return nil
	})
}
