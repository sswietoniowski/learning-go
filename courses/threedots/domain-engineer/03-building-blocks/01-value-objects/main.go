package main

// Money is a monetary amount with a currency.
//
// Turn this into a proper value object:
//  1. Reject invalid inputs in NewMoney (negative amount, empty currency).
//  2. Enforce the same-currency invariant in Add.
//
// Encapsulation (unexported fields) and immutability (Add returns a new Money)
// are already in place. The tests in main_test.go describe the target behavior.
type Money struct {
	Amount   int64
	Currency string
}

func NewMoney(amount int64, currency string) (Money, error) {
	// TODO: validate amount and currency.
	return Money{Amount: amount, Currency: currency}, nil
}
