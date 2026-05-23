package domain

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"eats/backend/common"
	"eats/backend/common/shared"
)

type TaxType struct {
	common.Enum[TaxTypeValues]
}

func (t TaxType) DisplayName() string {
	switch t.String() {
	case "vat":
		return "VAT"
	case "gst":
		return "GST"
	case "sales-tax":
		return "Sales Tax"
	default:
		return "Unknown"
	}
}

type TaxTypeValues string

func (TaxTypeValues) Values() []string {
	return []string{"vat", "gst", "sales-tax"}
}

var (
	TaxTypeVAT      = common.MustEnum[TaxType]("vat")
	TaxTypeGST      = common.MustEnum[TaxType]("gst")
	TaxTypeSalesTax = common.MustEnum[TaxType]("sales-tax")
)

type TaxRateRequest struct {
	BuyerCountryCode  shared.CountryCode
	BuyerTaxID        *shared.TaxID
	SellerCountryCode shared.CountryCode
	LineItemType      shared.LineItemType
	TransactionDate   time.Time
}

type TaxRateProvider interface {
	// In a real system, a single line item could have multiple tax rates applied to it,
	// e.g. in US, where state and local taxes apply.
	// For simplicity, we assume a single tax rate per line item here.
	GetTaxRate(ctx context.Context, input TaxRateRequest) (TaxRate, error)
}

type PriceBreakdown struct {
	rate TaxRate

	unitNetAmount   decimal.Decimal
	unitTaxAmount   decimal.Decimal
	unitGrossAmount decimal.Decimal

	netAmount   decimal.Decimal
	taxAmount   decimal.Decimal
	grossAmount decimal.Decimal
}

func NewPriceBreakdownFromNetAmount(
	rate TaxRate,
	unitNetAmount decimal.Decimal,
	currency shared.Currency,
	quantity int,
) (PriceBreakdown, error) {
	if rate.IsZero() {
		return PriceBreakdown{}, common.NewInvalidInputError("tax-rate-zero", "tax rate cannot be empty")
	}
	if unitNetAmount.IsNegative() {
		return PriceBreakdown{}, common.NewInvalidInputError("unit-net-amount-negative", "unit net amount cannot be negative")
	}
	if quantity <= 0 {
		return PriceBreakdown{}, common.NewInvalidInputError("quantity-not-positive", "quantity should be positive")
	}

	// Round just in case to avoid issues with repeating decimals
	unitNetAmount = roundInCurrency(unitNetAmount, currency)

	// Critical part: this is the only place where the rounding happens
	unitTaxAmount := roundInCurrency(unitNetAmount.Mul(rate.Rate()), currency)
	// No need to round again as we already work with rounded numbers
	unitGrossAmount := unitNetAmount.Add(unitTaxAmount)

	quantityDecimal := decimal.NewFromInt(int64(quantity))

	// We already work with rounded numbers at this point
	netAmount := unitNetAmount.Mul(quantityDecimal)
	taxAmount := unitTaxAmount.Mul(quantityDecimal)
	grossAmount := unitGrossAmount.Mul(quantityDecimal)

	return PriceBreakdown{
		rate:            rate,
		unitNetAmount:   unitNetAmount,
		unitTaxAmount:   unitTaxAmount,
		unitGrossAmount: unitGrossAmount,
		netAmount:       netAmount,
		taxAmount:       taxAmount,
		grossAmount:     grossAmount,
	}, nil
}

func NewPriceBreakdownFromGrossAmount(
	rate TaxRate,
	unitGrossAmount decimal.Decimal,
	currency shared.Currency,
	quantity int,
) (PriceBreakdown, error) {
	if rate.IsZero() {
		return PriceBreakdown{}, common.NewInvalidInputError("tax-rate-zero", "tax rate cannot be empty")
	}
	if unitGrossAmount.IsNegative() {
		return PriceBreakdown{}, common.NewInvalidInputError("unit-gross-amount-negative", "unit gross amount cannot be negative")
	}
	if quantity <= 0 {
		return PriceBreakdown{}, common.NewInvalidInputError("quantity-not-positive", "quantity should be positive")
	}

	// Round just in case to avoid issues with repeating decimals
	unitGrossAmount = roundInCurrency(unitGrossAmount, currency)

	// Critical part: this is the only place where the rounding happens
	// From now on, we operate on rounded numbers to avoid rounding issues later
	unitNetAmount := roundInCurrency(unitGrossAmount.Div(decimal.NewFromInt(1).Add(rate.Rate())), currency)

	// Gross price is the source of truth and we want line items gross price to add up to the correct total
	// so we calculate tax as the difference between gross and net
	unitTaxAmount := roundInCurrency(unitGrossAmount.Sub(unitNetAmount), currency)

	quantityDecimal := decimal.NewFromInt(int64(quantity))

	// We already work with rounded numbers at this point
	grossAmount := unitGrossAmount.Mul(quantityDecimal)
	netAmount := unitNetAmount.Mul(quantityDecimal)
	taxAmount := unitTaxAmount.Mul(quantityDecimal)

	return PriceBreakdown{
		rate:            rate,
		unitNetAmount:   unitNetAmount,
		unitTaxAmount:   unitTaxAmount,
		unitGrossAmount: unitGrossAmount,
		netAmount:       netAmount,
		taxAmount:       taxAmount,
		grossAmount:     grossAmount,
	}, nil
}

func (t PriceBreakdown) TaxRate() TaxRate {
	return t.rate
}

func (t PriceBreakdown) UnitNetAmount() decimal.Decimal {
	return t.unitNetAmount
}

func (t PriceBreakdown) UnitTaxAmount() decimal.Decimal {
	return t.unitTaxAmount
}

func (t PriceBreakdown) UnitGrossAmount() decimal.Decimal {
	return t.unitGrossAmount
}

func (t PriceBreakdown) NetAmount() decimal.Decimal {
	return t.netAmount
}

func (t PriceBreakdown) TaxAmount() decimal.Decimal {
	return t.taxAmount
}

func (t PriceBreakdown) GrossAmount() decimal.Decimal {
	return t.grossAmount
}

type TaxRate struct {
	rate    decimal.Decimal
	taxType TaxType
}

func NewTaxRate(rate decimal.Decimal, taxType TaxType) (TaxRate, error) {
	if rate.IsNegative() {
		return TaxRate{}, common.NewInvalidInputError("tax-rate-negative", "tax rate cannot be negative")
	}
	if taxType.IsZero() {
		return TaxRate{}, common.NewInvalidInputError("tax-type-zero", "tax type cannot be zero value")
	}

	return TaxRate{
		rate:    rate,
		taxType: taxType,
	}, nil
}

func (t TaxRate) IsZero() bool {
	return t.taxType.IsZero()
}

func (t TaxRate) Equal(other TaxRate) bool {
	return t.rate.Equal(other.rate) && t.taxType.String() == other.taxType.String()
}

func (t TaxRate) Rate() decimal.Decimal {
	return t.rate
}

func (t TaxRate) TaxType() TaxType {
	return t.taxType
}

func (t TaxRate) key() taxRateKey {
	return taxRateKey{
		rate:    t.rate.String(),
		taxType: t.taxType.String(),
	}
}

// roundInCurrency rounds the given amount according to the decimal places of the currency.
// Different currencies have different rules for decimal places (e.g., JPY has 0, USD has 2).
func roundInCurrency(amount decimal.Decimal, currency shared.Currency) decimal.Decimal {
	return amount.Round(int32(currency.DecimalPlaces()))
}
