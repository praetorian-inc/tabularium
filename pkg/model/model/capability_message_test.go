package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCapabilityMessage_BasicFields(t *testing.T) {
	message := &CapabilityMessage{
		JobKey: "#job#example.com#asset#portscan",
		Body:   "Please scan port 443",
		Sender: "user",
	}

	assert.NotEmpty(t, message.JobKey)
	assert.NotEmpty(t, message.Body)
	assert.Equal(t, "user", message.Sender)
}

func TestCapabilityMessage_SenderValidation(t *testing.T) {
	tests := []struct {
		name   string
		sender string
	}{
		{"user sender", "user"},
		{"capability sender", "capability"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := &CapabilityMessage{
				JobKey: "#job#test.com#asset#scan",
				Body:   "test message",
				Sender: tt.sender,
			}

			assert.Equal(t, tt.sender, message.Sender)
		})
	}
}

func TestCapabilityMessage_TTLCalculation(t *testing.T) {
	// Test that TTL is properly set to 1 day from now
	now := time.Now()
	expectedTTL := now.Add(24 * time.Hour).Unix()
	
	message := &CapabilityMessage{
		JobKey: "#job#test.com#asset#scan",
		Body:   "test message",
		Sender: "user",
	}

	// Simulate hook execution
	hooks := message.GetHooks()
	assert.Len(t, hooks, 1)
	
	// Execute the hook
	err := hooks[0].Call()
	assert.NoError(t, err)

	// Verify TTL is set to approximately 1 day from now (within 1 minute tolerance)
	assert.InDelta(t, expectedTTL, message.TTL, 60) // 60 second tolerance
}

func TestCapabilityMessage_CompositeKeyGeneration(t *testing.T) {
	message := &CapabilityMessage{
		JobKey: "#job#example.com#asset#portscan",
		Body:   "test message",
		Sender: "user",
	}

	// Execute hooks to generate key and KSUID
	hooks := message.GetHooks()
	err := hooks[0].Call()
	assert.NoError(t, err)

	// Verify composite key format
	assert.Contains(t, message.Key, "#message#")
	assert.Contains(t, message.Key, message.JobKey)
	assert.NotEmpty(t, message.MessageID)
	
	// Verify key contains both job key and message ID
	expectedKeyPrefix := "#message#" + message.JobKey + "#"
	assert.Contains(t, message.Key, expectedKeyPrefix)
}