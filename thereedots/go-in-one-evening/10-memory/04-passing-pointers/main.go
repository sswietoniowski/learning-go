package main

import (
	"fmt"
)

func main() {
	addresses := []string{
		"127.0.0.1",
		"10.0.0.0",
		"10.0.0.0",
		"127.0.0.1",
		"10.0.0.2",
	}

	Deduplicate(&addresses)
	fmt.Println(addresses)
}

func Deduplicate(addresses *[]string) {
	m := map[string]bool{}

	for _, a := range *addresses {
		m[a] = true
	}

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	*addresses = keys
}
