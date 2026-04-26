package module

import (
	"context"
	"fmt"

	"eats/backend/delivery/api/module/client"
	"eats/backend/delivery/app"
)

type Delivery struct {
	service *app.Service
}

func New(service *app.Service) *Delivery {
	if service == nil {
		panic("service cannot be nil")
	}

	return &Delivery{service: service}
}

func (i Delivery) CalculateDeliveryFee(ctx context.Context, req client.CalculateDeliveryFeeRequest) (client.CalculateDeliveryFeeResponse, error) {
	fee, err := i.service.CalculateDeliveryFee(
		ctx,
		req.RestaurantAddress,
		req.DeliveryAddress,
		req.Currency,
		req.When,
	)
	if err != nil {
		return client.CalculateDeliveryFeeResponse{}, fmt.Errorf("failed to calculate delivery fee: %w", err)
	}

	return client.CalculateDeliveryFeeResponse{
		GrossFee: fee,
	}, nil
}
