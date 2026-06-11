package db

import (
	"context"
	"fmt"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/settlements/adapters/db/dbmodels"
	"eats/backend/settlements/app"
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

// CloseBillingCycle closes the current billing cycle and creates the next one in a single
// serializable transaction. Settlement (invoices) is handled separately to make the
// operation idempotent. With an event-driven approach, closing would emit a
// "BillingCycleClosed" event via the outbox pattern, and settlement would be a separate
// subscriber. See https://threedots.tech/event-driven/
func (r *BillingCycleRepository) CloseBillingCycle(ctx context.Context, partnerUUID domain.LegalEntityUUID) (*domain.BillingCycle, []models.Order, error) {
	var closedCycle *domain.BillingCycle
	var orders []models.Order

	// Serializable isolation is required here to prevent race conditions with
	// AddOrderToCurrentBillingCycle. See TestAddOrderToCurrentBillingCycle_RaceWithClose.
	err := common.UpdateInSerializableTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		currentDBCycle, err := queries.CurrentBillingCycle(ctx, partnerUUID)
		if err != nil {
			return fmt.Errorf("error getting last billing cycle: %w", err)
		}

		currentCycle := newBillingCycleFromDBModel(currentDBCycle)

		// Read orders within the same transaction to ensure consistency.
		// This prevents race conditions where an order is added after we read
		// but before we close the billing cycle.
		orders, err = r.billingCycleOrdersTx(ctx, queries, currentCycle.UUID())
		if err != nil {
			return fmt.Errorf("error getting billing cycle orders: %w", err)
		}

		// Close inside the serializable transaction to guarantee no orders are added
		// between reading orders and marking the cycle as closed.
		// Close is idempotent at the DB level: SaveBillingCycle uses ON CONFLICT DO UPDATE.
		// Documents issued against the closed cycle are immutable and idempotent via
		// external references.
		err = currentCycle.Close()
		if err != nil {
			return fmt.Errorf("error closing billing cycle: %w", err)
		}

		err = queries.SaveBillingCycle(ctx, newBillingCycleSaveParams(currentCycle))
		if err != nil {
			return fmt.Errorf("error saving billing cycle: %w", err)
		}

		nextCycle, err := domain.NewNextBillingCycle(currentCycle)
		if err != nil {
			return fmt.Errorf("error creating next billing cycle: %w", err)
		}

		err = queries.SaveBillingCycle(ctx, newBillingCycleSaveParams(nextCycle))
		if err != nil {
			return fmt.Errorf("error saving next billing cycle: %w", err)
		}

		closedCycle = currentCycle
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return closedCycle, orders, nil
}

func (r *BillingCycleRepository) UnsettledClosedCycles(ctx context.Context, partnerUUID domain.LegalEntityUUID) ([]*domain.BillingCycle, error) {
	queries := dbmodels.New(r.db)

	dbCycles, err := queries.UnsettledClosedCycles(ctx, partnerUUID)
	if err != nil {
		return nil, fmt.Errorf("error fetching unsettled closed cycles: %w", err)
	}

	cycles := make([]*domain.BillingCycle, len(dbCycles))
	for i, dbCycle := range dbCycles {
		cycles[i] = newBillingCycleFromDBModel(dbCycle)
	}

	return cycles, nil
}

func (r *BillingCycleRepository) SettleBillingCycle(ctx context.Context, billingCycleUUID domain.BillingCycleUUID) error {
	return common.UpdateInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		dbCycle, err := queries.GetBillingCycleByUUID(ctx, billingCycleUUID)
		if err != nil {
			return fmt.Errorf("error getting billing cycle: %w", err)
		}

		cycle := newBillingCycleFromDBModel(dbCycle)

		err = cycle.Settle()
		if err != nil {
			return fmt.Errorf("error settling billing cycle: %w", err)
		}

		return queries.SaveBillingCycle(ctx, newBillingCycleSaveParams(cycle))
	})
}

func (r *BillingCycleRepository) CalculateCommissionInvoiceData(ctx context.Context, billingCycleUUID domain.BillingCycleUUID, platformUUID models.PlatformEntityUUID) (app.NewInvoiceData, error) {
	queries := dbmodels.New(r.db)

	invoice, err := queries.CommissionInvoiceByBillingCycleUUID(ctx, billingCycleUUID)
	if err != nil {
		return app.NewInvoiceData{}, fmt.Errorf("error getting commission invoice for billing cycle: %w", err)
	}

	// We need to provide a unique external reference to guarantee idempotency
	externalRef := fmt.Sprintf("settlements-commission-invoice-%v", billingCycleUUID)

	return app.NewInvoiceData{
		ExternalReference: externalRef,
		BuyerUUID:         invoice.BuyerUuid,
		SellerUUID:        platformUUID.LegalEntityUUID,
		LineItems: []app.NewInvoiceDataLineItem{
			{
				Name:      fmt.Sprintf("Platform commission (%d orders)", invoice.Quantity),
				Type:      shared.LineItemTypeService,
				Quantity:  1,
				NetAmount: invoice.NetAmount,
			},
		},
	}, nil
}

func (r *BillingCycleRepository) CalculateDeliveryInvoicesData(ctx context.Context, billingCycleUUID domain.BillingCycleUUID) ([]app.NewInvoiceData, error) {
	queries := dbmodels.New(r.db)

	dbInvoices, err := queries.DeliveryInvoicesByBillingCycleUUID(ctx, billingCycleUUID)
	if err != nil {
		return nil, fmt.Errorf("error getting delivery invoices for billing cycle: %w", err)
	}

	invoices := make([]app.NewInvoiceData, len(dbInvoices))
	for i, dbInvoice := range dbInvoices {
		// We need to provide a unique external reference to guarantee idempotency
		externalRef := fmt.Sprintf("settlements-delivery-invoice-%v-%v-%v", billingCycleUUID, dbInvoice.SellerUuid, dbInvoice.BuyerUuid)

		invoices[i] = app.NewInvoiceData{
			ExternalReference: externalRef,
			BuyerUUID:         dbInvoice.BuyerUuid,
			SellerUUID:        dbInvoice.SellerUuid,
			LineItems: []app.NewInvoiceDataLineItem{
				{
					Name:      fmt.Sprintf("Delivery (%d orders)", dbInvoice.Quantity),
					Type:      shared.LineItemTypeDelivery,
					Quantity:  1,
					NetAmount: dbInvoice.NetAmount,
				},
			},
		}
	}

	return invoices, nil
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
