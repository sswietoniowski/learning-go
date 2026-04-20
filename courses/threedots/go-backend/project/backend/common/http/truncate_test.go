// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package http

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTruncateBodyForLog_PassThrough(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty_body", input: "", want: ""},
		{name: "short_string", input: "hello world", want: "hello world"},
		{name: "short_JSON_object", input: `{"key":"value"}`, want: `{"key":"value"}`},
		{name: "short_JSON_array", input: `[1,2,3]`, want: `[1,2,3]`},
		{name: "JSON_null", input: `null`, want: `null`},
		{name: "JSON_number", input: `42`, want: `42`},
		{name: "JSON_string", input: `"hello"`, want: `"hello"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, truncateBodyForLog(tt.input))
		})
	}
}

func TestTruncateBodyForLog_StringMode(t *testing.T) {
	t.Run("long_plain_text", func(t *testing.T) {
		input := strings.Repeat("abcdefghij", 100) // 1000 bytes
		result := truncateBodyForLog(input)

		assert.Contains(t, result, "bytes truncated")
		assert.True(t, strings.HasPrefix(result, "abcdefghij"))
		assert.True(t, strings.HasSuffix(result, "abcdefghij"))
		assert.Less(t, len(result), len(input))
	})

	t.Run("long_invalid_JSON", func(t *testing.T) {
		input := "{not valid json " + strings.Repeat("x", 600)
		result := truncateBodyForLog(input)

		assert.Contains(t, result, "bytes truncated")
		assert.True(t, strings.HasPrefix(result, "{not valid json"))
	})
}

func TestTruncateBodyForLog_JSONArray(t *testing.T) {
	t.Run("array_at_limit_unchanged", func(t *testing.T) {
		arr := make([]map[string]string, maxArrayLogItems)
		for i := range arr {
			arr[i] = map[string]string{"id": fmt.Sprintf("item-%d", i), "data": strings.Repeat("x", 80)}
		}
		input, err := json.Marshal(arr)
		require.NoError(t, err)
		require.Greater(t, len(string(input)), maxBodyLogBytes, "input must exceed threshold to exercise truncation path")

		result := truncateBodyForLog(string(input))
		assert.Equal(t, string(input), result)
	})

	t.Run("large_array_truncated", func(t *testing.T) {
		arr := make([]map[string]string, 20)
		for i := range arr {
			arr[i] = map[string]string{"id": fmt.Sprintf("item-%d", i), "data": strings.Repeat("x", 30)}
		}
		input, err := json.Marshal(arr)
		require.NoError(t, err)

		result := truncateBodyForLog(string(input))
		require.True(t, json.Valid([]byte(result)), "result must be valid JSON, got: %s", result)

		var parsed []any
		require.NoError(t, json.Unmarshal([]byte(result), &parsed))

		// 5 kept + 1 marker + 1 last = 7
		assert.Len(t, parsed, maxArrayLogItems+1)

		// First items preserved.
		for i := 0; i < maxArrayLogItems-1; i++ {
			m, ok := parsed[i].(map[string]any)
			require.True(t, ok)
			assert.Equal(t, fmt.Sprintf("item-%d", i), m["id"])
		}

		// Marker present.
		marker, ok := parsed[maxArrayLogItems-1].(map[string]any)
		require.True(t, ok)
		assert.Contains(t, marker["..."], "items truncated")

		// Last item preserved.
		last, ok := parsed[maxArrayLogItems].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "item-19", last["id"])
	})
}

func TestTruncateBodyForLog_JSONObject(t *testing.T) {
	t.Run("large_object_stays_valid_JSON", func(t *testing.T) {
		obj := make(map[string]string)
		for i := 0; i < 50; i++ {
			obj[fmt.Sprintf("key_%03d", i)] = strings.Repeat("v", 20)
		}
		input, err := json.Marshal(obj)
		require.NoError(t, err)
		require.Greater(t, len(string(input)), maxBodyLogBytes)

		result := truncateBodyForLog(string(input))
		assert.True(t, json.Valid([]byte(result)), "large JSON object must remain valid JSON")
	})
}

func TestTruncateBodyForLog_JSONWithLargeItems(t *testing.T) {
	t.Run("array_with_large_items_stays_valid_JSON", func(t *testing.T) {
		// Simulates a restaurant onboarding payload with large menu items.
		items := make([]map[string]string, 10)
		for i := range items {
			items[i] = map[string]string{
				"name":  fmt.Sprintf("Menu item %d with a long description %s", i, strings.Repeat("x", 50)),
				"price": "16.30",
				"uuid":  fmt.Sprintf("019cba2b-0f72-7dc7-943c-%012d", i),
			}
		}
		obj := map[string]any{
			"name":       "Test Restaurant",
			"menu_items": items,
		}
		input, err := json.Marshal(obj)
		require.NoError(t, err)
		require.Greater(t, len(string(input)), maxBodyLogBytes)

		result := truncateBodyForLog(string(input))
		assert.True(t, json.Valid([]byte(result)), "result must be valid JSON, got: %s", result)
		assert.NotContains(t, result, "bytes truncated", "valid JSON should not use head-tail truncation")
	})
}

func TestTruncateBodyForLog_NestedArrays(t *testing.T) {
	t.Run("nested_long_array_truncated", func(t *testing.T) {
		inner := make([]map[string]string, 20)
		for i := range inner {
			inner[i] = map[string]string{"v": strings.Repeat("x", 30)}
		}
		obj := map[string]any{"items": inner, "name": "test"}
		input, err := json.Marshal(obj)
		require.NoError(t, err)

		result := truncateBodyForLog(string(input))
		require.True(t, json.Valid([]byte(result)), "result must be valid JSON, got: %s", result)

		var parsed map[string]any
		require.NoError(t, json.Unmarshal([]byte(result), &parsed))

		items, ok := parsed["items"].([]any)
		require.True(t, ok)
		assert.Len(t, items, maxArrayLogItems+1)
	})
}

func TestTruncateBodyForLog_NoNewlines(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"long_plain_text", strings.Repeat("abcdefghij", 100)},
		{"long_invalid_JSON", "{bad" + strings.Repeat("x", 600)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateBodyForLog(tt.input)
			assert.NotContains(t, result, "\n", "truncated output must not contain newlines")
		})
	}
}

func TestTruncateBodyForLog_MalformedJSONOverLimit(t *testing.T) {
	input := `{"key": "` + strings.Repeat("x", 600)
	result := truncateBodyForLog(input)

	assert.Contains(t, result, "bytes truncated")
}
