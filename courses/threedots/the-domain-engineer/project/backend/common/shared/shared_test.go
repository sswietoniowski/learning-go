// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package shared

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"eats/backend/common"
)

func TestNewAddress_ValidAddress(t *testing.T) {
	countryCode := MustNewCountryCode("US")

	addr, err := NewAddress("123 Main St", "Apt 4", "12345", "New York", countryCode)

	require.NoError(t, err)
	require.Equal(t, "123 Main St", addr.Line1())
	require.Equal(t, "Apt 4", addr.Line2())
	require.Equal(t, "12345", addr.PostalCode())
	require.Equal(t, "New York", addr.City())
	require.Equal(t, "US", addr.CountryCode().Code())
}

func TestNewAddress_ValidAddressWithoutLine2(t *testing.T) {
	countryCode := MustNewCountryCode("US")

	addr, err := NewAddress("123 Main St", "", "12345", "New York", countryCode)

	require.NoError(t, err)
	require.Equal(t, "123 Main St", addr.Line1())
	require.Equal(t, "", addr.Line2())
}

func TestNewAddress_MissingLine1(t *testing.T) {
	countryCode := MustNewCountryCode("US")

	_, err := NewAddress("", "Apt 4", "12345", "New York", countryCode)

	requireAddressError(t, err, "address-line1-required")
}

func TestNewAddress_MissingPostalCode(t *testing.T) {
	countryCode := MustNewCountryCode("US")

	_, err := NewAddress("123 Main St", "Apt 4", "", "New York", countryCode)

	requireAddressError(t, err, "address-postal-code-required")
}

func TestNewAddress_MissingCity(t *testing.T) {
	countryCode := MustNewCountryCode("US")

	_, err := NewAddress("123 Main St", "Apt 4", "12345", "", countryCode)

	requireAddressError(t, err, "address-city-required")
}

func TestNewAddress_MissingCountryCode(t *testing.T) {
	_, err := NewAddress("123 Main St", "Apt 4", "12345", "New York", CountryCode{})

	requireAddressError(t, err, "address-country-code-required")
}

func TestNewAddress_MultipleFieldsMissing(t *testing.T) {
	_, err := NewAddress("", "", "", "", CountryCode{})

	var domainErr common.Error
	require.True(t, errors.As(err, &domainErr))
	require.Equal(t, "invalid-address", domainErr.ErrorSlug)
	require.Len(t, domainErr.Details, 4)
}

func requireAddressError(t *testing.T, err error, expectedSlug string) {
	t.Helper()

	var domainErr common.Error
	require.True(t, errors.As(err, &domainErr), "expected common.Error, got %T", err)
	require.Equal(t, "invalid-address", domainErr.ErrorSlug)

	found := false
	for _, d := range domainErr.Details {
		if d.ErrorSlug == expectedSlug {
			found = true
			break
		}
	}
	require.True(t, found, "expected detail with slug %q, got %v", expectedSlug, domainErr.Details)
}

func TestAddress_MarshalJSON(t *testing.T) {
	countryCode := MustNewCountryCode("US")

	addr, err := NewAddress("123 Main St", "Apt 4", "12345", "New York", countryCode)
	require.NoError(t, err)

	data, err := json.Marshal(addr)
	require.NoError(t, err)

	expected, err := json.Marshal(map[string]any{
		"line_1":       "123 Main St",
		"line_2":       "Apt 4",
		"postal_code":  "12345",
		"city":         "New York",
		"country_code": "US",
	})
	require.NoError(t, err)

	require.JSONEq(t, string(expected), string(data))
}

func TestAddress_Scan(t *testing.T) {
	input, err := json.Marshal(map[string]any{
		"line_1":       "123 Main St",
		"line_2":       "Apt 4",
		"postal_code":  "12345",
		"city":         "New York",
		"country_code": "US",
	})
	require.NoError(t, err)

	var addr Address
	err = addr.Scan(string(input))

	require.NoError(t, err)
	require.Equal(t, "123 Main St", addr.Line1())
	require.Equal(t, "Apt 4", addr.Line2())
	require.Equal(t, "12345", addr.PostalCode())
	require.Equal(t, "New York", addr.City())
	require.Equal(t, "US", addr.CountryCode().Code())
}
