package tax

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	"eats/backend/billing/domain"
	"eats/backend/common/shared"
)

// StubClient is a test double for the tax rate provider.
// It always returns a fixed 23% VAT rate unless configured otherwise.
type StubClient struct {
	rates map[shared.LineItemType]domain.TaxRate
}

func NewStub() *StubClient {
	return &StubClient{}
}

func NewConfiguredStub(rates map[shared.LineItemType]domain.TaxRate) *StubClient {
	return &StubClient{
		rates: rates,
	}
}

func (s *StubClient) GetTaxRate(_ context.Context, input domain.TaxRateRequest) (domain.TaxRate, error) {
	if len(s.rates) == 0 {
		return domain.UnmarshalTaxRate(decimal.NewFromFloat(0.23), domain.TaxTypeVAT), nil
	}

	rate, ok := s.rates[input.LineItemType]
	if !ok {
		return domain.TaxRate{}, fmt.Errorf("no tax rate configured for line item type %s", input.LineItemType)
	}

	return rate, nil
}
