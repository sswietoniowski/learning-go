package main

import (
	"fmt"
	"time"
)

func main() {
	metrics := &Metrics{}

	err := Execute(func() error {
		fmt.Println("Executing...")
		return nil
	}, metrics)

	fmt.Println(err)
}

func Execute(f func() error, metrics *Metrics) (err error) {
	metrics.StoreExecution()

	defer func() {
		if err != nil {
			metrics.StoreFailure()
		} else {
			metrics.StoreSuccess()
		}
	}()

	return f()
}

type Metrics struct {
	execution []time.Time
	success   []time.Time
	failure   []time.Time
}

func (m *Metrics) StoreExecution() {
	m.execution = append(m.execution, time.Now())
}

func (m *Metrics) StoreSuccess() {
	m.success = append(m.success, time.Now())
}

func (m *Metrics) StoreFailure() {
	m.failure = append(m.failure, time.Now())
}
