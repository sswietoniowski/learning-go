package main

import "fmt"

func Alphabet(length int) []string {
	alphabet := make([]string, length)
	for i := 0; i < length; i++ {
		alphabet[i] = characterByIndex(i)
	}

	return alphabet
}

func main() {
	alphabet := Alphabet(26)
	fmt.Println(alphabet)
}

func characterByIndex(i int) string {
	return string(rune('a' + i))
}
