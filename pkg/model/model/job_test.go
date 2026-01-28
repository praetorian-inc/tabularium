package model

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"net/url"
	"strings"
	"testing"
)

func TestJob_ImportAssets(t *testing.T) {
	// Create a buffer to capture log output
	var logBuffer strings.Builder
	slog.SetDefault(slog.New(slog.NewTextHandler(&logBuffer, nil)))

	tests := []struct {
		name          string
		jobConfig     map[string]string
		want          bool
		expectedError string
	}{
		{
			name:          "no config key returns true",
			jobConfig:     map[string]string{},
			want:          true,
			expectedError: "",
		},
		{
			name:          "config set to true returns true",
			jobConfig:     map[string]string{"importAssets": "true"},
			want:          true,
			expectedError: "",
		},
		{
			name:          "config set to false returns false",
			jobConfig:     map[string]string{"importAssets": "false"},
			want:          false,
			expectedError: "",
		},
		{
			name:          "invalid boolean value returns false",
			jobConfig:     map[string]string{"importAssets": "invalid"},
			want:          false,
			expectedError: "Error parsing importAssets config value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear the buffer before each test
			logBuffer.Reset()

			job := &Job{Config: tt.jobConfig}
			context := job.ToContext()
			if got := context.ImportAssets(); got != tt.want {
				t.Errorf("Job.ImportAssets() = %v, want %v", got, tt.want)
			}

			// Check error logging
			logOutput := logBuffer.String()
			if tt.expectedError != "" && !strings.Contains(logOutput, tt.expectedError) {
				t.Errorf("Expected error log containing %q, got %q", tt.expectedError, logOutput)
			} else if tt.expectedError == "" && logOutput != "" {
				t.Errorf("Expected no error log, got %q", logOutput)
			}
		})
	}
}

func TestJob_ImportVulnerabilities(t *testing.T) {
	// Create a buffer to capture log output
	var logBuffer strings.Builder
	slog.SetDefault(slog.New(slog.NewTextHandler(&logBuffer, nil)))

	tests := []struct {
		name          string
		jobConfig     map[string]string
		want          bool
		expectedError string
	}{
		{
			name:          "no config key returns true",
			jobConfig:     map[string]string{},
			want:          true,
			expectedError: "",
		},
		{
			name:          "config set to true returns true",
			jobConfig:     map[string]string{"importVulnerabilities": "true"},
			want:          true,
			expectedError: "",
		},
		{
			name:          "config set to false returns false",
			jobConfig:     map[string]string{"importVulnerabilities": "false"},
			want:          false,
			expectedError: "",
		},
		{
			name:          "invalid boolean value returns false",
			jobConfig:     map[string]string{"importVulnerabilities": "invalid"},
			want:          false,
			expectedError: "Error parsing importVulnerabilities config value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear the buffer before each test
			logBuffer.Reset()

			job := &Job{Config: tt.jobConfig}
			context := job.ToContext()
			if got := context.ImportVulnerabilities(); got != tt.want {
				t.Errorf("Job.ImportVulnerabilities() = %v, want %v", got, tt.want)
			}

			// Check error logging
			logOutput := logBuffer.String()
			if tt.expectedError != "" && !strings.Contains(logOutput, tt.expectedError) {
				t.Errorf("Expected error log containing %q, got %q", tt.expectedError, logOutput)
			} else if tt.expectedError == "" && logOutput != "" {
				t.Errorf("Expected no error log, got %q", logOutput)
			}
		})
	}
}

