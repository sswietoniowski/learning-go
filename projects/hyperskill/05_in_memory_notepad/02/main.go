package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	NotesQuantity = 5
)

type notepad struct {
	notes     [NotesQuantity]string
	lastIndex int
}

func newNotepad() *notepad {

	return &notepad{lastIndex: -1}
}

func (n *notepad) create(note string) error {
	if n.lastIndex+1 == NotesQuantity {
		return fmt.Errorf("Notepad is full")
	}

	n.lastIndex++

	n.notes[n.lastIndex] = note

	return nil
}

func (n *notepad) clear() {
	n.lastIndex = -1
}

func (n *notepad) list() []string {
	if n.lastIndex == -1 {
		return nil
	}

	return n.notes[:n.lastIndex+1]
}

func main() {
	const (
		CommandCreate = "create"
		CommandClear  = "clear"
		CommandList   = "list"
		CommandExit   = "exit"
	)

	readCommand := func() (string, string) {
		fmt.Print("Enter a command and data: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()

		input := strings.Split(scanner.Text(), " ")
		command, data := input[0], strings.Join(input[1:], " ")
		return command, data
	}

	n := newNotepad()

	for {
		switch command, data := readCommand(); command {
		case CommandCreate:
			err := n.create(data)
			if err != nil {
				fmt.Printf("[Error] %s\n", err)
			} else {
				fmt.Printf("[OK] The note was successfully created\n")
			}
		case CommandClear:
			n.clear()
			fmt.Printf("[OK] All notes were successfully deleted\n")
		case CommandList:
			for i, note := range n.list() {
				fmt.Printf("[Info] %d: %s\n", i+1, note)
			}
		case CommandExit:
			fmt.Println("[Info] Bye!")
		}
	}
}
