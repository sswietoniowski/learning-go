// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eats/backend/billing/domain"
)

func TestNewDocumentSeries_EmptyRejected(t *testing.T) {
	_, err := domain.NewDocumentSeries("")

	assert.ErrorContains(t, err, "document series must not be empty")
}

func TestNewDocumentSeries_Valid(t *testing.T) {
	series, err := domain.NewDocumentSeries("R")

	require.NoError(t, err)
	assert.Equal(t, "R", series.String())
}

func TestDocumentSeries_IsZero(t *testing.T) {
	assert.True(t, domain.DocumentSeries{}.IsZero())

	series, err := domain.NewDocumentSeries("R")
	require.NoError(t, err)

	assert.False(t, series.IsZero())
}

func TestDocumentSeriesReceipt(t *testing.T) {
	assert.Equal(t, "R", domain.DocumentSeriesReceipt.String())
}

func TestNewDocumentNumber_ZeroSeriesRejected(t *testing.T) {
	_, err := domain.NewDocumentNumber(domain.DocumentSeries{}, 1)

	assert.ErrorContains(t, err, "document series must not be empty")
}

func TestNewDocumentNumber_NonPositiveNumberRejected(t *testing.T) {
	series, err := domain.NewDocumentSeries("R")
	require.NoError(t, err)

	tests := []struct {
		name   string
		number int
	}{
		{name: "zero", number: 0},
		{name: "negative", number: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewDocumentNumber(series, tt.number)

			assert.ErrorContains(t, err, "document number must be greater than zero")
		})
	}
}

func TestNewDocumentNumber_Valid(t *testing.T) {
	series, err := domain.NewDocumentSeries("R")
	require.NoError(t, err)

	num, err := domain.NewDocumentNumber(series, 42)

	require.NoError(t, err)
	assert.False(t, num.IsZero())
}

func TestDocumentNumber_IsZero(t *testing.T) {
	assert.True(t, domain.DocumentNumber{}.IsZero())

	series, err := domain.NewDocumentSeries("R")
	require.NoError(t, err)

	num, err := domain.NewDocumentNumber(series, 1)
	require.NoError(t, err)

	assert.False(t, num.IsZero())
}

func TestDocumentNumber_String(t *testing.T) {
	series, err := domain.NewDocumentSeries("R")
	require.NoError(t, err)

	num, err := domain.NewDocumentNumber(series, 42)
	require.NoError(t, err)

	assert.Equal(t, "R-00000042", num.String())
}
