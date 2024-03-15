package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sswietoniowski/learning-go/projects/00_mini/04_hrms/internal/data"
)

type Application struct {
	api        *fiber.App
	repository data.EmployeesRepository
}

func NewApplication(api *fiber.App, repository data.EmployeesRepository) *Application {
	return &Application{
		api:        api,
		repository: repository,
	}
}

func (a *Application) SetupRoutes() {
	a.api.Get("/api/v1/employees", a.getAllEmployees)
	a.api.Post("/api/v1/employees", a.addEmployee)
	a.api.Get("/api/v1/employees/:id", a.getEmployeeById)
	a.api.Put("/api/v1/employees/:id", a.modifyEmployeeById)
	a.api.Delete("/api/v1/employees/:id", a.removeEmployeeById)
}

func (a *Application) Close() {
	defer a.repository.Close()
}
