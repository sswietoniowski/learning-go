package db

import (
	"context"
	"errors"
	"fmt"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eats/backend/common"
	"eats/backend/common/log"
	"eats/backend/orders/adapters/db/dbmodels"
	"eats/backend/orders/app"
)

type RestaurantRepository struct {
	db *pgxpool.Pool
}

func NewRestaurantRepository(db *pgxpool.Pool) *RestaurantRepository {
	if db == nil {
		panic("db connection pool cannot be nil")
	}

	return &RestaurantRepository{
		db: db,
	}
}

func (r *RestaurantRepository) UpsertRestaurant(ctx context.Context, restaurantUUID app.RestaurantUUID, restaurant app.OnboardRestaurant) error {
	return common.UpdateInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		log.FromContext(ctx).With("restaurant_uuid", restaurantUUID).Info("Upserting restaurant")

		currentMenuItems, err := queries.GetRestaurantMenu(ctx, restaurantUUID)
		if err != nil {
			return fmt.Errorf("get current restaurant menu failed: %w", err)
		}

		currentMenuItemsUUIDs := make([]app.RestaurantMenuItemUUID, len(currentMenuItems))
		for i, menuItem := range currentMenuItems {
			currentMenuItemsUUIDs[i] = menuItem.OrdersRestaurantMenuItem.RestaurantMenuItemUuid
		}

		dbRestaurant, err := queries.UpsertRestaurant(ctx, dbmodels.UpsertRestaurantParams{
			RestaurantUuid: restaurantUUID,
			Name:           restaurant.Name,
			Description:    restaurant.Description,
			Address:        restaurant.Address,
			Currency:       restaurant.Currency,
		})
		if err != nil {
			return fmt.Errorf("upsert restaurant failed: %w", err)
		}

		// Currency is immutable after creation - the upsert doesn't update it.
		// Check here catches attempts to change it and returns a clear error.
		if dbRestaurant.Currency != restaurant.Currency {
			return common.NewInvalidInputError("cannot-change-currency", "cannot change restaurant currency once set")
		}

		for _, item := range restaurant.MenuItems {
			err = queries.UpsertRestaurantMenuItem(ctx, dbmodels.UpsertRestaurantMenuItemParams{
				RestaurantMenuItemUuid: item.MenuItemUUID,
				RestaurantUuid:         restaurantUUID,
				Name:                   item.Name,
				GrossPrice:             item.GrossPrice,
				Ordering:               item.Ordering,
				IsArchived:             false,
			})
			if err != nil {
				return fmt.Errorf("upsert restaurant menu position failed for menu position %s: %w", item.MenuItemUUID, err)
			}
		}

		menuItemsToArchive := make([]common.UUID, 0)
		for _, currentUUID := range currentMenuItemsUUIDs {
			found := false
			for _, newItem := range restaurant.MenuItems {
				if currentUUID == newItem.MenuItemUUID {
					found = true
					break
				}
			}
			if !found {
				menuItemsToArchive = append(menuItemsToArchive, currentUUID.UUID)
			}
		}

		if len(menuItemsToArchive) > 0 {
			if err = queries.ArchiveMenuItems(ctx, menuItemsToArchive); err != nil {
				return fmt.Errorf("archive old restaurant menu positions failed: %w", err)
			}
		}

		return nil
	})
}

func (r *RestaurantRepository) GetRestaurantMenu(ctx context.Context, restaurantUUID app.RestaurantUUID) (app.RestaurantMenu, error) {
	queries := dbmodels.New(r.db)

	dbItems, err := queries.GetRestaurantMenu(ctx, restaurantUUID)
	if err != nil {
		return app.RestaurantMenu{}, fmt.Errorf("get restaurant menu failed: %w", err)
	}

	log.FromContext(ctx).With("restaurant_uuid", restaurantUUID, "count", len(dbItems)).Info("Fetched menu items")

	items := make([]app.MenuItem, len(dbItems))
	for i, dbItem := range dbItems {
		items[i] = app.MenuItem{
			MenuItemUUID: dbItem.OrdersRestaurantMenuItem.RestaurantMenuItemUuid,
			Name:         dbItem.OrdersRestaurantMenuItem.Name,
			GrossPrice:   dbItem.OrdersRestaurantMenuItem.GrossPrice,
			Ordering:     dbItem.OrdersRestaurantMenuItem.Ordering,
		}
	}

	restaurant, err := queries.GetRestaurant(ctx, restaurantUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return app.RestaurantMenu{}, common.NewNotFoundError("restaurant-not-found", "restaurant not found")
		}
		return app.RestaurantMenu{}, fmt.Errorf("get restaurant %s failed: %w", restaurantUUID, err)
	}

	return app.RestaurantMenu{
		RestaurantName: restaurant.Name,
		Address:        restaurant.Address,
		Description:    restaurant.Description,
		Currency:       restaurant.Currency,
		Positions:      items,
	}, nil
}
