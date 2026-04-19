package main

import (
	"fmt"
	"strings"
)

func main() {
	planets := []string{"Mercury", "Venus", "Earth", "Mars", "Jupiter", "Saturn", "Uranus", "Neptune"}

	uppercase := func(s string) string {
		return strings.ToUpper(s)
	}

	uppercasePlanets := MapValues(planets, uppercase)
	fmt.Println(uppercasePlanets)
}

func MapValues(values []string, mapper func(string) string) []string {
	result := make([]string, len(values))
	for i, v := range values {
		result[i] = mapper(v)
	}
	return result
}
