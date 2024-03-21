package main

import (
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"

	"github.com/sswietoniowski/learning-go/projects/00_mini/05_aws_complete_serverless_stack/pkg/internal/handlers"
	"github.com/sswietoniowski/learning-go/projects/00_mini/05_aws_complete_serverless_stack/pkg/internal/users"
)

var dynaClient dynamodbiface.DynamoDBAPI

func main() {
	region := os.Getenv("AWS_REGION")

	awsSession, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})

	if err != nil {
		return
	}

	dynaClient = dynamodb.New(awsSession)

	lambda.Start(handler)
}

func handler(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	const tableName = "aws-complete-serverless-stack-users"

	usersRepository := users.NewUsersRepository(tableName, dynaClient)

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
