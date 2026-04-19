package main

import (
	"github.com/shopspring/decimal"
)

func CalculateTax(priceStr, taxRateStr string) (string, error) {
	price, err := decimal.NewFromString(priceStr)
	if err != nil {
		return "", err
	}
	taxRate, err := decimal.NewFromString(taxRateStr)
	if err != nil {
		return "", err
	}
	return price.Mul(taxRate).StringFixed(2), nil
}
