package module

import (
	"context"
	"fmt"

	"eats/backend/billing/api/module/client"
	"eats/backend/billing/app/command"
	"eats/backend/billing/app/query"
	"eats/backend/billing/domain"
	"eats/backend/common/shared"
)

type Billing struct {
	commandHandlers *command.Handlers
	queryHandlers   *query.Handlers
}

func New(
	commandHandlers *command.Handlers,
	queryHandlers *query.Handlers,
) *Billing {
	if commandHandlers == nil {
		panic("commandHandlers cannot be nil")
	}
	if queryHandlers == nil {
		panic("queryHandlers cannot be nil")
	}

	return &Billing{
		commandHandlers: commandHandlers,
		queryHandlers:   queryHandlers,
	}
}

func (b *Billing) IssueReceipt(ctx context.Context, req client.IssueReceiptRequest) (client.DocumentReadModel, error) {
	buyer, err := newDomainLegalEntityFromContract(req.Buyer)
	if err != nil {
		return client.DocumentReadModel{}, fmt.Errorf("could not create buyer domain legal entity: %w", err)
	}

	seller, err := newDomainLegalEntityFromContract(req.Seller)
	if err != nil {
		return client.DocumentReadModel{}, fmt.Errorf("could not create seller domain legal entity: %w", err)
	}

	lineItems := make([]domain.NewLineItemData, 0, len(req.LineItems))
	for _, lineItem := range req.LineItems {
		domainLineItem := domain.NewLineItemData{
			Name:         lineItem.Name,
			LineItemType: lineItem.Type,
			Quantity:     lineItem.Quantity,
			UnitAmount:   lineItem.UnitAmount,
		}
		lineItems = append(lineItems, domainLineItem)
	}

	uuid, err := b.commandHandlers.IssueReceipt(ctx, command.IssueReceipt{
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
		return client.DocumentReadModel{}, err
	}

	doc, err := b.queryHandlers.GetDocumentByUUID(ctx, query.GetDocumentByUUID{
		DocumentUUID: uuid,
	})
	if err != nil {
		return client.DocumentReadModel{}, fmt.Errorf("error getting document: %w", err)
	}

	return newDocumentReadModel(doc), nil
}

func (b *Billing) IssueInvoice(ctx context.Context, req client.IssueInvoiceRequest) (client.DocumentReadModel, error) {
	buyer, err := newDomainLegalEntityFromContract(req.Buyer)
	if err != nil {
		return client.DocumentReadModel{}, fmt.Errorf("could not create buyer domain legal entity: %w", err)
	}

	seller, err := newDomainLegalEntityFromContract(req.Seller)
	if err != nil {
		return client.DocumentReadModel{}, fmt.Errorf("could not create seller domain legal entity: %w", err)
	}

	lineItems := make([]domain.NewLineItemData, 0, len(req.LineItems))
	for _, lineItem := range req.LineItems {
		domainLineItem := domain.NewLineItemData{
			Name:         lineItem.Name,
			LineItemType: lineItem.Type,
			Quantity:     lineItem.Quantity,
			UnitAmount:   lineItem.UnitAmount,
		}
		lineItems = append(lineItems, domainLineItem)
	}

	uuid, err := b.commandHandlers.IssueInvoice(ctx, command.IssueInvoice{
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
		return client.DocumentReadModel{}, err
	}

	doc, err := b.queryHandlers.GetDocumentByUUID(ctx, query.GetDocumentByUUID{
		DocumentUUID: uuid,
	})
	if err != nil {
		return client.DocumentReadModel{}, fmt.Errorf("error getting document: %w", err)
	}

	return newDocumentReadModel(doc), nil
}

func (b *Billing) CalculateTaxes(ctx context.Context, req client.CalculateTaxesRequest) (client.CalculateTaxesResponse, error) {
	return b.queryHandlers.CalculateTaxes(ctx, req)
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

func newDocumentReadModel(doc *domain.Document) client.DocumentReadModel {
	var lineItems []client.LineItemReadModel
	for _, l := range doc.LineItems() {
		lineItems = append(lineItems, client.LineItemReadModel{
			Name:        l.Name(),
			Type:        l.LineItemType(),
			Quantity:    l.Quantity(),
			NetAmount:   l.PriceBreakdown().NetAmount(),
			TaxAmount:   l.PriceBreakdown().TaxAmount(),
			GrossAmount: l.PriceBreakdown().GrossAmount(),
		})
	}

	return client.DocumentReadModel{
		UUID:           doc.UUID().String(),
		DocumentNumber: doc.DocumentNumber().String(),
		LineItems:      lineItems,
		NetTotal:       doc.Summary().NetAmount(),
		TaxTotal:       doc.Summary().TaxAmount(),
		GrossTotal:     doc.Summary().GrossAmount(),
	}
}
