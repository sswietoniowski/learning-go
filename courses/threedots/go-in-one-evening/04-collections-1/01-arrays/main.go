package main

import (
	"fmt"
)

var roles = [4]string{
	"guest",
	"user",
	"moderator",
	"admin",
}

func main() {
	fmt.Println(roles)
}
