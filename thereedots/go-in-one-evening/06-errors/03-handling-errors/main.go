package main

import (
	"fmt"
	"os"
)

func CreateDirectory(name string) bool {
	err := os.Mkdir(name, 0755)
	if err != nil {
		return false
	}
	return true
}

func main() {
	ok := CreateDirectory("my-directory")
	if ok {
		fmt.Println("Directory created")
	} else {
		fmt.Println("Failed to create directory")
	}
}