func TestJob_GetParent(t *testing.T) {
	gladiator := NewAsset("gladiator.systems", "gladiator.systems")
	marcus := NewAsset("marcus.gladiator.systems", "marcus.gladiator.systems")
	preseed := NewPreseed("whois+company", "Chariot Systems", "Chariot Systems")

	tests := []struct {
		name   string
		target Target
		parent Target
		want   string
	}{
		{
			name:   "no parent key returns target key",
			target: &gladiator,
			parent: nil,
			want:   "#asset#gladiator.systems#gladiator.systems",
		},
		{
			name:   "parent key returns parent key",
			target: &marcus,
			parent: &gladiator,
			want:   "#asset#gladiator.systems#gladiator.systems",
		},
		{
			name:   "preseed target returns preseed key",
			target: &preseed,
			want:   "#preseed#whois+company#Chariot Systems#Chariot Systems",
		},
		{
			name:   "preseed parent returns preseed key",
			target: &gladiator,
			parent: &preseed,
			want:   "#preseed#whois+company#Chariot Systems#Chariot Systems",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := &ResultContext{Target: TargetWrapper{Model: tt.target}, Parent: TargetWrapper{Model: tt.parent}}
			if got := context.GetParent(); got.GetKey() != tt.want {
				t.Errorf("ResultContext.GetParent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJob_WebpageKeyCreationWithProtocol(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		source      string
		expectedKey string
	}{
		{
			name:        "HTTPS webpage should include protocol in key",
			url:         "https://example.com/path",
			source:      "test-source",
			expectedKey: "#job#https://example.com#/path#test-source",
		},
		{
			name:        "HTTP webpage should include protocol in key",
			url:         "http://example.com/path",
			source:      "test-source",
			expectedKey: "#job#http://example.com#/path#test-source",
		},
		{
			name:        "Ports should be included in key",
			url:         "https://example.com:8080/path",
			source:      "test-source",
			expectedKey: "#job#https://example.com:8080#/path#test-source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedURL, err := url.Parse(tt.url)
			if err != nil {
				t.Fatalf("Failed to parse URL: %v", err)
			}

			webpage := NewWebpage(*parsedURL, nil)
			job := NewJob(tt.source, &webpage)

			if job.Key != tt.expectedKey {
				t.Errorf("Expected key %q, got %q", tt.expectedKey, job.Key)
			}
		})
	}

	// Test that HTTP and HTTPS have different keys
	t.Run("HTTP and HTTPS protocols create different keys", func(t *testing.T) {
		httpsURL, _ := url.Parse("https://example.com/path")
		httpURL, _ := url.Parse("http://example.com/path")

		httpsWebpage := NewWebpage(*httpsURL, nil)
		httpWebpage := NewWebpage(*httpURL, nil)

		httpsJob := NewJob("test-source", &httpsWebpage)
		httpJob := NewJob("test-source", &httpWebpage)

		if httpsJob.Key == httpJob.Key {
			t.Errorf("HTTP and HTTPS jobs should have different keys, both got: %q", httpsJob.Key)
		}
	})
}

func TestJob_Parameters(t *testing.T) {
	dummy := NewAsset("example.com", "example.com")
	job := NewJob("test-source", &dummy)

	job.Config = map[string]string{"config1": "config-value1", "config2": "config-value2"}
	job.Secret = map[string]string{"secret1": "secret-value1", "secret2": "secret-value2"}

	encoded, err := json.Marshal(job)
	require.NoError(t, err)
	assert.Contains(t, string(encoded), "config1")
	assert.Contains(t, string(encoded), "config-value1")
	assert.Contains(t, string(encoded), "config2")
	assert.Contains(t, string(encoded), "config-value2")
	assert.Contains(t, string(encoded), "secret1")
	assert.Contains(t, string(encoded), "secret-value1")
	assert.Contains(t, string(encoded), "secret2")
	assert.Contains(t, string(encoded), "secret-value2")
}

func TestJob_Conversation(t *testing.T) {
	dummy := NewAsset("example.com", "example.com")

	tests := []struct {
		name         string
		conversation string
		shouldOmit   bool
	}{
		{
			name:         "conversation field with UUID",
			conversation: "550e8400-e29b-41d4-a716-446655440000",
			shouldOmit:   false,
		},
		{
			name:         "empty conversation field omitted",
			conversation: "",
			shouldOmit:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := NewJob("test-source", &dummy)
			job.Conversation = tt.conversation

			encoded, err := json.Marshal(job)
			require.NoError(t, err)

			if tt.shouldOmit {
				assert.NotContains(t, string(encoded), "conversation")
			} else {
				assert.Contains(t, string(encoded), "conversation")
				assert.Contains(t, string(encoded), tt.conversation)
				assert.Equal(t, tt.conversation, job.Conversation)
			}
		})
	}
}

// TestJob_AgentClientIDExtraction tests that agent client ID is correctly extracted from job config
func TestJob_AgentClientIDExtraction(t *testing.T) {
	dummy := NewAsset("example.com", "example.com")

	tests := []struct {
		name             string
		config           map[string]string
		expectedClientID string
	}{
		{
			name:             "client_id present in config",
			config:           map[string]string{"client_id": "agent-123"},
			expectedClientID: "agent-123",
		},
		{
			name:             "no client_id in config",
			config:           map[string]string{"other_key": "value"},
			expectedClientID: "",
		},
		{
			name:             "empty config",
			config:           map[string]string{},
			expectedClientID: "",
		},
		{
			name:             "client_id with special characters",
			config:           map[string]string{"client_id": "agent-abc-123-xyz"},
			expectedClientID: "agent-abc-123-xyz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := NewJob("test-source", &dummy)
			job.Config = tt.config

			context := job.ToContext()
			assert.Equal(t, tt.expectedClientID, context.AgentClientID, "AgentClientID should be extracted from config")
		})
	}
}

