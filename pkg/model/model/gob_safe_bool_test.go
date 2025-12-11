package model

import (
	"bytes"
	"encoding/gob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type GobSafeBoolWrapper struct {
	Value *GobSafeBool
}

func init() {
	gob.Register(GobSafeBoolWrapper{})
}

func TestGobSafeBool_FalseValue(t *testing.T) {
	// Test that nil encodes and decodes correctly
	original := GobSafeBool(false)

	// Encode
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(original); err != nil {
		t.Fatalf("Failed to encode false GobSafeBool: %v", err)
	}

	// Decode
	var decoded GobSafeBool
	decoder := gob.NewDecoder(&buf)
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("Failed to decode false GobSafeBool: %v", err)
	}

	assert.False(t, bool(decoded))
}

func TestGobSafeBool_TrueValue(t *testing.T) {
	// Test that nil encodes and decodes correctly
	original := GobSafeBool(true)

	// Encode
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(original); err != nil {
		t.Fatalf("Failed to encode nil GobSafeBool: %v", err)
	}

	// Decode
	var decoded GobSafeBool
	decoder := gob.NewDecoder(&buf)
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("Failed to decode nil GobSafeBool: %v", err)
	}

	assert.True(t, bool(decoded))
}

func TestGobSafeBool_FalseValuePointer(t *testing.T) {
	// Test that nil encodes and decodes correctly
	original := GobSafeBool(false)

	wrapper := GobSafeBoolWrapper{
		Value: &original,
	}

	// Encode
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(wrapper); err != nil {
		t.Fatalf("Failed to encode false GobSafeBool ptr: %v", err)
	}

	// Decode
	var decoded GobSafeBoolWrapper
	decoder := gob.NewDecoder(&buf)
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("Failed to decode false GobSafeBool ptr: %v", err)
	}

	require.NotNil(t, decoded.Value)
	assert.False(t, bool(*decoded.Value))
}

func TestGobSafeBool_TrueValuePointer(t *testing.T) {
	// Test that nil encodes and decodes correctly
	original := GobSafeBool(true)

	wrapper := GobSafeBoolWrapper{
		Value: &original,
	}

	// Encode
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(wrapper); err != nil {
		t.Fatalf("Failed to encode false GobSafeBool ptr: %v", err)
	}

	// Decode
	var decoded GobSafeBoolWrapper
	decoder := gob.NewDecoder(&buf)
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("Failed to decode false GobSafeBool ptr: %v", err)
	}

	require.NotNil(t, decoded.Value)
	assert.True(t, bool(*decoded.Value))
}
