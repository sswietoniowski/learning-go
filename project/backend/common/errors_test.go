// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestError_WithDetails_AppendsToExisting(t *testing.T) {
	initialDetails := []ErrorDetails{
		{EntityType: "order", EntityID: "123", ErrorSlug: "not-found", Message: "order not found"},
	}

	err := Error{
		HttpErrorCode: 400,
		PublicError:   "validation error",
		ErrorSlug:     "validation-failed",
		Details:       initialDetails,
	}

	newDetails := []ErrorDetails{
		{EntityType: "item", EntityID: "456", ErrorSlug: "invalid", Message: "item invalid"},
		{EntityType: "item", EntityID: "789", ErrorSlug: "missing", Message: "item missing"},
	}

	result := err.WithDetails(newDetails)

	require.Len(t, result.Details, 3)
	require.Equal(t, "order", result.Details[0].EntityType)
	require.Equal(t, "123", result.Details[0].EntityID)
	require.Equal(t, "item", result.Details[1].EntityType)
	require.Equal(t, "456", result.Details[1].EntityID)
	require.Equal(t, "item", result.Details[2].EntityType)
	require.Equal(t, "789", result.Details[2].EntityID)
}

func TestError_WithDetails_WorksWithEmptyExisting(t *testing.T) {
	err := Error{
		HttpErrorCode: 400,
		PublicError:   "validation error",
		ErrorSlug:     "validation-failed",
		Details:       nil,
	}

	newDetails := []ErrorDetails{
		{EntityType: "item", EntityID: "456", ErrorSlug: "invalid", Message: "item invalid"},
	}

	result := err.WithDetails(newDetails)

	require.Len(t, result.Details, 1)
	require.Equal(t, "item", result.Details[0].EntityType)
}
