package main

import (
	"fmt"
	"time"
)

func main() {
	SignUp("joe@example.com")

	time.Sleep(time.Second * 2)
}

func SignUp(email string) {
	go SaveUser(email)
	go SendNotification(email)
}

func SaveUser(email string) {
	fmt.Println("Saving user", email)

	// Takes a long time to process
	time.Sleep(time.Second)

	fmt.Println("User", email, "saved")
}

func SendNotification(email string) {
	fmt.Println("Sending notification to", email)

	// Takes a long time to process
	time.Sleep(time.Second)

	fmt.Println("Notification sent to", email)
}