// TestResultContext_GetAgentClientID tests the validation logic in GetAgentClientID
func TestResultContext_GetAgentClientID(t *testing.T) {
	// Create a buffer to capture log output
	var logBuffer strings.Builder
	slog.SetDefault(slog.New(slog.NewTextHandler(&logBuffer, nil)))

	tests := []struct {
		name          string
		agentClientID string
		expectedValue string
		expectWarning bool
	}{
		{
			name:          "valid client ID",
			agentClientID: "agent-123",
			expectedValue: "agent-123",
			expectWarning: false,
		},
		{
			name:          "empty client ID",
			agentClientID: "",
			expectedValue: "",
			expectWarning: false,
		},
		{
			name:          "client ID with whitespace",
			agentClientID: "  agent-123  ",
			expectedValue: "agent-123",
			expectWarning: false,
		},
		{
			name:          "client ID with only whitespace",
			agentClientID: "   ",
			expectedValue: "",
			expectWarning: true,
		},
		{
			name:          "client ID with tabs and spaces",
			agentClientID: "\t\n  \t",
			expectedValue: "",
			expectWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear the buffer before each test
			logBuffer.Reset()

			context := &ResultContext{
				AgentClientID: tt.agentClientID,
			}

			result := context.GetAgentClientID()
			assert.Equal(t, tt.expectedValue, result, "GetAgentClientID should return expected value")

			// Check for warning log
			logOutput := logBuffer.String()
			if tt.expectWarning {
				assert.Contains(t, logOutput, "AgentClientID contains only whitespace", "Should log warning for whitespace-only client ID")
			} else {
				assert.NotContains(t, logOutput, "AgentClientID contains only whitespace", "Should not log warning for valid client ID")
			}
		})
	}
}

