package main

import (
	"fmt"
	"time"
)

func main() {
	err := RunWithTimeout(func(errChan chan error) {
		time.Sleep(time.Second * 2)
		errChan <- nil
	})
	if err != nil {
		panic(err)
	}
}

func RunWithTimeout(f func(chan error)) error {
	errChan := make(chan error)
	go f(errChan)
	select {
	case err := <-errChan:
		return err
	case <-time.After(time.Second * 1):
		return fmt.Errorf("timeout")
	}
}
