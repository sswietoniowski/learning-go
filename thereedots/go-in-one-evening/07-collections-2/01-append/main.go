package main

var users []string

func main() {
	AddUser("Alice")
	AddUser("Bob")
}

func AddUser(name string) {
	users = append(users, name)
}
