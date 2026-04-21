// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package common_test

import (
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eats/backend/common"
)

// TestStatus is a concrete enum type for testing
type TestStatus string

const (
	TestStatusPending   TestStatus = "pending"
	TestStatusActive    TestStatus = "active"
	TestStatusCompleted TestStatus = "completed"
)

func (TestStatus) Values() []string {
	return []string{
		string(TestStatusPending),
		string(TestStatusActive),
		string(TestStatusCompleted),
	}
}

func TestEnum_UnmarshalText(t *testing.T) {
	testCases := []struct {
		Name             string
		Input            []byte
		ExpectError      bool
		ExpectValue      string
		ExpectErrorMatch string
	}{
		{
			Name:        "valid_pending_value",
			Input:       []byte("pending"),
			ExpectError: false,
			ExpectValue: "pending",
		},
		{
			Name:        "valid_active_value",
			Input:       []byte("active"),
			ExpectError: false,
			ExpectValue: "active",
		},
		{
			Name:        "valid_completed_value",
			Input:       []byte("completed"),
			ExpectError: false,
			ExpectValue: "completed",
		},
		{
			Name:             "invalid_value",
			Input:            []byte("invalid"),
			ExpectError:      true,
			ExpectErrorMatch: `invalid enum value for common_test.TestStatus: 'invalid', expected values ["pending" "active" "completed"]`,
		},
		{
			Name:        "empty_string",
			Input:       []byte(""),
			ExpectError: false,
			ExpectValue: "",
		},
		{
			Name:             "case_sensitive",
			Input:            []byte("PENDING"),
			ExpectError:      true,
			ExpectErrorMatch: `invalid enum value for common_test.TestStatus: 'PENDING', expected values ["pending" "active" "completed"]`,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			var enum common.Enum[TestStatus]

			err := enum.UnmarshalText(tc.Input)

			if tc.ExpectError {
				require.Error(t, err)
				assert.Equal(t, tc.ExpectErrorMatch, err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.ExpectValue, enum.String())
			}
		})
	}
}

func TestEnum_MarshalText(t *testing.T) {
	testCases := []struct {
		Name     string
		Setup    func() common.Enum[TestStatus]
		Expected []byte
	}{
		{
			Name: "marshal_pending",
			Setup: func() common.Enum[TestStatus] {
				var enum common.Enum[TestStatus]
				_ = enum.UnmarshalText([]byte("pending"))
				return enum
			},
			Expected: []byte("pending"),
		},
		{
			Name: "marshal_active",
			Setup: func() common.Enum[TestStatus] {
				var enum common.Enum[TestStatus]
				_ = enum.UnmarshalText([]byte("active"))
				return enum
			},
			Expected: []byte("active"),
		},
		{
			Name: "marshal_completed",
			Setup: func() common.Enum[TestStatus] {
				var enum common.Enum[TestStatus]
				_ = enum.UnmarshalText([]byte("completed"))
				return enum
			},
			Expected: []byte("completed"),
		},
		{
			Name: "marshal_zero_value",
			Setup: func() common.Enum[TestStatus] {
				return common.Enum[TestStatus]{}
			},
			Expected: []byte(""),
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			enum := tc.Setup()

			result, err := enum.MarshalText()

			require.NoError(t, err)
			assert.Equal(t, tc.Expected, result)
		})
	}
}

