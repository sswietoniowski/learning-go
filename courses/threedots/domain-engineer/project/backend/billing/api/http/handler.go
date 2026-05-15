package http

import (
	"context"

	"eats/backend/billing/app/command"
	"eats/backend/billing/app/query"
	"eats/backend/billing/domain"
	"eats/backend/common/shared"
)

type Handler struct {
	commands *command.Handlers
	queries  *query.Handlers
}

func (h Handler) CreateReceipt(ctx context.Context, request CreateReceiptRequestObject) (CreateReceiptResponseObject, error) {
	details, err := newDocumentDetailsFromCreateDocument(*request.Body)
	if err != nil {
		return nil, err
	}

	cmd := command.IssueReceipt{
		DocumentData: details,
	}
	uuid, err := h.commands.IssueReceipt(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return CreateReceipt201JSONResponse{
		DocumentUuid: uuid,
	}, nil
}

func (h Handler) GetDocument(ctx context.Context, request GetDocumentRequestObject) (GetDocumentResponseObject, error) {
	doc, err := h.queries.GetDocumentByUUID(ctx, query.GetDocumentByUUID{
		DocumentUUID: request.DocumentUuid,
	})
	if err != nil {
		return nil, err
	}

	return GetDocument200JSONResponse(documentToResponse(doc)), nil
}

func (h Handler) PrintDocument(ctx context.Context, request PrintDocumentRequestObject) (PrintDocumentResponseObject, error) {
	cmd := command.PrintDocument{
		DocumentUUID: request.DocumentUuid,
	}

	err := h.commands.PrintDocument(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return PrintDocument204Response{}, nil
}

func documentToResponse(doc *domain.Document) DocumentResponse {
	lineItems := make([]ResponseLineItem, 0, len(doc.LineItems()))
	for _, li := range doc.LineItems() {
		lineItems = append(lineItems, lineItemToResponse(li))
	}

	return DocumentResponse{
		Uuid:           doc.UUID(),
		DocumentType:   DocumentType(doc.DocumentType().String()),
		DocumentNumber: doc.DocumentNumber().String(),
		IssueDate:      doc.IssueDate(),
		Currency:       doc.Currency(),
		Seller:         legalEntityToResponse(doc.Seller()),
		Buyer:          legalEntityToResponse(doc.Buyer()),
		LineItems:      lineItems,
		Summary:        summaryToResponse(doc.Summary()),
	}
}

func legalEntityToResponse(le domain.LegalEntity) LegalEntity {
	return LegalEntity{
		Name: le.Name(),
		Address: Address{
			Line1:       le.Address().Line1(),
			Line2:       le.Address().Line2(),
			City:        le.Address().City(),
			PostalCode:  le.Address().PostalCode(),
			CountryCode: le.Address().CountryCode(),
		},
		TaxId: le.TaxID(),
	}
}

func lineItemToResponse(li domain.LineItem) ResponseLineItem {
	return ResponseLineItem{
		Name:      li.Name(),
		Quantity:  li.Quantity(),
		Breakdown: priceBreakdownToResponse(li.PriceBreakdown()),
	}
}

func priceBreakdownToResponse(pb domain.PriceBreakdown) PriceBreakdown {
	return PriceBreakdown{
		TaxRate:         pb.TaxRate().Rate(),
		TaxType:         TaxType(pb.TaxRate().TaxType().String()),
		UnitNetAmount:   pb.UnitNetAmount(),
		UnitTaxAmount:   pb.UnitTaxAmount(),
		UnitGrossAmount: pb.UnitGrossAmount(),
		NetAmount:       pb.NetAmount(),
		TaxAmount:       pb.TaxAmount(),
		GrossAmount:     pb.GrossAmount(),
	}
}

func summaryToResponse(s domain.PriceBreakdownSummary) PriceBreakdownSummary {
	taxes := make([]TaxSummary, 0, len(s.Taxes()))
	for _, t := range s.Taxes() {
		taxes = append(taxes, taxSummaryToResponse(t))
	}

	return PriceBreakdownSummary{
		NetAmount:   s.NetAmount(),
		TaxAmount:   s.TaxAmount(),
		GrossAmount: s.GrossAmount(),
		Taxes:       taxes,
	}
}

func taxSummaryToResponse(ts domain.TaxSummary) TaxSummary {
	return TaxSummary{
		TaxRate:   ts.TaxRate().Rate(),
		TaxType:   TaxType(ts.TaxRate().TaxType().String()),
		NetAmount: ts.NetAmount(),
		TaxAmount: ts.TaxAmount(),
	}
}

func newDocumentDetailsFromCreateDocument(cd CreateDocument) (domain.NewDocumentData, error) {
	seller, err := newLegalEntityFromHTTP(cd.Seller)
	if err != nil {
		return domain.NewDocumentData{}, err
	}

	buyer, err := newLegalEntityFromHTTP(cd.Buyer)
	if err != nil {
		return domain.NewDocumentData{}, err
	}

	var lineItems []domain.NewLineItemData
	for _, httpLineItem := range cd.LineItems {
		var unitAmount shared.LineAmount

		if httpLineItem.IsGross {
			unitAmount = shared.NewGrossAmount(httpLineItem.UnitAmount)
		} else {
			unitAmount = shared.NewNetAmount(httpLineItem.UnitAmount)
		}

		lineItem := domain.NewLineItemData{
			Name:       httpLineItem.Name,
			Quantity:   httpLineItem.Quantity,
			UnitAmount: unitAmount,
		}

		lineItems = append(lineItems, lineItem)
	}

	dd := domain.NewDocumentData{
		IssueDate: cd.IssueDate,
		Currency:  cd.Currency,
		Seller:    seller,
		Buyer:     buyer,
		LineItems: lineItems,
	}

	return dd, nil
}

func newLegalEntityFromHTTP(cd LegalEntity) (domain.LegalEntity, error) {
	address, err := shared.NewAddress(
		cd.Address.Line1,
		cd.Address.Line2,
		cd.Address.PostalCode,
		cd.Address.City,
		cd.Address.CountryCode,
	)
	if err != nil {
		return domain.LegalEntity{}, err
	}

	return domain.NewLegalEntity(cd.Name, address, cd.TaxId)
}

func Register(ctx context.Context, e EchoRouter, commands *command.Handlers, queries *query.Handlers) error {
	handler := Handler{
		commands: commands,
		queries:  queries,
	}

	RegisterHandlers(e, NewStrictHandler(handler, nil))

	return nil
}
