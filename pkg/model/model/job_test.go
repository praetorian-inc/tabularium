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
	parent := NewAsset("example", "com")

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

			webpage := NewWebpage(*parsedURL, &parent)
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

		httpsWebpage := NewWebpage(*httpsURL, &parent)
		httpWebpage := NewWebpage(*httpURL, &parent)

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
