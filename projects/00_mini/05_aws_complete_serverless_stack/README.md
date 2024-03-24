# AWS Complete Serverless Stack

This is a simple demonstration of how to create a serverless application using AWS: API Gateway + DynamoDB + Lambda.

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

For LocalStack, you can need to add to every command the `--endpoint-url=http://localhost:4566` flag (or define an alias `awslocal` for the AWS CLI with the same flag).

First you need to create a role for your lambda function. You can do this by running the following command:

```bash
aws iam create-role --role-name lambda-ex --assume-role-policy-document file://trust-policy.json
```

Then you need to attach the `AWSLambdaBasicExecutionRole` policy to the role:

```bash
aws iam attach-role-policy --role-name lambda-ex --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
```

Now you can build the application:

```bash
GOOS=linux GOARCH=amd64 go build -o ./build ./cmd/api/main.go
```

And then you can create a zip file with the application:

```bash
zip -jrm ./build/main.zip ./build/main
```

Now you can deploy the application to AWS or LocalStack:

```bash
awslocal lambda create-function --function-name aws-complete-serverless-stack --runtime go1.x --role arn:aws:iam::000000000000:role/lambda-role --handler main --zip-file fileb://./build/main.zip
```

Now you need to create a DynamoDB table:

```bash
aws dynamodb create-table \
    --table-name aws-complete-serverless-stack-users \
    --attribute-definitions AttributeName=email,AttributeType=S \
    --key-schema AttributeName=email,KeyType=HASH \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
```

And finally, you can create the API Gateway:

```bash
aws apigateway create-rest-api --name aws-complete-serverless-stack
```

Create any action that would use lambda integration:

```bash
aws apigateway put-method --rest-api-id <rest-api-id> --resource-id <resource-id> --http-method POST --authorization-type
```

And deploy the API:

```bash
aws apigateway create-deployment --rest-api-id <rest-api-id> --stage-name dev
```