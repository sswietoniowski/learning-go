package main

import "fmt"

func main() {
	result := Sum(1, 2, 3, 4, 5)
	fmt.Println(result)
}

func Sum(a int, b int, c int, d int, e int) int {
	return a + b + c + d + e
}
