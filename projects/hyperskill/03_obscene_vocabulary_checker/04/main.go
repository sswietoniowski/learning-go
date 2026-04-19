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

func (v *vocabulary) censorWord(word string) string {
	if v.isTaboo(word) {
		return strings.Repeat("*", len(word))
	}

	return word
}

func (v *vocabulary) censorSentence(sentence string) string {
	words := strings.Fields(sentence)

	for i, word := range words {
		words[i] = v.censorWord(word)
	}

	return strings.Join(words, " ")
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

	for sentence := readLine(); sentence != "exit"; sentence = readLine() {
		fmt.Println(vocabulary.censorSentence(sentence))
	}

	fmt.Println("Bye!")
}

func main() {
	checker()
}