// TestJob_ContextPropagation tests that context is correctly propagated through job processing
func TestJob_ContextPropagation(t *testing.T) {
	dummy := NewAsset("example.com", "example.com")

	t.Run("ToContext preserves all fields", func(t *testing.T) {
		job := NewJob("test-source", &dummy)
		job.Username = "testuser@example.com"
		job.Config = map[string]string{"client_id": "agent-456", "other_key": "value"}
		job.Secret = map[string]string{"api_key": "secret123"}
		job.Capabilities = []string{"nmap", "nuclei"}
		job.Queue = "priority"

		context := job.ToContext()

		assert.Equal(t, job.Username, context.Username, "Username should be preserved")
		assert.Equal(t, job.GetCapability(), context.Source, "Source should be preserved")
		assert.Equal(t, job.Config, context.Config, "Config should be preserved")
		assert.Equal(t, job.Secret, context.Secret, "Secret should be preserved")
		assert.Equal(t, job.Capabilities, context.Capabilities, "Capabilities should be preserved")
		assert.Equal(t, job.Queue, context.Queue, "Queue should be preserved")
		assert.Equal(t, "agent-456", context.AgentClientID, "AgentClientID should be extracted from config")
	})

	t.Run("SpawnJob propagates context", func(t *testing.T) {
		job := NewJob("parent-source", &dummy)
		job.Capabilities = []string{"capability1", "capability2"}
		job.Origin = TargetWrapper{Model: &dummy}

		context := job.ToContext()
		newTarget := NewAsset("new.example.com", "new.example.com")
		newConfig := map[string]string{"new_key": "new_value"}

		spawnedJob := context.SpawnJob("child-source", &newTarget, newConfig)

		assert.Equal(t, "child-source", spawnedJob.GetCapability(), "Spawned job should have new source capability")
		assert.True(t, strings.HasPrefix(spawnedJob.Source, "child-source#"), "Source should start with child-source# followed by timestamp")
		assert.Equal(t, newConfig, spawnedJob.Config, "Spawned job should have new config")
		assert.Equal(t, job.Capabilities, spawnedJob.Capabilities, "Spawned job should inherit capabilities")
		assert.Equal(t, job.Origin, spawnedJob.Origin, "Spawned job should inherit origin")
	})

	t.Run("GetAgentClientID returns validated ID", func(t *testing.T) {
		job := NewJob("test-source", &dummy)
		job.Config = map[string]string{"client_id": "  agent-789  "}

		context := job.ToContext()
		clientID := context.GetAgentClientID()

		assert.Equal(t, "agent-789", clientID, "GetAgentClientID should trim whitespace")
	})
}

// TestAegisAgent_Valid tests the validation logic for AegisAgent
func TestAegisAgent_Valid(t *testing.T) {
	tests := []struct {
		name     string
		agent    *AegisAgent
		expected bool
	}{
		{
			name: "valid agent with both ClientID and Key",
			agent: &AegisAgent{
				ClientID: "agent-123",
			},
			expected: true,
		},
		{
			name: "invalid agent with empty ClientID",
			agent: &AegisAgent{
				ClientID: "",
			},
			expected: false,
		},
		{
			name: "invalid agent with empty Key",
			agent: &AegisAgent{
				ClientID: "agent-123",
			},
			expected: false,
		},
		{
			name: "invalid agent with whitespace-only ClientID",
			agent: &AegisAgent{
				ClientID: "   ",
			},
			expected: false,
		},
		{
			name: "invalid agent with both fields empty",
			agent: &AegisAgent{
				ClientID: "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set Key for valid cases
			if tt.agent.ClientID != "" && tt.expected {
				tt.agent.Key = "#aegisagent#" + tt.agent.ClientID
			}

			result := tt.agent.Valid()
			assert.Equal(t, tt.expected, result, "Agent validation should match expected result")
		})
	}
}

func TestJob_Valid(t *testing.T) {
	target := NewAsset("example.com", "example.com")

	noKey := Job{}
	badKey := Job{Key: "malformed"}
	goodJob := NewJob("test", &target)

	emptyCredentials := NewJob("test", &target)
	emptyCredentials.CredentialIDs = []string{""}
	goodJobWithCredentials := NewJob("test", &target)
	goodJobWithCredentials.CredentialIDs = []string{"cred-id"}

	assert.False(t, noKey.Valid())
	assert.False(t, badKey.Valid())
	assert.False(t, emptyCredentials.Valid())

	assert.True(t, goodJob.Valid())
	assert.True(t, goodJobWithCredentials.Valid())
}

