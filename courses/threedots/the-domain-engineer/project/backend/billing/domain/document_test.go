// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package domain_test

import (
	"context"
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
	factory := domain.NewDocumentFactory(stubTaxProvider{})
	data := validReceiptData(t)
	docNumber := newTestDocumentNumber(t)

	builder, err := factory.NewReceiptBuilder(context.Background(), data)
	require.NoError(t, err)

	doc, err := builder.Build(docNumber)
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

	// 2x Cheeseburger @ $10.00 net + 1x Fries @ $3.50 net, 23% VAT (food).
	summary := doc.Summary()

	assertDecimalsEqual(t, decimal.NewFromFloat(23.50), summary.NetAmount(), "summary NetAmount")
	assertDecimalsEqual(t, decimal.NewFromFloat(5.41), summary.TaxAmount(), "summary TaxAmount")
	assertDecimalsEqual(t, decimal.NewFromFloat(28.91), summary.GrossAmount(), "summary GrossAmount")

	require.Len(t, summary.Taxes(), 1, "all food items share the same VAT rate")
	assertDecimalsEqual(t, decimal.NewFromFloat(23.50), summary.Taxes()[0].NetAmount(), "tax entry NetAmount")
	assertDecimalsEqual(t, decimal.NewFromFloat(5.41), summary.Taxes()[0].TaxAmount(), "tax entry TaxAmount")
}

