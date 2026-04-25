package app

import (
	"context"

	"eats/backend/delivery/api/module/client"
)

type ModulesContract interface {
	CalculateDeliveryFee(ctx context.Context, req client.CalculateDeliveryFeeRequest) (client.CalculateDeliveryFeeResponse, error)
}

type Service struct {
	restaurantRepository RestaurantRepository
	customerRepository   CustomerRepository
	orderRepository      OrderRepository
	modules              ModulesContract
}

func NewService(
	restaurantRepository RestaurantRepository,
	customerRepository CustomerRepository,
	orderRepository OrderRepository,
	modules ModulesContract,
) *Service {
	if restaurantRepository == nil {
		panic("restaurantRepository cannot be nil")
	}
	if customerRepository == nil {
		panic("customerRepository cannot be nil")
	}
	if orderRepository == nil {
		panic("orderRepository cannot be nil")
	}
	if modules == nil {
		panic("modules cannot be nil")
	}

	return &Service{
		restaurantRepository: restaurantRepository,
		customerRepository:   customerRepository,
		orderRepository:      orderRepository,
		modules:              modules,
	}
}
