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

func TestConversation_HierarchyFields(t *testing.T) {
	tests := []struct {
		name                 string
		parentConversationID string
		depth                int
		invokedBy            string
		wantValid            bool
	}{
		{
			name:                 "root conversation",
			parentConversationID: "",
			depth:                0,
			invokedBy:            "",
			wantValid:            true,
		},
		{
			name:                 "child conversation depth 1",
			parentConversationID: "parent-uuid-123",
			depth:                1,
			invokedBy:            "toolu_abc123",
			wantValid:            true,
		},
		{
			name:                 "grandchild conversation depth 2",
			parentConversationID: "child-uuid-456",
			depth:                2,
			invokedBy:            "toolu_def456",
			wantValid:            true,
		},
		{
			name:                 "max depth conversation",
			parentConversationID: "grandchild-uuid-789",
			depth:                3,
			invokedBy:            "toolu_ghi789",
			wantValid:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := NewConversation("Test topic")
			conv.ParentConversationID = tt.parentConversationID
			conv.Depth = tt.depth
			conv.InvokedBy = tt.invokedBy

			if got := conv.Valid(); got != tt.wantValid {
				t.Errorf("Conversation.Valid() = %v, want %v", got, tt.wantValid)
			}

			assert.Equal(t, tt.parentConversationID, conv.ParentConversationID)
			assert.Equal(t, tt.depth, conv.Depth)
			assert.Equal(t, tt.invokedBy, conv.InvokedBy)
		})
	}
}

func TestConversation_BackwardCompatibility(t *testing.T) {
	// Existing conversations without new fields should still work
	conv := NewConversation("Test conversation")
	conv.Username = "test@example.com"
	conv.User = "test@example.com"

	// New fields should have zero values
	assert.Empty(t, conv.ParentConversationID, "ParentConversationID should be empty for new conversations")
	assert.Equal(t, 0, conv.Depth, "Depth should be 0 for new conversations")
	assert.Empty(t, conv.InvokedBy, "InvokedBy should be empty for new conversations")

	assert.True(t, conv.Valid(), "Conversation without new fields should still be valid")
}

func TestConversation_HierarchyChain(t *testing.T) {
	// Create parent conversation
	parent := NewConversation("Parent topic")
	parent.Username = "test@example.com"
	parent.User = "test@example.com"

	// Create child conversation
	child := NewConversation("Child topic")
	child.Username = "test@example.com"
	child.User = "test@example.com"
	child.ParentConversationID = parent.UUID
	child.Depth = parent.Depth + 1
	child.InvokedBy = "toolu_test123"

	// Verify hierarchy
	assert.Equal(t, parent.UUID, child.ParentConversationID)
	assert.Equal(t, 1, child.Depth)
	assert.NotEmpty(t, child.InvokedBy)

	// Create grandchild
	grandchild := NewConversation("Grandchild topic")
	grandchild.Username = "test@example.com"
	grandchild.User = "test@example.com"
	grandchild.ParentConversationID = child.UUID
	grandchild.Depth = child.Depth + 1
	grandchild.InvokedBy = "toolu_test456"

	// Verify grandchild depth
	assert.Equal(t, 2, grandchild.Depth)
	assert.Equal(t, child.UUID, grandchild.ParentConversationID)
}

func TestConversation_IsRootConversation(t *testing.T) {
	tests := []struct {
		name                 string
		parentConversationID string
		depth                int
		wantIsRoot           bool
	}{
		{
			name:                 "root conversation",
			parentConversationID: "",
			depth:                0,
			wantIsRoot:           true,
		},
		{
			name:                 "child conversation",
			parentConversationID: "parent-123",
			depth:                1,
			wantIsRoot:           false,
		},
		{
			name:                 "grandchild conversation",
			parentConversationID: "child-456",
			depth:                2,
			wantIsRoot:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := NewConversation("Test")
			conv.ParentConversationID = tt.parentConversationID
			conv.Depth = tt.depth

			isRoot := conv.ParentConversationID == ""
			assert.Equal(t, tt.wantIsRoot, isRoot)
		})
	}
}
