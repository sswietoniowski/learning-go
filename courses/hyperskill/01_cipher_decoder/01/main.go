package main

import (
	"fmt"
)

func main() {
	var generator, prime int

	_, err := fmt.Scanf("g is %d and p is %d", &generator, &prime)
	if err != nil {
		panic(err)
	}

	fmt.Printf("g=%d and p=%d\n", generator, prime)
}
