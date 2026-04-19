package main

import "fmt"

func main() {
	rect := Rectangle{
		Width:  10,
		Length: 5,
	}

	area := Area(rect)

	fmt.Println("Area:", area)
}

type Rectangle struct {
	Width  int
	Length int
}

func Area(r Rectangle) int {
	return r.Width * r.Length
}
