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
