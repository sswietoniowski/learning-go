package module

import (
	"context"
	"fmt"

	"eats/backend/billing/api/module/client"
	"eats/backend/billing/app/command"
	"eats/backend/billing/domain"
	"eats/backend/common/shared"
)

type Billing struct {
	commandHandlers *command.Handlers
}

func New(
	commandHandlers *command.Handlers,
) *Billing {
	if commandHandlers == nil {
		panic("commandHandlers cannot be nil")
	}

	return &Billing{
		commandHandlers: commandHandlers,
	}
}

func (b *Billing) IssueReceipt(ctx context.Context, req client.IssueReceiptRequest) error {
	buyer, err := newDomainLegalEntityFromContract(req.Buyer)
	if err != nil {
		return fmt.Errorf("could not create buyer domain legal entity: %w", err)
	}

	seller, err := newDomainLegalEntityFromContract(req.Seller)
	if err != nil {
		return fmt.Errorf("could not create seller domain legal entity: %w", err)
	}

	lineItems := make([]domain.NewLineItemData, 0, len(req.LineItems))
	for _, lineItem := range req.LineItems {
		domainLineItem := domain.NewLineItemData{
			Name:       lineItem.Name,
			Quantity:   lineItem.Quantity,
			UnitAmount: lineItem.UnitAmount,
		}
		lineItems = append(lineItems, domainLineItem)
	}

	_, err = b.commandHandlers.IssueReceipt(ctx, command.IssueReceipt{
		DocumentData: domain.NewDocumentData{
			ExternalReference: req.ExternalReference,
			IssueDate:         req.IssueDate,
			Currency:          req.Currency,
			Seller:            *seller,
			Buyer:             *buyer,
			LineItems:         lineItems,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func newDomainLegalEntityFromContract(le client.LegalEntity) (*domain.LegalEntity, error) {
	address, err := shared.NewAddress(
		le.Address.Line1(),
		le.Address.Line2(),
		le.Address.PostalCode(),
		le.Address.City(),
		le.Address.CountryCode(),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating address from contract: %w", err)
	}

	domainLe, err := domain.NewLegalEntity(
		le.Name,
		address,
		le.TaxID,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating domain le: %w", err)
	}

	return &domainLe, nil
}
