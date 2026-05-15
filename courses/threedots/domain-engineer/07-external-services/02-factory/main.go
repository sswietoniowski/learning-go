package main

import (
	"errors"
	"fmt"
)

type PaymentDetails struct {
	Amount   int
	Currency string
}

type PaymentsClient interface {
	GetPaymentDetails(nonce string) (PaymentDetails, error)
}

type Order struct {
	nonce    string
	amount   int
	currency string
}

func (o *Order) Nonce() string    { return o.nonce }
func (o *Order) Amount() int      { return o.amount }
func (o *Order) Currency() string { return o.currency }

type OrderFactory struct {
	paymentsClient PaymentsClient
}

func NewOrderFactory(paymentsClient PaymentsClient) OrderFactory {
	if paymentsClient == nil {
		panic("payments client is required")
	}
	return OrderFactory{
		paymentsClient: paymentsClient,
	}
}

func (f OrderFactory) NewOrder(nonce string) (*Order, error) {
	if nonce == "" {
		return nil, errors.New("nonce is required")
	}

	details, err := f.paymentsClient.GetPaymentDetails(nonce)
	if err != nil {
		return nil, fmt.Errorf("getting payment details: %w", err)
	}

	if details.Amount <= 0 {
		return nil, errors.New("payment amount must be positive")
	}

	return &Order{
		nonce:    nonce,
		amount:   details.Amount,
		currency: details.Currency,
	}, nil
}
