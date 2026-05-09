// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package domain_test

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eats/backend/billing/domain"
	"eats/backend/common"
	"eats/backend/common/shared"
)

func TestNewReceipt_ValidReceipt(t *testing.T) {
	data := validReceiptData(t)
	docNumber := newTestDocumentNumber(t)

	doc, err := domain.NewReceipt(data, docNumber)
	require.NoError(t, err)

	assert.Equal(t, domain.DocumentTypeReceipt, doc.DocumentType())
	assert.Equal(t, docNumber, doc.DocumentNumber())
	assert.Equal(t, data.Currency, doc.Currency())
	assert.Equal(t, data.IssueDate, doc.IssueDate())
	assert.Equal(t, data.Seller, doc.Seller())
	assert.Equal(t, data.Buyer, doc.Buyer())
	assert.Equal(t, data.ExternalReference, doc.ExternalReference())
	assert.False(t, doc.UUID().IsZero(), "document UUID should be generated")

	require.Len(t, doc.LineItems(), len(data.LineItems))
	for i, want := range data.LineItems {
		got := doc.LineItems()[i]
		assert.Equal(t, want.Name, got.Name(), "line item %d name", i)
		assert.Equal(t, want.Quantity, got.Quantity(), "line item %d quantity", i)
		assert.False(t, got.PriceBreakdown().TaxRate().IsZero(), "line item %d should have a tax rate", i)
		assert.True(t, got.PriceBreakdown().UnitNetAmount().Equal(want.UnitAmount.Amount()),
			"line item %d unit net should equal input net amount", i)
	}

	// Summary aggregates the line items.
	// 2x Cheeseburger @ $10.00 net + 1x Fries @ $3.50 net, 10% tax.
	summary := doc.Summary()

	assertDecimalsEqual(t, decimal.NewFromFloat(23.50), summary.NetAmount(), "summary NetAmount")
	assertDecimalsEqual(t, decimal.NewFromFloat(2.35), summary.TaxAmount(), "summary TaxAmount")
	assertDecimalsEqual(t, decimal.NewFromFloat(25.85), summary.GrossAmount(), "summary GrossAmount")

	require.Len(t, summary.Taxes(), 1, "all items share the same tax rate")
	assertDecimalsEqual(t, decimal.NewFromFloat(23.50), summary.Taxes()[0].NetAmount(), "tax entry NetAmount")
	assertDecimalsEqual(t, decimal.NewFromFloat(2.35), summary.Taxes()[0].TaxAmount(), "tax entry TaxAmount")
}

