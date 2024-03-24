package main

import (
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/sswietoniowski/learning-go/projects/00_mini/05_aws_complete_serverless_stack/internal/database"
	"github.com/sswietoniowski/learning-go/projects/00_mini/05_aws_complete_serverless_stack/internal/handlers"
)

var usersRepository *database.UsersRepository

func main() {
	region := os.Getenv("AWS_REGION")

	log.Printf("AWS Region: %s", region)

	awsSession, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Endpoint:    aws.String("http://host.docker.internal:4566"),                               // LocalStack runs on this endpoint by default
		Credentials: credentials.NewStaticCredentials("dummy-access-key", "dummy-secret-key", ""), // LocalStack uses dummy credentials by default
	})

	if err != nil {
		log.Printf("Error creating AWS session: %s", err.Error())
		return
	}

	usersRepository = database.NewUsersRepository(awsSession)

	lambda.Start(handler)
}

func handler(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	const get = "GET"
	const post = "POST"
	const put = "PUT"
	const delete = "DELETE"

	method := req.HTTPMethod

	log.Printf("Lambda Method: %s", method)

	switch method {
	case get:
		return handlers.GetUser(req, usersRepository)
	case post:
		return handlers.CreateUser(req, usersRepository)
	case put:
		return handlers.UpdateUser(req, usersRepository)
	case delete:
		return handlers.DeleteUser(req, usersRepository)
	default:
		return handlers.UnhandledMethod()
	}
}
