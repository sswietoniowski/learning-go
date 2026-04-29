package payments

import (
	"context"
	"fmt"
	"net/http"

	commonclients "github.com/ThreeDotsLabs/the-domain-engineer/clients"
	"github.com/ThreeDotsLabs/the-domain-engineer/clients/bank"
	"github.com/shopspring/decimal"
)

type Client struct {
	clients *commonclients.Clients
}

func NewClient(clients *commonclients.Clients) *Client {
	return &Client{
		clients: clients,
	}
}

func (c *Client) CapturePayment(ctx context.Context, nonce string, amount decimal.Decimal, merchantID string) error {
	resp, err := c.clients.Bank.CapturePaymentWithResponse(ctx, nonce, bank.CapturePaymentJSONRequestBody{
		Amount:     amount,
		MerchantId: merchantID,
	})
	if err != nil {
		return fmt.Errorf("could not capture payment: %w", err)
	}

	if resp.StatusCode() != http.StatusCreated {
		return fmt.Errorf("unexpected status code for capture payment: %v", resp.StatusCode())
	}

	return nil
}
