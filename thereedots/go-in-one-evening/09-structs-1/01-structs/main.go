package main

import "fmt"

func main() {
	point := Point{
		X: 10,
		Y: 5,
	}

	fmt.Println("A point:", point)
}

type Point struct {
	X int
	Y int
}
