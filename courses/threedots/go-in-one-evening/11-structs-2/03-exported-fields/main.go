package main

import (
	"encoding/json"
	"fmt"
)

type Account struct {
	Name     string `json:"name"`
	password string
}

func main() {
	account := Account{
		Name:     "John",
		password: "top-secret",
	}

	marshaled, err := json.Marshal(account)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(marshaled))
}
