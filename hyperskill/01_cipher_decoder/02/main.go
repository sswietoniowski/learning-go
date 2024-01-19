package main

import (
	"fmt"
	"math/rand"
)

func ModularPower(base, exponent, modulus int) int {
	if modulus == 1 {
		return 0
	}

	result := 1
	for i := 0; i < exponent; i++ {
		result = (result * base) % modulus
	}

	return result
}

func main() {
	var generator, prime int

	_, err := fmt.Scanf("g is %d and p is %d\n", &generator, &prime)
	if err != nil {
		panic(err)
	}

	fmt.Println("OK")

	var secretB = rand.Intn(prime-1) + 1
	var publicB = ModularPower(generator, secretB, prime)

	var publicA int

	_, err = fmt.Scanf("A is %d", &publicA)

	if err != nil {
		panic(err)
	}

	var sharedSecret = ModularPower(publicA, secretB, prime)

	fmt.Printf("B is %d\n", publicB)
	fmt.Printf("A is %d\n", publicA)
	fmt.Printf("S is %d\n", sharedSecret)
}
