package client

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"eats/backend/common/shared"
)

type Billing interface {
	CalculateTaxes(ctx context.Context, req CalculateTaxesRequest) (CalculateTaxesResponse, error)
	IssueReceipt(ctx context.Context, req IssueReceiptRequest) error
}

type CalculateTaxesRequest struct {
	Currency          shared.Currency
	BuyerCountryCode  shared.CountryCode
	BuyerTaxID        *shared.TaxID
	SellerCountryCode shared.CountryCode
	LineItems         []LineItem
}

type CalculateTaxesResponse struct {
	LineItems []LineItemReadModel

	NetTotal   decimal.Decimal
	TaxTotal   decimal.Decimal
	GrossTotal decimal.Decimal
}

type LineItem struct {
	Name       string
	Type       shared.LineItemType
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

type LineItemReadModel struct {
	Name     string
	Type     shared.LineItemType
	Quantity int

	NetAmount   decimal.Decimal
	TaxAmount   decimal.Decimal
	GrossAmount decimal.Decimal
}