func TestNewReceipt_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(t *testing.T, d *domain.NewDocumentData)
		wantErr string
	}{
		{
			name: "zero_buyer_rejected",
			mutate: func(t *testing.T, d *domain.NewDocumentData) {
				d.Buyer = domain.LegalEntity{}
			},
			wantErr: "buyer can't be empty",
		},
		{
			name: "zero_seller_rejected",
			mutate: func(t *testing.T, d *domain.NewDocumentData) {
				d.Seller = domain.LegalEntity{}
			},
			wantErr: "seller can't be empty",
		},
		{
			name: "empty_currency_rejected",
			mutate: func(t *testing.T, d *domain.NewDocumentData) {
				d.Currency = shared.Currency{}
			},
			wantErr: "currency can't be empty",
		},
		{
			name: "empty_issue_date_rejected",
			mutate: func(t *testing.T, d *domain.NewDocumentData) {
				d.IssueDate = time.Time{}
			},
			wantErr: "issue date can't be empty",
		},
		{
			name: "future_issue_date_rejected",
			mutate: func(t *testing.T, d *domain.NewDocumentData) {
				d.IssueDate = time.Now().Add(48 * time.Hour)
			},
			wantErr: "issue date can't be in the future",
		},
		{
			name: "buyer_with_tax_id_rejected",
			mutate: func(t *testing.T, d *domain.NewDocumentData) {
				taxID, err := shared.NewTaxID("99999")
				require.NoError(t, err)

				buyer, err := domain.NewLegalEntity(d.Buyer.Name(), d.Buyer.Address(), &taxID)
				require.NoError(t, err)

				d.Buyer = buyer
			},
			wantErr: "receipts cannot be issued to buyers with a tax ID",
		},
		{
			name: "seller_without_tax_id_rejected",
			mutate: func(t *testing.T, d *domain.NewDocumentData) {
				seller, err := domain.NewLegalEntity(d.Seller.Name(), d.Seller.Address(), nil)
				require.NoError(t, err)

				d.Seller = seller
			},
			wantErr: "seller must have a tax ID",
		},
		{
			name: "empty_line_items_rejected",
			mutate: func(t *testing.T, d *domain.NewDocumentData) {
				d.LineItems = nil
			},
			wantErr: "at least one line item",
		},
		{
			name: "line_item_empty_name_rejected",
			mutate: func(t *testing.T, d *domain.NewDocumentData) {
				d.LineItems[0].Name = ""
			},
			wantErr: "name can't be empty",
		},
		{
			name: "line_item_zero_quantity_rejected",
			mutate: func(t *testing.T, d *domain.NewDocumentData) {
				d.LineItems[0].Quantity = 0
			},
			wantErr: "quantity must be positive",
		},
		{
			name: "line_item_negative_unit_amount_rejected",
			mutate: func(t *testing.T, d *domain.NewDocumentData) {
				d.LineItems[0].UnitAmount = shared.NewNetAmount(decimal.NewFromFloat(-1.00))
			},
			wantErr: "unit amount can't be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := validReceiptData(t)
			tt.mutate(t, &data)

			_, err := domain.NewReceipt(data, newTestDocumentNumber(t))

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestNewReceipt_SummarizesMultipleItems(t *testing.T) {
	data := validReceiptData(t)
	data.LineItems = []domain.NewLineItemData{
		{Name: "Item A", Quantity: 2, UnitAmount: shared.NewNetAmount(decimal.NewFromFloat(10.00))},
		{Name: "Item B", Quantity: 1, UnitAmount: shared.NewNetAmount(decimal.NewFromFloat(5.00))},
		{Name: "Item C", Quantity: 1, UnitAmount: shared.NewNetAmount(decimal.NewFromFloat(15.00))},
		{Name: "Item D", Quantity: 3, UnitAmount: shared.NewNetAmount(decimal.NewFromFloat(5.00))},
		{Name: "Item E", Quantity: 2, UnitAmount: shared.NewNetAmount(decimal.NewFromFloat(10.00))},
	}

	doc, err := domain.NewReceipt(data, newTestDocumentNumber(t))
	require.NoError(t, err)

	// All items share the default 10% tax rate.
	// Total: 20 + 5 + 15 + 15 + 20 = 75.00 net, 7.50 tax, 82.50 gross.
	summary := doc.Summary()

	assertDecimalsEqual(t, decimal.NewFromFloat(75.00), summary.NetAmount(), "total NetAmount")
	assertDecimalsEqual(t, decimal.NewFromFloat(7.50), summary.TaxAmount(), "total TaxAmount")
	assertDecimalsEqual(t, decimal.NewFromFloat(82.50), summary.GrossAmount(), "total GrossAmount")

	require.Len(t, summary.Taxes(), 1, "all items share one tax rate")
	assertDecimalsEqual(t, decimal.NewFromFloat(75.00), summary.Taxes()[0].NetAmount(), "tax group NetAmount")
	assertDecimalsEqual(t, decimal.NewFromFloat(7.50), summary.Taxes()[0].TaxAmount(), "tax group TaxAmount")
}

func validReceiptData(t *testing.T) domain.NewDocumentData {
	t.Helper()

	sellerTaxID, err := shared.NewTaxID("1234567890")
	require.NoError(t, err)

	seller, err := domain.NewLegalEntity("Food Delivery Inc.", newTestAddress(t), &sellerTaxID)
	require.NoError(t, err)

	buyer, err := domain.NewLegalEntity("John Doe", newTestAddress(t), nil)
	require.NoError(t, err)

	return domain.NewDocumentData{
		ExternalReference: common.ToPtr("EXT-REF-001"),
		IssueDate:         time.Date(2025, 3, 14, 0, 0, 0, 0, time.UTC),
		Currency:          shared.MustNewCurrency("USD"),
		Seller:            seller,
		Buyer:             buyer,
		LineItems: []domain.NewLineItemData{
			{
				Name:       "Cheeseburger",
				Quantity:   2,
				UnitAmount: shared.NewNetAmount(decimal.NewFromFloat(10.00)),
			},
			{
				Name:       "Fries",
				Quantity:   1,
				UnitAmount: shared.NewNetAmount(decimal.NewFromFloat(3.50)),
			},
		},
	}
}

func newTestDocumentNumber(t *testing.T) domain.DocumentNumber {
	t.Helper()

	series, err := domain.NewDocumentSeries("R")
	require.NoError(t, err)

	num, err := domain.NewDocumentNumber(series, 1)
	require.NoError(t, err)

	return num
}
