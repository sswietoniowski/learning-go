package db

import (
	"context"
	"fmt"

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

func (r *CourierRepository) RegisterCourier(ctx context.Context, courierOrUUID any, maybeCourier ...app.Courier) error {
	courier, err := normalizeCourierRegistrationInput(courierOrUUID, maybeCourier...)
	if err != nil {
		return err
	}

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

func normalizeCourierRegistrationInput(courierOrUUID any, maybeCourier ...app.Courier) (app.Courier, error) {
	switch value := courierOrUUID.(type) {
	case app.Courier:
		if len(maybeCourier) != 0 {
			return app.Courier{}, fmt.Errorf("unexpected extra courier arguments: %d", len(maybeCourier))
		}
		return value, nil
	case app.CourierUUID:
		if len(maybeCourier) != 1 {
			return app.Courier{}, fmt.Errorf("expected courier payload alongside courier UUID")
		}
		courier := maybeCourier[0]
		courier.CourierUUID = value
		return courier, nil
	default:
		return app.Courier{}, fmt.Errorf("unsupported courier registration input type %T", courierOrUUID)
	}
}

func (r *CourierRepository) GetCourierCity(ctx context.Context, courierUUID app.CourierUUID) (string, error) {
	queries := dbmodels.New(r.db)
	return queries.GetCourierCity(ctx, courierUUID)
}
