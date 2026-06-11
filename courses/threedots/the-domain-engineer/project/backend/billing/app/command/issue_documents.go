package command

import (
	"context"
	"fmt"

	"eats/backend/billing/domain"
)

type IssueReceipt struct {
	DocumentData domain.NewDocumentData
}

func (h *Handlers) IssueReceipt(ctx context.Context, cmd IssueReceipt) (domain.DocumentUUID, error) {
	// Build the document outside the transaction to avoid holding a database connection
	// during inter-module calls (tax rate lookups for each line item). If other modules share the
	// same database, exhausting the pool with slow inter-module calls is a self-inflicted DDoS.
	// In production, use a separate database user per module with its own connection limit.
	builder, err := h.documentFactory.NewReceiptBuilder(ctx, cmd.DocumentData)
	if err != nil {
		return domain.DocumentUUID{}, fmt.Errorf("error building receipt: %w", err)
	}

	return h.documentRepository.CreateDocument(
		ctx,
		domain.DocumentSeriesReceipt,
		func(documentNumber domain.DocumentNumber) (*domain.Document, error) {
			// No inter-module calls here: this just assigns the document number.
			return builder.Build(documentNumber)
		},
	)
}

type IssueInvoice struct {
	DocumentData domain.NewDocumentData
}

func (h *Handlers) IssueInvoice(ctx context.Context, cmd IssueInvoice) (domain.DocumentUUID, error) {
	// Build the document outside the transaction to avoid holding a database connection
	// during inter-module calls (tax rate lookups for each line item). If other modules share the
	// same database, exhausting the pool with slow inter-module calls is a self-inflicted DDoS.
	// In production, use a separate database user per module with its own connection limit.
	builder, err := h.documentFactory.NewInvoiceBuilder(ctx, cmd.DocumentData)
	if err != nil {
		return domain.DocumentUUID{}, fmt.Errorf("error building invoice: %w", err)
	}

	return h.documentRepository.CreateDocument(
		ctx,
		domain.DocumentSeriesInvoice,
		func(documentNumber domain.DocumentNumber) (*domain.Document, error) {
			// No inter-module calls here: this just assigns the document number.
			return builder.Build(documentNumber)
		},
	)
}