func TestJob_SetStatus_StartedAndFinishedTimes(t *testing.T) {
	target := NewAsset("example.com", "example.com")

	t.Run("queued status clears started and finished times", func(t *testing.T) {
		job := NewJob("test-source", &target)

		// Set some initial times
		job.Started = "2023-10-27T10:00:00Z"
		job.Finished = "2023-10-27T10:05:00Z"

		// Set status to queued
		job.SetStatus(Queued)

		// Verify both times are cleared
		assert.Empty(t, job.Started, "Started time should be cleared when status is set to Queued")
		assert.Empty(t, job.Finished, "Finished time should be cleared when status is set to Queued")
		assert.True(t, strings.HasPrefix(job.Status, "JQ#"), "Status should start with JQ# followed by timestamp")
		assert.Equal(t, Queued, job.GetStatus(), "GetStatus should return Queued")
	})

	t.Run("running status sets started time when unset", func(t *testing.T) {
		job := NewJob("test-source", &target)

		// Ensure started is empty
		job.Started = ""

		// Set status to running
		job.SetStatus(Running)

		// Verify started time is set
		assert.NotEmpty(t, job.Started, "Started time should be set when status changes to Running")
		assert.Empty(t, job.Finished, "Finished time should remain empty")
		assert.True(t, strings.HasPrefix(job.Status, "JR#"), "Status should start with JR# followed by timestamp")
		assert.Equal(t, Running, job.GetStatus(), "GetStatus should return Running")
	})

	t.Run("running status does not overwrite existing started time", func(t *testing.T) {
		job := NewJob("test-source", &target)

		// Set an existing started time
		existingStarted := "2023-10-27T09:00:00Z"
		job.Started = existingStarted

		// Set status to running
		job.SetStatus(Running)

		// Verify started time is not overwritten
		assert.Equal(t, existingStarted, job.Started, "Started time should not be overwritten when already set")
		assert.True(t, strings.HasPrefix(job.Status, "JR#"), "Status should start with JR# followed by timestamp")
		assert.Equal(t, Running, job.GetStatus(), "GetStatus should return Running")
	})

	t.Run("pass status sets finished time when unset", func(t *testing.T) {
		job := NewJob("test-source", &target)

		// Ensure finished is empty
		job.Finished = ""

		// Set status to pass
		job.SetStatus(Pass)

		// Verify finished time is set
		assert.NotEmpty(t, job.Finished, "Finished time should be set when status changes to Pass")
		assert.True(t, strings.HasPrefix(job.Status, "JP#"), "Status should start with JP# followed by timestamp")
		assert.Equal(t, Pass, job.GetStatus(), "GetStatus should return Pass")
	})

	t.Run("fail status sets finished time when unset", func(t *testing.T) {
		job := NewJob("test-source", &target)

		// Ensure finished is empty
		job.Finished = ""

		// Set status to fail
		job.SetStatus(Fail)

		// Verify finished time is set
		assert.NotEmpty(t, job.Finished, "Finished time should be set when status changes to Fail")
		assert.True(t, strings.HasPrefix(job.Status, "JF#"), "Status should start with JF# followed by timestamp")
		assert.Equal(t, Fail, job.GetStatus(), "GetStatus should return Fail")
	})

	t.Run("pass status does not overwrite existing finished time", func(t *testing.T) {
		job := NewJob("test-source", &target)

		// Set an existing finished time
		existingFinished := "2023-10-27T10:05:00Z"
		job.Finished = existingFinished

		// Set status to pass
		job.SetStatus(Pass)

		// Verify finished time is not overwritten
		assert.Equal(t, existingFinished, job.Finished, "Finished time should not be overwritten when already set")
		assert.True(t, strings.HasPrefix(job.Status, "JP#"), "Status should start with JP# followed by timestamp")
		assert.Equal(t, Pass, job.GetStatus(), "GetStatus should return Pass")
	})

	t.Run("fail status does not overwrite existing finished time", func(t *testing.T) {
		job := NewJob("test-source", &target)

		// Set an existing finished time
		existingFinished := "2023-10-27T10:05:00Z"
		job.Finished = existingFinished

		// Set status to fail
		job.SetStatus(Fail)

		// Verify finished time is not overwritten
		assert.Equal(t, existingFinished, job.Finished, "Finished time should not be overwritten when already set")
		assert.True(t, strings.HasPrefix(job.Status, "JF#"), "Status should start with JF# followed by timestamp")
		assert.Equal(t, Fail, job.GetStatus(), "GetStatus should return Fail")
	})

	t.Run("full job lifecycle manages times correctly", func(t *testing.T) {
		job := NewJob("test-source", &target)

		// Initial state: no times set
		assert.Empty(t, job.Started)
		assert.Empty(t, job.Finished)

		// Move to running: started should be set
		job.SetStatus(Running)
		assert.NotEmpty(t, job.Started, "Started should be set when moving to Running")
		assert.Empty(t, job.Finished, "Finished should still be empty")
		startedTime := job.Started

		// Move to pass: finished should be set
		job.SetStatus(Pass)
		assert.Equal(t, startedTime, job.Started, "Started should remain unchanged")
		assert.NotEmpty(t, job.Finished, "Finished should be set when moving to Pass")

		// Re-queue: both times should be cleared
		job.SetStatus(Queued)
		assert.Empty(t, job.Started, "Started should be cleared when re-queued")
		assert.Empty(t, job.Finished, "Finished should be cleared when re-queued")

		// Run again: new started time should be set
		job.SetStatus(Running)
		assert.NotEmpty(t, job.Started, "Started should be set again after re-queuing")
		secondStartedTime := job.Started

		// Fail this time: finished should be set
		job.SetStatus(Fail)
		assert.NotEmpty(t, job.Finished, "Finished should be set when failing")
		assert.Equal(t, secondStartedTime, job.Started, "Started should remain unchanged during failure")
	})
}

