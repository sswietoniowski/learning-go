package db

import (
	"context"
	"fmt"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eats/backend/common"
	"eats/backend/settlements/adapters/db/dbmodels"
	"eats/backend/settlements/app/models"
	"eats/backend/settlements/app/query"
	"eats/backend/settlements/domain"
)

type BillingCycleRepository struct {
	db *pgxpool.Pool
}

func NewBillingCycleRepository(db *pgxpool.Pool) *BillingCycleRepository {
	if db == nil {
		panic("db is nil")
	}

	return &BillingCycleRepository{
		db: db,
	}
}

func (r *BillingCycleRepository) AddOrderToCurrentBillingCycle(ctx context.Context, partnerUUID domain.LegalEntityUUID, orderUUID models.OrderUUID) error {
	return common.UpdateInSerializableTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		alreadyExists, err := queries.OrderInPartnerBillingCycleExists(ctx, dbmodels.OrderInPartnerBillingCycleExistsParams{
			OrderUuid:   orderUUID,
			PartnerUuid: partnerUUID,
		})
		if err != nil {
			return fmt.Errorf("error checking if order exists in billing cycle: %w", err)
		}
		if alreadyExists {
			return nil
		}

		lastCycle, err := queries.CurrentBillingCycle(ctx, partnerUUID)
		if err != nil {
			return fmt.Errorf("error getting last billing cycle: %w", err)
		}

		err = queries.AddOrderToBillingCycle(ctx, dbmodels.AddOrderToBillingCycleParams{
			BillingCycleUuid: lastCycle.BillingCycleUuid,
			OrderUuid:        orderUUID,
		})
		if err != nil {
			return fmt.Errorf("error adding order to billing cycle: %w", err)
		}

		return nil
	})
}

func (r *BillingCycleRepository) BillingCycleOrders(ctx context.Context, billingCycleUUID domain.BillingCycleUUID) ([]models.Order, error) {
	queries := dbmodels.New(r.db)
	return r.billingCycleOrdersTx(ctx, queries, billingCycleUUID)
}

func (r *BillingCycleRepository) billingCycleOrdersTx(ctx context.Context, queries *dbmodels.Queries, billingCycleUUID domain.BillingCycleUUID) ([]models.Order, error) {
	dbOrders, err := queries.OrdersByBillingCycleUUID(ctx, billingCycleUUID)
	if err != nil {
		return nil, fmt.Errorf("error getting orders for billing cycle: %w", err)
	}

	dbBreakdowns, err := queries.OrderBreakdownsByBillingCycleUUID(ctx, billingCycleUUID)
	if err != nil {
		return nil, fmt.Errorf("error getting order breakdowns for billing cycle: %w", err)
	}

	breakdownsMap := make(map[models.OrderUUID]map[dbmodels.SettlementsBreakdownType]models.AmountBreakdown)
	for _, br := range dbBreakdowns {
		breakdowns, ok := breakdownsMap[br.OrderUuid]
		if !ok {
			breakdowns = make(map[dbmodels.SettlementsBreakdownType]models.AmountBreakdown)
		}

		breakdowns[br.BreakdownType] = models.AmountBreakdown{
			Net:   br.NetAmount,
			Tax:   br.TaxAmount,
			Gross: br.GrossAmount,
		}

		breakdownsMap[br.OrderUuid] = breakdowns
	}

	orders := make([]models.Order, len(dbOrders))
	for i, dbOrder := range dbOrders {
		breakdowns, ok := breakdownsMap[dbOrder.OrderUuid]
		if !ok {
			// This indicates data inconsistency: an order exists in the billing cycle
			// but an associated breakdown is missing. This should not happen in normal operation
			// and likely requires manual database intervention to fix (either add missing
			// breakdowns or remove the orphaned order from the billing cycle).
			return nil, fmt.Errorf("no breakdowns found for order %v", dbOrder.OrderUuid)
		}

		order := models.UnmarshalOrder(
			dbOrder.OrderUuid,
			dbOrder.RestaurantUuid,
			dbOrder.CourierUuid,
			dbOrder.Currency,
			breakdowns[dbmodels.SettlementsBreakdownTypeItems],
			breakdowns[dbmodels.SettlementsBreakdownTypeDelivery],
			breakdowns[dbmodels.SettlementsBreakdownTypeTotal],
			dbOrder.CommissionNetAmount,
			dbOrder.OrderedAt,
		)

		orders[i] = order
	}

	return orders, nil
}

func (r *BillingCycleRepository) BillingCyclesForPartner(ctx context.Context, partnerUUID domain.LegalEntityUUID) ([]query.BillingCycleReadModel, error) {
	queries := dbmodels.New(r.db)

	rows, err := queries.BillingCyclesByPartnerUUID(ctx, partnerUUID)
	if err != nil {
		return nil, fmt.Errorf("error fetching billing cycles: %w", err)
	}

	if len(rows) == 0 {
		return []query.BillingCycleReadModel{}, nil
	}

	readModels := make([]query.BillingCycleReadModel, 0, len(rows))
	for _, bc := range rows {
		readModel := query.BillingCycleReadModel{
			BillingCycleUUID:   bc.BillingCycleUuid,
			PartnerUUID:        bc.PartnerUuid,
			BillingCycleNumber: int(bc.BillingCycleNumber),
			StartDate:          bc.StartDate,
			EndDate:            bc.EndDate,
			Closed:             bc.Closed,
			Settled:            bc.Settled,
		}

		readModels = append(readModels, readModel)
	}

	return readModels, nil
}

func newBillingCycleSaveParams(billingCycle *domain.BillingCycle) dbmodels.SaveBillingCycleParams {
	return dbmodels.SaveBillingCycleParams{
		BillingCycleUuid:   billingCycle.UUID(),
		PartnerUuid:        billingCycle.PartnerUUID(),
		PartnerType:        billingCycle.PartnerType(),
		BillingCycleNumber: int32(billingCycle.Number()),
		Closed:             billingCycle.Closed(),
		Settled:            billingCycle.Settled(),
		StartDate:          billingCycle.StartDate(),
		EndDate:            billingCycle.EndDate(),
	}
}

func newBillingCycleFromDBModel(dbModel dbmodels.SettlementsBillingCycle) *domain.BillingCycle {
	return domain.UnmarshalBillingCycle(
		dbModel.BillingCycleUuid,
		dbModel.PartnerUuid,
		dbModel.PartnerType,
		int(dbModel.BillingCycleNumber),
		dbModel.Closed,
		dbModel.Settled,
		dbModel.StartDate,
		dbModel.EndDate,
	)
}
