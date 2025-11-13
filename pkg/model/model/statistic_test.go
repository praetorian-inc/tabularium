package model

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatistic_KeyLengthLimit(t *testing.T) {
	// Create a value that is greater than 1024 characters
	longValue := strings.Repeat("a", 1500)

	stat := NewStatistic("test_key", "test_name", longValue, Now())

	require.LessOrEqual(t, len(stat.Key), 1024, "Expected key length to be <= 1024, but got %d", len(stat.Key))
}
