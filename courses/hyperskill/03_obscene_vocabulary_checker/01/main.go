package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func readFileName() string {
	var name string

	_, err := fmt.Scan(&name)
	handleError(err)

	return name
}

func readWords(name string) []string {
	file, err := os.Open(name)
	handleError(err)

	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)

	var words []string
	for scanner.Scan() {
		words = append(words, scanner.Text())
	}

	return words
}

func printWords(words []string) {
	for _, word := range words {
		fmt.Println(word)
	}
}

func main() {
	name := readFileName()
	words := readWords(name)
	printWords(words)
}
