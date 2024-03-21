package users

import "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"

type UsersRepository struct {
	tableName  string
	dynaClient dynamodbiface.DynamoDBAPI
}

func NewUsersRepository(tableName string, dynaClient dynamodbiface.DynamoDBAPI) *UsersRepository {
	return &UsersRepository{
		tableName:  tableName,
		dynaClient: dynaClient,
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

func (ur *UsersRepository) DeleteUser(email string) error {
	return nil
}
