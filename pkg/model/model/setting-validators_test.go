package model

import (
	"testing"
)

func TestScanLevelValidator(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{
			name:    "valid active scan level",
			input:   Active,
			wantErr: false,
		},
		{
			name:    "valid active-low scan level",
			input:   ActiveLow,
			wantErr: false,
		},
		{
			name:    "valid active-passive scan level",
			input:   ActivePassive,
			wantErr: false,
		},
		{
			name:    "invalid scan level string",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "non-string input",
			input:   123,
			wantErr: true,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := scanLevelValidator(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("scanLevelValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRateLimitValidator(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{
			name: "valid simultaneous hosts - minimum",
			input: map[string]any{
				"capabilityRateLimit": 50,
				"simultaneousHosts":   10,
			},
			wantErr: false,
		},
		{
			name: "valid simultaneous hosts - maximum",
			input: map[string]any{
				"capabilityRateLimit": 50,
				"simultaneousHosts":   500,
			},
			wantErr: false,
		},
		{
			name: "valid simultaneous hosts - middle value",
			input: map[string]any{
				"capabilityRateLimit": 50,
				"simultaneousHosts":   250,
			},
			wantErr: false,
		},
		{
			name: "invalid simultaneous hosts - below minimum",
			input: map[string]any{
				"capabilityRateLimit": 50,
				"simultaneousHosts":   5,
			},
			wantErr: true,
		},
		{
			name: "invalid simultaneous hosts - above maximum",
			input: map[string]any{
				"capabilityRateLimit": 50,
				"simultaneousHosts":   505,
			},
			wantErr: true,
		},
		{
			name:    "invalid input - not a JSON object",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rateLimitValidator(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("rateLimitValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockedCapabilitiesValidator(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{
			name: "not a list",
			input: map[string]any{
				"test": "test",
			},
			wantErr: true,
		},
		{
			name:    "empty list",
			input:   []string{},
			wantErr: true,
		},
		{
			name: "an int list with entries",
			input: map[string]any{
				"capabilities": []int{1, 2},
			},
			wantErr: true,
		},
		{
			name: "a string list with entries",
			input: map[string]any{
				"capabilities": []string{"nuclei", "fingerprint"},
			},
			wantErr: false,
		},
		{
			name: "an any list with entries",
			input: map[string]any{
				"capabilities": []any{"nuclei", "fingerprint"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := blockedCapabilitiesValidator(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("blockedCapabilitiesValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
