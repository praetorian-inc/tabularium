package filters

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestSliceOrValue_UnmarshalJSON_SingleValue_Object(t *testing.T) {
	input := `{"name": "test", "value": 123}`
	expected := SliceOrValue[TestStruct]{
		{Name: "test", Value: 123},
	}

	var result SliceOrValue[TestStruct]
	err := json.Unmarshal([]byte(input), &result)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestSliceOrValue_UnmarshalJSON_SingleValue_String(t *testing.T) {
	input := `"test"`
	expected := SliceOrValue[string]{
		"test",
	}

	var result SliceOrValue[string]
	err := json.Unmarshal([]byte(input), &result)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestSliceOrValue_UnmarshalJSON_SingleValue_Number(t *testing.T) {
	input := `42`
	expected := SliceOrValue[int]{
		42,
	}

	var result SliceOrValue[int]
	err := json.Unmarshal([]byte(input), &result)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestSliceOrValue_UnmarshalJSON_Slice_Objects(t *testing.T) {
	input := `[
		{"name": "test1", "value": 123},
		{"name": "test2", "value": 456}
	]`
	expected := SliceOrValue[TestStruct]{
		{Name: "test1", Value: 123},
		{Name: "test2", Value: 456},
	}

	var result SliceOrValue[TestStruct]
	err := json.Unmarshal([]byte(input), &result)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestSliceOrValue_UnmarshalJSON_Slice_Strings(t *testing.T) {
	input := `["test1", "test2"]`
	expected := SliceOrValue[string]{
		"test1",
		"test2",
	}

	var result SliceOrValue[string]
	err := json.Unmarshal([]byte(input), &result)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestSliceOrValue_UnmarshalJSON_Slice_MixedTypes(t *testing.T) {
	input := `["test1", 2]`
	expected := SliceOrValue[any]{
		"test1",
		2.0,
	}

	var result SliceOrValue[any]
	err := json.Unmarshal([]byte(input), &result)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestSliceOrValue_UnmarshalJSON_EmptyArray(t *testing.T) {
	input := `[]`
	expected := SliceOrValue[TestStruct]{}

	var result SliceOrValue[TestStruct]
	err := json.Unmarshal([]byte(input), &result)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestSliceOrValue_UnmarshalJSON_TestStruct_Error(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "invalid json syntax",
			input:         `{bad json`,
			expectedError: "invalid character",
		},
		{
			name:          "mixed types in array",
			input:         `[{"name": "test"}, 123, "string"]`,
			expectedError: "failed to unmarshal",
		},
		{
			name:          "number for struct",
			input:         `42`,
			expectedError: "failed to unmarshal",
		},
		{
			name:          "string for struct",
			input:         `"test"`,
			expectedError: "failed to unmarshal",
		},
		{
			name:          "bool for struct",
			input:         `true`,
			expectedError: "failed to unmarshal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result SliceOrValue[TestStruct]
			err := json.Unmarshal([]byte(tt.input), &result)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestSliceOrValue_MarshalJSON_Objects(t *testing.T) {
	input := SliceOrValue[TestStruct]{
		{Name: "test1", Value: 123},
		{Name: "test2", Value: 456},
	}
	expected := `[{"name":"test1","value":123},{"name":"test2","value":456}]`

	result, err := json.Marshal(input)
	require.NoError(t, err)
	assert.JSONEq(t, expected, string(result))
}

func TestSliceOrValue_MarshalJSON_Strings(t *testing.T) {
	input := SliceOrValue[string]{
		"test1",
		"test2",
	}
	expected := `["test1","test2"]`

	result, err := json.Marshal(input)
	require.NoError(t, err)
	assert.JSONEq(t, expected, string(result))
}

func TestSliceOrValue_MarshalJSON_Empty(t *testing.T) {
	input := SliceOrValue[TestStruct]{}
	expected := `null`

	result, err := json.Marshal(input)
	require.NoError(t, err)
	assert.JSONEq(t, expected, string(result))
}

func TestSliceOrValue_FullCycle_Objects(t *testing.T) {
	input := SliceOrValue[TestStruct]{
		{Name: "test1", Value: 123},
		{Name: "test2", Value: 456},
	}

	// Marshal
	marshaled, err := json.Marshal(input)
	require.NoError(t, err)

	// Unmarshal
	var result SliceOrValue[TestStruct]
	err = json.Unmarshal(marshaled, &result)
	require.NoError(t, err)

	// Compare
	assert.Equal(t, input, result)
}

func TestSliceOrValue_FullCycle_Strings(t *testing.T) {
	input := SliceOrValue[string]{
		"test1",
		"test2",
	}

	// Marshal
	marshaled, err := json.Marshal(input)
	require.NoError(t, err)

	// Unmarshal
	var result SliceOrValue[string]
	err = json.Unmarshal(marshaled, &result)
	require.NoError(t, err)

	// Compare
	assert.Equal(t, input, result)
}

func TestSliceOrValue_Atomic(t *testing.T) {
	t.Run("unmarshal single string", func(t *testing.T) {
		input := `"hello"`
		var result SliceOrValue[string]
		err := json.Unmarshal([]byte(input), &result)
		require.NoError(t, err)
		assert.Equal(t, SliceOrValue[string]{"hello"}, result)
	})

	t.Run("unmarshal string array", func(t *testing.T) {
		input := `["hello", "world"]`
		var result SliceOrValue[string]
		err := json.Unmarshal([]byte(input), &result)
		require.NoError(t, err)
		assert.Equal(t, SliceOrValue[string]{"hello", "world"}, result)
	})

	t.Run("marshal single string", func(t *testing.T) {
		input := SliceOrValue[string]{"hello"}
		result, err := json.Marshal(input)
		require.NoError(t, err)
		assert.JSONEq(t, `"hello"`, string(result))
	})

	t.Run("marshal string array", func(t *testing.T) {
		input := SliceOrValue[string]{"hello", "world"}
		result, err := json.Marshal(input)
		require.NoError(t, err)
		assert.JSONEq(t, `["hello", "world"]`, string(result))
	})

	t.Run("full cycle single string", func(t *testing.T) {
		input := SliceOrValue[string]{"hello"}
		marshaled, err := json.Marshal(input)
		require.NoError(t, err)

		var result SliceOrValue[string]
		err = json.Unmarshal(marshaled, &result)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("full cycle string array", func(t *testing.T) {
		input := SliceOrValue[string]{"hello", "world"}
		marshaled, err := json.Marshal(input)
		require.NoError(t, err)

		var result SliceOrValue[string]
		err = json.Unmarshal(marshaled, &result)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("error cases", func(t *testing.T) {
		tests := []struct {
			name          string
			input         string
			expectedError string
		}{
			{
				name:          "invalid json",
				input:         `{"bad json"`,
				expectedError: "unexpected end of JSON input",
			},
			{
				name:          "number for string",
				input:         `42`,
				expectedError: "failed to unmarshal",
			},
			{
				name:          "object for string",
				input:         `{"key": "value"}`,
				expectedError: "failed to unmarshal",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var result SliceOrValue[string]
				err := json.Unmarshal([]byte(tt.input), &result)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			})
		}
	})
}

func TestSliceOrValue_DifferentTypes(t *testing.T) {
	t.Run("bool type", func(t *testing.T) {
		input := `true`
		var result SliceOrValue[bool]
		err := json.Unmarshal([]byte(input), &result)
		require.NoError(t, err)
		assert.Equal(t, SliceOrValue[bool]{true}, result)
	})

	t.Run("float type", func(t *testing.T) {
		input := `3.14`
		var result SliceOrValue[float64]
		err := json.Unmarshal([]byte(input), &result)
		require.NoError(t, err)
		assert.Equal(t, SliceOrValue[float64]{3.14}, result)
	})

	t.Run("bool array", func(t *testing.T) {
		input := `[true, false, true]`
		var result SliceOrValue[bool]
		err := json.Unmarshal([]byte(input), &result)
		require.NoError(t, err)
		assert.Equal(t, SliceOrValue[bool]{true, false, true}, result)
	})

	t.Run("float array", func(t *testing.T) {
		input := `[3.14, 2.718, 1.414]`
		var result SliceOrValue[float64]
		err := json.Unmarshal([]byte(input), &result)
		require.NoError(t, err)
		assert.Equal(t, SliceOrValue[float64]{3.14, 2.718, 1.414}, result)
	})
}

func TestSliceOrValue_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    SliceOrValue[string]
		expected []byte
	}{
		{
			name:     "empty",
			input:    SliceOrValue[string]{},
			expected: []byte(`null`),
		},
		{
			name:     "one",
			input:    SliceOrValue[string]{"test"},
			expected: []byte(`"test"`),
		},
		{
			name:     "two",
			input:    SliceOrValue[string]{"test", "test"},
			expected: []byte(`["test","test"]`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := json.Marshal(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
