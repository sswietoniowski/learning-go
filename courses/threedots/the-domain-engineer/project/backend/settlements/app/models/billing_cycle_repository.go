package models

import (
	"context"

	"eats/backend/settlements/domain"
)

type BillingCycleRepository interface {
	AddOrderToCurrentBillingCycle(ctx context.Context, partnerUUID domain.LegalEntityUUID, orderUUID OrderUUID) error
	BillingCycleOrders(ctx context.Context, billingCycleUUID domain.BillingCycleUUID) ([]Order, error)
	CloseBillingCycle(ctx context.Context, partnerUUID domain.LegalEntityUUID) (*domain.BillingCycle, []Order, error)
	UnsettledClosedCycles(ctx context.Context, partnerUUID domain.LegalEntityUUID) ([]*domain.BillingCycle, error)
	SettleBillingCycle(ctx context.Context, billingCycleUUID domain.BillingCycleUUID) error
}
