package model

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlannerEvent_NewPlannerEvent(t *testing.T) {
	conversationID := "550e8400-e29b-41d4-a716-446655440000"
	jobKey := "#job#example.com#10.0.1.5#nuclei#1698422400"
	source := "nuclei"
	target := "#asset#example.com#10.0.1.5"
	status := "JP"
	username := "user@example.com"
	
	event := NewPlannerEvent(conversationID, jobKey, source, target, status, username)
	
	assert.Equal(t, conversationID, event.ConversationID)
	assert.Equal(t, jobKey, event.JobKey)
	assert.Equal(t, source, event.Source)
	assert.Equal(t, target, event.Target)
	assert.Equal(t, status, event.Status)
	assert.Equal(t, username, event.Username)
	assert.NotEmpty(t, event.Key)
	assert.NotEmpty(t, event.CompletedAt)
	assert.NotZero(t, event.TTL)
	assert.True(t, strings.HasPrefix(event.Key, "#plannerevent#"))
	assert.True(t, event.Valid())
}

func TestPlannerEvent_GetKey(t *testing.T) {
	event := NewPlannerEvent("conv-id", "job-key", "source", "target", "JP", "user@example.com")
	assert.Equal(t, event.Key, event.GetKey())
	assert.NotEmpty(t, event.GetKey())
}

func TestPlannerEvent_Valid(t *testing.T) {
	testCases := []struct {
		name     string
		event    PlannerEvent
		expected bool
	}{
		{
			name: "valid event",
			event: PlannerEvent{
				ConversationID: "conv-id",
				JobKey:         "job-key",
				Source:         "nuclei",
			},
			expected: true,
		},
		{
			name: "missing conversation ID",
			event: PlannerEvent{
				JobKey: "job-key",
				Source: "nuclei",
			},
			expected: false,
		},
		{
			name: "missing job key",
			event: PlannerEvent{
				ConversationID: "conv-id",
				Source:         "nuclei",
			},
			expected: false,
		},
		{
			name: "missing source",
			event: PlannerEvent{
				ConversationID: "conv-id",
				JobKey:         "job-key",
			},
			expected: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.event.Valid())
		})
	}
}

func TestPlannerEvent_Hooks(t *testing.T) {
	event := &PlannerEvent{
		ConversationID: "conv-id",
		JobKey:         "job-key",
		Source:         "nuclei",
	}
	
	hooks := event.GetHooks()
	require.Len(t, hooks, 1)
	
	err := hooks[0].Call()
	assert.NoError(t, err)
	assert.NotEmpty(t, event.Key)
	assert.True(t, strings.HasPrefix(event.Key, "#plannerevent#conv-id#job-key"))
}