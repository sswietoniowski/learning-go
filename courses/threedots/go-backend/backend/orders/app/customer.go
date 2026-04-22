package app

import (
	"context"

	"eats/backend/common"
	"eats/backend/common/shared"
)

type CustomerUUID struct {
	common.UUID
}

type Customer struct {
	CustomerUUID CustomerUUID
	Name         string
	Email        string
	Address      shared.Address
	PhoneNumber  string
}

type CustomerRepository interface {
	RegisterCustomer(ctx context.Context, customer Customer) error
}

func (s *Service) RegisterCustomer(ctx context.Context, customer Customer) error {
	return s.customerRepository.RegisterCustomer(ctx, customer)
}
