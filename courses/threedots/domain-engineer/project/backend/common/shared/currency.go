package shared

import (
	"fmt"

	"eats/backend/common"
)

type Currency struct {
	common.Enum[CurrencyType]
}

func (c Currency) Code() string {
	return c.String()
}

func (c Currency) DecimalPlaces() int {
	switch c.String() {
	case "JPY":
		return 0
	default:
		return 2
	}
}

type CurrencyType string

func (c CurrencyType) Values() []string {
	return []string{"USD", "EUR", "GBP", "JPY", "PLN"}
}

func MustNewCurrency(value string) Currency {
	c := Currency{}
	err := c.UnmarshalText([]byte(value))
	if err != nil {
		panic(fmt.Errorf("error unmarshalling currency value: %s", value))
	}

	return c
}
