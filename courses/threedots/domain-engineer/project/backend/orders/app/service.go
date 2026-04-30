package app

import (
	"context"

	"github.com/shopspring/decimal"

	"eats/backend/delivery/api/module/client"
)

type ModulesContract interface {
	CalculateDeliveryFee(ctx context.Context, req client.CalculateDeliveryFeeRequest) (client.CalculateDeliveryFeeResponse, error)
}

type PaymentsService interface {
	CapturePayment(ctx context.Context, nonce string, amount decimal.Decimal, merchantID string) error
}

type Service struct {
	restaurantRepository RestaurantRepository
	customerRepository   CustomerRepository
	orderRepository      OrderRepository
	courierRepository    CourierRepository
	paymentsService      PaymentsService
	modules              ModulesContract
}

func NewService(
	restaurantRepository RestaurantRepository,
	customerRepository CustomerRepository,
	orderRepository OrderRepository,
	courierRepository CourierRepository,
	paymentsService PaymentsService,
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
	if courierRepository == nil {
		panic("courierRepository cannot be nil")
	}
	if paymentsService == nil {
		panic("paymentsService cannot be nil")
	}
	if modules == nil {
		panic("modules cannot be nil")
	}

	return &Service{
		restaurantRepository: restaurantRepository,
		customerRepository:   customerRepository,
		orderRepository:      orderRepository,
		courierRepository:    courierRepository,
		paymentsService:      paymentsService,
		modules:              modules,
	}
}
