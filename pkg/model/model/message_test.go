package model

import (
	"github.com/praetorian-inc/tabularium/pkg/registry/model"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessage_NewMessage(t *testing.T) {
	conversationID := "550e8400-e29b-41d4-a716-446655440000"
	role := "user"
	content := "Hello, world!"
	username := "gladiator@praetorian.com"

	msg := NewMessage(conversationID, role, content, username)

	assert.Equal(t, conversationID, msg.ConversationID)
	assert.Equal(t, role, msg.Role)
	assert.Equal(t, content, msg.Content)
	assert.Equal(t, username, msg.Username)
	assert.NotEmpty(t, msg.Timestamp)
	assert.NotZero(t, msg.TTL)
	assert.NotEmpty(t, msg.MessageID)
	assert.NotEmpty(t, msg.Key)
	assert.True(t, strings.HasPrefix(msg.Key, "#message#"+conversationID+"#"))
	assert.True(t, msg.Valid())
}

func TestMessage_GetKey(t *testing.T) {
	msg := NewMessage("conv-id", "user", "content", "user@example.com")
	assert.Equal(t, msg.Key, msg.GetKey())
	assert.NotEmpty(t, msg.GetKey())
}

func TestMessage_GetDescription(t *testing.T) {
	msg := &Message{}
	expected := "Represents a message within a conversation, with UUIDv7 ordering for proper sequencing."
	assert.Equal(t, expected, msg.GetDescription())
}

func TestMessage_Defaulted(t *testing.T) {
	msg := &Message{}
	msg.Defaulted()

	assert.NotEmpty(t, msg.Timestamp)
	assert.NotZero(t, msg.TTL)
	assert.NotEmpty(t, msg.MessageID)

	// Verify TTL is approximately 30 days from now
	future30Days := Future(24 * 30)
	assert.InDelta(t, future30Days, msg.TTL, 60) // Allow 60 seconds tolerance

	// Verify UUIDv7 format
	_, err := uuid.Parse(msg.MessageID)
	assert.NoError(t, err, "MessageID should be a valid UUID")
}

func TestMessage_Hooks(t *testing.T) {
	conversationID := "conv-123"
	msg := &Message{
		ConversationID: conversationID,
		Role:           "user",
		Content:        "test message",
		Username:       "user@example.com",
	}

	// Call hooks manually
	model.CallHooks(msg)

	assert.NotEmpty(t, msg.Key)
	assert.NotEmpty(t, msg.MessageID)
	assert.True(t, strings.HasPrefix(msg.Key, "#message#"+conversationID+"#"))

	// Verify UUID format in key
	keyParts := strings.Split(msg.Key, "#")
	require.Len(t, keyParts, 4)
	messageIDFromKey := keyParts[3]
	assert.Equal(t, msg.MessageID, messageIDFromKey)

	// Verify it's a valid UUID
	_, err := uuid.Parse(messageIDFromKey)
	assert.NoError(t, err)
}

func TestMessage_Hooks_ExistingKey(t *testing.T) {
	existingKey := "#message#conv-123#existing-key"
	msg := &Message{
		Key:            existingKey,
		ConversationID: "conv-123",
		Role:           "user",
		Content:        "test",
		Username:       "user@example.com",
	}

	model.CallHooks(msg)

	// Should not change existing key
	assert.Equal(t, existingKey, msg.Key)
}

func TestMessage_Hooks_ExistingMessageID(t *testing.T) {
	existingMessageID := "01234567-89ab-7def-0123-456789abcdef"
	msg := &Message{
		ConversationID: "conv-123",
		MessageID:      existingMessageID,
		Role:           "user",
		Content:        "test",
		Username:       "user@example.com",
	}

	model.CallHooks(msg)

	assert.NotEmpty(t, msg.Key)
	assert.Equal(t, existingMessageID, msg.MessageID)
	assert.True(t, strings.Contains(msg.Key, existingMessageID))
}

func TestMessage_Valid(t *testing.T) {
	testCases := []struct {
		name     string
		msg      Message
		expected bool
	}{
		{
			name: "valid message",
			msg: Message{
				ConversationID: "conv-123",
				Role:           "user",
				Content:        "Hello",
				Username:       "user@example.com",
			},
			expected: true,
		},
		{
			name: "missing conversation ID",
			msg: Message{
				Role:     "user",
				Content:  "Hello",
				Username: "user@example.com",
			},
			expected: false,
		},
		{
			name: "missing role",
			msg: Message{
				ConversationID: "conv-123",
				Content:        "Hello",
				Username:       "user@example.com",
			},
			expected: false,
		},
		{
			name: "missing content",
			msg: Message{
				ConversationID: "conv-123",
				Role:           "user",
				Username:       "user@example.com",
			},
			expected: false,
		},
		{
			name: "missing username",
			msg: Message{
				ConversationID: "conv-123",
				Role:           "user",
				Content:        "Hello",
			},
			expected: true,
		},
		{
			name: "empty conversation ID",
			msg: Message{
				ConversationID: "",
				Role:           "user",
				Content:        "Hello",
				Username:       "user@example.com",
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.msg.Valid())
		})
	}
}

