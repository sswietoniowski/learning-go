package database

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

var (
	ErrorFailedToUnmarshalRecord = "failed to unmarshal record"
	ErrorFailedToFetchRecord     = "failed to fetch record"
	ErrorInvalidUserData         = "invalid user data"
	ErrorInvalidEmail            = "invalid email"
	ErrorCouldNotMarshalItem     = "could not marshal item"
	ErrorCouldNotDeleteItem      = "could not delete item"
	ErrorCouldNotDynamoPutItem   = "could not dynamo put item"
	ErrorUserAlreadyExists       = "user.User already exists"
	ErrorUserDoesNotExist        = "user.User does not exist"
)

type UsersRepository struct {
	dynaClient dynamodbiface.DynamoDBAPI
	tableName  string
}

const tableName = "aws-complete-serverless-stack-users"

func NewUsersRepository(session *session.Session) *UsersRepository {
	dynaClient := dynamodb.New(session)

	return &UsersRepository{
		dynaClient: dynaClient,
		tableName:  tableName,
	}
}

func (ur *UsersRepository) GetUsers() ([]User, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(ur.tableName),
	}

	result, err := ur.dynaClient.Scan(input)
	if err != nil {
		return nil, errors.New(ErrorFailedToFetchRecord)
	}

	users := make([]User, 0)
	if err := dynamodbattribute.UnmarshalListOfMaps(result.Items, users); err != nil {
		return nil, errors.New(ErrorFailedToUnmarshalRecord)
	}

	return users, nil
}

func (ur *UsersRepository) GetUserByEmail(email string) (*User, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
		TableName: aws.String(ur.tableName),
	}

	result, err := ur.dynaClient.GetItem(input)
	if err != nil {
		return nil, errors.New(ErrorFailedToFetchRecord)
	}

	user := new(User)
	if err := dynamodbattribute.UnmarshalMap(result.Item, user); err != nil {
		return nil, errors.New(ErrorFailedToUnmarshalRecord)
	}

	return user, nil
}

func (ur *UsersRepository) CreateUser(user *User) error {
	return nil
}

func (ur *UsersRepository) UpdateUser(user *User) error {
	return nil
}

func (ur *UsersRepository) DeleteUserByEmail(email string) error {
	return nil
}
