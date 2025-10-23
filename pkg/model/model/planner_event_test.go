package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlannerEvent_AgentCompletion(t *testing.T) {
	event := PlannerEvent{
		Type:           "agent_completion",
		Username:       "test@example.com",
		User:           "test@example.com",
		ConversationID: "parent-123",
		AgentCompletion: &PlannerAgentCompletion{
			ParentConversationID: "parent-123",
			ChildConversationID:  "child-456",
			ParentToolUseID:      "toolu_abc123",
			Response:             "Child completed successfully",
			ToolsUsed:            []string{"query", "schema"},
			Success:              true,
			CompletedAt:          "2025-10-22T16:00:00Z",
		},
	}

	assert.Equal(t, "agent_completion", event.Type)
	assert.NotNil(t, event.AgentCompletion)
	assert.Equal(t, "parent-123", event.AgentCompletion.ParentConversationID)
	assert.Equal(t, "child-456", event.AgentCompletion.ChildConversationID)
	assert.Equal(t, "toolu_abc123", event.AgentCompletion.ParentToolUseID)
	assert.True(t, event.AgentCompletion.Success)
	assert.Empty(t, event.AgentCompletion.Error)
}

func TestPlannerEvent_AgentCompletionWithError(t *testing.T) {
	event := PlannerEvent{
		Type:           "agent_completion",
		Username:       "test@example.com",
		User:           "test@example.com",
		ConversationID: "parent-123",
		AgentCompletion: &PlannerAgentCompletion{
			ParentConversationID: "parent-123",
			ChildConversationID:  "child-456",
			ParentToolUseID:      "toolu_abc123",
			Response:             "",
			ToolsUsed:            []string{"query"},
			Success:              false,
			Error:                "Child execution failed: timeout",
			CompletedAt:          "2025-10-22T16:00:00Z",
		},
	}

	assert.Equal(t, "agent_completion", event.Type)
	assert.NotNil(t, event.AgentCompletion)
	assert.False(t, event.AgentCompletion.Success)
	assert.NotEmpty(t, event.AgentCompletion.Error)
	assert.Contains(t, event.AgentCompletion.Error, "timeout")
}

func TestPlannerEvent_BackwardCompatibility(t *testing.T) {
	// Test that existing event types still work
	tests := []struct {
		name      string
		eventType string
		setupFunc func() PlannerEvent
	}{
		{
			name:      "planner_execution event",
			eventType: "planner_execution",
			setupFunc: func() PlannerEvent {
				return PlannerEvent{
					Type:           "planner_execution",
					Username:       "test@example.com",
					User:           "test@example.com",
					ConversationID: "conv-123",
					UserMessage: &PlannerUserMessage{
						Message: "Test message",
						Mode:    "query",
					},
				}
			},
		},
		{
			name:      "job_completion event",
			eventType: "job_completion",
			setupFunc: func() PlannerEvent {
				return PlannerEvent{
					Type:           "job_completion",
					Username:       "test@example.com",
					User:           "test@example.com",
					ConversationID: "conv-123",
					JobCompletion: &PlannerJobCompletion{
						JobKey:      "job-123",
						Source:      "nuclei",
						Target:      "example.com",
						Status:      "JP",
						TotalCount:  5,
						CompletedAt: "2025-10-22T16:00:00Z",
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := tt.setupFunc()

			assert.Equal(t, tt.eventType, event.Type)
			assert.Nil(t, event.AgentCompletion, "New field should be nil for existing event types")

			// Should serialize/deserialize correctly
			eventBytes, err := json.Marshal(event)
			require.NoError(t, err)

			var decoded PlannerEvent
			err = json.Unmarshal(eventBytes, &decoded)
			require.NoError(t, err)

			assert.Equal(t, event.Type, decoded.Type)
			assert.Equal(t, event.Username, decoded.Username)
		})
	}
}

func TestPlannerAgentCompletion_Serialization(t *testing.T) {
	completion := PlannerAgentCompletion{
		ParentConversationID: "parent-123",
		ChildConversationID:  "child-456",
		ParentToolUseID:      "toolu_abc123",
		Response:             "Test response",
		ToolsUsed:            []string{"query", "schema", "job"},
		Success:              true,
		CompletedAt:          "2025-10-22T16:00:00Z",
	}

	// Serialize
	data, err := json.Marshal(completion)
	require.NoError(t, err)

	// Deserialize
	var decoded PlannerAgentCompletion
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify all fields
	assert.Equal(t, completion.ParentConversationID, decoded.ParentConversationID)
	assert.Equal(t, completion.ChildConversationID, decoded.ChildConversationID)
	assert.Equal(t, completion.ParentToolUseID, decoded.ParentToolUseID)
	assert.Equal(t, completion.Response, decoded.Response)
	assert.Equal(t, completion.ToolsUsed, decoded.ToolsUsed)
	assert.Equal(t, completion.Success, decoded.Success)
	assert.Equal(t, completion.CompletedAt, decoded.CompletedAt)
}

func TestPlannerEvent_AllEventTypes(t *testing.T) {
	// Test that an event can have any one of the completion types
	event := PlannerEvent{
		Type:           "agent_completion",
		Username:       "test@example.com",
		User:           "test@example.com",
		ConversationID: "conv-123",
		AgentCompletion: &PlannerAgentCompletion{
			ParentConversationID: "parent-123",
			ChildConversationID:  "child-456",
			ParentToolUseID:      "toolu_abc123",
			Response:             "Success",
			ToolsUsed:            []string{"query"},
			Success:              true,
			CompletedAt:          Now(),
		},
	}

	// Should serialize correctly
	data, err := json.Marshal(event)
	require.NoError(t, err)

	// Should deserialize correctly
	var decoded PlannerEvent
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "agent_completion", decoded.Type)
	assert.NotNil(t, decoded.AgentCompletion)
	assert.Nil(t, decoded.JobCompletion)
	assert.Nil(t, decoded.UserMessage)
}

func TestPlannerAgentCompletion_RequiredFields(t *testing.T) {
	completion := PlannerAgentCompletion{
		ParentConversationID: "parent-123",
		ChildConversationID:  "child-456",
		ParentToolUseID:      "toolu_abc123",
		Response:             "Test",
		ToolsUsed:            []string{},  // Initialize to empty slice
		Success:              true,
		CompletedAt:          Now(),
	}

	// ToolsUsed can be empty but should be initialized
	assert.NotNil(t, completion.ToolsUsed)
	assert.Len(t, completion.ToolsUsed, 0)

	// Error should be empty on success
	assert.Empty(t, completion.Error)

	// Required fields must be set
	assert.NotEmpty(t, completion.ParentConversationID)
	assert.NotEmpty(t, completion.ChildConversationID)
	assert.NotEmpty(t, completion.ParentToolUseID)
	assert.NotEmpty(t, completion.CompletedAt)
}
