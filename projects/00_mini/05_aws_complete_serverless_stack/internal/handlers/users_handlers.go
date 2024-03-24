package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/sswietoniowski/learning-go/projects/00_mini/05_aws_complete_serverless_stack/internal/database"
	"github.com/sswietoniowski/learning-go/projects/00_mini/05_aws_complete_serverless_stack/internal/validators"
)

const (
	ErrorInvalidUserData = "invalid user data"
	ErrorInvalidEmail    = "invalid email"
)

func GetUser(req events.APIGatewayProxyRequest, usersRepository *database.UsersRepository) (*events.APIGatewayProxyResponse, error) {
	email := req.QueryStringParameters["email"]
	if email == "" {
		result, err := usersRepository.GetAll()
		if err != nil {
			return errorResponse(http.StatusBadRequest, err)
		}

		return apiResponse(http.StatusOK, result)
	}

	result, err := usersRepository.GetByEmail(email)
	if err != nil {
		return errorResponse(http.StatusBadRequest, err)
	}

	return apiResponse(http.StatusOK, result)
}

func CreateUser(req events.APIGatewayProxyRequest, usersRepository *database.UsersRepository) (*events.APIGatewayProxyResponse, error) {
	user := new(database.User)

	if err := json.Unmarshal([]byte(req.Body), user); err != nil {
		log.Printf("Error: %s", err)

		return errorResponse(http.StatusBadRequest, errors.New(ErrorInvalidUserData))
	}

	if !validators.IsValidEmail(user.Email) {
		return errorResponse(http.StatusBadRequest, errors.New(ErrorInvalidEmail))
	}

	user, err := usersRepository.Create(user)
	if err != nil {
		return errorResponse(http.StatusBadRequest, err)
	}

	return apiResponse(http.StatusCreated, user)
}

func UpdateUser(req events.APIGatewayProxyRequest, usersRepository *database.UsersRepository) (*events.APIGatewayProxyResponse, error) {
	user := new(database.User)

	if err := json.Unmarshal([]byte(req.Body), user); err != nil {
		log.Printf("Error: %s", err)

		return errorResponse(http.StatusBadRequest, errors.New(ErrorInvalidUserData))
	}

	if !validators.IsValidEmail(user.Email) {
		return errorResponse(http.StatusBadRequest, errors.New(ErrorInvalidEmail))
	}

	_, err := usersRepository.Update(user)
	if err != nil {
		return errorResponse(http.StatusBadRequest, err)
	}

	return apiResponse(http.StatusNoContent, nil)
}

func DeleteUser(req events.APIGatewayProxyRequest, usersRepository *database.UsersRepository) (*events.APIGatewayProxyResponse, error) {
	email := req.QueryStringParameters["email"]
	if email == "" {
		return errorResponse(http.StatusBadRequest, errors.New(ErrorInvalidEmail))
	}

	_, err := usersRepository.DeleteByEmail(email)

	if err != nil {
		return errorResponse(http.StatusBadRequest, err)
	}

	return apiResponse(http.StatusNoContent, nil)
}
