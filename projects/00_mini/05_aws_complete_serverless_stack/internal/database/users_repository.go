package database

import (
	"errors"
	"log"

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

func (ur *UsersRepository) GetAll() ([]User, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(ur.tableName),
	}

	result, err := ur.dynaClient.Scan(input)
	if err != nil {
		log.Printf("UsersRepository.GetAll - Error: %s", err)

		return nil, errors.New(ErrorFailedToFetchRecord)
	}

	users := new([]User)
	if err := dynamodbattribute.UnmarshalListOfMaps(result.Items, users); err != nil {
		log.Printf("UsersRepository.GetAll - Error: %s", err)

		return nil, errors.New(ErrorFailedToUnmarshalRecord)
	}

	return *users, nil
}

func (ur *UsersRepository) GetByEmail(email string) (*User, error) {
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
		log.Printf("UsersRepository.GetByEmail - Error: %s", err)

		return nil, errors.New(ErrorFailedToFetchRecord)
	}

	user := new(User)
	if err := dynamodbattribute.UnmarshalMap(result.Item, user); err != nil {
		log.Printf("UsersRepository.GetByEmail - Error: %s", err)

		return nil, errors.New(ErrorFailedToUnmarshalRecord)
	}

	return user, nil
}

func (ur *UsersRepository) Create(user *User) (*User, error) {
	if user == nil {
		return nil, errors.New(ErrorInvalidUserData)
	}

	existingUser, err := ur.GetByEmail(user.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New(ErrorUserAlreadyExists)
	}

	av, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		log.Printf("UsersRepository.Create - Error: %s", err)

		return nil, errors.New(ErrorCouldNotMarshalItem)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(ur.tableName),
	}

	_, err = ur.dynaClient.PutItem(input)
	if err != nil {
		log.Printf("UsersRepository.Create - Error: %s", err)

		return nil, errors.New(ErrorCouldNotDynamoPutItem)
	}

	return user, nil
}

func (ur *UsersRepository) Update(user *User) (*User, error) {
	if user == nil {
		return nil, errors.New(ErrorInvalidUserData)
	}

	databaseUser, err := ur.GetByEmail(user.Email)
	if err != nil {
		return nil, err
	}
	if databaseUser == nil {
		return nil, errors.New(ErrorUserDoesNotExist)
	}

	av, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		log.Printf("UsersRepository.Update - Error: %s", err)

		return nil, errors.New(ErrorCouldNotMarshalItem)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(ur.tableName),
	}

	_, err = ur.dynaClient.PutItem(input)
	if err != nil {
		log.Printf("UsersRepository.Update - Error: %s", err)

		return nil, errors.New(ErrorCouldNotDynamoPutItem)
	}

	return user, nil
}

func (ur *UsersRepository) DeleteByEmail(email string) (*User, error) {
	if email == "" {
		return nil, errors.New(ErrorInvalidEmail)
	}

	databaseUser, err := ur.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
		TableName: aws.String(ur.tableName),
	}

	_, err = ur.dynaClient.DeleteItem(input)
	if err != nil {
		log.Printf("UsersRepository.DeleteByEmail - Error: %s", err)

		return nil, errors.New(ErrorCouldNotDeleteItem)
	}

	return databaseUser, nil
}
