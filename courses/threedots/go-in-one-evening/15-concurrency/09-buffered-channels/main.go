package main

import (
	"errors"
	"fmt"
	"time"
)

func main() {
	payments := []int{100, 200, 50, 150, 205, 500, 400}
	r, err := CalculatePayments(payments)
	if err != nil {
		panic(err)
	}

	fmt.Println("Total paid:", r)
}

func CalculatePayments(payments []int) (int, error) {
	paymentsChan := make(chan int, 5)
	resultChan := make(chan int)

	go Aggregate(paymentsChan, resultChan)

	err := SendPayments(payments, paymentsChan)
	if err != nil {
		return 0, err
	}

	r := <-resultChan
	return r, nil
}

func Aggregate(payments <-chan int, result chan<- int) {
	total := 0

	for p := range payments {
		total += p
		time.Sleep(time.Millisecond * 5)
	}

	result <- total
}

func SendPayments(payments []int, requests chan<- int) error {
	timeout := time.After(time.Millisecond * 25)

	for _, p := range payments {
		select {
		case requests <- p:
		case <-timeout:
			return errors.New("timed out")
		}
	}

	close(requests)
	return nil
}
