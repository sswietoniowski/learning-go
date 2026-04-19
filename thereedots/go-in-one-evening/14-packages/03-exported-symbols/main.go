package main

import (
	"fmt"
	"os"

	"login/account"
)

func main() {
	acc, err := account.New("kate@example.com", "t0ps3cr3t")
	if err != nil {
		panic(err)
	}

	if len(os.Args) < 3 {
		fmt.Println("Usage:", os.Args[0], "<email> <password>")
		os.Exit(1)
	}

	email := os.Args[1]
	password := os.Args[2]

	ok := acc.Login(email, password)
	if ok {
		fmt.Println("Login successful")
	} else {
		fmt.Println("Login failed")
	}
}
