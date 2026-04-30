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

func (r *OrdersRepo) OrderByID(ctx context.Context, orderUUID app.OrderUUID) (app.Order, error) {
	queries := dbmodels.New(r.db)

	dbOrder, err := queries.GetOrder(ctx, orderUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return app.Order{}, common.NewNotFoundError("order_not_found", "Order not found")
		}
		return app.Order{}, fmt.Errorf("error getting order: %w", err)
	}

	return dbOrderToAppOrder(dbOrder), nil
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

func (r *OrdersRepo) CreateQuote(
	ctx context.Context,
	restaurantID app.RestaurantUUID,
	menuItems app.CreateQuoteItems,
	updateFn func(
		ctx context.Context,
		menuItems map[app.RestaurantMenuItemUUID]app.MenuItem,
		restaurant app.Restaurant,
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

		dbRestaurant, err := queries.GetRestaurant(ctx, restaurantID)
		if err != nil {
			return fmt.Errorf("failed to get restaurant currency for restaurant %s: %w", restaurantID, err)
		}

		var items []app.QuoteMenuItem
		quote, items, err = updateFn(ctx, appMenuItems, appRestaurantFromDB(dbRestaurant))
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

func appRestaurantFromDB(r dbmodels.OrdersRestaurant) app.Restaurant {
	return app.Restaurant{
		RestaurantUUID: r.RestaurantUuid,
		Name:           r.Name,
		Description:    r.Description,
		Address:        r.Address,
		Currency:       r.Currency,
	}
}

func (r *OrdersRepo) QuoteWithMenuItems(ctx context.Context, quoteUUID app.QuoteUUID) (app.Quote, map[app.RestaurantMenuItemUUID]app.MenuItem, error) {
	queries := dbmodels.New(r.db)

	dbQuote, err := queries.GetQuote(ctx, quoteUUID)
	if err != nil {
		return app.Quote{}, nil, fmt.Errorf("failed to get quote %s: %w", quoteUUID, err)
	}

	dbPositions, err := queries.GetMenuItemsForQuote(ctx, quoteUUID)
	if err != nil {
		return app.Quote{}, nil, fmt.Errorf("failed to get menu positions for quote %s: %w", quoteUUID, err)
	}

	return appQuoteFromDbQuote(dbQuote), appMenuItemsFromDbMenuItems(dbPositions), nil
}

func (r *OrdersRepo) SaveOrder(ctx context.Context, order app.Order) error {
	return common.UpdateInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		err := queries.AddOrder(ctx, dbOrderFromAppOrder(order))
		if err != nil {
			return fmt.Errorf("failed to add order %s: %w", order.OrderUUID, err)
		}

		return nil
	})
}

func dbOrderFromAppOrder(order app.Order) dbmodels.AddOrderParams {
	return dbmodels.AddOrderParams{
		order.OrderUUID,
		order.QuoteUUID,
		order.CustomerUUID,
		order.RestaurantUUID,
		order.DeliveryAddress,
		order.ItemsSubtotal,
		order.ServiceFeeGross,
		order.DeliveryFeeGross,
		order.TotalAmountGross,
		order.TotalTax,
		order.OrderedAt,
		order.Currency,
	}
}

func appQuoteFromDbQuote(dbQuote dbmodels.OrdersQuote) app.Quote {
	return app.Quote{
		dbQuote.QuoteUuid,
		dbQuote.CustomerUuid,
		dbQuote.RestaurantUuid,
		dbQuote.DeliveryAddress,
		dbQuote.ItemsSubtotalGross,
		dbQuote.ServiceFeeGross,
		dbQuote.DeliveryFeeGross,
		dbQuote.TotalAmountGross,
		dbQuote.TotalTax,
		dbQuote.Currency,
		dbQuote.CreatedAt,
	}
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
			return fmt.Errorf("failed to get order %s: %w", orderUUID, err)
		}

		appOrder := dbOrderToAppOrder(dbOrder)

		updatedOrder, err := updateFn(ctx, appOrder)
		if err != nil {
			return fmt.Errorf("failed to update order %s using updateFn: %w", orderUUID, err)
		}

		err = queries.UpdateOrder(ctx, dbmodels.UpdateOrderParams{
			orderUUID,
			updatedOrder.CourierUUID,
			&updatedOrder.OrderedAt,
			updatedOrder.RestaurantConfirmedAt,
			updatedOrder.CourierAcceptedAt,
			updatedOrder.RestaurantPreparedAt,
			updatedOrder.PickedUpAt,
			updatedOrder.DeliveredAt,
		})
		if err != nil {
			return fmt.Errorf("failed to update order %s: %w", orderUUID, err)
		}

		return nil
	})
}

func dbOrderToAppOrder(dbOrder dbmodels.OrdersOrder) app.Order {
	return app.Order{
		dbOrder.OrderUuid,
		dbOrder.QuoteUuid,
		dbOrder.CustomerUuid,
		dbOrder.RestaurantUuid,
		dbOrder.CourierUuid,
		dbOrder.DeliveryAddress,
		dbOrder.OrderedAt,
		dbOrder.RestaurantConfirmedAt,
		dbOrder.CourierAcceptedAt,
		dbOrder.RestaurantPreparedAt,
		dbOrder.PickedUpAt,
		dbOrder.DeliveredAt,
		dbOrder.ItemsSubtotalGross,
		dbOrder.ServiceFeeGross,
		dbOrder.DeliveryFeeGross,
		dbOrder.TotalAmountGross,
		dbOrder.TotalTax,
		dbOrder.Currency,
	}
}
