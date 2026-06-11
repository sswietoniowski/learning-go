package app

import (
	"github.com/shopspring/decimal"

	"eats/backend/common/shared"
	"eats/backend/settlements/domain"
)

type NewInvoiceData struct {
	ExternalReference string
	BuyerUUID         domain.LegalEntityUUID
	SellerUUID        domain.LegalEntityUUID
	LineItems         []NewInvoiceDataLineItem
}

type NewInvoiceDataLineItem struct {
	Name      string
	Type      shared.LineItemType
	Quantity  int
	NetAmount decimal.Decimal
}
