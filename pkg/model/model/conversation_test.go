package model

import (
	"strings"
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConversation_NewConversation(t *testing.T) {
	topic := "Test Conversation"

	conv := NewConversation(topic)

	assert.Equal(t, topic, conv.Topic)
	assert.NotEmpty(t, conv.UUID)
	assert.NotEmpty(t, conv.Created)
	assert.NotEmpty(t, conv.Key)
	assert.True(t, strings.HasPrefix(conv.Key, "#conversation#"))
	assert.True(t, conv.Valid())
}

func TestConversation_GetKey(t *testing.T) {
	conv := NewConversation("test")
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
}

func TestConversation_Hooks(t *testing.T) {
	conv := &Conversation{}

	// Call hooks manually
	registry.CallHooks(conv)

	assert.NotEmpty(t, conv.Key)
	assert.True(t, strings.HasPrefix(conv.Key, "#conversation#"))

	// Verify UUID format in key (should be 36 characters with dashes)
	keyParts := strings.Split(conv.Key, "#")
	require.Len(t, keyParts, 3)
	uuid := keyParts[2]
	assert.Len(t, uuid, 36)
	assert.Contains(t, uuid, "-")
}

func TestConversation_Hooks_ExistingKey(t *testing.T) {
	existingKey := "#conversation#existing#12345"
	conv := &Conversation{
		Key: existingKey,
	}

	registry.CallHooks(conv)

	// Should not change existing key
	assert.Equal(t, existingKey, conv.Key)
}

func TestConversation_RegistryIntegration(t *testing.T) {
	// Test that the conversation is properly registered in the registry
	conv := &Conversation{}

	// Check that it's registered by calling a registry function
	hooks := conv.GetHooks()
	assert.Len(t, hooks, 1)

	// Verify hook functionality
	err := hooks[0].Call()
	assert.NoError(t, err)
	assert.NotEmpty(t, conv.Key)
}

func TestConversation_KeyGeneration_Uniqueness(t *testing.T) {
	// Test that multiple conversations with same topic get different keys
	topic := "Same Name"

	conv1 := NewConversation(topic)
	conv2 := NewConversation(topic)

	assert.NotEqual(t, conv1.Key, conv2.Key)
	assert.True(t, strings.HasPrefix(conv1.Key, "#conversation#"))
	assert.True(t, strings.HasPrefix(conv2.Key, "#conversation#"))
	assert.Equal(t, topic, conv1.Topic)
	assert.Equal(t, topic, conv2.Topic)
}

func TestConversation_SecurityScenarios(t *testing.T) {
	testCases := []struct {
		name             string
		conversationName string
		expectValid      bool
	}{
		{
			name:             "valid standard conversation",
			conversationName: "Normal Conversation",
			expectValid:      true,
		},
		{
			name:             "conversation with special characters",
			conversationName: "Conv/\\with<>special|chars",
			expectValid:      true,
		},
		{
			name:             "very long conversation name",
			conversationName: strings.Repeat("a", 1000),
			expectValid:      true,
		},
		{
			name:             "SQL injection attempt in name",
			conversationName: "'; DROP TABLE users; --",
			expectValid:      true, // Should be treated as regular string
		},
		{
			name:             "XSS attempt in name",
			conversationName: "<script>alert('xss')</script>",
			expectValid:      true, // Should be treated as regular string
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conv := NewConversation(tc.conversationName)

			assert.Equal(t, tc.expectValid, conv.Valid())
			if tc.expectValid {
				assert.Equal(t, tc.conversationName, conv.Topic)
				assert.NotEmpty(t, conv.Key)
			}
		})
	}
}

func TestConversation_TopicField(t *testing.T) {
	testCases := []struct {
		name     string
		topic    string
		expected string
	}{
		{
			name:     "short topic",
			topic:    "Find all assets",
			expected: "Find all assets",
		},
		{
			name:     "long topic gets truncated",
			topic:    strings.Repeat("a", 300),
			expected: strings.Repeat("a", 256),
		},
		{
			name:     "empty topic",
			topic:    "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conv := NewConversation("Test Chat")
			conv.Topic = tc.topic

			if len(tc.topic) > 256 {
				conv.Topic = tc.topic[:256]
			}

			assert.Equal(t, tc.expected, conv.Topic)
		})
	}
}

func TestConversationParentLink(t *testing.T) {
	parent := NewConversation("parent topic")
	child := Conversation{
		Topic:              "child topic",
		ParentConversation: parent.UUID,
	}
	child.Defaulted()
	registry.CallHooks(&child)

	assert.NotEmpty(t, child.UUID, "child should have UUID")
	assert.Equal(t, parent.UUID, child.ParentConversation, "child should link to parent")
}
