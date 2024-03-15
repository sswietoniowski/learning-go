package data

import (
	"fmt"
	"strconv"

	"github.com/Rhymond/go-money"
	"github.com/sswietoniowski/learning-go/projects/00_mini/04_hrms/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const currencyCode = money.USD // as of now, we only support USD

func MongoDbEmployeeToEmployee(e *MongoDbEmployee) (*domain.Employee, error) {
	id := e.ID.Hex()
	if id == "" {
		return nil, fmt.Errorf("ID is not a valid hex string")
	}

	num, err := strconv.ParseFloat(e.Salary.String(), 64)
	if err != nil {
		return nil, err
	}
	salary := money.New(int64(num), currencyCode)

	employee := &domain.Employee{
		ID:     id,
		Name:   e.Name,
		Salary: *salary,
		Age:    e.Age,
	}

	return employee, nil
}

func EmployeeToMongoDbEmployee(e *domain.Employee) (*MongoDbEmployee, error) {
	id := primitive.ObjectID{}
	if e.ID != "" {
		objectId, err := primitive.ObjectIDFromHex(e.ID)
		if err != nil {
			return nil, err
		}

		id = objectId
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

func MongoDbEmployeesToEmployees(mongoDbEmployees []MongoDbEmployee) ([]domain.Employee, error) {
	employees := make([]domain.Employee, 0, len(mongoDbEmployees))
	for _, e := range mongoDbEmployees {
		employee, err := MongoDbEmployeeToEmployee(&e)
		if err != nil {
			return nil, err
		}

		employees = append(employees, *employee)
	}

	return employees, nil
}
