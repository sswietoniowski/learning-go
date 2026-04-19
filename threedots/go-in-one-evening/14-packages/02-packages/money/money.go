package money

type Money struct {
	Amount   int
	Currency string
}

func New(amount int, currency string) Money {
	return Money{
		Amount:   amount,
		Currency: currency,
	}
}
