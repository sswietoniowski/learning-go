package database

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
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

func (ur *UsersRepository) GetUserByEmail(email string) (*User, error) {
	return nil, nil
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
