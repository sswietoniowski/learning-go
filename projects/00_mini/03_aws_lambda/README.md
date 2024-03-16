# AWS Lambda

This is a simple demonstration of how to create a serverless application using AWS Lambda.

- [AWS Lambda](#aws-lambda)
  - [Features](#features)
  - [Technologies](#technologies)
  - [Setup](#setup)

## Features

This application has the following features:

- just returns a simple message.

## Technologies

The application is built using the following technologies, libraries, frameworks, and tools:

- [Go](https://golang.org/),
- [AWS Lambda for Go](https://github.com/aws/aws-lambda-go).

## Setup

To run this application, you need to have the following installed on your system:

- [Go](https://golang.org/),
- [AWS CLI](https://aws.amazon.com/cli/).

Of course, you also need an AWS account.

First you need to create a role for your Lambda function. You can do this by running the following command:

```bash
aws iam create-role --role-name lambda-ex --assume-role-policy-document file://trust-policy.json
```

Then you need to attach the `AWSLambdaBasicExecutionRole` policy to the role:

```bash
aws iam attach-role-policy --role-name lambda-ex --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
```

Now you can build the application:

```powershell
set GOOS=linux
set GOARCH=amd64
go build -o bootstrap main.go
```

And then you can create a zip file with the application:

```powershell
Compress-Archive -Path .\bootstrap -DestinationPath .\bootstrap.zip
```

Then you can create the Lambda function (you need to replace the role ARN with the one you created):

```bash
aws lambda create-function --function-name aws-lambda --zip-file fileb://./bootstrap.zip --handler bootstrap --runtime provided.al2 --role arn:aws:iam::PUT_YOUR_ID_HERE:role/lambda-ex
```

To list the functions, you can run the following command:

```bash
aws lambda list-functions
```

Finally, you can invoke the function:

```bash
aws lambda invoke --invocation-type RequestResponse --function-name aws-lambda --cli-binary-format raw-in-base64-out --payload '{"What is your name?":"John Doe","What is your year of birth?":2000}' response.json
```

In the `response.json` file, you should see something like this:

```json
{
  "message": "Hello, John Doe! You are 24 years old."
}
```

To remove the function, you can run the following command:

```bash
aws lambda delete-function --function-name aws-lambda
```

To remove the role, you can run the following command:

```bash
aws iam detach-role-policy --role-name lambda-ex --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
aws iam delete-role --role-name lambda-ex
```