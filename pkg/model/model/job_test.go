package model

import (
	"encoding/json"
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
		name              string
		config            map[string]string
		expectedClientID  string
	}{
		{
			name:              "client_id present in config",
			config:            map[string]string{"client_id": "agent-123"},
			expectedClientID:  "agent-123",
		},
		{
			name:              "no client_id in config",
			config:            map[string]string{"other_key": "value"},
			expectedClientID:  "",
		},
		{
			name:              "empty config",
			config:            map[string]string{},
			expectedClientID:  "",
		},
		{
			name:              "client_id with special characters",
			config:            map[string]string{"client_id": "agent-abc-123-xyz"},
			expectedClientID:  "agent-abc-123-xyz",
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
		assert.Equal(t, job.Source, context.Source, "Source should be preserved")
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

		assert.Equal(t, "child-source", spawnedJob.Source, "Spawned job should have new source")
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
