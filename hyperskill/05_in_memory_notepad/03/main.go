package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type notepad struct {
	notes   []string
	maxSize int
}

func newNotepad(maxSize int) *notepad {
	return &notepad{
		notes:   make([]string, 0, maxSize),
		maxSize: maxSize,
	}
}

func (n *notepad) create(note string) error {
	if len(note) == 0 || strings.TrimSpace(note) == "" {
		return fmt.Errorf("Missing note argument")
	}

	if len(n.notes) == n.maxSize {
		return fmt.Errorf("Notepad is full")
	}

	n.notes = append(n.notes, note)

	return nil
}

func (n *notepad) clear() {
	n.notes = make([]string, 0, n.maxSize)
}

func (n *notepad) list() []string {
	return n.notes[:len(n.notes)]
}

func main() {
	const (
		CommandCreate = "create"
		CommandClear  = "clear"
		CommandList   = "list"
		CommandExit   = "exit"
	)

	readMaxSize := func() int {
		fmt.Print("Enter the maximum number of notes: ")
		var size int
		fmt.Scan(&size)
		return size
	}

	readCommand := func() (string, string) {
		fmt.Print("Enter a command and data: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()

		input := strings.Split(scanner.Text(), " ")
		command, data := input[0], strings.Join(input[1:], " ")
		return command, data
	}

	maxSize := readMaxSize()

	n := newNotepad(maxSize)

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
			notes := n.list()
			if len(notes) == 0 {
				fmt.Println("[Info] Notepad is empty")
				continue
			} else {
				for i, note := range notes {
					fmt.Printf("[Info] %d: %s\n", i+1, note)
				}
			}
		case CommandExit:
			fmt.Println("[Info] Bye!")
		default:
			fmt.Println("[Error] Unknown command")
		}
	}
}
