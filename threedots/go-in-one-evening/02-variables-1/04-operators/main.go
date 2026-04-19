package main

import "fmt"

func main() {
	numberOfProducts := 3
	productCost := 100
	shippingCost := 10
	totalCost := numberOfProducts*productCost + shippingCost

	fmt.Println("Total order cost:", totalCost)
}