func TestMessage_UUID_Ordering(t *testing.T) {
	// Test that messages created in sequence have proper UUIDv7 ordering
	conversationID := "conv-ordering-test"
	username := "user@example.com"

	var messages []Message
	var uuids []uuid.UUID

	// Create messages with small delays to ensure ordering
	for i := 0; i < 5; i++ {
		msg := NewMessage(conversationID, "user", "Message "+string(rune('A'+i)), username)
		messages = append(messages, msg)

		parsed, err := uuid.Parse(msg.MessageID)
		require.NoError(t, err)
		uuids = append(uuids, parsed)

		// Small delay to ensure different timestamps
		time.Sleep(1 * time.Millisecond)
	}

	// Verify UUIDs are in ascending order (chronological for UUIDv7)
	for i := 1; i < len(uuids); i++ {
		// UUIDv7 has timestamp in most significant bits, so string comparison works
		assert.True(t, uuids[i-1].String() <= uuids[i].String(),
			"UUID %d should be <= UUID %d chronologically", i-1, i)
	}

	// Test sorting by UUID works correctly
	shuffled := make([]Message, len(messages))
	copy(shuffled, messages)

	// Shuffle the messages (simple reverse to test sorting)
	for i := 0; i < len(shuffled)/2; i++ {
		j := len(shuffled) - 1 - i
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	// Sort by MessageID (UUID) - this should put them back in chronological order
	sort.Slice(shuffled, func(i, j int) bool {
		return shuffled[i].MessageID < shuffled[j].MessageID
	})

	// Verify they're back in chronological order by comparing with original order
	// Sort original messages by UUID too for comparison
	originalSorted := make([]Message, len(messages))
	copy(originalSorted, messages)
	sort.Slice(originalSorted, func(i, j int) bool {
		return originalSorted[i].MessageID < originalSorted[j].MessageID
	})

	for i, msg := range shuffled {
		assert.Equal(t, originalSorted[i].MessageID, msg.MessageID)
		assert.Equal(t, originalSorted[i].Content, msg.Content)
	}
}

func TestMessage_UUID_Uniqueness(t *testing.T) {
	// Test that multiple messages get unique UUIDs
	conversationID := "conv-unique-test"
	username := "user@example.com"

	messageIDs := make(map[string]bool)
	keys := make(map[string]bool)

	for i := 0; i < 10; i++ {
		msg := NewMessage(conversationID, "user", "Message", username)

		// Check MessageID uniqueness
		assert.False(t, messageIDs[msg.MessageID], "MessageID should be unique")
		messageIDs[msg.MessageID] = true

		// Check Key uniqueness
		assert.False(t, keys[msg.Key], "Key should be unique")
		keys[msg.Key] = true

		// Verify UUID format
		_, err := uuid.Parse(msg.MessageID)
		assert.NoError(t, err)
	}
}

func TestMessage_RegistryIntegration(t *testing.T) {
	// Test that the message is properly registered in the registry
	msg := &Message{}

	// Check that it's registered by calling a registry function
	hooks := msg.GetHooks()
	assert.Len(t, hooks, 1)

	// Verify hook functionality
	msg.ConversationID = "conv-123"
	err := hooks[0].Call()
	assert.NoError(t, err)
	assert.NotEmpty(t, msg.Key)
	assert.NotEmpty(t, msg.MessageID)
}

func TestMessage_SecurityScenarios(t *testing.T) {
	conversationID := "conv-security-test"
	username := "gladiator@praetorian.com"

	testCases := []struct {
		name        string
		role        string
		content     string
		expectValid bool
	}{
		{
			name:        "valid user message",
			role:        "user",
			content:     "Hello, how are you?",
			expectValid: true,
		},
		{
			name:        "valid assistant message",
			role:        "assistant",
			content:     "I'm doing well, thank you!",
			expectValid: true,
		},
		{
			name:        "valid system message",
			role:        "system",
			content:     "System notification",
			expectValid: true,
		},
		{
			name:        "SQL injection in content",
			role:        "user",
			content:     "'; DROP TABLE messages; --",
			expectValid: true, // Should be treated as regular string
		},
		{
			name:        "XSS attempt in content",
			role:        "user",
			content:     "<script>alert('xss')</script>",
			expectValid: true, // Should be treated as regular string
		},
		{
			name:        "very long content",
			role:        "user",
			content:     strings.Repeat("a", 10000),
			expectValid: true,
		},
		{
			name:        "empty content",
			role:        "user",
			content:     "",
			expectValid: false,
		},
		{
			name:        "malicious role",
			role:        "<script>",
			content:     "test content",
			expectValid: true, // Role validation is application-level
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := NewMessage(conversationID, tc.role, tc.content, username)

			assert.Equal(t, tc.expectValid, msg.Valid())
			if tc.expectValid {
				assert.Equal(t, tc.role, msg.Role)
				assert.Equal(t, tc.content, msg.Content)
				assert.NotEmpty(t, msg.Key)
				assert.NotEmpty(t, msg.MessageID)

				// Verify UUID is still valid
				_, err := uuid.Parse(msg.MessageID)
				assert.NoError(t, err)
			}
		})
	}
}

