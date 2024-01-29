package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type notepad struct {
	notes []string
}

func newNotepad(maxSize int) *notepad {
	return &notepad{
		notes: make([]string, 0, maxSize),
	}
}

func (n *notepad) create(note string) error {
	if len(note) == 0 || strings.TrimSpace(note) == "" {
		return fmt.Errorf("Missing note argument")
	}

	if len(n.notes) == cap(n.notes) {
		return fmt.Errorf("Notepad is full")
	}

	n.notes = append(n.notes, note)

	return nil
}

func (n *notepad) clear() {
	n.notes = make([]string, 0, cap(n.notes))
}

func (n *notepad) list() []string {
	return n.notes[:len(n.notes)]
}

func (n *notepad) update(index int, note string) error {
	if index > len(n.notes) && index <= cap(n.notes) {
		return fmt.Errorf("There is nothing to update")
	}

	if index < 1 || index > cap(n.notes) {
		return fmt.Errorf("Position %d is out of the boundaries [1, %d]", index, cap(n.notes))
	}

	if len(note) == 0 || strings.TrimSpace(note) == "" {
		return fmt.Errorf("Missing note argument")
	}

	n.notes[index-1] = note

	return nil
}

func (n *notepad) delete(index int) error {
	if index > len(n.notes) && index <= cap(n.notes) {
		return fmt.Errorf("There is nothing to delete")
	}

	if index < 1 || index > cap(n.notes) {
		return fmt.Errorf("Position %d is out of the boundaries [1, %d]", index, cap(n.notes))
	}

	n.notes = append(n.notes[:index-1], n.notes[index:]...)

	return nil
}

func main() {
	const (
		CommandCreate = "create"
		CommandClear  = "clear"
		CommandList   = "list"
		CommandExit   = "exit"
		CommandUpdate = "update"
		CommandDelete = "delete"
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

	parseUpdate := func(data string) (int, string, error) {
		input := strings.Split(data, " ")
		indexStr, note := input[0], strings.Join(input[1:], " ")

		if indexStr == "" {
			return -1, "", fmt.Errorf("Missing position argument")
		}

		if note == "" {
			return -1, "", fmt.Errorf("Missing note argument")
		}

		index, err := strconv.Atoi(indexStr)

		if err != nil {
			return -1, "", fmt.Errorf("Invalid position: %s", indexStr)
		}

		return index, note, nil
	}

	parseDelete := func(data string) (int, error) {
		if data == "" {
			return -1, fmt.Errorf("Missing position argument")
		}

		index, err := strconv.Atoi(data)

		if err != nil {
			return -1, fmt.Errorf("Invalid position: %s", data)
		}

		return index, nil
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
		case CommandUpdate:
			index, note, err := parseUpdate(data)
			if err != nil {
				fmt.Printf("[Error] %s\n", err)
			} else {
				err := n.update(index, note)

				if err != nil {
					fmt.Printf("[Error] %s\n", err)
				} else {
					fmt.Printf("[OK] The note at position %d was successfully updated\n", index)
				}
			}
		case CommandDelete:
			index, err := parseDelete(data)

			if err != nil {
				fmt.Printf("[Error] %s\n", err)
			} else {
				err := n.delete(index)
				if err != nil {
					fmt.Printf("[Error] %s\n", err)
				} else {
					fmt.Printf("[OK] The note at position %d was successfully deleted\n", index)
				}
			}
		default:
			fmt.Println("[Error] Unknown command")
		}
	}
}
