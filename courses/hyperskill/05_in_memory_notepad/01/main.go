package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	const (
		CommandExit = "exit"
	)

	readCommand := func() (string, string) {
		fmt.Print("Enter a command and data: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()

		input := strings.Split(scanner.Text(), " ")
		command, data := input[0], strings.Join(input[1:], " ")
		return command, data
	}

	for {
		switch command, data := readCommand(); command {
		case CommandExit:
			fmt.Println("[Info] Bye!")
		default:
			fmt.Println(command, data)
		}
	}
}
