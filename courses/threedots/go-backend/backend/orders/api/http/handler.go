package http

import (
	"context"

	"eats/backend/common"
)

type CustomerRepository interface {
	RegisterCustomer(ctx context.Context, customerUUID common.UUID, customer RegisterCustomer) error
}

type Handler struct {
	customerRepository CustomerRepository
}

func NewHandler(
	customerRepository CustomerRepository,
) Handler {
	if customerRepository == nil {
		panic("customerRepository cannot be nil")
	}

	return Handler{
		customerRepository: customerRepository,
	}
}

func (h Handler) RegisterCustomer(ctx context.Context, request RegisterCustomerRequestObject) (RegisterCustomerResponseObject, error) {
	customerUUID := common.NewUUIDv7()

	err := h.customerRepository.RegisterCustomer(ctx, customerUUID, *request.Body)

	if err != nil {
		return nil, err
	}

	return RegisterCustomer201JSONResponse{
		CustomerUuid: customerUUID,
	}, nil
}

func Register(ctx context.Context, e common.EchoRouter, handler Handler) error {
	RegisterHandlers(e, NewStrictHandler(handler, nil))

	return nil
}