func TestEnum_Scan(t *testing.T) {
	testCases := []struct {
		Name             string
		Input            any
		ExpectError      bool
		ExpectValue      string
		ExpectErrorMatch string
	}{
		{
			Name:        "scan_valid_string",
			Input:       "pending",
			ExpectError: false,
			ExpectValue: "pending",
		},
		{
			Name:        "scan_active_string",
			Input:       "active",
			ExpectError: false,
			ExpectValue: "active",
		},
		{
			Name:             "scan_invalid_enum_value",
			Input:            "invalid",
			ExpectError:      true,
			ExpectErrorMatch: `invalid enum value for common_test.TestStatus: 'invalid', expected values ["pending" "active" "completed"]`,
		},
		{
			Name:             "scan_int_type",
			Input:            42,
			ExpectError:      true,
			ExpectErrorMatch: "invalid type for enum: int, expected string",
		},
		{
			Name:             "scan_byte_slice",
			Input:            []byte("pending"),
			ExpectError:      true,
			ExpectErrorMatch: "invalid type for enum: []uint8, expected string",
		},
		{
			Name:             "scan_nil",
			Input:            nil,
			ExpectError:      true,
			ExpectErrorMatch: "invalid type for enum: <nil>, expected string",
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			var enum common.Enum[TestStatus]

			err := enum.Scan(tc.Input)

			if tc.ExpectError {
				require.Error(t, err)
				assert.Equal(t, tc.ExpectErrorMatch, err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.ExpectValue, enum.String())
			}
		})
	}
}

func TestEnum_Value(t *testing.T) {
	testCases := []struct {
		Name     string
		Setup    func() common.Enum[TestStatus]
		Expected driver.Value
	}{
		{
			Name: "value_pending",
			Setup: func() common.Enum[TestStatus] {
				var enum common.Enum[TestStatus]
				_ = enum.UnmarshalText([]byte("pending"))
				return enum
			},
			Expected: "pending",
		},
		{
			Name: "value_active",
			Setup: func() common.Enum[TestStatus] {
				var enum common.Enum[TestStatus]
				_ = enum.UnmarshalText([]byte("active"))
				return enum
			},
			Expected: "active",
		},
		{
			Name: "value_zero_value",
			Setup: func() common.Enum[TestStatus] {
				return common.Enum[TestStatus]{}
			},
			Expected: "",
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			enum := tc.Setup()

			result, err := enum.Value()

			require.NoError(t, err)
			assert.Equal(t, tc.Expected, result)
		})
	}
}

func TestEnum_String(t *testing.T) {
	testCases := []struct {
		Name     string
		Setup    func() common.Enum[TestStatus]
		Expected string
	}{
		{
			Name: "string_pending",
			Setup: func() common.Enum[TestStatus] {
				var enum common.Enum[TestStatus]
				_ = enum.UnmarshalText([]byte("pending"))
				return enum
			},
			Expected: "pending",
		},
		{
			Name: "string_active",
			Setup: func() common.Enum[TestStatus] {
				var enum common.Enum[TestStatus]
				_ = enum.UnmarshalText([]byte("active"))
				return enum
			},
			Expected: "active",
		},
		{
			Name: "string_zero_value",
			Setup: func() common.Enum[TestStatus] {
				return common.Enum[TestStatus]{}
			},
			Expected: "",
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			enum := tc.Setup()

			result := enum.String()

			assert.Equal(t, tc.Expected, result)
		})
	}
}

func TestEnum_Scan_UnmarshalText_Integration(t *testing.T) {
	t.Run("scan_and_retrieve_value", func(t *testing.T) {
		var enum common.Enum[TestStatus]

		// Scan from database
		err := enum.Scan("active")
		require.NoError(t, err)

		// Verify value can be retrieved
		value, err := enum.Value()
		require.NoError(t, err)
		assert.Equal(t, driver.Value("active"), value)

		// Verify string representation
		assert.Equal(t, "active", enum.String())
	})

	t.Run("marshal_and_unmarshal_roundtrip", func(t *testing.T) {
		var enum1 common.Enum[TestStatus]
		err := enum1.UnmarshalText([]byte("completed"))
		require.NoError(t, err)

		// Marshal to bytes
		marshaled, err := enum1.MarshalText()
		require.NoError(t, err)

		// Unmarshal to new enum
		var enum2 common.Enum[TestStatus]
		err = enum2.UnmarshalText(marshaled)
		require.NoError(t, err)

		// Should be equal
		assert.Equal(t, enum1.String(), enum2.String())
	})
}
