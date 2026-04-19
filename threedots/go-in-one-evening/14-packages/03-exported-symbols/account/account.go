package account

import "errors"

var (
	ErrEmptyEmail    = errors.New("empty email")
	ErrEmptyPassword = errors.New("empty password")
)

type Account struct {
	email    string
	password string
}

func (a Account) Login(email string, password string) bool {
	return a.email == email && a.password == password
}

func New(email string, password string) (Account, error) {
	if email == "" {
		return Account{}, ErrEmptyEmail
	}

	if password == "" {
		return Account{}, ErrEmptyPassword
	}

	return Account{
		email:    email,
		password: password,
	}, nil
}
