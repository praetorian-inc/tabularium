package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlannerEventSubagentExecution(t *testing.T) {
	event := PlannerEvent{
		Type:               "subagent_execution",
		ConversationID:     "child-conv-123",
		ParentConversation: "parent-conv-456",
		SubagentExecution: &SubagentExecution{
			Agent:   "query",
			Task:    "Analyze security posture",
			Context: map[string]any{"priority": "high"},
		},
	}

	// Verify JSON marshaling
	data, err := json.Marshal(event)
	require.NoError(t, err)

	var decoded PlannerEvent
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "subagent_execution", decoded.Type)
	assert.Equal(t, "child-conv-123", decoded.ConversationID)
	assert.Equal(t, "parent-conv-456", decoded.ParentConversation)
	assert.NotNil(t, decoded.SubagentExecution)
	assert.Equal(t, "query", decoded.SubagentExecution.Agent)
}

func TestPlannerEventSubagentCompletion(t *testing.T) {
	event := PlannerEvent{
		Type:           "subagent_completion",
		ConversationID: "parent-conv-456",
		SubagentCompletion: &SubagentCompletion{
			Agent:   "query",
			Task:    "Analyze security posture",
			Status:  "success",
			Results: "Analysis complete: 3 risks found",
		},
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var decoded PlannerEvent
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "subagent_completion", decoded.Type)
	assert.NotNil(t, decoded.SubagentCompletion)
	assert.Equal(t, "success", decoded.SubagentCompletion.Status)
}
