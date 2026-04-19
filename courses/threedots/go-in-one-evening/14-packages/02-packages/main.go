package main

import (
	"fmt"

	"shop/money"
)

type Product struct {
	Name  string
	Price money.Money
}

func main() {
	laptop := Product{
		Name:  "Laptop",
		Price: money.New(1000, "EUR"),
	}

	fmt.Println(laptop.Name, "costs", laptop.Price)
}
