package command

import (
	"context"

	"eats/backend/billing/api/module/client"
	"eats/backend/settlements/app/models"
)

type ModulesContract interface {
	IssueReceipt(ctx context.Context, req client.IssueReceiptRequest) (client.DocumentReadModel, error)
	IssueInvoice(ctx context.Context, req client.IssueInvoiceRequest) (client.DocumentReadModel, error)
}

type Handlers struct {
	billingCycleRepository models.BillingCycleRepository
	orderRepository        models.OrderRepository
	legalEntityRepository  models.LegalEntityRepository
	invoiceDataGenerator   invoiceDataGenerator

	modules ModulesContract
}

func NewHandlers(
	billingCycleRepository models.BillingCycleRepository,
	orderRepository models.OrderRepository,
	legalEntityRepository models.LegalEntityRepository,
	invoiceDataGenerator invoiceDataGenerator,
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
	if invoiceDataGenerator == nil {
		panic("invoiceDataGenerator is required")
	}
	if modules == nil {
		panic("modules is required")
	}

	return &Handlers{
		billingCycleRepository: billingCycleRepository,
		orderRepository:        orderRepository,
		legalEntityRepository:  legalEntityRepository,
		invoiceDataGenerator:   invoiceDataGenerator,
		modules:                modules,
	}
}
