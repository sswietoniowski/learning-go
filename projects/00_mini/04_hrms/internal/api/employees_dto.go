package api

import (
	"github.com/Rhymond/go-money"
	"github.com/sswietoniowski/learning-go/projects/00_mini/04_hrms/internal/domain"
)

type CreateEmployeeDTO struct {
	Name   string `json:"name"`
	Salary int64  `json:"salary"`
	Age    int    `json:"age"`
}

const currencyCode = money.USD

func (dto *CreateEmployeeDTO) ToEmployee() *domain.Employee {
	return &domain.Employee{
		Name:   dto.Name,
		Salary: *money.New(dto.Salary, currencyCode),
		Age:    dto.Age,
	}
}

type EmployeeDTO struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Salary int64  `json:"salary"`
	Age    int    `json:"age"`
}

func EmployeeToEmployeeDTO(e *domain.Employee) *EmployeeDTO {
	return &EmployeeDTO{
		Id:     e.ID,
		Name:   e.Name,
		Salary: e.Salary.Amount(),
		Age:    e.Age,
	}
}

type ModifyEmployeeDTO struct {
	Name   string `json:"name"`
	Salary int64  `json:"salary"`
	Age    int    `json:"age"`
}

func (dto *ModifyEmployeeDTO) ToEmployee(id string) *domain.Employee {
	return &domain.Employee{
		ID:     id,
		Name:   dto.Name,
		Salary: *money.New(dto.Salary, currencyCode),
		Age:    dto.Age,
	}
}
