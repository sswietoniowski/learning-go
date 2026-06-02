package db

import (
	"context"
	"fmt"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eats/backend/common"
	"eats/backend/orders/adapters/db/dbmodels"
	"eats/backend/orders/app"
)

type CourierRepository struct {
	db *pgxpool.Pool
}

func NewCourierRepository(db *pgxpool.Pool) *CourierRepository {
	if db == nil {
		panic("db connection pool cannot be nil")
	}

	return &CourierRepository{
		db: db,
	}
}

func (r *CourierRepository) RegisterCourier(ctx context.Context, courierUUID app.CourierUUID, courier app.RegisterCourier) error {
	return common.UpdateInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		err := queries.InsertCourier(ctx, dbmodels.InsertCourierParams{
			CourierUuid: courierUUID,
			Name:        courier.Name,
			PhoneNumber: courier.PhoneNumber,
			City:        courier.City,
		})
		if err != nil {
			return fmt.Errorf("insert courier failed: %w", err)
		}

		return nil
	})
}

func (r *CourierRepository) GetCourierCity(ctx context.Context, courierUUID app.CourierUUID) (string, error) {
	queries := dbmodels.New(r.db)

	city, err := queries.GetCourierCity(ctx, courierUUID)
	if err != nil {
		return "", fmt.Errorf("failed to get courier city: %w", err)
	}

	return city, nil
}
