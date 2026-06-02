// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package printer_test

import (
	"context"
	_ "embed"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"eats/backend/billing/adapters/printer"
	"eats/backend/billing/adapters/tax"
	"eats/backend/billing/domain"
	"eats/backend/common"
	"eats/backend/common/shared"
)

//go:embed testdata/receipt_golden.html
var receiptGoldenHTML string

func TestPrintReceipt_Golden(t *testing.T) {
	printer := printer.NewPrinter()

	taxProvider := tax.NewConfiguredStub(map[shared.LineItemType]domain.TaxRate{
		shared.LineItemTypeFood:     domain.UnmarshalTaxRate(decimal.NewFromFloat(0.23), domain.TaxTypeVAT),
		shared.LineItemTypeBeverage: domain.UnmarshalTaxRate(decimal.NewFromFloat(0.10), domain.TaxTypeVAT),
		shared.LineItemTypeDelivery: domain.UnmarshalTaxRate(decimal.NewFromFloat(0.05), domain.TaxTypeSalesTax),
	})

	factory := domain.NewDocumentFactory(taxProvider)

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
	docNumber, err := domain.NewDocumentNumber(series, 1)
	require.NoError(t, err)

	builder, err := factory.NewReceiptBuilder(context.Background(), domain.NewDocumentData{
		ExternalReference: common.ToPtr("ORDER-456"),
		IssueDate:         time.Date(2025, 3, 14, 0, 0, 0, 0, time.UTC),
		Currency:          shared.MustNewCurrency("USD"),
		Seller:            seller,
		Buyer:             buyer,
		LineItems: []domain.NewLineItemData{
			{
				Name:         "Pepperoni Pizza",
				LineItemType: shared.LineItemTypeFood,
				Quantity:     1,
				UnitAmount:   shared.NewGrossAmount(decimal.NewFromFloat(15.99)),
			},
			{
				Name:         "Garlic Bread",
				LineItemType: shared.LineItemTypeFood,
				Quantity:     2,
				UnitAmount:   shared.NewGrossAmount(decimal.NewFromFloat(2.50)),
			},
			{
				Name:         "Delivery",
				LineItemType: shared.LineItemTypeDelivery,
				Quantity:     1,
				UnitAmount:   shared.NewGrossAmount(decimal.NewFromFloat(5.00)),
			},
		},
	})
	require.NoError(t, err)

	doc, err := builder.Build(docNumber)
	require.NoError(t, err)

	output, err := printer.PrintDocument(context.Background(), doc)
	require.NoError(t, err)

	if os.Getenv("UPDATE_GOLDEN_TESTS") == "1" {
		err = os.MkdirAll("testdata", 0o755)
		require.NoError(t, err)

		err = os.WriteFile("testdata/receipt_golden.html", output, 0o644)
		require.NoError(t, err)
		return
	}

	if diff := cmp.Diff(strings.Split(receiptGoldenHTML, "\n"), strings.Split(string(output), "\n")); diff != "" {
		t.Errorf("golden file mismatch (-want +got):\n%s", diff)
	}
}
