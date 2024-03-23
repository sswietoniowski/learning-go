# AWS Complete Serverless Stack

This is a simple demonstration of how to create a serverless application using AWS: API Gateway + DynamoDB + Lambda.

- [AWS Complete Serverless Stack](#aws-complete-serverless-stack)
  - [Features](#features)
  - [Technologies](#technologies)
  - [Setup](#setup)

## Features

This application has the following features:

- TODO:.

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
