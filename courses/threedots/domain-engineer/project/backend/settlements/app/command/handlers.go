package command

import (
	"context"

	"eats/backend/billing/api/module/client"
	"eats/backend/settlements/app/models"
)

type ModulesContract interface {
	IssueReceipt(ctx context.Context, req client.IssueReceiptRequest) (client.DocumentReadModel, error)
}

type Handlers struct {
	orderRepository       models.OrderRepository
	legalEntityRepository models.LegalEntityRepository

	modules ModulesContract
}

func NewHandlers(
	orderRepository models.OrderRepository,
	legalEntityRepository models.LegalEntityRepository,
	modules ModulesContract,
) *Handlers {
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
		orderRepository:       orderRepository,
		legalEntityRepository: legalEntityRepository,
		modules:               modules,
	}
}
