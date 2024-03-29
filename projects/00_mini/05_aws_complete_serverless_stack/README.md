# AWS Complete Serverless Stack

This is a simple demonstration of how to create a serverless application using AWS: API Gateway, DynamoDB, and Lambda.

- [AWS Complete Serverless Stack](#aws-complete-serverless-stack)
  - [Features](#features)
  - [Technologies](#technologies)
  - [Setup](#setup)

## Features

This application has the following features.

Functional:

- store, retrieve, update, and delete users' data in a **DynamoDB** table,
- expose the CRUD operations on users' data through an **API Gateway** and **Lambda**.

Non-functional:

- use **AWS** cloud or **LocalStack** to run the application,
- use the **AWS SDK for Go** and **AWS Lambda for Go** to interact with the AWS services.

## Technologies

The application is built using the following technologies, libraries, frameworks, and tools:

- [Go](https://golang.org/),
- [AWS Lambda for Go](https://github.com/aws/aws-lambda-go),
- [AWS SDK for Go](https://github.com/aws/aws-sdk-go),
- [LocalStack](https://github.com/localstack/localstack) [📖](https://docs.localstack.cloud/user-guide/integrations/aws-cli/#localstack-aws-cli-awslocal).

## Setup

To run this application, you need to have the following installed on your system:

- [Go](https://golang.org/),
- [AWS CLI](https://aws.amazon.com/cli/).

Of course, you also need an AWS account or a **LocalStack** instance running (requires Docker).

To install LocalStack, you can run the following command:

```bash
pip install localstack
```

To start LocalStack, you can run the following command:

```bash
localstack start -d
```

Alternatively, you can use the `docker-compose` file provided in this repository:

```bash
docker-compose up -d
```

If you want to use the AWS CLI with LocalStack, please follow [this](https://docs.localstack.cloud/user-guide/integrations/aws-cli/#localstack-aws-cli-awslocal) guide.

For LocalStack, you need to add to every command the `--endpoint-url=http://localhost:4566` flag (or define an alias `awslocal` for the AWS CLI with the same flag).

I've tested everything under LocalStack, and if you want to deploy this application to real AWS, you need to modify `main.go` and replace the existing code for the DynamoDB client with the following code:

```go
awsSession, err := session.NewSession(&aws.Config{
  Region:      aws.String(region),
})
```

First, you need to create a role for your lambda function. You can do this by running the following command:

```bash
aws iam create-role --role-name lambda-ex --assume-role-policy-document file://trust-policy.json
```

Then you need to attach the `AWSLambdaBasicExecutionRole` policy to the role:

```bash
aws iam attach-role-policy --role-name lambda-ex --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
```

To allow access to the DynamoDB table, you need to attach the `AmazonDynamoDBFullAccess` policy to the role (you can also create a custom policy with the required permissions):

```bash
aws iam attach-role-policy --role-name lambda-ex --policy-arn arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess
```

Now you can build the application:

```bash
GOOS=linux GOARCH=amd64 go build -o ./build ./cmd/api/main.go
```

Then you can create a zip file with the application:

```bash
zip -jrm ./build/main.zip ./build/main
```

Now you can deploy the application to AWS or LocalStack:

```bash
aws lambda create-function --function-name aws-complete-serverless-stack --runtime go1.x --role arn:aws:iam::PUT_YOUR_ID_HERE:role/lambda-ex --handler main --zip-file fileb://./build/main.zip --timeout 900
```

Now, you need to create a DynamoDB table:

```bash
aws dynamodb create-table \
    --table-name aws-complete-serverless-stack-users \
    --attribute-definitions AttributeName=email,AttributeType=S \
    --key-schema AttributeName=email,KeyType=HASH \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
```

You can also add some data to the table:

```bash
aws dynamodb put-item --table-name aws-complete-serverless-stack-users --item '{"email": {"S": "jdoe@unknown.com"}}'
aws dynamodb put-item --table-name aws-complete-serverless-stack-users --item '{"email": {"S": "asmith@unknown.com"}}'
```

To retrieve the data from the table, you can run the following command:

```bash
aws dynamodb scan --table-name aws-complete-serverless-stack-users
```

You can invoke the lambda function using GET:

```bash
aws lambda invoke --function-name aws-complete-serverless-stack --cli-binary-format raw-in-base64-out --payload file://get-request.json response.json
```

Where `get-request.json` is:

```json
{
  "httpMethod": "GET",
  "path": "/users",
  "queryStringParameters": {
    "email": "jdoe@unknown.com"
  }
}
```

Alternatively, you can use the following command:

```bash
aws lambda invoke --function-name aws-complete-serverless-stack --payload "{\"httpMethod\": \"GET\", \"path\": \"/users\", \"queryStringParameters\": {}}" --endpoint-url=http://localhost:4566 response.json
```

You can also invoke the lambda function using POST:

```bash
aws lambda invoke --function-name aws-complete-serverless-stack --cli-binary-format raw-in-base64-out --payload file://post-request.json response.json
```

Where `post-request.json` is:

```json
{
  "httpMethod": "POST",
  "path": "/users",
  "body": "{\"email\": \"asmith@unknown.com\"}"
}
```

To see logs from the lambda function, you can run the following command:

```bash
aws logs tail /aws/lambda/aws-complete-serverless-stack --follow
```

And finally, you can create the API Gateway, to simplify that part I've create a script that you can run (adjust it to your needs):

```bash
./setup_api_gateway.sh
```

Great explanation of what this script is doing can be found [here](https://conor-aspell.medium.com/creating-an-api-gateway-using-the-aws-cli-dbed2849a773).

If that succeeded, you can list all the APIs:

```bash
aws apigateway get-rest-apis
```

to remove the API Gateway:

```bash
aws apigateway delete-rest-api --rest-api-id <rest-api-id>
```

You can retrieve endpoint URL:

```bash
region=$(aws configure get region)
endpoint_url=http://localhost:4566
echo "https://$(aws apigateway get-rest-apis --query 'items[0].id' --output text --region $region --endpoint-url=$endpoint_url).execute-api.$region.amazonaws.com/prod"
```

Now you can test the API Gateway (LocalStack):

```bash
rest_api_id=$(aws apigateway get-rest-apis --query 'items[0].id' --output text --region $region --endpoint-url=$endpoint_url)
stage_name=prod
curl -X GET -H "Content-Type: application/json" http://localhost:4566/restapis/$rest_api_id/$stage_name/_user_request_/
```

If it works, you should see something like this:

```json
[
  { "firstName": "", "lastName": "", "email": "jdoe@unknown.com" },
  { "firstName": "", "lastName": "", "email": "asmith@unknown.com" }
]
```

And it means that you have successfully deployed the complete serverless stack application using AWS :-).
