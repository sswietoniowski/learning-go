package data

import (
	"fmt"
	"strconv"

	"github.com/Rhymond/go-money"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MongoDbEmployee struct {
	ID     primitive.ObjectID   `json:"id" bson:"_id"`
	Name   string               `json:"name"`
	Salary primitive.Decimal128 `json:"salary"`
	Age    int                  `json:"age"`
}

const moneyCode = money.USD

func MongoDbEmployeeToEmployee(e *MongoDbEmployee) (*Employee, error) {
	id := e.ID.Hex()
	if id == "" {
		return nil, fmt.Errorf("ID is not a valid hex string")
	}

	num, err := strconv.ParseFloat(e.Salary.String(), 64)
	if err != nil {
		return nil, err
	}
	salary := money.New(int64(num), moneyCode)

	employee := &Employee{
		ID:     id,
		Name:   e.Name,
		Salary: *salary,
		Age:    e.Age,
	}

	return employee, nil
}

func EmployeeToMongoDbEmployee(e *Employee) (*MongoDbEmployee, error) {
	id, err := primitive.ObjectIDFromHex(e.ID)
	if err != nil {
		return nil, err
	}

	salary, err := primitive.ParseDecimal128(fmt.Sprintf("%d", e.Salary.Amount()))
	if err != nil {
		return nil, err
	}

	mongoDbEmployee := &MongoDbEmployee{
		ID:     id,
		Name:   e.Name,
		Salary: salary,
		Age:    e.Age,
	}

	return mongoDbEmployee, nil
}