func TestJob_Traced(t *testing.T) {
	target := NewAsset("example.com", "example.com")

	t.Run("NewJob does not auto-initialize TraceID", func(t *testing.T) {
		job := NewJob("test-source", &target)
		assert.Empty(t, job.TraceID, "NewJob should not auto-initialize TraceID")
		assert.False(t, job.IsTraced(), "IsTraced should return false for untraced job")
	})

	t.Run("Traced initializes TraceID when empty", func(t *testing.T) {
		job := NewJob("test-source", &target)
		assert.Empty(t, job.TraceID, "TraceID should be empty before calling Traced")

		result := job.Traced()

		assert.NotEmpty(t, job.TraceID, "TraceID should be set after calling Traced")
		assert.Same(t, &job, result, "Traced should return pointer to same job for chaining")
		assert.True(t, job.IsTraced(), "IsTraced should return true after Traced is called")
	})

	t.Run("Traced does not overwrite existing TraceID", func(t *testing.T) {
		job := NewJob("test-source", &target)
		existingTraceID := "existing-trace-id-12345"
		job.TraceID = existingTraceID

		job.Traced()

		assert.Equal(t, existingTraceID, job.TraceID, "Traced should not overwrite existing TraceID")
	})

	t.Run("Traced enables method chaining", func(t *testing.T) {
		job := NewJob("test-source", &target)

		// Should be able to chain Traced with other operations
		result := job.Traced()
		assert.NotNil(t, result, "Traced should return non-nil for chaining")
		assert.NotEmpty(t, result.TraceID, "Chained result should have TraceID set")
	})

	t.Run("IsTraced returns correct state", func(t *testing.T) {
		job := NewJob("test-source", &target)

		// Before tracing
		assert.False(t, job.IsTraced(), "IsTraced should return false when TraceID is empty")

		// After tracing
		job.Traced()
		assert.True(t, job.IsTraced(), "IsTraced should return true when TraceID is set")
	})

	t.Run("SpawnJob does not propagate trace when parent is untraced", func(t *testing.T) {
		job := NewJob("parent-source", &target)
		// Do NOT call Traced - parent should have empty TraceID

		context := job.ToContext()
		newTarget := NewAsset("new.example.com", "new.example.com")
		spawnedJob := context.SpawnJob("child-source", &newTarget, nil)

		assert.Empty(t, spawnedJob.TraceID, "Spawned job should not have TraceID when parent is untraced")
		assert.Empty(t, spawnedJob.ParentSpanID, "Spawned job should not have ParentSpanID when parent is untraced")
	})

	t.Run("SpawnJob propagates trace when parent is traced", func(t *testing.T) {
		job := NewJob("parent-source", &target)
		job.Traced() // Enable tracing on parent
		job.CurrentSpanID = "parent-span-123"

		context := job.ToContext()
		newTarget := NewAsset("new.example.com", "new.example.com")
		spawnedJob := context.SpawnJob("child-source", &newTarget, nil)

		assert.Equal(t, job.TraceID, spawnedJob.TraceID, "Spawned job should inherit TraceID from traced parent")
		assert.Equal(t, "parent-span-123", spawnedJob.ParentSpanID, "Spawned job should have parent's CurrentSpanID as ParentSpanID")
	})
}

