package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

type helloEvent struct {
	Name string `json:"What is your name?"`
	YOB  int    `json:"What is your year of birth?"`
}

type helloResponse struct {
	Message string `json:"Answer:"`
}

func handleLambdaEvent(event helloEvent) (helloResponse, error) {
	age := time.Now().Year() - event.YOB
	message := fmt.Sprintf("Hello, %s! You are %d years old.", event.Name, age)

	return helloResponse{Message: message}, nil
}

func main() {
	lambda.Start(handleLambdaEvent)
}
