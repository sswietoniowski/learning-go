package data

import "github.com/Rhymond/go-money"

type Employee struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Salary money.Money `json:"salary"`
	Age    int         `json:"age"`
}
