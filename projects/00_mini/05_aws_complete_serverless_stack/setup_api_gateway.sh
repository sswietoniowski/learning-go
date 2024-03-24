#!/bin/bash

# Set the AWS region and profile
region=eu-north-1 # adjust to your region
endpoint_url=http://localhost:4566 # adjust to your localstack or AWS endpoint

# Set lambda function name and IAM role
lambda_name="aws-complete-serverless-stack"
role_arn=$(aws iam get-role --role-name lambda-ex --query 'Role.Arn' --output text --endpoint-url=$endpoint_url)

# Create the API Gateway
rest_api_name="aws-complete-serverless-stack"
description="API for retrieving users' data"
rest_api_id=$(aws apigateway create-rest-api --name $rest_api_name --description "$description" --query 'id' --output text --region $region --endpoint-url=$endpoint_url)
resource_id=$(aws apigateway get-resources --rest-api-id $rest_api_id --query 'items[0].id' --output text --region $region --endpoint-url=$endpoint_url)
aws apigateway put-method --rest-api-id $rest_api_id --resource-id $resource_id --http-method GET --authorization-type NONE --region $region --endpoint-url=$endpoint_url
aws apigateway put-method --rest-api-id $rest_api_id --resource-id $resource_id --http-method POST --authorization-type NONE --region $region --endpoint-url=$endpoint_url
aws apigateway put-method --rest-api-id $rest_api_id --resource-id $resource_id --http-method PUT --authorization-type NONE --region $region --endpoint-url=$endpoint_url
aws apigateway put-method --rest-api-id $rest_api_id --resource-id $resource_id --http-method DELETE --authorization-type NONE --region $region --endpoint-url=$endpoint_url
uri="arn:aws:apigateway:$region:lambda:path/2015-03-31/functions/arn:aws:lambda:$region:$(aws sts get-caller-identity --query 'Account' --output text --endpoint-url=$endpoint_url):function:$lambda_name/invocations" 
aws apigateway put-integration --rest-api-id $rest_api_id --resource-id $resource_id --http-method GET --type AWS_PROXY --integration-http-method GET --uri $uri --region $region --endpoint-url=$endpoint_url
aws apigateway put-integration --rest-api-id $rest_api_id --resource-id $resource_id --http-method POST --type AWS_PROXY --integration-http-method POST --uri $uri --region $region --endpoint-url=$endpoint_url
aws apigateway put-integration --rest-api-id $rest_api_id --resource-id $resource_id --http-method PUT --type AWS_PROXY --integration-http-method PUT --uri $uri --region $region --endpoint-url=$endpoint_url
aws apigateway put-integration --rest-api-id $rest_api_id --resource-id $resource_id --http-method DELETE --type AWS_PROXY --integration-http-method DELETE --uri $uri --region $region --endpoint-url=$endpoint_url
aws apigateway create-deployment --rest-api-id $rest_api_id --stage-name prod --region $region --endpoint-url=$endpoint_url
#endpoint_url="https://$(aws apigateway get-rest-apis --query 'items[0].id' --output text --region $region --endpoint-url=$endpoint_url).execute-api.$region.amazonaws.com/prod"

# Update the Lambda function with the API Gateway trigger
#aws lambda add-permission --function-name $lambda_name --statement-id apigateway-test-2 --action lambda:InvokeFunction --principal apigateway.amazonaws.com --source-arn "arn:aws:execute-api:$region:$(aws sts get-caller-identity --query 'Account' --output text --endpoint-url=$endpoint_url):$rest_api_id/*/GET/" --region $region --endpoint-url=$endpoint_url
#aws lambda create-event-source-mapping --function-name $lambda_name --batch-size 1 --event-source-arn "arn:aws:execute-api:$region:$(aws sts get-caller-identity --query 'Account' --output text --endpoint-url=$endpoint_url):$rest_api_id/*/GET/" --region $region --endpoint-url=$endpoint_url --