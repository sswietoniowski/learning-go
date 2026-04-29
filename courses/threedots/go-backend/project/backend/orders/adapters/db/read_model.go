package db

import (
	"context"
	"errors"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eats/backend/common"
	"eats/backend/common/shared"
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

func (r ReadModel) RestaurantName(ctx context.Context, restaurantUUID app.RestaurantUUID) (string, error) {
	queries := dbmodels.New(r.db)
	restaurant, err := queries.GetRestaurant(ctx, restaurantUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", common.NewNotFoundError("restaurant-not-found", "restaurant not found")
		}
		return "", err
	}
	return restaurant.Name, nil
}

func (r ReadModel) ListRestaurants(ctx context.Context) ([]app.Restaurant, error) {
	queries := dbmodels.New(r.db)
	rows, err := queries.ListRestaurants(ctx)
	if err != nil {
		return nil, err
	}

	restaurants := make([]app.Restaurant, 0, len(rows))
	for _, row := range rows {
		restaurants = append(restaurants, app.Restaurant{
			RestaurantUUID: row.RestaurantUuid,
			Name:           row.Name,
			Description:    row.Description,
			Address:        row.Address,
			Currency:       row.Currency,
		})
	}
	return restaurants, nil
}

func (r ReadModel) GetRestaurantMenu(ctx context.Context, restaurantUUID app.RestaurantUUID) (app.RestaurantMenu, error) {
	queries := dbmodels.New(r.db)

	restaurant, err := queries.GetRestaurant(ctx, restaurantUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return app.RestaurantMenu{}, common.NewNotFoundError("restaurant-not-found", "restaurant not found")
		}
		return app.RestaurantMenu{}, err
	}

	dbItems, err := queries.GetRestaurantMenu(ctx, restaurantUUID)
	if err != nil {
		return app.RestaurantMenu{}, err
	}

	items := make([]app.MenuItem, len(dbItems))
	for i, dbItem := range dbItems {
		items[i] = app.MenuItem{
			MenuItemUUID: dbItem.OrdersRestaurantMenuItem.RestaurantMenuItemUuid,
			Name:         dbItem.OrdersRestaurantMenuItem.Name,
			GrossPrice:   dbItem.OrdersRestaurantMenuItem.GrossPrice,
			Ordering:     dbItem.OrdersRestaurantMenuItem.Ordering,
		}
	}

	return app.RestaurantMenu{
		RestaurantName: restaurant.Name,
		Address:        restaurant.Address,
		Description:    restaurant.Description,
		Currency:       restaurant.Currency,
		Positions:      items,
	}, nil
}

func (r ReadModel) ListCustomerOrders(ctx context.Context, customerUUID app.CustomerUUID) ([]http.CustomerOrder, error) {
	queries := dbmodels.New(r.db)
	rows, err := queries.ListCustomerOrders(ctx, customerUUID)
	if err != nil {
		return nil, err
	}

	orders := make([]http.CustomerOrder, 0, len(rows))
	for _, row := range rows {
		orders = append(orders, http.CustomerOrder{
			OrderUuid:             row.OrderUuid,
			RestaurantUuid:        row.RestaurantUuid,
			RestaurantName:        row.RestaurantName,
			CourierUuid:           row.CourierUuid,
			DeliveryAddress:       sharedAddressToHTTP(row.DeliveryAddress),
			OrderedAt:             row.OrderedAt,
			RestaurantConfirmedAt: row.RestaurantConfirmedAt,
			CourierAcceptedAt:     row.CourierAcceptedAt,
			RestaurantPreparedAt:  row.RestaurantPreparedAt,
			PickedUpAt:            row.PickedUpAt,
			DeliveredAt:           row.DeliveredAt,
			ItemsSubtotalGross:    row.ItemsSubtotalGross,
			ServiceFeeGross:       row.ServiceFeeGross,
			DeliveryFeeGross:      row.DeliveryFeeGross,
			TotalGross:            row.TotalAmountGross,
			TotalTax:              row.TotalTax,
			Currency:              row.Currency,
		})
	}
	return orders, nil
}

