package command

import (
	"context"

	"eats/backend/billing/api/module/client"
	"eats/backend/settlements/app/models"
	"eats/backend/settlements/domain"
)

type ModulesContract interface {
	IssueReceipt(ctx context.Context, req client.IssueReceiptRequest) (client.DocumentReadModel, error)
}

type billingCycleRepository interface {
	AddOrderToCurrentBillingCycle(ctx context.Context, partnerUUID domain.LegalEntityUUID, orderUUID models.OrderUUID) error
}

type Handlers struct {
	billingCycleRepository billingCycleRepository
	orderRepository        models.OrderRepository
	legalEntityRepository  models.LegalEntityRepository

	modules ModulesContract
}

func NewHandlers(
	billingCycleRepository billingCycleRepository,
	orderRepository models.OrderRepository,
	legalEntityRepository models.LegalEntityRepository,
	modules ModulesContract,
) *Handlers {
	if billingCycleRepository == nil {
		panic("billingCycleRepository is required")
	}
	if orderRepository == nil {
		panic("orderRepository is required")
	}
	if legalEntityRepository == nil {
		panic("legalEntityRepository is required")
	}
	if modules == nil {
		panic("modules is required")
	}

	return &Handlers{
		billingCycleRepository: billingCycleRepository,
		orderRepository:        orderRepository,
		legalEntityRepository:  legalEntityRepository,
		modules:                modules,
	}
}
