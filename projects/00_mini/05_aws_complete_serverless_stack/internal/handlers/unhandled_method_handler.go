package handlers

import "github.com/aws/aws-lambda-go/events"

func UnhandledMethod() (*events.APIGatewayProxyResponse, error) {
	return nil, nil
}
