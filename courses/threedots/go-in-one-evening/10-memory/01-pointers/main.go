package main

import "fmt"

type Order struct {
	Products []int
}

type User struct {
	Name   string
	Orders []*Order
}

func main() {
	firstOrder := Order{
		Products: []int{545, 490},
	}

	secondOrder := Order{
		Products: []int{98, 829, 245},
	}

	alice := User{
		Name:   "Alice",
		Orders: []*Order{&firstOrder, &secondOrder},
	}

	// This change will be visible in 'alice.Orders'
	firstOrder.Products = append(firstOrder.Products, 99)

	fmt.Println(alice.Name, "placed", len(alice.Orders), "orders")

	for _, o := range alice.Orders {
		fmt.Println("Ordered products:", o.Products)
	}
}
