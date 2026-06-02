// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package domain_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eats/backend/billing/domain"
	"eats/backend/common/shared"
)

func TestNewPriceBreakdownFromGrossAmount(t *testing.T) {
	vatRate, err := domain.NewTaxRate(decimal.NewFromFloat(0.21), domain.TaxTypeVAT)
	require.NoError(t, err)

	tests := []struct {
		name            string
		rate            domain.TaxRate
		unitGrossAmount string
		quantity        int
		wantErr         string
		wantUnitNet     string
		wantUnitTax     string
		wantUnitGross   string
		wantNet         string
		wantTax         string
		wantGross       string
	}{
		{
			name:            "single item",
			rate:            vatRate,
			unitGrossAmount: "16.03",
			quantity:        1,
			wantUnitNet:     "13.25",
			wantUnitTax:     "2.78",
			wantUnitGross:   "16.03",
			wantNet:         "13.25",
			wantTax:         "2.78",
			wantGross:       "16.03",
		},
		{
			name:            "multiple items - unit price times quantity",
			rate:            vatRate,
			unitGrossAmount: "16.03",
			quantity:        4,
			wantUnitNet:     "13.25",
			wantUnitTax:     "2.78",
			wantUnitGross:   "16.03",
			wantNet:         "53.00",
			wantTax:         "11.12",
			wantGross:       "64.12",
		},
		{
			name:            "rounding on VAT",
			rate:            vatRate,
			unitGrossAmount: "16.03",
			quantity:        4,
			wantUnitNet:     "13.25",
			wantUnitTax:     "2.78",
			wantUnitGross:   "16.03",
			wantNet:         "53.00",
			wantTax:         "11.12",
			wantGross:       "64.12",
		},
		{
			name:            "rounding needed on unit net",
			rate:            vatRate,
			unitGrossAmount: "1.00",
			quantity:        1,
			wantUnitNet:     "0.83",
			wantUnitTax:     "0.17",
			wantUnitGross:   "1.00",
			wantNet:         "0.83",
			wantTax:         "0.17",
			wantGross:       "1.00",
		},
		{
			name:            "total tax derived from gross minus net",
			rate:            vatRate,
			unitGrossAmount: "1.00",
			quantity:        3,
			wantUnitNet:     "0.83",
			wantUnitTax:     "0.17",
			wantUnitGross:   "1.00",
			wantNet:         "2.49",
			wantTax:         "0.51",
			wantGross:       "3.00",
		},
		{
			name:            "zero gross amount",
			rate:            vatRate,
			unitGrossAmount: "0.00",
			quantity:        1,
			wantUnitNet:     "0.00",
			wantUnitTax:     "0.00",
			wantUnitGross:   "0.00",
			wantNet:         "0.00",
			wantTax:         "0.00",
			wantGross:       "0.00",
		},
		{
			name:            "large quantity",
			rate:            vatRate,
			unitGrossAmount: "9.99",
			quantity:        100,
			wantUnitNet:     "8.26",
			wantUnitTax:     "1.73",
			wantUnitGross:   "9.99",
			wantNet:         "826.00",
			wantTax:         "173.00",
			wantGross:       "999.00",
		},
		{
			name:            "negative gross amount",
			rate:            vatRate,
			unitGrossAmount: "-10.00",
			quantity:        1,
			wantErr:         "unit-gross-amount-negative",
		},
		{
			name:            "zero quantity",
			rate:            vatRate,
			unitGrossAmount: "10.00",
			quantity:        0,
			wantErr:         "quantity-not-positive",
		},
		{
			name:            "negative quantity",
			rate:            vatRate,
			unitGrossAmount: "10.00",
			quantity:        -1,
			wantErr:         "quantity-not-positive",
		},
		{
			name:            "zero tax rate",
			rate:            domain.TaxRate{},
			unitGrossAmount: "10.00",
			quantity:        1,
			wantErr:         "tax-rate-zero",
		},
	}

	currency := shared.MustNewCurrency("USD")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unitGross := decimal.RequireFromString(tt.unitGrossAmount)

			got, err := domain.NewPriceBreakdownFromGrossAmount(tt.rate, unitGross, currency, tt.quantity)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)

			wantUnitNet := decimal.RequireFromString(tt.wantUnitNet)
			wantUnitTax := decimal.RequireFromString(tt.wantUnitTax)
			wantUnitGross := decimal.RequireFromString(tt.wantUnitGross)
			wantNet := decimal.RequireFromString(tt.wantNet)
			wantTax := decimal.RequireFromString(tt.wantTax)
			wantGross := decimal.RequireFromString(tt.wantGross)

			assertDecimalsEqual(t, wantUnitNet, got.UnitNetAmount(), "UnitNetAmount")
			assertDecimalsEqual(t, wantUnitTax, got.UnitTaxAmount(), "UnitTaxAmount")
			assertDecimalsEqual(t, wantUnitGross, got.UnitGrossAmount(), "UnitGrossAmount")

			assertDecimalsEqual(t, wantNet, got.NetAmount(), "NetAmount")
			assertDecimalsEqual(t, wantTax, got.TaxAmount(), "TaxAmount")
			assertDecimalsEqual(t, wantGross, got.GrossAmount(), "GrossAmount")

			// Invariant: totals must be consistent
			assert.True(t, got.GrossAmount().Equal(got.NetAmount().Add(got.TaxAmount())),
				"Invariant violated: GrossAmount (%s) != NetAmount (%s) + TaxAmount (%s)",
				got.GrossAmount(), got.NetAmount(), got.TaxAmount())
		})
	}
}

