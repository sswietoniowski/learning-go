package main

import (
	"fmt"
	"time"
)

type DateRange struct {
	start time.Time
	end   time.Time
}

func NewDateRange(start, end time.Time) (DateRange, error) {
	if start.IsZero() || end.IsZero() {
		return DateRange{}, fmt.Errorf("start and end dates must be valid")
	}
	if end.Before(start) {
		return DateRange{}, fmt.Errorf("end date must be after start date")
	}
	return DateRange{start: start, end: end}, nil
}

func (d DateRange) Hours() float64 {
	return d.end.Sub(d.start).Hours()
}

func main() {
	lifetime, err := NewDateRange(
		time.Date(1815, 12, 10, 0, 0, 0, 0, time.UTC),
		time.Date(1852, 11, 27, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(lifetime.Hours())

	travelInTime, err := NewDateRange(
		time.Date(1852, 11, 27, 0, 0, 0, 0, time.UTC),
		time.Date(1815, 12, 10, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(travelInTime.Hours())
}
