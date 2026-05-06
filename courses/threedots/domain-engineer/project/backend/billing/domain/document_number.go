package domain

import "fmt"

type DocumentSeries struct {
	series string
}

func NewDocumentSeries(series string) (DocumentSeries, error) {
	if series == "" {
		return DocumentSeries{}, fmt.Errorf("document series must not be empty")
	}

	return DocumentSeries{
		series: series,
	}, nil
}

func (s DocumentSeries) IsZero() bool {
	return s.series == ""
}

func (s DocumentSeries) String() string {
	return s.series
}

var DocumentSeriesReceipt = DocumentSeries{"R"}

type DocumentNumber struct {
	series DocumentSeries
	number int
}

func NewDocumentNumber(series DocumentSeries, number int) (DocumentNumber, error) {
	if series.IsZero() {
		return DocumentNumber{}, fmt.Errorf("document series must not be empty")
	}

	if number <= 0 {
		return DocumentNumber{}, fmt.Errorf("document number must be greater than zero")
	}

	return DocumentNumber{
		series: series,
		number: number,
	}, nil
}

func (d DocumentNumber) IsZero() bool {
	return d.series.IsZero() && d.number == 0
}

func (d DocumentNumber) String() string {
	return fmt.Sprintf("%s-%08d", d.series, d.number)
}
