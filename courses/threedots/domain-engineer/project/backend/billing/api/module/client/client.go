package client

import (
	"context"
	"time"

	"eats/backend/common/shared"
)

type Billing interface {
	IssueReceipt(ctx context.Context, req IssueReceiptRequest) error
}

type LineItem struct {
	Name       string
	UnitAmount shared.LineAmount
	Quantity   int
}

type IssueReceiptRequest struct {
	ExternalReference *string
	IssueDate         time.Time
	Currency          shared.Currency

	Seller    LegalEntity
	Buyer     LegalEntity
	LineItems []LineItem
}

type LegalEntity struct {
	Name    string
	Address shared.Address
	TaxID   *shared.TaxID
}
