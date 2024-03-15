package data

import (
	"context"
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
	GetAll(ctx context.Context) ([]domain.Employee, error)
	Add(ctx context.Context, employee domain.Employee) (*domain.Employee, error)
	GetById(ctx context.Context, id string) (*domain.Employee, error)
	ModifyById(ctx context.Context, id string, employee domain.Employee) (*domain.Employee, error)
	RemoveById(ctx context.Context, id string) (*domain.Employee, error)
	Close() error
}
