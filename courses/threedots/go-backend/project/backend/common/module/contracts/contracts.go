// Package contracts lives in a sub-package of common/module (not in common/module itself)
// to break the import cycle: common/module/module.go imports this package,
// and each module's module.go implements RegisterContracts — which means
// common/module cannot also define contracts (that would create a cycle).
package contracts

import (
	"errors"

	ordersModule "eats/backend/orders/api/module/client"
)

type Contracts struct {
	ordersModule.Orders
}

func (c *Contracts) Verify() error {
	var err error

	if c.Orders == nil {
		err = errors.Join(err, errors.New("orders module contract is empty"))
	}

	return err
}
