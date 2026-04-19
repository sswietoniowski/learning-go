package main

import "fmt"

func main() {
	fmt.Println(DirectionFromNumber(1))
	fmt.Println(DirectionFromNumber(3))
}

func DirectionFromNumber(number int) (direction string) {
	switch number {
	case 0:
		direction = "up"
	case 1:
		direction = "right"
	case 2:
		direction = "down"
	case 3:
		direction = "left"
	default:
		direction = "invalid"
	}
	return
}
