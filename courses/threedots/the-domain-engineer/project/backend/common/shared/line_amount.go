package shared

import (
	"github.com/shopspring/decimal"
)

type LineAmount struct {
	amount  decimal.Decimal
	isGross bool
}

func (l LineAmount) Amount() decimal.Decimal {
	return l.amount
}

func (l LineAmount) IsNet() bool {
	return !l.isGross
}

func (l LineAmount) IsGross() bool {
	return l.isGross
}

func NewNetAmount(amount decimal.Decimal) LineAmount {
	return LineAmount{
		amount:  amount,
		isGross: false,
	}
}

func NewGrossAmount(amount decimal.Decimal) LineAmount {
	return LineAmount{
		amount:  amount,
		isGross: true,
	}
}
