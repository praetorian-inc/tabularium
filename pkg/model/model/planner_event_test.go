package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFinalizeOutput_Marshal(t *testing.T) {
	output := FinalizeOutput{
		Summary: "Analysis complete",
		Data: map[string]interface{}{
			"count":  5,
			"status": "success",
		},
		Recommendations: []string{
			"Review findings",
			"Update configuration",
		},
	}

	bytes, err := json.Marshal(output)
	require.NoError(t, err)

	var decoded FinalizeOutput
	err = json.Unmarshal(bytes, &decoded)
	require.NoError(t, err)

	assert.Equal(t, output.Summary, decoded.Summary)
	assert.Equal(t, 2, len(decoded.Data))
	assert.Equal(t, 2, len(decoded.Recommendations))
}

func TestPlannerSubagentCompletion_Marshal(t *testing.T) {
	completion := PlannerSubagentCompletion{
		SubagentID:           "subagent_abc123",
		ParentConversationID: "conv_parent_456",
		AgentMode:            "query",
		Status:               "finalized",
		FinalResponse:        "Analysis complete. Found 3 critical issues.",
		ToolCallCount:        5,
		ExecutionTime:        2500,
		StructuredOutput: &FinalizeOutput{
			Summary: "Critical issues found",
			Data: map[string]interface{}{
				"issue_count": 3,
				"severity":    "critical",
			},
			Recommendations: []string{"Patch immediately"},
		},
		Error: "",
	}

	bytes, err := json.Marshal(completion)
	require.NoError(t, err)

	var decoded PlannerSubagentCompletion
	err = json.Unmarshal(bytes, &decoded)
	require.NoError(t, err)

	assert.Equal(t, completion.SubagentID, decoded.SubagentID)
	assert.Equal(t, completion.Status, decoded.Status)
	assert.NotNil(t, decoded.StructuredOutput)
	assert.Equal(t, 3, int(decoded.StructuredOutput.Data["issue_count"].(float64)))
}

func TestPlannerSubagentCompletion_NoStructuredOutput(t *testing.T) {
	completion := PlannerSubagentCompletion{
		SubagentID:           "subagent_def789",
		ParentConversationID: "conv_parent_012",
		AgentMode:            "query",
		Status:               "completed",
		FinalResponse:        "Task completed successfully.",
		ToolCallCount:        2,
		ExecutionTime:        1200,
		StructuredOutput:     nil, // Natural completion without finalize
		Error:                "",
	}

	bytes, err := json.Marshal(completion)
	require.NoError(t, err)

	var decoded PlannerSubagentCompletion
	err = json.Unmarshal(bytes, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "completed", decoded.Status)
	assert.Nil(t, decoded.StructuredOutput)
}

func TestPlannerEvent_SubagentCompletion(t *testing.T) {
	event := PlannerEvent{
		Type:           "subagent_completion",
		Username:       "test-user",
		User:           "test@example.com",
		ConversationID: "conv_parent_123",
		SubagentCompletion: &PlannerSubagentCompletion{
			SubagentID:           "subagent_abc",
			ParentConversationID: "conv_parent_123",
			AgentMode:            "query",
			Status:               "finalized",
			FinalResponse:        "Task completed",
			ToolCallCount:        3,
			ExecutionTime:        1500,
			StructuredOutput: &FinalizeOutput{
				Summary: "Completed successfully",
				Data: map[string]interface{}{
					"result": "success",
				},
				Recommendations: []string{"Next step A", "Next step B"},
			},
		},
	}

	bytes, err := json.Marshal(event)
	require.NoError(t, err)

	var decoded PlannerEvent
	err = json.Unmarshal(bytes, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "subagent_completion", decoded.Type)
	assert.NotNil(t, decoded.SubagentCompletion)
	assert.Equal(t, "subagent_abc", decoded.SubagentCompletion.SubagentID)
	assert.Equal(t, "finalized", decoded.SubagentCompletion.Status)
}

func TestPlannerEvent_MultipleEventTypes(t *testing.T) {
	// Test that old event types still work
	jobEvent := PlannerEvent{
		Type:           "job_completion",
		Username:       "test-user",
		ConversationID: "conv_123",
		JobCompletion: &PlannerJobCompletion{
			JobKey:      "job_key",
			Source:      "nuclei",
			Target:      "example.com",
			Status:      "JC",
			CompletedAt: "2026-01-08T15:00:00Z",
		},
	}

	bytes, err := json.Marshal(jobEvent)
	require.NoError(t, err)
	assert.NotNil(t, bytes)

	// Test new subagent event type
	subagentEvent := PlannerEvent{
		Type:           "subagent_completion",
		Username:       "test-user",
		ConversationID: "conv_456",
		SubagentCompletion: &PlannerSubagentCompletion{
			SubagentID:    "sub_789",
			AgentMode:     "query",
			Status:        "completed",
			FinalResponse: "Done",
		},
	}

	bytes2, err := json.Marshal(subagentEvent)
	require.NoError(t, err)
	assert.NotNil(t, bytes2)
}