func TestNewReceipt_ValidationErrors(t *testing.T) {
	factory := domain.NewDocumentFactory(stubTaxProvider{})

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

			_, err := factory.NewReceiptBuilder(context.Background(), data)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestNewReceipt_SummarizesMultipleItems(t *testing.T) {
	factory := domain.NewDocumentFactory(stubTaxProvider{})
	data := validReceiptData(t)
	data.LineItems = []domain.NewLineItemData{
		{Name: "Item A", LineItemType: shared.LineItemTypeFood, Quantity: 2, UnitAmount: shared.NewNetAmount(decimal.NewFromFloat(10.00))},
		{Name: "Item B", LineItemType: shared.LineItemTypeFood, Quantity: 1, UnitAmount: shared.NewNetAmount(decimal.NewFromFloat(5.00))},
		{Name: "Item C", LineItemType: shared.LineItemTypeBeverage, Quantity: 1, UnitAmount: shared.NewNetAmount(decimal.NewFromFloat(15.00))},
		{Name: "Item D", LineItemType: shared.LineItemTypeDelivery, Quantity: 3, UnitAmount: shared.NewNetAmount(decimal.NewFromFloat(5.00))},
		{Name: "Item E", LineItemType: shared.LineItemTypeService, Quantity: 2, UnitAmount: shared.NewNetAmount(decimal.NewFromFloat(10.00))},
	}

	builder, err := factory.NewReceiptBuilder(context.Background(), data)
	require.NoError(t, err)

	doc, err := builder.Build(newTestDocumentNumber(t))
	require.NoError(t, err)

	// Food: 25.00 net, 23% VAT = 5.75 tax
	// Beverage: 15.00 net, 8% VAT = 1.20 tax
	// Delivery: 15.00 net, 10% GST = 1.50 tax
	// Service: 20.00 net, 0% Sales Tax = 0.00 tax
	summary := doc.Summary()

	assertDecimalsEqual(t, decimal.NewFromFloat(75.00), summary.NetAmount(), "total NetAmount")
	assertDecimalsEqual(t, decimal.NewFromFloat(8.45), summary.TaxAmount(), "total TaxAmount")
	assertDecimalsEqual(t, decimal.NewFromFloat(83.45), summary.GrossAmount(), "total GrossAmount")

	require.Len(t, summary.Taxes(), 4)

	vat23 := findTax(t, summary.Taxes(), domain.UnmarshalTaxRate(decimal.NewFromFloat(0.23), domain.TaxTypeVAT))
	assertDecimalsEqual(t, decimal.NewFromFloat(25.00), vat23.NetAmount(), "VAT 23% net")
	assertDecimalsEqual(t, decimal.NewFromFloat(5.75), vat23.TaxAmount(), "VAT 23% tax")

	vat8 := findTax(t, summary.Taxes(), domain.UnmarshalTaxRate(decimal.NewFromFloat(0.08), domain.TaxTypeVAT))
	assertDecimalsEqual(t, decimal.NewFromFloat(15.00), vat8.NetAmount(), "VAT 8% net")
	assertDecimalsEqual(t, decimal.NewFromFloat(1.20), vat8.TaxAmount(), "VAT 8% tax")

	gst10 := findTax(t, summary.Taxes(), domain.UnmarshalTaxRate(decimal.NewFromFloat(0.10), domain.TaxTypeGST))
	assertDecimalsEqual(t, decimal.NewFromFloat(15.00), gst10.NetAmount(), "GST 10% net")
	assertDecimalsEqual(t, decimal.NewFromFloat(1.50), gst10.TaxAmount(), "GST 10% tax")

	sales0 := findTax(t, summary.Taxes(), domain.UnmarshalTaxRate(decimal.Zero, domain.TaxTypeSalesTax))
	assertDecimalsEqual(t, decimal.NewFromFloat(20.00), sales0.NetAmount(), "Sales Tax 0% net")
	assertDecimalsEqual(t, decimal.NewFromFloat(0.00), sales0.TaxAmount(), "Sales Tax 0% tax")
}

func TestNewReceipt_FoodAndBeverage(t *testing.T) {
	factory := domain.NewDocumentFactory(stubTaxProvider{})

	sellerAddress, err := shared.NewAddress("123 Food St.", "Suite 100", "98765", "Gourmet City", shared.MustNewCountryCode("US"))
	require.NoError(t, err)

	sellerTaxID, err := shared.NewTaxID("1234567890")
	require.NoError(t, err)

	seller, err := domain.NewLegalEntity("Food Delivery Inc.", sellerAddress, &sellerTaxID)
	require.NoError(t, err)

	buyerAddress, err := shared.NewAddress("456 Main St.", "", "12345", "Hometown", shared.MustNewCountryCode("US"))
	require.NoError(t, err)

	buyer, err := domain.NewLegalEntity("John Doe", buyerAddress, nil)
	require.NoError(t, err)

	series, err := domain.NewDocumentSeries("TEST-RECEIPT")
	require.NoError(t, err)
	docNumber, err := domain.NewDocumentNumber(series, 2)
	require.NoError(t, err)

	builder, err := factory.NewReceiptBuilder(context.Background(), domain.NewDocumentData{
		ExternalReference: common.ToPtr("EXTERNAL-REF-456"),
		IssueDate:         time.Date(2025, 3, 14, 0, 0, 0, 0, time.UTC),
		Currency:          shared.MustNewCurrency("USD"),
		Seller:            seller,
		Buyer:             buyer,
		LineItems: []domain.NewLineItemData{
			{
				Name:         "Cheeseburger",
				LineItemType: shared.LineItemTypeFood,
				Quantity:     2,
				UnitAmount:   shared.NewNetAmount(decimal.NewFromFloat(10.00)),
			},
			{
				Name:         "Mineral Water",
				LineItemType: shared.LineItemTypeBeverage,
				Quantity:     5,
				UnitAmount:   shared.NewNetAmount(decimal.NewFromFloat(3.00)),
			},
		},
	})
	require.NoError(t, err)

	doc, err := builder.Build(docNumber)
	require.NoError(t, err)

	// Cheeseburger: 2 * 10.00 = 20.00 net, 23% VAT = 4.60 tax
	// Mineral Water: 5 * 3.00 = 15.00 net, 8% VAT = 1.20 tax
	assertDecimalsEqual(t, decimal.NewFromFloat(35.00), doc.Summary().NetAmount(), "total net amount")
	assertDecimalsEqual(t, decimal.NewFromFloat(5.80), doc.Summary().TaxAmount(), "total tax amount")
	assertDecimalsEqual(t, decimal.NewFromFloat(40.80), doc.Summary().GrossAmount(), "total gross amount")

	require.Len(t, doc.LineItems(), 2)

	assertDecimalsEqual(t, decimal.NewFromFloat(20.00), doc.LineItems()[0].PriceBreakdown().NetAmount(), "cheeseburger net amount")
	assertDecimalsEqual(t, decimal.NewFromFloat(4.60), doc.LineItems()[0].PriceBreakdown().TaxAmount(), "cheeseburger tax amount")
	assertDecimalsEqual(t, decimal.NewFromFloat(24.60), doc.LineItems()[0].PriceBreakdown().GrossAmount(), "cheeseburger gross amount")

	assertDecimalsEqual(t, decimal.NewFromFloat(15.00), doc.LineItems()[1].PriceBreakdown().NetAmount(), "mineral water net amount")
	assertDecimalsEqual(t, decimal.NewFromFloat(1.20), doc.LineItems()[1].PriceBreakdown().TaxAmount(), "mineral water tax amount")
	assertDecimalsEqual(t, decimal.NewFromFloat(16.20), doc.LineItems()[1].PriceBreakdown().GrossAmount(), "mineral water gross amount")

	require.Len(t, doc.Summary().Taxes(), 2)

	vat8 := findTax(t, doc.Summary().Taxes(), domain.UnmarshalTaxRate(decimal.NewFromFloat(0.08), domain.TaxTypeVAT))
	vat23 := findTax(t, doc.Summary().Taxes(), domain.UnmarshalTaxRate(decimal.NewFromFloat(0.23), domain.TaxTypeVAT))

	assertDecimalsEqual(t, decimal.NewFromFloat(15.00), vat8.NetAmount(), "VAT 8% net amount")
	assertDecimalsEqual(t, decimal.NewFromFloat(1.20), vat8.TaxAmount(), "VAT 8% tax amount")

	assertDecimalsEqual(t, decimal.NewFromFloat(20.00), vat23.NetAmount(), "VAT 23% net amount")
	assertDecimalsEqual(t, decimal.NewFromFloat(4.60), vat23.TaxAmount(), "VAT 23% tax amount")
}

func findTax(t *testing.T, taxes []domain.TaxSummary, rate domain.TaxRate) domain.TaxSummary {
	for _, tt := range taxes {
		if tt.TaxRate().Equal(rate) {
			return tt
		}
	}

	t.Errorf("findTax: tax rate not found in tax summary")

	return domain.TaxSummary{}
}

func TestNewReceipt_GrossPriceBreakdown(t *testing.T) {
	factory := domain.NewDocumentFactory(stubTaxProvider{})
	data := validReceiptData(t)
	data.LineItems = []domain.NewLineItemData{
		{Name: "Food qty=1", LineItemType: shared.LineItemTypeFood, Quantity: 1, UnitAmount: shared.NewGrossAmount(decimal.RequireFromString("25.00"))},
		{Name: "Food qty=3", LineItemType: shared.LineItemTypeFood, Quantity: 3, UnitAmount: shared.NewGrossAmount(decimal.RequireFromString("25.00"))},
		{Name: "Cheap Item", LineItemType: shared.LineItemTypeFood, Quantity: 1, UnitAmount: shared.NewGrossAmount(decimal.RequireFromString("1.00"))},
		{Name: "Delivery", LineItemType: shared.LineItemTypeDelivery, Quantity: 2, UnitAmount: shared.NewGrossAmount(decimal.RequireFromString("15.00"))},
	}

	builder, err := factory.NewReceiptBuilder(context.Background(), data)
	require.NoError(t, err)

	doc, err := builder.Build(newTestDocumentNumber(t))
	require.NoError(t, err)

	require.Len(t, doc.LineItems(), 4)

	// Food qty=1: 25.00 gross, 23% VAT
	assertDecimalsEqual(t, decimal.NewFromFloat(20.33), doc.LineItems()[0].PriceBreakdown().NetAmount(), "food qty=1 net")
	assertDecimalsEqual(t, decimal.NewFromFloat(4.67), doc.LineItems()[0].PriceBreakdown().TaxAmount(), "food qty=1 tax")
	assertDecimalsEqual(t, decimal.NewFromFloat(25.00), doc.LineItems()[0].PriceBreakdown().GrossAmount(), "food qty=1 gross")
	assertLineReconciles(t, doc.LineItems()[0])

	// Food qty=3: 3 * 25.00 = 75.00 gross, 23% VAT
	assertDecimalsEqual(t, decimal.NewFromFloat(60.99), doc.LineItems()[1].PriceBreakdown().NetAmount(), "food qty=3 net")
	assertDecimalsEqual(t, decimal.NewFromFloat(14.01), doc.LineItems()[1].PriceBreakdown().TaxAmount(), "food qty=3 tax")
	assertDecimalsEqual(t, decimal.NewFromFloat(75.00), doc.LineItems()[1].PriceBreakdown().GrossAmount(), "food qty=3 gross")
	assertLineReconciles(t, doc.LineItems()[1])

	// Cheap Item: 1.00 gross, 23% VAT (rounding edge case)
	assertDecimalsEqual(t, decimal.RequireFromString("0.81"), doc.LineItems()[2].PriceBreakdown().NetAmount(), "cheap net")
	assertDecimalsEqual(t, decimal.RequireFromString("0.19"), doc.LineItems()[2].PriceBreakdown().TaxAmount(), "cheap tax")
	assertDecimalsEqual(t, decimal.RequireFromString("1.00"), doc.LineItems()[2].PriceBreakdown().GrossAmount(), "cheap gross")
	assertLineReconciles(t, doc.LineItems()[2])

	// Delivery qty=2: 2 * 15.00 = 30.00 gross, 10% GST
	assertDecimalsEqual(t, decimal.RequireFromString("27.28"), doc.LineItems()[3].PriceBreakdown().NetAmount(), "delivery net")
	assertDecimalsEqual(t, decimal.RequireFromString("2.72"), doc.LineItems()[3].PriceBreakdown().TaxAmount(), "delivery tax")
	assertDecimalsEqual(t, decimal.RequireFromString("30.00"), doc.LineItems()[3].PriceBreakdown().GrossAmount(), "delivery gross")
	assertLineReconciles(t, doc.LineItems()[3])
}

func assertLineReconciles(t *testing.T, li domain.LineItem) {
	sum := li.PriceBreakdown().NetAmount().Add(li.PriceBreakdown().TaxAmount())
	assert.True(t, sum.Equal(li.PriceBreakdown().GrossAmount()),
		"net %s + tax %s = %s, expected gross %s",
		li.PriceBreakdown().NetAmount(), li.PriceBreakdown().TaxAmount(), sum, li.PriceBreakdown().GrossAmount())
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
				LineItemType: shared.LineItemTypeFood,
				Name:         "Cheeseburger",
				Quantity:     2,
				UnitAmount:   shared.NewNetAmount(decimal.NewFromFloat(10.00)),
			},
			{
				LineItemType: shared.LineItemTypeFood,
				Name:         "Fries",
				Quantity:     1,
				UnitAmount:   shared.NewNetAmount(decimal.NewFromFloat(3.50)),
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

type stubTaxProvider struct{}

func (s stubTaxProvider) GetTaxRate(ctx context.Context, input domain.TaxRateRequest) (domain.TaxRate, error) {
	switch input.LineItemType {
	case shared.LineItemTypeFood:
		return domain.NewTaxRate(decimal.NewFromFloat(0.23), domain.TaxTypeVAT)
	case shared.LineItemTypeBeverage:
		return domain.NewTaxRate(decimal.NewFromFloat(0.08), domain.TaxTypeVAT)
	case shared.LineItemTypeDelivery:
		return domain.NewTaxRate(decimal.NewFromFloat(0.10), domain.TaxTypeGST)
	default:
		return domain.NewTaxRate(decimal.Zero, domain.TaxTypeSalesTax)
	}
}