func TestJob_SourceAndStatusConsistency(t *testing.T) {
	target := NewAsset("example.com", "example.com")

	t.Run("SetStatus updates status with timestamp", func(t *testing.T) {
		job := NewJob("test-source", &target)

		// Set a specific Updated time and verify exact format
		job.Updated = "2023-10-27T10:00:00Z"
		job.SetStatus(Running)

		// Status should have the exact format: {status}#{timestamp}
		assert.Equal(t, "JR#2023-10-27T10:00:00Z", job.Status, "Status should be exactly JR#{Updated}")

		// Verify GetStatus extracts just the status part
		assert.Equal(t, Running, job.GetStatus(), "GetStatus should return just the status without timestamp")
	})

	t.Run("SetCapability updates source with timestamp", func(t *testing.T) {
		job := NewJob("test-source", &target)

		// Set a specific Updated time and verify exact format
		job.Updated = "2023-10-27T10:00:00Z"
		job.SetCapability("new-capability")

		// Source should have the exact format: {capability}#{timestamp}
		assert.Equal(t, "new-capability#2023-10-27T10:00:00Z", job.Source, "Source should be exactly {capability}#{Updated}")

		// Verify GetCapability extracts just the capability part
		assert.Equal(t, "new-capability", job.GetCapability(), "GetCapability should return just the capability without timestamp")
	})

	t.Run("Update sets timestamp then calls SetStatus and SetCapability", func(t *testing.T) {
		job := NewJob("test-source", &target)

		// Get the initial capability
		initialCapability := job.GetCapability()

		// Call Update which should:
		// 1. Set job.Updated to Now()
		// 2. Call SetStatus(status) which uses job.Updated
		// 3. Call SetCapability(capability) which uses job.Updated
		job.Update(Running)

		// Verify Updated was set
		assert.NotEmpty(t, job.Updated, "Update should set the Updated field")

		// Extract timestamps from Status and Source
		statusParts := strings.Split(job.Status, "#")
		sourceParts := strings.Split(job.Source, "#")

		require.Len(t, statusParts, 2, "Status should have format {status}#{timestamp}")
		require.Len(t, sourceParts, 2, "Source should have format {capability}#{timestamp}")

		// Verify the format is exactly {status}#{job.Updated} and {capability}#{job.Updated}
		assert.Equal(t, fmt.Sprintf("%s#%s", Running, job.Updated), job.Status, "Status should be {status}#{Updated}")
		assert.Equal(t, fmt.Sprintf("%s#%s", initialCapability, job.Updated), job.Source, "Source should be {capability}#{Updated}")

		// Both timestamps should match job.Updated
		assert.Equal(t, job.Updated, statusParts[1], "Status timestamp should match Updated field")
		assert.Equal(t, job.Updated, sourceParts[1], "Source timestamp should match Updated field")
	})

	t.Run("Update keeps Source and Status timestamps consistent", func(t *testing.T) {
		job := NewJob("test-source", &target)

		// Call Update which should update both Status and Source with the same timestamp
		job.Update(Running)

		// Extract timestamps from Status and Source
		statusParts := strings.Split(job.Status, "#")
		sourceParts := strings.Split(job.Source, "#")

		require.Len(t, statusParts, 2, "Status should have format {status}#{timestamp}")
		require.Len(t, sourceParts, 2, "Source should have format {capability}#{timestamp}")

		statusTimestamp := statusParts[1]
		sourceTimestamp := sourceParts[1]

		// Both should have the same timestamp
		assert.Equal(t, statusTimestamp, sourceTimestamp, "Status and Source should have the same timestamp after Update()")
		assert.Equal(t, job.Updated, statusTimestamp, "Status timestamp should match Updated field")
		assert.Equal(t, job.Updated, sourceTimestamp, "Source timestamp should match Updated field")
	})

	t.Run("Multiple state changes maintain timestamp consistency", func(t *testing.T) {
		job := NewJob("test-source", &target)

		// First update
		job.Update(Running)
		statusParts1 := strings.Split(job.Status, "#")
		sourceParts1 := strings.Split(job.Source, "#")
		assert.Equal(t, statusParts1[1], sourceParts1[1], "Timestamps should match after first Update")
		assert.Equal(t, Running, job.GetStatus(), "Status should be Running after first Update")

		// Second update with different status
		job.Update(Pass)
		statusParts2 := strings.Split(job.Status, "#")
		sourceParts2 := strings.Split(job.Source, "#")
		assert.Equal(t, statusParts2[1], sourceParts2[1], "Timestamps should match after second Update")
		assert.Equal(t, Pass, job.GetStatus(), "Status should be Pass after second Update")

		// Verify both updates have valid timestamp format
		assert.NotEmpty(t, statusParts1[1], "First update should have a timestamp")
		assert.NotEmpty(t, statusParts2[1], "Second update should have a timestamp")
	})

	t.Run("SetStatus alone does not update Source timestamp", func(t *testing.T) {
		job := NewJob("test-source", &target)

		// Get initial source
		initialSource := job.Source

		// Set a specific Updated time and call SetStatus
		job.Updated = "2023-10-27T10:00:00Z"
		job.SetStatus(Running)

		// Source should remain unchanged since SetStatus doesn't modify Source
		assert.Equal(t, initialSource, job.Source, "SetStatus should not modify Source field")

		// But Status should have the new timestamp
		assert.Contains(t, job.Status, "2023-10-27T10:00:00Z", "Status should have the new timestamp")
	})

	t.Run("GetStatus and GetCapability correctly extract values", func(t *testing.T) {
		job := NewJob("test-source", &target)

		// Manually set Status and Source with known timestamps
		job.Status = "JR#2023-10-27T10:00:00Z"
		job.Source = "portscan#2023-10-27T10:00:00Z"

		assert.Equal(t, "JR", job.GetStatus(), "GetStatus should extract status without timestamp")
		assert.Equal(t, "portscan", job.GetCapability(), "GetCapability should extract capability without timestamp")
	})

	t.Run("Is and Was methods work with timestamped status", func(t *testing.T) {
		job := NewJob("test-source", &target)

		// Set initial status
		job.SetStatus(Queued)
		job.originalStatus = job.Status

		// Verify Is method
		assert.True(t, job.Is(Queued), "Is(Queued) should return true")
		assert.False(t, job.Is(Running), "Is(Running) should return false")

		// Change status
		job.SetStatus(Running)

		// Verify Was method checks original status
		assert.True(t, job.Was(Queued), "Was(Queued) should return true")
		assert.False(t, job.Was(Running), "Was(Running) should return false")

		// Verify Is method checks current status
		assert.True(t, job.Is(Running), "Is(Running) should return true")
		assert.False(t, job.Is(Queued), "Is(Queued) should return false")
	})
}