func (r ReadModel) ListRestaurantOrders(ctx context.Context, restaurantUUID app.RestaurantUUID) ([]http.RestaurantOrder, error) {
	queries := dbmodels.New(r.db)
	rows, err := queries.ListRestaurantOrders(ctx, restaurantUUID)
	if err != nil {
		return nil, err
	}

	orders := make([]http.RestaurantOrder, 0, len(rows))
	for _, row := range rows {
		orders = append(orders, http.RestaurantOrder{
			OrderUuid:             row.OrderUuid,
			CustomerUuid:          row.CustomerUuid,
			CourierUuid:           row.CourierUuid,
			OrderedAt:             row.OrderedAt,
			RestaurantConfirmedAt: row.RestaurantConfirmedAt,
			CourierAcceptedAt:     row.CourierAcceptedAt,
			RestaurantPreparedAt:  row.RestaurantPreparedAt,
			PickedUpAt:            row.PickedUpAt,
			DeliveredAt:           row.DeliveredAt,
			ItemsSubtotalGross:    row.ItemsSubtotalGross,
		})
	}
	return orders, nil
}

func (r ReadModel) ListAssignedCourierOrders(ctx context.Context, courierUUID app.CourierUUID) ([]http.CourierOrder, error) {
	queries := dbmodels.New(r.db)
	rows, err := queries.ListAssignedCourierOrders(ctx, &courierUUID)
	if err != nil {
		return nil, err
	}

	orders := make([]http.CourierOrder, 0, len(rows))
	for _, row := range rows {
		orders = append(orders, http.CourierOrder{
			OrderUuid:             row.OrderUuid,
			CustomerUuid:          row.CustomerUuid,
			CourierUuid:           row.CourierUuid,
			RestaurantUuid:        row.RestaurantUuid,
			RestaurantName:        row.RestaurantName,
			DeliveryAddress:       sharedAddressToHTTP(row.DeliveryAddress),
			OrderedAt:             row.OrderedAt,
			RestaurantConfirmedAt: row.RestaurantConfirmedAt,
			AcceptedByCourierAt:   row.CourierAcceptedAt,
			RestaurantPreparedAt:  row.RestaurantPreparedAt,
			PickedUpAt:            row.PickedUpAt,
			DeliveredAt:           row.DeliveredAt,
			ItemsSubtotalGross:    row.ItemsSubtotalGross,
		})
	}
	return orders, nil
}

func (r ReadModel) ListAvailableOrdersForCourier(ctx context.Context, courierUUID app.CourierUUID) ([]http.CourierOrder, error) {
	queries := dbmodels.New(r.db)
	rows, err := queries.ListAvailableOrdersForCourier(ctx, courierUUID)
	if err != nil {
		return nil, err
	}

	orders := make([]http.CourierOrder, 0, len(rows))
	for _, row := range rows {
		orders = append(orders, http.CourierOrder{
			OrderUuid:             row.OrderUuid,
			CustomerUuid:          row.CustomerUuid,
			CourierUuid:           row.CourierUuid,
			RestaurantUuid:        row.RestaurantUuid,
			RestaurantName:        row.RestaurantName,
			DeliveryAddress:       sharedAddressToHTTP(row.DeliveryAddress),
			OrderedAt:             row.OrderedAt,
			RestaurantConfirmedAt: row.RestaurantConfirmedAt,
			AcceptedByCourierAt:   row.CourierAcceptedAt,
			RestaurantPreparedAt:  row.RestaurantPreparedAt,
			PickedUpAt:            row.PickedUpAt,
			DeliveredAt:           row.DeliveredAt,
			ItemsSubtotalGross:    row.ItemsSubtotalGross,
		})
	}
	return orders, nil
}

func sharedAddressToHTTP(addr shared.Address) http.Address {
	return http.Address{
		Line1:       addr.Line1,
		Line2:       addr.Line2,
		PostalCode:  addr.PostalCode,
		City:        addr.City,
		CountryCode: addr.CountryCode,
	}
}

func (r ReadModel) ListMenuItemsWithRestaurant(ctx context.Context, filter http.ListMenuItemsFilter) ([]http.MenuItemWithRestaurant, error) {
	queries := dbmodels.New(r.db)

	rows, err := queries.ListMenuItems(ctx, dbmodels.ListMenuItemsParams{
		SearchTerm:           filter.Search,
		RestaurantNameFilter: filter.RestaurantName,
		OrderBy:              filter.OrderBy,
	})
	if err != nil {
		return nil, err
	}

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
