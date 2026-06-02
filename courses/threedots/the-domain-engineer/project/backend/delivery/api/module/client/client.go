package client

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"eats/backend/common/shared"
)

type Delivery interface {
	CalculateDeliveryFee(ctx context.Context, req CalculateDeliveryFeeRequest) (CalculateDeliveryFeeResponse, error)
}

type CalculateDeliveryFeeRequest struct {
	RestaurantAddress shared.Address
	DeliveryAddress   shared.Address
	Currency          shared.Currency
	When              time.Time
}

type CalculateDeliveryFeeResponse struct {
	GrossFee decimal.Decimal
}
