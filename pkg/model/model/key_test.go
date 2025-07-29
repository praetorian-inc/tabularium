package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewKey(t *testing.T) {
	tests := []struct {
		name     string
		keyName  string
		wantName string
	}{
		{
			name:     "simple key name",
			keyName:  "test-key",
			wantName: "test-key",
		},
		{
			name:     "empty key name",
			keyName:  "",
			wantName: "",
		},
		{
			name:     "key with special characters",
			keyName:  "api-key-2024",
			wantName: "api-key-2024",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := NewKey(tt.keyName)

			assert.Equal(t, tt.wantName, key.Name)
			assert.NotEmpty(t, key.ID, "ID should be generated")
			assert.NotEmpty(t, key.Created, "Created timestamp should be set")
			assert.Equal(t, Active, key.Status, "Status should be Active by default")
			assert.Equal(t, fmt.Sprintf("#key#%s", key.ID), key.Key, "Key should follow expected format")
		})
	}
}

func TestKey_GetDescription(t *testing.T) {
	key := &Key{}
	expected := "An API key for the Chariot platform"
	
	assert.Equal(t, expected, key.GetDescription())
}

func TestKey_GetHooks(t *testing.T) {
	key := &Key{ID: "test-id-123"}
	hooks := key.GetHooks()

	require.Len(t, hooks, 1, "Should have exactly one hook")
	
	err := hooks[0].Call()
	require.NoError(t, err, "Hook should execute without error")
	
	expectedKey := "#key#test-id-123"
	assert.Equal(t, expectedKey, key.Key, "Hook should set Key field correctly")
}

func TestKey_Defaulted(t *testing.T) {
	key := &Key{}
	
	assert.Empty(t, key.Created)
	assert.Empty(t, key.Status)
	assert.Empty(t, key.ID)
	
	key.Defaulted()
	
	assert.NotEmpty(t, key.Created, "Created should be set")
	assert.Equal(t, Active, key.Status, "Status should be Active")
	assert.NotEmpty(t, key.ID, "ID should be generated")
	
	assert.Regexp(t, "^[A-Z2-7=]+$", key.ID, "ID should be valid base32")
}

func TestKey_IDGeneration(t *testing.T) {
	key1 := NewKey("test1")
	key2 := NewKey("test2")
	
	assert.NotEqual(t, key1.ID, key2.ID, "Different keys should have different IDs")
	assert.NotEqual(t, key1.Key, key2.Key, "Different keys should have different Key values")
}

func TestKey_JSONSerialization(t *testing.T) {
	key := NewKey("test-api-key")
	key.Username = "user@example.com"
	key.Creator = "admin@example.com"
	key.Secret = "secret-value"

	jsonData, err := json.Marshal(key)
	require.NoError(t, err, "Should marshal to JSON without error")
	
	jsonStr := string(jsonData)
	assert.Contains(t, jsonStr, "secret-value", "Secret should be included in JSON")
	assert.Contains(t, jsonStr, "user@example.com", "Username should be included in JSON")
	
	var unmarshaled Key
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err, "Should unmarshal from JSON without error")
	
	assert.Equal(t, key.Name, unmarshaled.Name)
	assert.Equal(t, key.Username, unmarshaled.Username)
	assert.Equal(t, key.ID, unmarshaled.ID)
	assert.Equal(t, key.Secret, unmarshaled.Secret)
	assert.Equal(t, key.Creator, unmarshaled.Creator)
}

func TestKey_FieldValidation(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() Key
		validate func(t *testing.T, key Key)
	}{
		{
			name: "all fields populated",
			setup: func() Key {
				key := NewKey("full-test-key")
				key.Username = "user@example.com"
				key.Creator = "admin@example.com"
				key.Deleter = "admin2@example.com"
				key.Deleted = "2024-01-01T00:00:00Z"
				key.Secret = "secret-123"
				return key
			},
			validate: func(t *testing.T, key Key) {
				assert.Equal(t, "full-test-key", key.Name)
				assert.Equal(t, "user@example.com", key.Username)
				assert.Equal(t, "admin@example.com", key.Creator)
				assert.Equal(t, "admin2@example.com", key.Deleter)
				assert.Equal(t, "2024-01-01T00:00:00Z", key.Deleted)
				assert.Equal(t, "secret-123", key.Secret)
				assert.NotEmpty(t, key.Created)
				assert.Equal(t, Active, key.Status)
			},
		},
		{
			name: "minimal key",
			setup: func() Key {
				return NewKey("minimal")
			},
			validate: func(t *testing.T, key Key) {
				assert.Equal(t, "minimal", key.Name)
				assert.Empty(t, key.Username)
				assert.Empty(t, key.Creator)
				assert.Empty(t, key.Deleter)
				assert.Empty(t, key.Deleted)
				assert.Empty(t, key.Secret)
				assert.NotEmpty(t, key.Created)
				assert.Equal(t, Active, key.Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := tt.setup()
			tt.validate(t, key)
		})
	}
}

func TestKey_KeyFormat(t *testing.T) {
	key := NewKey("format-test")
	
	expectedPrefix := "#key#"
	assert.True(t, strings.HasPrefix(key.Key, expectedPrefix), 
		"Key should start with '#key#' prefix")
	
	keyPart := strings.TrimPrefix(key.Key, expectedPrefix)
	assert.Equal(t, key.ID, keyPart, "Key suffix should match ID field")
}

func TestKey_EmptyNameHandling(t *testing.T) {
	key := NewKey("")
	
	assert.Empty(t, key.Name, "Name should remain empty")
	assert.NotEmpty(t, key.ID, "ID should still be generated")
	assert.NotEmpty(t, key.Created, "Created should still be set")
	assert.Equal(t, Active, key.Status, "Status should still be Active")
	assert.Equal(t, fmt.Sprintf("#key#%s", key.ID), key.Key, "Key should still be formatted correctly")
}
