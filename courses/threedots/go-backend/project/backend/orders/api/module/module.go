package module

import (
	"eats/backend/orders/api/module/client"
)

type Orders struct{}

func (o Orders) PingOrders(request client.PingOrdersRequest) error {
	return nil
}
