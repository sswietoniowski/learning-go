package data

import (
	"fmt"

	"github.com/sswietoniowski/learning-go/projects/00_mini/04_hrms/internal/domain"
)

type NotFoundError struct {
	ID string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("Record with ID %s not found", e.ID)
}

type DatabaseError struct {
	Operation string
	Err       error
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("Database error during %s: %v", e.Operation, e.Err)
}

type EmployeesRepository interface {
	GetAll() ([]domain.Employee, error)
	Add(employee domain.Employee) (*domain.Employee, error)
	GetById(id string) (*domain.Employee, error)
	ModifyById(id string, employee domain.Employee) (*domain.Employee, error)
	RemoveById(id string) (*domain.Employee, error)
}
