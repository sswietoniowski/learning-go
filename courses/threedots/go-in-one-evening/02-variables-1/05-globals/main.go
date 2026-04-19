package main

import "fmt"

var price = 40.0

const taxRate = 0.1

func main() {
	tax := taxRate * price
	fmt.Println("tax:", tax)
}
