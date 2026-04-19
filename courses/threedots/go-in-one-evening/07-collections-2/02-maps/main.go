package main

import "fmt"

var (
	Stats = map[string]int{}
)

func CreateUser(user string) {
	fmt.Println("Creating user", user)
	if _, ok := Stats["create"]; !ok {
		Stats["create"] = 1
	} else {
		Stats["create"]++
	}
}

func UpdateUser(user string) {
	fmt.Println("Updating user", user)
	if _, ok := Stats["update"]; !ok {
		Stats["update"] = 1
	} else {
		Stats["update"]++
	}
}

func PurgeStats() {
	Stats["create"] = 0
	Stats["update"] = 0
}
