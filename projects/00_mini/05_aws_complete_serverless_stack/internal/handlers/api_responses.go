package handlers

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
)

func apiResponse(statusCode int, body interface{}) (*events.APIGatewayProxyResponse, error) {
	resp := events.APIGatewayProxyResponse{
		Headers:    map[string]string{"Content-Type": "application/json"},
		StatusCode: statusCode,
	}

	if body != nil {
		stringBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}

		resp.Body = string(stringBody)
	}

	return &resp, nil
}

func errorResponse(statusCode int, err error) (*events.APIGatewayProxyResponse, error) {
	return apiResponse(statusCode, aws.String(err.Error()))
}
