package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"eats/backend/orders/api/http"
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

// TODO: Implement this method using sqlc to query menu items joined with restaurants.
func (r ReadModel) ListMenuItemsWithRestaurant(ctx context.Context) ([]http.MenuItemWithRestaurant, error) {
	return nil, errors.New("not implemented")
}
