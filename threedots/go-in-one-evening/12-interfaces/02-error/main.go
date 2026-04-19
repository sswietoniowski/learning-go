package main

import "fmt"

type User struct {
	ID   int
	Name string
}

type UserNotFoundError struct {
	ID int
}

func (e UserNotFoundError) Error() string {
	return fmt.Sprintf("user not found: %d", e.ID)
}

type MemoryStorage struct {
	users []User
}

func (m MemoryStorage) FindUser(id int) (User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}

	return User{}, UserNotFoundError{ID: id}
}

func main() {
	storage := MemoryStorage{users: []User{
		{ID: 1, Name: "Eve"},
		{ID: 2, Name: "John"},
	}}

	eve, err := storage.FindUser(1)
	if err != nil {
		panic(err)
	}

	fmt.Println(eve)

	alice, err := storage.FindUser(3)
	if err != nil {
		if _, ok := err.(UserNotFoundError); ok {
			fmt.Println(err)
		} else {
			panic(err)
		}
	}

	fmt.Println(alice)
}
