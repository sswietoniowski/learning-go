package db

import (
	"context"

	"eats/backend/common"
	"eats/backend/orders/adapters/db/dbmodels"
	"eats/backend/orders/app"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

func (r *CourierRepository) RegisterCourier(ctx context.Context, courier app.Courier) error {
	return common.UpdateInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		err := queries.InsertCourier(ctx, dbmodels.InsertCourierParams{
			CourierUuid: courier.CourierUUID,
			Name:        courier.Name,
			PhoneNumber: courier.PhoneNumber,
			City:        courier.City,
		})
		if err != nil {
			return err
		}

		return nil
	})
}
