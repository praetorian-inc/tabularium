package model

import (
	"log/slog"
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
