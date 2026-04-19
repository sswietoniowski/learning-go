package main

import (
	"fmt"
	"time"
)

func main() {
	emails := []string{
		"alice@example.com",
		"kate@example.com",
		"joe@example.com",
		"rob@example.com",
		"patrick@example.com",
	}
	SendNewsletter(emails)
}

func SendNewsletter(emails []string) {
	done := make(chan bool)

	for _, email := range emails {
		go SendNewsletterToEmail(email, done)
	}

	for i := 0; i < len(emails); i++ {
		<-done
	}

	fmt.Println("All newsletters sent")
}

func SendNewsletterToEmail(email string, done chan bool) {
	fmt.Println("Sending newsletter to", email)

	time.Sleep(time.Second)

	done <- true
}
