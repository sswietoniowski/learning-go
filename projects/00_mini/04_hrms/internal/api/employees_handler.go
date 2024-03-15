package api

import (
	"encoding/json"

	"github.com/gofiber/fiber/v3"
	"github.com/sswietoniowski/learning-go/projects/00_mini/04_hrms/internal/data"
)

func (a *Application) getAllEmployees(ctx fiber.Ctx) error {
	employees, err := a.repository.GetAll(ctx.Context())
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(err)
		return err
	}

	ctx.Status(fiber.StatusOK).JSON(EmployeesToEmployeesDto(employees))
	return nil
}

func (a *Application) addEmployee(ctx fiber.Ctx) error {
	createEmployeeDto := &CreateEmployeeDTO{}
	body := ctx.Body()
	err := json.Unmarshal(body, &createEmployeeDto)
	if err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(err)
		return nil
	}

	employee, err := a.repository.Add(ctx.Context(), *createEmployeeDto.ToEmployee())
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(err)
		return err
	}

	ctx.Status(fiber.StatusCreated).JSON(EmployeeToEmployeeDTO(employee))
	return nil
}

func (a *Application) getEmployeeById(ctx fiber.Ctx) error {
	id := ctx.Params("id")
	employee, err := a.repository.GetById(ctx.Context(), id)
	if err != nil {
		switch err.(type) {
		case *data.NotFoundError:
			ctx.Status(fiber.StatusNotFound).JSON(err)
			return nil
		default:
			ctx.Status(fiber.StatusInternalServerError).JSON(err)
			return err
		}
	}

	ctx.Status(fiber.StatusOK).JSON(EmployeeToEmployeeDTO(employee))
	return nil
}

func (a *Application) modifyEmployeeById(ctx fiber.Ctx) error {
	id := ctx.Params("id")
	modifyEmployeeDto := &ModifyEmployeeDTO{}
	body := ctx.Body()
	err := json.Unmarshal(body, &modifyEmployeeDto)
	if err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(err)
		return nil
	}

	_, err = a.repository.ModifyById(ctx.Context(), id, *modifyEmployeeDto.ToEmployee(id))
	if err != nil {
		switch err.(type) {
		case *data.NotFoundError:
			ctx.Status(fiber.StatusNotFound).JSON(err)
			return nil
		default:
			ctx.Status(fiber.StatusInternalServerError).JSON(err)
			return err
		}
	}

	ctx.Status(fiber.StatusNoContent)
	return nil
}

func (a *Application) removeEmployeeById(ctx fiber.Ctx) error {
	id := ctx.Params("id")
	_, err := a.repository.RemoveById(ctx.Context(), id)
	if err != nil {
		switch err.(type) {
		case *data.NotFoundError:
			ctx.Status(fiber.StatusNotFound).JSON(err)
			return nil
		default:
			ctx.Status(fiber.StatusInternalServerError).JSON(err)
			return err
		}
	}

	ctx.Status(fiber.StatusNoContent)
	return nil
}
