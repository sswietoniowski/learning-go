package tax

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ThreeDotsLabs/the-domain-engineer/clients"
	"github.com/ThreeDotsLabs/the-domain-engineer/clients/tax"

	"eats/backend/billing/domain"
	"eats/backend/common"
	"eats/backend/common/shared"
)

type Client struct {
	clients *clients.Clients
}

func NewClient(clients *clients.Clients) *Client {
	if clients == nil {
		panic("nil clients")
	}
	return &Client{clients: clients}
}

func (c *Client) GetTaxRate(ctx context.Context, input domain.TaxRateRequest) (domain.TaxRate, error) {
	// We map the line item type to the external API's tax class explicitly.
	// This is an anti-corruption layer: internal and external types evolve independently.
	var taxClass tax.TaxRateRequestTaxClass
	switch input.LineItemType {
	case shared.LineItemTypeFood:
		taxClass = tax.TaxRateRequestTaxClassFOOD
	case shared.LineItemTypeBeverage:
		taxClass = tax.TaxRateRequestTaxClassBEVERAGE
	case shared.LineItemTypeDelivery:
		taxClass = tax.TaxRateRequestTaxClassDELIVERY
	case shared.LineItemTypeService:
		taxClass = tax.TaxRateRequestTaxClassSERVICE
	default:
		return domain.TaxRate{}, fmt.Errorf("unknown line item type: %s", input.LineItemType.String())
	}

	var buyerTaxID *string
	if input.BuyerTaxID != nil {
		buyerTaxID = common.ToPtr(input.BuyerTaxID.String())
	}

	resp, err := c.clients.Tax.GetTaxRateWithResponse(ctx, tax.GetTaxRateJSONRequestBody{
		BuyerCountryCode:  input.BuyerCountryCode.String(),
		BuyerTaxId:        buyerTaxID,
		SellerCountryCode: input.SellerCountryCode.String(),
		TaxClass:          &taxClass,
		TransactionDate:   input.TransactionDate,
	})
	if err != nil {
		return domain.TaxRate{}, fmt.Errorf("failed to get tax rate: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return domain.TaxRate{}, fmt.Errorf("failed to get tax rate: unexpected status %d", resp.StatusCode())
	}

	// We map the tax type explicitly, rather than unmarshalling it to the enum type.
	// It comes from an external system, and it uses slightly different values than our internal enum.
	var domainTax domain.TaxType
	switch resp.JSON200.TaxType {
	case tax.VAT:
		domainTax = domain.TaxTypeVAT
	case tax.GST:
		domainTax = domain.TaxTypeGST
	case tax.SALES:
		domainTax = domain.TaxTypeSalesTax
	default:
		return domain.TaxRate{}, fmt.Errorf("unknown tax type: %s", resp.JSON200.TaxType)
	}

	return domain.NewTaxRate(resp.JSON200.Rate, domainTax)
}
