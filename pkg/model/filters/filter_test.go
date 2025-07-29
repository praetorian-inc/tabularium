package filters

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilter_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name   string
		filter Filter
	}{
		{
			name:   "single value equals filter",
			filter: NewFilter("status", OperatorEqual, "active"),
		},
		{
			name:   "array value contains filter",
			filter: NewFilter("tags", OperatorContains, []any{"tag1", "tag2"}),
		},
		{
			name:   "negated equals filter",
			filter: NewFilter("status", OperatorEqual, "deleted", WithNot()),
		},
		{
			name: "OR operator with filters",
			filter: Filter{
				Field:    "status",
				Operator: OperatorOr,
				Value: SliceOrValue[any]{
					NewFilter("status", OperatorEqual, "active"),
					NewFilter("status", OperatorEqual, "pending"),
				},
			},
		},
		{
			name: "AND operator with filters",
			filter: Filter{
				Field:    "status",
				Operator: OperatorAnd,
				Value: SliceOrValue[any]{
					NewFilter("status", OperatorEqual, "active"),
					NewFilter("type", OperatorEqual, "high"),
				},
			},
		},
		{
			name:   "mixed type values",
			filter: NewFilter("mixed", OperatorContains, []any{float64(42), "test", true}),
		},
		{
			name:   "starts with filter",
			filter: NewFilter("name", OperatorStartsWith, "test-"),
		},
		{
			name:   "ends with filter",
			filter: NewFilter("name", OperatorEndsWith, ".com"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			bytes, err := json.Marshal(tt.filter)
			require.NoError(t, err, "Marshal failed")

			// Test unmarshaling
			var result Filter
			err = json.Unmarshal(bytes, &result)
			require.NoError(t, err, "Unmarshal failed")

			// Compare fields
			assert.Equal(t, tt.filter.Field, result.Field)
			assert.Equal(t, tt.filter.Operator, result.Operator)
			assert.Equal(t, tt.filter.Not, result.Not)

			// Special handling for nested filters in OR/AND operators
			if tt.filter.Operator == OperatorOr || tt.filter.Operator == OperatorAnd {
				require.Len(t, result.Value, len(tt.filter.Value))

				for i := range tt.filter.Value {
					expectedFilter := tt.filter.Value[i].(Filter)
					resultFilter := result.Value[i].(Filter)
					assert.Equal(t, expectedFilter.Field, resultFilter.Field)
					assert.Equal(t, expectedFilter.Operator, resultFilter.Operator)
					assert.Equal(t, expectedFilter.Value, resultFilter.Value)
					assert.Equal(t, expectedFilter.Not, resultFilter.Not)
				}
			} else {
				assert.Equal(t, tt.filter.Value, result.Value)
			}

			// Marshal again and compare with original
			resultBytes, err := json.Marshal(result)
			require.NoError(t, err, "Re-marshal failed")

			originalBytes, err := json.Marshal(tt.filter)
			require.NoError(t, err, "Original re-marshal failed")

			assert.JSONEq(t, string(originalBytes), string(resultBytes),
				"Marshal/Unmarshal cycle produced different result")
		})
	}
}

func TestFilter_Constructors(t *testing.T) {
	t.Run("NewFilter single value", func(t *testing.T) {
		filter := NewFilter("name", OperatorEqual, "test")
		assert.Equal(t, "name", filter.Field)
		assert.Equal(t, OperatorEqual, filter.Operator)
		assert.Equal(t, SliceOrValue[any]{"test"}, filter.Value)
		assert.False(t, filter.Not)
	})

	t.Run("NewFilter multiple values", func(t *testing.T) {
		filter := NewFilter("status", OperatorContains, []string{"active", "pending"})
		assert.Equal(t, "status", filter.Field)
		assert.Equal(t, OperatorContains, filter.Operator)
		assert.Equal(t, SliceOrValue[any]{[]string{"active", "pending"}}, filter.Value)
		assert.False(t, filter.Not)
	})

	t.Run("NewNotFilter", func(t *testing.T) {
		filter := NewFilter("status", OperatorEqual, "deleted", WithNot())
		assert.Equal(t, "status", filter.Field)
		assert.Equal(t, OperatorEqual, filter.Operator)
		assert.Equal(t, SliceOrValue[any]{"deleted"}, filter.Value)
		assert.True(t, filter.Not)
	})

	t.Run("NewFilter with OR operator", func(t *testing.T) {
		filter := Filter{
			Field:    "combined",
			Operator: OperatorOr,
			Value: SliceOrValue[any]{
				NewFilter("status", OperatorEqual, "active"),
				NewFilter("type", OperatorEqual, "user"),
			},
		}
		require.Len(t, filter.Value, 2)
		assert.Equal(t, OperatorOr, filter.Operator)
		assert.Equal(t, "combined", filter.Field)
		assert.False(t, filter.Not)

		f1 := filter.Value[0].(Filter)
		assert.Equal(t, "status", f1.Field)
		assert.Equal(t, OperatorEqual, f1.Operator)
		assert.Equal(t, SliceOrValue[any]{"active"}, f1.Value)

		f2 := filter.Value[1].(Filter)
		assert.Equal(t, "type", f2.Field)
		assert.Equal(t, OperatorEqual, f2.Operator)
		assert.Equal(t, SliceOrValue[any]{"user"}, f2.Value)
	})
}
