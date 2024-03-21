package main

import (
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/sswietoniowski/learning-go/projects/00_mini/05_aws_complete_serverless_stack/pkg/internal/handlers"
	"github.com/sswietoniowski/learning-go/projects/00_mini/05_aws_complete_serverless_stack/pkg/internal/users"
)

var usersRepository *users.UsersRepository

func main() {
	region := os.Getenv("AWS_REGION")

	awsSession, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})

	if err != nil {
		return
	}

	usersRepository = users.NewUsersRepository(awsSession)

	lambda.Start(handler)
}

func handler(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		return handlers.GetUser(req, usersRepository)
	case "POST":
		return handlers.CreateUser(req, usersRepository)
	case "PUT":
		return handlers.UpdateUser(req, usersRepository)
	case "DELETE":
		return handlers.DeleteUser(req, usersRepository)
	default:
		return handlers.UnhandledMethod()
	}
}
