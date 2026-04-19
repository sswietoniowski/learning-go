package main

import "fmt"

func WordGenerator(words []string) func() string {
	index := 0

	return func() string {
		word := words[index]
		index = (index + 1) % len(words)
		return word
	}
}

func main() {
	continents := []string{
		"Africa",
		"Antarctica",
		"Asia",
		"Australia",
		"Europe",
		"North America",
		"South America",
	}

	generator := WordGenerator(continents)

	for i := 0; i < 10; i++ {
		fmt.Println(generator())
	}
}