func TestMessage_RoleConstants(t *testing.T) {
	// Test that the role constants are defined correctly
	assert.Equal(t, "user", RoleUser)
	assert.Equal(t, "chariot", RoleChariot)
	assert.Equal(t, "system", RoleSystem)
	assert.Equal(t, "tool call", RoleToolCall)
	assert.Equal(t, "tool response", RoleToolResponse)
	assert.Equal(t, "planner-output", RolePlannerOutput)

	// Test using role constants
	msg := NewMessage("conv-123", RoleUser, "test", "user@example.com")
	assert.Equal(t, RoleUser, msg.Role)
	assert.True(t, msg.Valid())

	// Test tool roles
	toolCallMsg := NewMessage("conv-123", RoleToolCall, "tool call content", "user@example.com")
	assert.Equal(t, RoleToolCall, toolCallMsg.Role)
	assert.True(t, toolCallMsg.Valid())

	toolResponseMsg := NewMessage("conv-123", RoleToolResponse, "tool response content", "user@example.com")
	assert.Equal(t, RoleToolResponse, toolResponseMsg.Role)
	assert.True(t, toolResponseMsg.Valid())

	plannerOutputMsg := NewMessage("conv-123", RolePlannerOutput, "planner output content", "user@example.com")
	assert.Equal(t, RolePlannerOutput, plannerOutputMsg.Role)
	assert.True(t, plannerOutputMsg.Valid())
}
