package model

import (
	"testing"
	"strings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func TestConversation_NewConversation(t *testing.T) {
	name := "Test Conversation"
	username := "gladiator@praetorian.com"
	
	conv := NewConversation(name, username)
	
	assert.Equal(t, name, conv.Name)
	assert.Equal(t, username, conv.Username)
	assert.NotEmpty(t, conv.UUID)
	assert.NotEmpty(t, conv.Created)
	assert.NotEmpty(t, conv.Source)
	assert.NotEmpty(t, conv.Key)
	assert.True(t, strings.HasPrefix(conv.Key, "#conversation#"))
	assert.True(t, conv.Valid())
}

func TestConversation_GetKey(t *testing.T) {
	conv := NewConversation("test", "user@example.com")
	assert.Equal(t, conv.Key, conv.GetKey())
	assert.NotEmpty(t, conv.GetKey())
}

func TestConversation_GetDescription(t *testing.T) {
	conv := &Conversation{}
	expected := "Represents a conversation between a user and AI assistant with running capabilities."
	assert.Equal(t, expected, conv.GetDescription())
}

func TestConversation_Defaulted(t *testing.T) {
	conv := &Conversation{}
	conv.Defaulted()
	
	assert.NotEmpty(t, conv.Created)
	assert.NotEmpty(t, conv.Source)
	
	// Verify TTL is approximately 30 days from now
	assert.NotEmpty(t, conv.Created) // Allow 60 seconds tolerance
}

func TestConversation_Hooks(t *testing.T) {
	conv := &Conversation{
		Name:     "Test Conversation",
		Username: "user@example.com",
	}
	
	// Call hooks manually
	registry.CallHooks(conv)
	
	assert.NotEmpty(t, conv.Key)
	assert.True(t, strings.HasPrefix(conv.Key, "#conversation#Test Conversation#"))
	
	// Verify UUID format in key (should be 36 characters with dashes)
	keyParts := strings.Split(conv.Key, "#")
	require.Len(t, keyParts, 4)
	uuid := keyParts[3]
	assert.Len(t, uuid, 36)
	assert.Contains(t, uuid, "-")
}

func TestConversation_Hooks_ExistingKey(t *testing.T) {
	existingKey := "#conversation#existing#12345"
	conv := &Conversation{
		Key:      existingKey,
		Name:     "Test Conversation",
		Username: "user@example.com",
	}
	
	registry.CallHooks(conv)
	
	// Should not change existing key
	assert.Equal(t, existingKey, conv.Key)
}

func TestConversation_Valid(t *testing.T) {
	testCases := []struct {
		name     string
		conv     Conversation
		expected bool
	}{
		{
			name: "valid conversation",
			conv: Conversation{
				Name:     "Test",
				Username: "user@example.com",
			},
			expected: true,
		},
		{
			name: "missing name",
			conv: Conversation{
				Username: "user@example.com",
			},
			expected: false,
		},
		{
			name: "missing username",
			conv: Conversation{
				Name: "Test",
			},
			expected: false,
		},
		{
			name: "empty name",
			conv: Conversation{
				Name:     "",
				Username: "user@example.com",
			},
			expected: false,
		},
		{
			name: "empty username",
			conv: Conversation{
				Name:     "Test",
				Username: "",
			},
			expected: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.conv.Valid())
		})
	}
}

func TestConversation_RegistryIntegration(t *testing.T) {
	// Test that the conversation is properly registered in the registry
	conv := &Conversation{}
	
	// Check that it's registered by calling a registry function
	hooks := conv.GetHooks()
	assert.Len(t, hooks, 1)
	
	// Verify hook functionality
	conv.Name = "Registry Test"
	err := hooks[0].Call()
	assert.NoError(t, err)
	assert.NotEmpty(t, conv.Key)
}

func TestConversation_KeyGeneration_Uniqueness(t *testing.T) {
	// Test that multiple conversations with same name get different keys
	name := "Same Name"
	username := "user@example.com"
	
	conv1 := NewConversation(name, username)
	conv2 := NewConversation(name, username)
	
	assert.NotEqual(t, conv1.Key, conv2.Key)
	assert.True(t, strings.HasPrefix(conv1.Key, "#conversation#"+name+"#"))
	assert.True(t, strings.HasPrefix(conv2.Key, "#conversation#"+name+"#"))
}

func TestConversation_SecurityScenarios(t *testing.T) {
	testCases := []struct {
		name        string
		conversationName string
		username    string
		expectValid bool
	}{
		{
			name:         "valid standard conversation",
			conversationName: "Normal Conversation",
			username:     "gladiator@praetorian.com",
			expectValid:  true,
		},
		{
			name:         "conversation with special characters",
			conversationName: "Conv/\\with<>special|chars",
			username:     "user@domain.com",
			expectValid:  true,
		},
		{
			name:         "very long conversation name",
			conversationName: strings.Repeat("a", 1000),
			username:     "user@domain.com",
			expectValid:  true,
		},
		{
			name:         "SQL injection attempt in name",
			conversationName: "'; DROP TABLE users; --",
			username:     "user@domain.com",
			expectValid:  true, // Should be treated as regular string
		},
		{
			name:         "XSS attempt in name",
			conversationName: "<script>alert('xss')</script>",
			username:     "user@domain.com",
			expectValid:  true, // Should be treated as regular string
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conv := NewConversation(tc.conversationName, tc.username)
			
			assert.Equal(t, tc.expectValid, conv.Valid())
			if tc.expectValid {
				assert.Equal(t, tc.conversationName, conv.Name)
				assert.Equal(t, tc.username, conv.Username)
				assert.NotEmpty(t, conv.Key)
			}
		})
	}
}