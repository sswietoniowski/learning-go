package data

import (
	"fmt"
	"testing"

	"github.com/Rhymond/go-money"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestEmployeeToMongoDbEmploee(t *testing.T) {
	// Arrange
	id := primitive.NewObjectID().Hex()
	name := "John Doe"
	amount := int64(1000)
	salary := money.New(amount, money.USD)
	age := 30
	employee := &Employee{
		ID:     id,
		Name:   name,
		Salary: *salary,
		Age:    age,
	}
	wantId, _ := primitive.ObjectIDFromHex(id)
	wantSalary, _ := primitive.ParseDecimal128(fmt.Sprintf("%d", amount))

	// Act
	got, err := EmployeeToMongoDbEmployee(employee)

	// Assert
	if err != nil {
		t.Errorf("EmployeeToMongoDbEmployee() failed: %v", err)
	}

	if got.ID != wantId {
		t.Errorf("Expected ID to be %v, but got %v", wantId, got.ID)
	}

	if got.Name != name {
		t.Errorf("Expected Name to be %v, but got %v", name, got.Name)
	}

	if got.Salary != wantSalary {
		t.Errorf("Expected Salary to be %v, but got %v", wantSalary, got.Salary)
	}

	if got.Age != age {
		t.Errorf("Expected Age to be %v, but got %v", age, got.Age)
	}
}

func TestMongoDbEmployeeToEmployee(t *testing.T) {
	// Arrange
	id := primitive.NewObjectID()
	name := "John Doe"
	amount := int64(1000)
	salary, _ := primitive.ParseDecimal128(fmt.Sprintf("%d", amount))
	age := 30
	mongoDbEmployee := &MongoDbEmployee{
		ID:     id,
		Name:   name,
		Salary: salary,
		Age:    age,
	}
	wantId := id.Hex()
	wantSalary := *money.New(1000, money.USD)

	// Act
	got, err := MongoDbEmployeeToEmployee(mongoDbEmployee)

	// Assert
	if err != nil {
		t.Errorf("MongoDbEmployeeToEmployee() failed: %v", err)
	}

	if got.ID != wantId {
		t.Errorf("Expected ID to be %v, but got %v", wantId, got.ID)
	}

	if got.Name != name {
		t.Errorf("Expected Name to be %v, but got %v", name, got.Name)
	}

	if got.Salary != wantSalary {
		t.Errorf("Expected Salary to be %v, but got %v", wantSalary, got.Salary)
	}
}
