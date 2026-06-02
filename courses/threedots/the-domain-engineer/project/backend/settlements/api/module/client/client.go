package client

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"eats/backend/common"
	"eats/backend/common/shared"
)

type Settlements interface {
	StartSettlement(ctx context.Context, cmd StartSettlementRequest) error
	GetPlatformEntity(ctx context.Context, req GetPlatformEntityRequest) (GetPlatformEntityResponse, error)
}

type GetPlatformEntityRequest struct {
	PartnerUUID common.UUID
}

type GetPlatformEntityResponse struct {
	PlatformUUID common.UUID
}

type StartSettlementRequest struct {
	OrderUUID      common.UUID
	RestaurantUUID common.UUID
	CourierUUID    common.UUID
	Currency       shared.Currency

	CustomerName    string
	CustomerAddress shared.Address

	LineItems []LineItem

	TotalAmount decimal.Decimal

	OrderedAt time.Time
}

func (r StartSettlementRequest) Validate() error {
	if r.OrderUUID.IsZero() {
		return common.NewInvalidInputError("order-uuid-empty", "order UUID cannot be empty")
	}

	if r.RestaurantUUID.IsZero() {
		return common.NewInvalidInputError("restaurant-uuid-empty", "restaurant UUID cannot be empty")
	}

	if r.CourierUUID.IsZero() {
		return common.NewInvalidInputError("courier-uuid-empty", "courier UUID cannot be empty")
	}

	if len(r.LineItems) == 0 {
		return common.NewInvalidInputError("line-items-empty", "at least one line item is required")
	}

	return nil
}

type LineItem struct {
	Name        string
	Type        shared.LineItemType
	Quantity    int
	GrossAmount decimal.Decimal
}
