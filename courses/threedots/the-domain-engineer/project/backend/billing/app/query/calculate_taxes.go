package query

import (
	"context"
	"fmt"

	"eats/backend/billing/api/module/client"
	"eats/backend/billing/domain"
)

func (h *Handlers) CalculateTaxes(ctx context.Context, query client.CalculateTaxesRequest) (client.CalculateTaxesResponse, error) {
	lineItems := make([]domain.NewLineItemData, 0, len(query.LineItems))
	for _, li := range query.LineItems {
		lineItem := domain.NewLineItemData{
			Name:         li.Name,
			LineItemType: li.Type,
			Quantity:     li.Quantity,
			UnitAmount:   li.UnitAmount,
		}
		lineItems = append(lineItems, lineItem)
	}

	taxCalc, err := h.documentFactory.NewTaxCalculation(ctx, domain.TaxCalculationInput{
		BuyerCountryCode:  query.BuyerCountryCode,
		BuyerTaxID:        query.BuyerTaxID,
		SellerCountryCode: query.SellerCountryCode,
		LineItems:         lineItems,
		Currency:          query.Currency,
	})
	if err != nil {
		return client.CalculateTaxesResponse{}, fmt.Errorf("tax calculation failed: %w", err)
	}

	respItems := make([]client.LineItemReadModel, 0, len(taxCalc.LineItems()))
	for _, lineItem := range taxCalc.LineItems() {
		respItem := client.LineItemReadModel{
			Name:        lineItem.Name(),
			Type:        lineItem.LineItemType(),
			Quantity:    lineItem.Quantity(),
			NetAmount:   lineItem.PriceBreakdown().NetAmount(),
			TaxAmount:   lineItem.PriceBreakdown().TaxAmount(),
			GrossAmount: lineItem.PriceBreakdown().GrossAmount(),
		}
		respItems = append(respItems, respItem)
	}

	return client.CalculateTaxesResponse{
		LineItems:  respItems,
		NetTotal:   taxCalc.Summary().NetAmount(),
		TaxTotal:   taxCalc.Summary().TaxAmount(),
		GrossTotal: taxCalc.Summary().GrossAmount(),
	}, nil
}
