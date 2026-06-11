package command

import (
	"context"
	"fmt"
	"time"

	"eats/backend/billing/api/module/client"
	"eats/backend/common/shared"
	"eats/backend/settlements/app"
	"eats/backend/settlements/app/models"
	"eats/backend/settlements/domain"
)

type invoiceDataGenerator interface {
	CalculateDeliveryInvoicesData(ctx context.Context, billingCycleUUID domain.BillingCycleUUID) ([]app.NewInvoiceData, error)
	CalculateCommissionInvoiceData(ctx context.Context, billingCycleUUID domain.BillingCycleUUID, platformUUID models.PlatformEntityUUID) (app.NewInvoiceData, error)
}

type CloseBillingCycle struct {
	PartnerUUID domain.LegalEntityUUID
}

func (h *Handlers) CloseBillingCycle(ctx context.Context, cmd CloseBillingCycle) error {
	// Step 1: Close the current cycle + create next (atomic serializable tx).
	// This is safe to retry: closing an empty cycle just adds a settled-immediately cycle.
	_, _, err := h.billingCycleRepository.CloseBillingCycle(ctx, cmd.PartnerUUID)
	if err != nil {
		return fmt.Errorf("error closing billing cycle: %w", err)
	}

	// Step 2: Settle all closed-but-not-settled cycles.
	// Separating close from settlement makes the operation idempotent: retrying
	// always picks up where it left off. Each step within settlement is also
	// idempotent (invoices via external references).
	//
	// With an event-driven approach, closing the billing cycle would emit a
	// "BillingCycleClosed" event (via outbox pattern), and settlement would be
	// a separate subscriber with at-least-once delivery guaranteeing completion.
	// See https://threedots.tech/event-driven/
	unsettledCycles, err := h.billingCycleRepository.UnsettledClosedCycles(ctx, cmd.PartnerUUID)
	if err != nil {
		return fmt.Errorf("error fetching unsettled cycles: %w", err)
	}

	for _, cycle := range unsettledCycles {
		orders, err := h.billingCycleRepository.BillingCycleOrders(ctx, cycle.UUID())
		if err != nil {
			return fmt.Errorf("error fetching orders for cycle %v: %w", cycle.UUID(), err)
		}

		if len(orders) > 0 {
			// Issue invoices for the closed cycle.
			// Invoices are idempotent via external reference. Retrying is safe.
			switch cycle.PartnerType() {
			case domain.PartnerTypeRestaurant:
				err = h.issueRestaurantInvoices(ctx, cycle.UUID(), cycle.PartnerUUID())
			case domain.PartnerTypeCourier:
				err = h.issueCourierInvoices(ctx, cycle.UUID())
			default:
				err = fmt.Errorf("unsupported partner type for billing cycle: %s", cycle.PartnerType())
			}
			if err != nil {
				return err
			}
		}

		err = h.billingCycleRepository.SettleBillingCycle(ctx, cycle.UUID())
		if err != nil {
			return fmt.Errorf("error marking cycle as settled: %w", err)
		}
	}

	return nil
}

func (h *Handlers) issueRestaurantInvoices(ctx context.Context, billingCycleUUID domain.BillingCycleUUID, partnerUUID domain.LegalEntityUUID) error {
	partner, err := h.legalEntityRepository.PartnerByUUID(ctx, partnerUUID)
	if err != nil {
		return fmt.Errorf("error fetching partner: %w", err)
	}

	invoice, err := h.invoiceDataGenerator.CalculateCommissionInvoiceData(ctx, billingCycleUUID, partner.PlatformEntityUUID)
	if err != nil {
		return fmt.Errorf("error generating commission invoice data: %w", err)
	}

	_, err = h.issueInvoice(ctx, h.legalEntityRepository, invoice)
	if err != nil {
		return fmt.Errorf("error issuing commission invoice: %w", err)
	}

	return nil
}

func (h *Handlers) issueCourierInvoices(ctx context.Context, billingCycleUUID domain.BillingCycleUUID) error {
	deliveryInvoices, err := h.invoiceDataGenerator.CalculateDeliveryInvoicesData(ctx, billingCycleUUID)
	if err != nil {
		return fmt.Errorf("error generating delivery invoices: %w", err)
	}

	// Use cached finder as we issue many invoices for the same seller
	cachedFinder := app.NewCachedLegalEntityFinder(h.legalEntityRepository)

	for _, invoice := range deliveryInvoices {
		_, err = h.issueInvoice(ctx, cachedFinder, invoice)
		if err != nil {
			return fmt.Errorf("error issuing delivery invoice: %w", err)
		}
	}

	return nil
}

type legalEntityFinder interface {
	LegalEntityByUUID(ctx context.Context, uuid domain.LegalEntityUUID) (models.LegalEntity, error)
}

func (h *Handlers) issueInvoice(ctx context.Context, legalEntityFinder legalEntityFinder, inv app.NewInvoiceData) (client.DocumentReadModel, error) {
	buyer, err := legalEntityFinder.LegalEntityByUUID(ctx, inv.BuyerUUID)
	if err != nil {
		return client.DocumentReadModel{}, fmt.Errorf("error getting buyer legal entity: %w", err)
	}

	seller, err := legalEntityFinder.LegalEntityByUUID(ctx, inv.SellerUUID)
	if err != nil {
		return client.DocumentReadModel{}, fmt.Errorf("error getting seller legal entity: %w", err)
	}

	if buyer.Currency != seller.Currency {
		return client.DocumentReadModel{}, fmt.Errorf("buyer and seller currency mismatch: %s vs %s", buyer.Currency.Code(), seller.Currency.Code())
	}

	var lineItems []client.LineItem

	for _, lineItem := range inv.LineItems {
		billingLineItem := client.LineItem{
			Name:       lineItem.Name,
			Type:       lineItem.Type,
			Quantity:   lineItem.Quantity,
			UnitAmount: shared.NewNetAmount(lineItem.NetAmount),
		}

		lineItems = append(lineItems, billingLineItem)
	}

	req := client.IssueInvoiceRequest{
		ExternalReference: &inv.ExternalReference,
		IssueDate:         time.Now(),
		Currency:          buyer.Currency,
		Seller:            newModuleLegalEntity(seller),
		Buyer:             newModuleLegalEntity(buyer),
		LineItems:         lineItems,
	}

	doc, err := h.modules.IssueInvoice(ctx, req)
	if err != nil {
		return client.DocumentReadModel{}, fmt.Errorf("error issuing invoice: %w", err)
	}

	return doc, nil
}

func newModuleLegalEntity(p models.LegalEntity) client.LegalEntity {
	return client.LegalEntity{
		Name:    p.BusinessName,
		Address: p.Address,
		TaxID:   &p.TaxID,
	}
}