func TestNewPriceBreakdownFromNetAmount(t *testing.T) {
	vatRate, err := domain.NewTaxRate(decimal.NewFromFloat(0.21), domain.TaxTypeVAT)
	require.NoError(t, err)

	tests := []struct {
		name          string
		rate          domain.TaxRate
		unitNetAmount string
		quantity      int
		wantErr       string
		wantUnitNet   string
		wantUnitTax   string
		wantUnitGross string
		wantNet       string
		wantTax       string
		wantGross     string
	}{
		{
			name:          "single item",
			rate:          vatRate,
			unitNetAmount: "13.25",
			quantity:      1,
			wantUnitNet:   "13.25",
			wantUnitTax:   "2.78",
			wantUnitGross: "16.03",
			wantNet:       "13.25",
			wantTax:       "2.78",
			wantGross:     "16.03",
		},
		{
			name:          "multiple items - unit price times quantity",
			rate:          vatRate,
			unitNetAmount: "13.25",
			quantity:      4,
			wantUnitNet:   "13.25",
			wantUnitTax:   "2.78",
			wantUnitGross: "16.03",
			wantNet:       "53.00",
			wantTax:       "11.12",
			wantGross:     "64.12",
		},
		{
			name:          "zero net amount",
			rate:          vatRate,
			unitNetAmount: "0.00",
			quantity:      1,
			wantUnitNet:   "0.00",
			wantUnitTax:   "0.00",
			wantUnitGross: "0.00",
			wantNet:       "0.00",
			wantTax:       "0.00",
			wantGross:     "0.00",
		},
		{
			name:          "negative net amount",
			rate:          vatRate,
			unitNetAmount: "-10.00",
			quantity:      1,
			wantErr:       "unit-net-amount-negative",
		},
		{
			name:          "zero quantity",
			rate:          vatRate,
			unitNetAmount: "10.00",
			quantity:      0,
			wantErr:       "quantity-not-positive",
		},
		{
			name:          "zero tax rate",
			rate:          domain.TaxRate{},
			unitNetAmount: "10.00",
			quantity:      1,
			wantErr:       "tax-rate-zero",
		},
	}

	currency := shared.MustNewCurrency("USD")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unitNet := decimal.RequireFromString(tt.unitNetAmount)

			got, err := domain.NewPriceBreakdownFromNetAmount(tt.rate, unitNet, currency, tt.quantity)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)

			assertDecimalsEqual(t, decimal.RequireFromString(tt.wantUnitNet), got.UnitNetAmount(), "UnitNetAmount")
			assertDecimalsEqual(t, decimal.RequireFromString(tt.wantUnitTax), got.UnitTaxAmount(), "UnitTaxAmount")
			assertDecimalsEqual(t, decimal.RequireFromString(tt.wantUnitGross), got.UnitGrossAmount(), "UnitGrossAmount")
			assertDecimalsEqual(t, decimal.RequireFromString(tt.wantNet), got.NetAmount(), "NetAmount")
			assertDecimalsEqual(t, decimal.RequireFromString(tt.wantTax), got.TaxAmount(), "TaxAmount")
			assertDecimalsEqual(t, decimal.RequireFromString(tt.wantGross), got.GrossAmount(), "GrossAmount")

			assert.True(t, got.GrossAmount().Equal(got.NetAmount().Add(got.TaxAmount())),
				"Invariant violated: GrossAmount (%s) != NetAmount (%s) + TaxAmount (%s)",
				got.GrossAmount(), got.NetAmount(), got.TaxAmount())
		})
	}
}

