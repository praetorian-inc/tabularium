package model

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestNewSetting(t *testing.T) {
	testName := "testSetting"
	testValue := "testValue"
	expectedKey := fmt.Sprintf("#setting#%s", testName)

	setting := NewSetting(testName, testValue)

	if setting.Key != expectedKey {
		t.Errorf("Expected Key to be %s, but got %s", expectedKey, setting.Key)
	}
	if setting.Name != testName {
		t.Errorf("Expected Name to be %s, but got %s", testName, setting.Name)
	}
	if setting.Value != testValue {
		t.Errorf("Expected Value to be %s, but got %v", testValue, setting.Value)
	}
}

func TestNewConfiguration(t *testing.T) {
	testName := "testConfig"
	testValue := 123
	expectedKey := fmt.Sprintf("#configuration#%s", testName)

	config := NewConfiguration(testName, testValue)

	if config.Key != expectedKey {
		t.Errorf("Expected Key to be %s, but got %s", expectedKey, config.Key)
	}
	if config.Name != testName {
		t.Errorf("Expected Name to be %s, but got %s", testName, config.Name)
	}
	if config.Value != testValue {
		t.Errorf("Expected Value to be %v, but got %v", testValue, config.Value)
	}
}

func TestIntOrDefault_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    IntOrDefault
		wantJSON string
		wantErr  bool
	}{
		{
			name:     "Positive integer",
			input:    IntOrDefault(42),
			wantJSON: "42",
			wantErr:  false,
		},
		{
			name:     "Zero",
			input:    IntOrDefault(0),
			wantJSON: "0",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotJSON, err := json.Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(gotJSON) != tt.wantJSON {
				t.Errorf("MarshalJSON() gotJSON = %s, want %s", gotJSON, tt.wantJSON)
			}
		})
	}
}

func TestIntOrDefault_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		inputJSON string
		wantVal   IntOrDefault
		wantErr   bool
	}{
		{
			name:      "Positive integer",
			inputJSON: "42",
			wantVal:   IntOrDefault(42),
			wantErr:   false,
		},
		{
			name:      "Zero",
			inputJSON: "0",
			wantVal:   IntOrDefault(0),
			wantErr:   false,
		},
		{
			name:      "Negative integer",
			inputJSON: "-5",
			wantVal:   IntOrDefault(0), // Expect zero value on error
			wantErr:   true,            // Expect error because type is uint
		},
		{
			name:      "Invalid JSON",
			inputJSON: `"abc"`,
			wantVal:   IntOrDefault(0), // Expect zero value on error
			wantErr:   true,            // Expect error because type is uint
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotVal IntOrDefault
			err := json.Unmarshal([]byte(tt.inputJSON), &gotVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && gotVal != tt.wantVal {
				t.Errorf("UnmarshalJSON() gotVal = %v, want %v", gotVal, tt.wantVal)
			}
			// Optionally check if gotVal is the zero value when an error occurs
			if tt.wantErr && gotVal != 0 {
				t.Errorf("UnmarshalJSON() gotVal = %v, want 0 on error", gotVal)
			}
		})
	}
}
