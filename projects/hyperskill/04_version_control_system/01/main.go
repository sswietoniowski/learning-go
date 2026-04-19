package main

import (
	"fmt"
	"os"
)

func printHelp() {
	printCommandDescription("--help")
}

func printError(command string) {
	fmt.Printf("'%s' is not a SVCS command.\n", command)
}

var commands = map[string]string{
	"config":   "Get and set a username.",
	"add":      "Add a file to the index.",
	"log":      "Show commit logs.",
	"commit":   "Save changes.",
	"checkout": "Restore a file.",
	"--help": `These are SVCS commands:
config     Get and set a username.
add        Add a file to the index.
log        Show commit logs.
commit     Save changes.
checkout   Restore a file.`,
}

func printCommandDescription(command string) {
	fmt.Println(commands[command])
}

func main() {
	numberOfArgs := len(os.Args) - 1

	if numberOfArgs == 0 {
		printHelp()
		return
	}

	command := os.Args[1]

	if _, ok := commands[command]; !ok {
		printError(command)
		return
	}

	printCommandDescription(command)
}
