package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
)

const (
	LowercaseAAscii = 97
	LowercaseZAscii = 122
	UppercaseAAscii = 65
	UppercaseZAscii = 90
	AlphabetLength  = 26
)

type Operation bool

const (
	OperationEncrypt Operation = true
	OperationDecrypt           = false
)

func shiftChar(char rune, shift int, operation Operation) string {
	var base int
	if char >= LowercaseAAscii && char <= LowercaseZAscii {
		base = LowercaseAAscii
	} else if char >= UppercaseAAscii && char <= UppercaseZAscii {
		base = UppercaseAAscii
	} else {
		return string(char)
	}

	if operation == OperationDecrypt {
		shift = -shift
	}

	return string(rune((int(char)-base+shift+AlphabetLength)%AlphabetLength + base))
}

func cipher(input string, shift int, operation Operation) string {
	var output string
	for _, char := range input {
		output += shiftChar(char, shift, operation)
	}
	return output
}

func encrypt(plainText string, shift int) string {
	return cipher(plainText, shift, OperationEncrypt)
}

func decrypt(cipherText string, shift int) string {
	return cipher(cipherText, shift, OperationDecrypt)
}

func generatePrivateB(prime int) int {
	return rand.Intn(prime-1) + 1
}

func modularPower(base, exponent, modulus int) int {
	if modulus == 1 {
		return 0
	}

	result := 1
	for i := 0; i < exponent; i++ {
		result = (result * base) % modulus
	}

	return result
}

func calculatePublicB(generator, privateB, prime int) int {
	return modularPower(generator, privateB, prime)
}

func calculateSharedSecret(publicA, privateB, prime int) int {
	return modularPower(publicA, privateB, prime)
}

func calculateShift(sharedSecret, prime int) int {
	return sharedSecret % AlphabetLength
}

func handleError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func readLine() string {
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	handleError(err)
	return strings.TrimSpace(text)
}

func readGAndP() (int, int) {
	line := readLine()
	var generator, prime int
	_, err := fmt.Sscanf(line, "g is %d and p is %d", &generator, &prime)
	handleError(err)
	return generator, prime
}

func readPublicA() int {
	line := readLine()
	var publicA int
	_, err := fmt.Sscanf(line, "A is %d", &publicA)
	handleError(err)
	return publicA
}

func diffieHellmanKeyExchange() int {
	generator, prime := readGAndP()

	fmt.Println("OK")

	secretB := generatePrivateB(prime)
	publicB := calculatePublicB(generator, secretB, prime)
	publicA := readPublicA()
	sharedSecret := calculateSharedSecret(publicA, secretB, prime)

	fmt.Printf("B is %d\n", publicB)

	shift := calculateShift(sharedSecret, prime)
	return shift
}

func bobAsks(shift int) {
	const bobQuestion = "Will you marry me?"
	bobEncryptedQuestion := encrypt(bobQuestion, shift)
	fmt.Printf("%s\n", bobEncryptedQuestion)
}

func aliceAnswers(shift int) string {
	aliceEncryptedResponse := readLine()
	aliceResponse := decrypt(aliceEncryptedResponse, shift)
	return aliceResponse
}

func bobAnswers(aliceResponse string, shift int) {
	const (
		AlicePositiveResponse = "Yeah, okay!"
		AliceNegativeResponse = "Let's be friends."
		BobPositiveResponse   = "Great!"
		BobNegativeResponse   = "What a pity!"
	)

	var bobResponse string
	switch aliceResponse {
	case AlicePositiveResponse:
		bobResponse = BobPositiveResponse
	case AliceNegativeResponse:
		bobResponse = BobNegativeResponse
	}

	if bobResponse != "" {
		bobEncryptedResponse := encrypt(bobResponse, shift)
		fmt.Printf("%s\n", bobEncryptedResponse)
	}
}

func conversation() {
	shift := diffieHellmanKeyExchange()
	bobAsks(shift)
	aliceResponse := aliceAnswers(shift)
	bobAnswers(aliceResponse, shift)
}

func main() {
	conversation()
}
