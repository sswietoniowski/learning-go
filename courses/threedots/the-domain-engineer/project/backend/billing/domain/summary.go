package domain

import "github.com/shopspring/decimal"

type taxRateKey struct {
	rate    string
	taxType string
}

type PriceBreakdownSummary struct {
	netAmount   decimal.Decimal
	taxAmount   decimal.Decimal
	grossAmount decimal.Decimal
	taxes       []TaxSummary
}

func summarizeLineItems(lineItems []LineItem) PriceBreakdownSummary {
	netAmount := decimal.Zero
	taxAmount := decimal.Zero
	grossAmount := decimal.Zero

	for _, l := range lineItems {
		price := l.PriceBreakdown()
		netAmount = netAmount.Add(price.NetAmount())
		taxAmount = taxAmount.Add(price.TaxAmount())
		grossAmount = grossAmount.Add(price.GrossAmount())
	}

	return PriceBreakdownSummary{
		netAmount:   netAmount,
		taxAmount:   taxAmount,
		grossAmount: grossAmount,
		taxes:       newGroupedTaxes(lineItems),
	}
}

func (p PriceBreakdownSummary) NetAmount() decimal.Decimal {
	return p.netAmount
}

func (p PriceBreakdownSummary) TaxAmount() decimal.Decimal {
	return p.taxAmount
}

func (p PriceBreakdownSummary) GrossAmount() decimal.Decimal {
	return p.grossAmount
}

func (p PriceBreakdownSummary) Taxes() []TaxSummary {
	return p.taxes
}

type TaxSummary struct {
	taxRate   TaxRate
	netAmount decimal.Decimal
	taxAmount decimal.Decimal
}

func newTaxSummary(price PriceBreakdown) *TaxSummary {
	return &TaxSummary{
		taxRate:   price.TaxRate(),
		netAmount: price.NetAmount(),
		taxAmount: price.TaxAmount(),
	}
}

func (t TaxSummary) TaxRate() TaxRate {
	return t.taxRate
}

func (t TaxSummary) NetAmount() decimal.Decimal {
	return t.netAmount
}

func (t TaxSummary) TaxAmount() decimal.Decimal {
	return t.taxAmount
}

// TaxSummary is immutable except for add.
// Note that it's unexported, so it can't be modified from outside the package.
func (t *TaxSummary) add(price PriceBreakdown) {
	t.netAmount = t.netAmount.Add(price.NetAmount())
	t.taxAmount = t.taxAmount.Add(price.TaxAmount())
}

func newGroupedTaxes(lineItems []LineItem) []TaxSummary {
	taxes := map[taxRateKey]*TaxSummary{}

	for _, l := range lineItems {
		price := l.PriceBreakdown()
		key := price.rate.key()

		summary, ok := taxes[key]
		if ok {
			summary.add(price)
		} else {
			taxes[key] = newTaxSummary(price)
		}
	}

	keys := make([]taxRateKey, 0, len(taxes))
	for k := range taxes {
		keys = append(keys, k)
	}

	result := make([]TaxSummary, 0, len(taxes))
	for _, k := range keys {
		result = append(result, *taxes[k])
	}

	return result
}
