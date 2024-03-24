package handlers

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

const ErrorMethodNotAllowed = "method not allowed"

func UnhandledMethod() (*events.APIGatewayProxyResponse, error) {
	return apiResponse(http.StatusMethodNotAllowed, ErrorMethodNotAllowed)
}