func TestNewTaxRate(t *testing.T) {
	tests := []struct {
		name    string
		rate    decimal.Decimal
		taxType domain.TaxType
		wantErr string
	}{
		{
			name:    "valid VAT rate",
			rate:    decimal.NewFromFloat(0.21),
			taxType: domain.TaxTypeVAT,
		},
		{
			name:    "zero rate allowed",
			rate:    decimal.Zero,
			taxType: domain.TaxTypeSalesTax,
		},
		{
			name:    "negative rate rejected",
			rate:    decimal.NewFromFloat(-0.01),
			taxType: domain.TaxTypeVAT,
			wantErr: "tax-rate-negative",
		},
		{
			name:    "zero tax type rejected",
			rate:    decimal.NewFromFloat(0.21),
			taxType: domain.TaxType{},
			wantErr: "tax-type-zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.NewTaxRate(tt.rate, tt.taxType)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.True(t, tt.rate.Equal(got.Rate()))
			assert.Equal(t, tt.taxType, got.TaxType())
		})
	}
}

func TestTaxRate_Equal(t *testing.T) {
	vat21, err := domain.NewTaxRate(decimal.NewFromFloat(0.21), domain.TaxTypeVAT)
	require.NoError(t, err)

	vat21Again, err := domain.NewTaxRate(decimal.NewFromFloat(0.21), domain.TaxTypeVAT)
	require.NoError(t, err)

	vat23, err := domain.NewTaxRate(decimal.NewFromFloat(0.23), domain.TaxTypeVAT)
	require.NoError(t, err)

	gst21, err := domain.NewTaxRate(decimal.NewFromFloat(0.21), domain.TaxTypeGST)
	require.NoError(t, err)

	assert.True(t, vat21.Equal(vat21Again), "same rate and type should be equal")
	assert.False(t, vat21.Equal(vat23), "different rate should not be equal")
	assert.False(t, vat21.Equal(gst21), "different tax type should not be equal")
}

func TestTaxRate_IsZero(t *testing.T) {
	assert.True(t, domain.TaxRate{}.IsZero(), "zero value is zero")

	nonZero, err := domain.NewTaxRate(decimal.NewFromFloat(0.21), domain.TaxTypeVAT)
	require.NoError(t, err)
	assert.False(t, nonZero.IsZero(), "constructed rate is not zero")
}

func TestTaxType_DisplayName(t *testing.T) {
	tests := []struct {
		taxType domain.TaxType
		want    string
	}{
		{domain.TaxTypeVAT, "VAT"},
		{domain.TaxTypeGST, "GST"},
		{domain.TaxTypeSalesTax, "Sales Tax"},
		{domain.TaxType{}, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.taxType.DisplayName())
		})
	}
}

func assertDecimalsEqual(t *testing.T, expected, actual decimal.Decimal, message string) {
	assert.True(t, expected.Equal(actual), "%v: expected %s to equal %s", message, actual.String(), expected.String())
}
