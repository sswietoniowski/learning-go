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

func readConfiguration() (int, int, int) {
	line := readLine()
	var size, seed, generations int
	_, err := fmt.Sscanf(line, "%d %d %d", &size, &seed, &generations)
	handleError(err)
	if size < 1 {
		fmt.Println("N must be greater than 0")
		os.Exit(1)
	}
	if seed < 0 {
		fmt.Println("Seed must be greater than or equal to 0")
		os.Exit(1)
	}
	if generations < 0 {
		fmt.Println("M must be greater than or equal to 0")
		os.Exit(1)
	}
	return size, seed, generations
}

type cell struct {
	alive bool
}

func newCell() *cell {
	return &cell{false}
}

func (c *cell) isAlive() bool {
	return c.alive
}

func (c *cell) setAlive() {
	c.alive = true
}

func (c *cell) String() string {
	const (
		cellSymbolDead  = " "
		cellSymbolAlive = "O"
	)

	if c.isAlive() {
		return cellSymbolAlive
	}
	return cellSymbolDead
}

type board struct {
	cells [][]cell
}

func newBoard(size int) *board {
	cells := make([][]cell, size)
	for i := range cells {
		cells[i] = make([]cell, size)
		for j := range cells[i] {
			cells[i][j] = *newCell()
		}
	}

	return &board{cells}
}

func (b *board) fill(seed int) {
	src := rand.NewSource(int64(seed))
	r := rand.New(src)

	for i := range b.cells {
		for j := range b.cells[i] {
			if r.Intn(2) == 1 {
				b.cells[i][j].setAlive()
			}
		}
	}
}

func (b *board) countAliveNeighbours(i int, j int) int {
	size := len(b.cells)

	neighbours := make([]cell, 0, 8)

	dx := []int{-1, -1, -1, 0, 0, 1, 1, 1}
	dy := []int{-1, 0, 1, -1, 1, -1, 0, 1}

	for k := 0; k < 8; k++ {
		neighbours = append(neighbours, b.cells[(i+dx[k]+size)%size][(j+dy[k]+size)%size])
	}

	alive := 0
	for _, neighbour := range neighbours {
		if neighbour.isAlive() {
			alive++
		}
	}

	return alive
}

func (b *board) generate() {
	next := newBoard(len(b.cells))

	for i := range b.cells {
		for j := range b.cells[i] {
			alive := b.countAliveNeighbours(i, j)

			if b.cells[i][j].isAlive() {
				if alive == 2 || alive == 3 {
					next.cells[i][j].setAlive()
				}

			} else {
				if alive == 3 {
					next.cells[i][j].setAlive()
				}
			}
		}

	}

	b.cells = next.cells
}

func (b *board) String() string {
	var sb strings.Builder
	for i := range b.cells {
		for j := range b.cells[i] {
			sb.WriteString(b.cells[i][j].String())
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func (b *board) print() {
	fmt.Print(b)
}

func game() {
	size, seed, generations := readConfiguration()
	board := newBoard(size)
	board.fill(seed)
	for i := 1; i <= generations; i++ {
		board.generate()
	}
	board.print()
}

func main() {
	game()
}
