package db

import (
	"context"
	"fmt"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eats/backend/common"
	"eats/backend/settlements/adapters/db/dbmodels"
	"eats/backend/settlements/app/models"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

func (o *OrderRepository) SaveOrder(ctx context.Context, order models.Order) error {
	return common.UpdateInTx(ctx, o.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		err := queries.SaveOrder(ctx, dbmodels.SaveOrderParams{
			OrderUuid:           order.UUID(),
			RestaurantUuid:      order.RestaurantUUID(),
			CourierUuid:         order.CourierUUID(),
			Currency:            order.Currency(),
			CommissionNetAmount: order.CommissionNetAmount(),
			OrderedAt:           order.OrderedAt(),
		})
		if err != nil {
			return fmt.Errorf("error saving order: %w", err)
		}

		breakdowns := map[dbmodels.SettlementsBreakdownType]models.AmountBreakdown{
			dbmodels.SettlementsBreakdownTypeItems:    order.ItemsBreakdown(),
			dbmodels.SettlementsBreakdownTypeDelivery: order.DeliveryBreakdown(),
			dbmodels.SettlementsBreakdownTypeTotal:    order.TotalBreakdown(),
		}

		for brType, breakdown := range breakdowns {
			err := queries.SaveOrderBreakdown(ctx, dbmodels.SaveOrderBreakdownParams{
				OrderUuid:     order.UUID(),
				BreakdownType: brType,
				NetAmount:     breakdown.Net,
				TaxAmount:     breakdown.Tax,
				GrossAmount:   breakdown.Gross,
			})
			if err != nil {
				return fmt.Errorf("error saving order breakdown: %w", err)
			}
		}

		return nil
	})
}
