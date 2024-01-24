package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

type vocabulary struct {
	words map[string]bool
}

func newVocabulary() *vocabulary {
	return &vocabulary{words: make(map[string]bool)}
}

func cleanWord(word string) string {
	return strings.ToLower(strings.TrimSpace(word))
}

func (v *vocabulary) addWord(word string) {
	v.words[cleanWord(word)] = true
}

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func (v *vocabulary) readWords(fileName string) {
	file, err := os.Open(fileName)
	handleError(err)

	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		word := scanner.Text()
		v.addWord(word)
	}
}

func (v *vocabulary) isTaboo(word string) bool {
	return v.words[cleanWord(word)]
}

func checker() {
	readLine := func() string {
		var line string

		_, err := fmt.Scan(&line)
		handleError(err)

		return line
	}

	fileName := readLine()

	vocabulary := newVocabulary()
	vocabulary.readWords(fileName)

	word := readLine()

	if vocabulary.isTaboo(word) {
		fmt.Println("True")
	} else {
		fmt.Println("False")
	}
}

func main() {
	checker()
}
