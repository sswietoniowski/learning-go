package handlers

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/sswietoniowski/learning-go/projects/00_mini/05_aws_complete_serverless_stack/pkg/internal/users"
)

func GetUser(req events.APIGatewayProxyRequest, usersRepository users.UsersRepository) (*events.APIGatewayProxyResponse, error) {
	return nil, nil
}

func CreateUser(req events.APIGatewayProxyRequest, usersRepository users.UsersRepository) (*events.APIGatewayProxyResponse, error) {
	return nil, nil
}

func UpdateUser(req events.APIGatewayProxyRequest, usersRepository users.UsersRepository) (*events.APIGatewayProxyResponse, error) {
	return nil, nil
}

func DeleteUser(req events.APIGatewayProxyRequest, usersRepository users.UsersRepository) (*events.APIGatewayProxyResponse, error) {
	return nil, nil
}

func UnhandledMethod() (*events.APIGatewayProxyResponse, error) {
	return nil, nil
}
