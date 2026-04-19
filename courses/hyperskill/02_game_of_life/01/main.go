package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
)

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

func readNAndSeed() (int, int) {
	line := readLine()
	var n, seed int
	_, err := fmt.Sscanf(line, "%d %d", &n, &seed)
	handleError(err)
	return n, seed
}

type CellState int

const (
	CellStateDead CellState = iota
	CellStateAlive
)

func generateBoard(n int) [][]CellState {
	board := make([][]CellState, n)
	for i := range board {
		board[i] = make([]CellState, n)
		for j := range board[i] {
			board[i][j] = CellStateDead
		}
	}
	return board
}

func initBoard(board [][]CellState, seed int) {
	src := rand.NewSource(int64(seed))
	r := rand.New(src)
	for i := range board {
		for j := range board[i] {
			board[i][j] = CellState(r.Intn(2))
		}
	}
}

const (
	CellSymbolDead  = " "
	CellSymbolAlive = "O"
)

func printBoard(board [][]CellState) {
	for i := range board {
		for j := range board[i] {
			if board[i][j] == CellStateAlive {
				fmt.Print(CellSymbolAlive)
			} else if board[i][j] == CellStateDead {
				fmt.Print(CellSymbolDead)
			} else {
				fmt.Println("Unknown cell state")
				os.Exit(1)
			}
		}
		fmt.Println()
	}
}

func game() {
	n, seed := readNAndSeed()
	board := generateBoard(n)
	initBoard(board, seed)
	printBoard(board)
}

func main() {
	game()
}
