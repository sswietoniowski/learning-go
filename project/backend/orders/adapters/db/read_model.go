package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"eats/backend/orders/adapters/db/dbmodels"
	"eats/backend/orders/api/http"
	"eats/backend/orders/app"
)

// ReadModel provides read-optimized queries that return HTTP response types directly.
type ReadModel struct {
	db *pgxpool.Pool
}

func NewReadModel(db *pgxpool.Pool) *ReadModel {
	if db == nil {
		panic("db connection pool cannot be nil")
	}
	return &ReadModel{db: db}
}

func (r ReadModel) ListMenuItemsWithRestaurant(ctx context.Context, filter http.ListMenuItemsFilter) ([]http.MenuItemWithRestaurant, error) {
	queries := dbmodels.New(r.db)

	rows, err := queries.ListMenuItemsWithRestaurant(ctx, dbmodels.ListMenuItemsWithRestaurantParams{
		SearchTerm:           filter.Search,
		RestaurantNameFilter: filter.RestaurantName,
		OrderBy:              filter.OrderBy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query menu items: %w", err)
	}

	// Map directly to HTTP response types - no domain objects needed for reads
	items := make([]http.MenuItemWithRestaurant, 0, len(rows))
	for _, row := range rows {
		items = append(items, http.MenuItemWithRestaurant{
			MenuItemUuid:   row.MenuItemUuid,
			MenuItemName:   row.MenuItemName,
			GrossPrice:     row.GrossPrice,
			Currency:       row.Currency,
			RestaurantUuid: row.RestaurantUuid,
			RestaurantName: row.RestaurantName,
		})
	}

	return items, nil
}

func (r ReadModel) ListCustomerOrders(ctx context.Context, customerUUID app.CustomerUUID) ([]http.CustomerOrder, error) {
	queries := dbmodels.New(r.db)

	dbOrders, err := queries.GetCustomerOrders(ctx, customerUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer orders: %w", err)
	}

	orders := make([]http.CustomerOrder, 0, len(dbOrders))
	for _, dbOrder := range dbOrders {
		orders = append(orders, http.CustomerOrder{
			dbOrder.CourierAcceptedAt,
			dbOrder.CourierUuid,
			dbOrder.Currency,
			dbOrder.DeliveredAt,
			http.Address{
				dbOrder.DeliveryAddress.City,
				dbOrder.DeliveryAddress.CountryCode,
				dbOrder.DeliveryAddress.Line1,
				dbOrder.DeliveryAddress.Line2,
				dbOrder.DeliveryAddress.PostalCode,
			},
			dbOrder.DeliveryFeeGross,
			dbOrder.ItemsSubtotalGross,
			dbOrder.OrderUuid,
			dbOrder.OrderedAt,
			dbOrder.RestaurantConfirmedAt,
			dbOrder.RestaurantConfirmedAt,
			dbOrder.RestaurantName,
			dbOrder.RestaurantPreparedAt,
			dbOrder.RestaurantUuid,
			dbOrder.ServiceFeeGross,
			dbOrder.TotalAmountGross,
			dbOrder.TotalTax,
		})
	}

	return orders, nil
}

func (r ReadModel) ListRestaurantOrders(ctx context.Context, restaurantUUID app.RestaurantUUID) ([]http.RestaurantOrder, error) {
	queries := dbmodels.New(r.db)

	dbOrders, err := queries.GetRestaurantOrders(ctx, restaurantUUID)
	if err != nil {
		return nil, err
	}

	orders := make([]http.RestaurantOrder, 0, len(dbOrders))
	for _, dbOrder := range dbOrders {
		orders = append(orders, http.RestaurantOrder{
			dbOrder.CourierAcceptedAt,
			dbOrder.CourierUuid,
			dbOrder.CustomerUuid,
			dbOrder.DeliveredAt,
			dbOrder.ItemsSubtotalGross,
			dbOrder.OrderUuid,
			dbOrder.OrderedAt,
			dbOrder.PickedUpAt,
			dbOrder.RestaurantConfirmedAt,
			dbOrder.RestaurantPreparedAt,
		})
	}

	return orders, nil
}

func (r ReadModel) ListAssignedCourierOrders(ctx context.Context, courierUUID app.CourierUUID) ([]http.CourierOrder, error) {
	queries := dbmodels.New(r.db)

	// Get orders assigned to this courier
	dbOrders, err := queries.GetCourierOrders(ctx, &courierUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assigned orders: %w", err)
	}

	orders := make([]http.CourierOrder, 0, len(dbOrders))
	for _, dbOrder := range dbOrders {
		orders = append(orders, http.CourierOrder{
			dbOrder.CourierAcceptedAt,
			dbOrder.CourierUuid,
			dbOrder.CustomerUuid,
			dbOrder.DeliveredAt,
			http.Address{
				dbOrder.DeliveryAddress.City,
				dbOrder.DeliveryAddress.CountryCode,
				dbOrder.DeliveryAddress.Line1,
				dbOrder.DeliveryAddress.Line2,
				dbOrder.DeliveryAddress.PostalCode,
			},
			dbOrder.ItemsSubtotalGross,
			dbOrder.OrderUuid,
			dbOrder.OrderedAt,
			dbOrder.PickedUpAt,
			dbOrder.RestaurantConfirmedAt,
			dbOrder.RestaurantName,
			dbOrder.RestaurantPreparedAt,
			dbOrder.RestaurantUuid,
		})
	}

	return orders, nil
}

func (r ReadModel) ListAvailableOrdersForCourier(ctx context.Context, courierUUID app.CourierUUID) ([]http.CourierOrder, error) {
	queries := dbmodels.New(r.db)

	dbOrders, err := queries.GetAvailableOrdersForCourier(ctx, courierUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get available orders: %w", err)
	}

	orders := make([]http.CourierOrder, 0, len(dbOrders))
	for _, dbOrder := range dbOrders {
		orders = append(orders, http.CourierOrder{
			dbOrder.CourierAcceptedAt,
			dbOrder.CourierUuid,
			dbOrder.CustomerUuid,
			dbOrder.DeliveredAt,
			http.Address{
				dbOrder.DeliveryAddress.City,
				dbOrder.DeliveryAddress.CountryCode,
				dbOrder.DeliveryAddress.Line1,
				dbOrder.DeliveryAddress.Line2,
				dbOrder.DeliveryAddress.PostalCode,
			},
			dbOrder.ItemsSubtotalGross,
			dbOrder.OrderUuid,
			dbOrder.OrderedAt,
			dbOrder.PickedUpAt,
			dbOrder.RestaurantConfirmedAt,
			dbOrder.RestaurantName,
			dbOrder.RestaurantPreparedAt,
			dbOrder.RestaurantUuid,
		})
	}

	return orders, nil
}
