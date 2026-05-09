// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package shared_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eats/backend/common/shared"
)

func TestNewTaxID(t *testing.T) {
	id, err := shared.NewTaxID("ABC-123")

	require.NoError(t, err)
	assert.Equal(t, "ABC-123", id.String())
	assert.False(t, id.IsZero())
}

func TestNewTaxID_Empty(t *testing.T) {
	_, err := shared.NewTaxID("")

	require.Error(t, err)
}

func TestNewTaxID_TooShort(t *testing.T) {
	_, err := shared.NewTaxID("ab")

	require.Error(t, err)
}

func TestNewTaxID_InvalidChars(t *testing.T) {
	_, err := shared.NewTaxID("abc!@#")

	require.Error(t, err)
}

func TestTaxID_ZeroValue(t *testing.T) {
	var id shared.TaxID

	assert.True(t, id.IsZero())
	assert.Equal(t, "", id.String())
}

func TestTaxID_ScanString(t *testing.T) {
	var id shared.TaxID

	err := id.Scan("ABC-123")
	require.NoError(t, err)
	assert.Equal(t, "ABC-123", id.String())
	assert.False(t, id.IsZero())
}

func TestTaxID_ScanNil(t *testing.T) {
	var id shared.TaxID

	err := id.Scan(nil)
	require.NoError(t, err)
	assert.True(t, id.IsZero())
}

func TestTaxID_Value(t *testing.T) {
	id, err := shared.NewTaxID("ABC-123")
	require.NoError(t, err)

	v, err := id.Value()
	require.NoError(t, err)
	assert.Equal(t, "ABC-123", v)
}

func TestTaxID_Value_Zero(t *testing.T) {
	var id shared.TaxID

	v, err := id.Value()
	require.NoError(t, err)
	assert.Nil(t, v)
}

func TestTaxID_InSharedTypes(t *testing.T) {
	found := false
	for _, st := range shared.SharedTypes {
		if _, ok := st.(shared.TaxID); ok {
			found = true
			break
		}
	}
	require.True(t, found, "TaxID{} should be in SharedTypes slice")
}
