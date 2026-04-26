package app

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"eats/backend/common/shared"
)

func (s *Service) CalculateDeliveryFee(
	ctx context.Context,
	restaurantAddress shared.Address,
	deliveryAddress shared.Address,
	currency shared.Currency,
	when time.Time,
) (decimal.Decimal, error) {
	// In real world, here would be some complex logic to calculate delivery fee.
	return decimal.NewFromInt(10), nil
}
