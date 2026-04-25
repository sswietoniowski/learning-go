package shared

import (
	"fmt"

	"eats/backend/common"
)

type CurrencyType string

func (c CurrencyType) Values() []string {
	return []string{"USD", "EUR", "GBP", "JPY", "PLN"}
}

type Currency struct {
	common.Enum[CurrencyType]
}

func (c Currency) Code() string {
	return c.String()
}

func MustNewCurrency(value string) Currency {
	c := Currency{}
	err := c.UnmarshalText([]byte(value))
	if err != nil {
		panic(fmt.Errorf("error unmarshalling currency value: %s", value))
	}
	return c
}
