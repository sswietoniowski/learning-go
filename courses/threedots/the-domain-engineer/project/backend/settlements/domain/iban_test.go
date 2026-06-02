// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eats/backend/settlements/domain"
)

func TestNewIBAN_Valid(t *testing.T) {
	testCases := []struct {
		name string
		iban string
	}{
		{name: "minimum_length_norway", iban: "NO9386011117947"},                   // 15 chars
		{name: "germany_22_chars", iban: "DE89370400440532013000"},                 // 22 chars
		{name: "uk_22_chars", iban: "GB29NWBK60161331926819"},                      // 22 chars
		{name: "maximum_length_34_chars", iban: "MT84MALT011000012345MTLCAST001S"}, // 31 chars
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			iban, err := domain.NewIBAN(tc.iban)
			require.NoError(t, err)
			assert.Equal(t, tc.iban, iban.String())
			assert.False(t, iban.IsZero())
		})
	}
}

func TestNewIBAN_Invalid(t *testing.T) {
	testCases := []struct {
		name string
		iban string
	}{
		{name: "empty", iban: ""},
		{name: "too_short_below_15_chars", iban: "DE893704004"},
		{name: "too_long_above_34_chars", iban: "DE893704004405320130001234567890123456"},
		{name: "country_code_not_letters", iban: "12893704004405320130000"},
		{name: "country_code_partially_numeric", iban: "1E893704004405320130000"},
		{name: "non_alphanumeric_body", iban: "DE89370400440532!13000"},
		{name: "spaces_in_body", iban: "DE89 3704 0044 0532 1300"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := domain.NewIBAN(tc.iban)
			require.Error(t, err)
		})
	}
}

func TestIBAN_IsZero(t *testing.T) {
	var zero domain.IBAN
	assert.True(t, zero.IsZero())

	iban := domain.UnmarshalIBAN("DE89370400440532013000")
	assert.False(t, iban.IsZero())
}

func TestIBAN_String(t *testing.T) {
	const value = "DE89370400440532013000"

	iban, err := domain.NewIBAN(value)
	require.NoError(t, err)

	assert.Equal(t, value, iban.String())
	assert.True(t, strings.HasPrefix(iban.String(), "DE"))
}
