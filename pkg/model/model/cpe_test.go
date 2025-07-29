package model

import (
	"strings"
	"testing"

	"github.com/knqyf263/go-cpe/common"
)

func TestCPE_NewCPE(t *testing.T) {
	tests := []struct {
		name          string
		cpe           string
		wantWFN       common.WellFormedName
		expectedError string
	}{
		{
			name: "Valid CPE",
			cpe:  "cpe:2.3:a:microsoft:internet_explorer:8.0.6001:beta:*:sp2:*:*:*:*",
			wantWFN: common.WellFormedName{
				"part":     "a",
				"vendor":   "microsoft",
				"product":  "internet_explorer",
				"version":  "8\\.0\\.6001",
				"update":   "beta",
				"language": "sp2",
			},
			expectedError: "",
		},
		{
			name: "Valid CPE with extra components",
			cpe:  "cpe:2.3:a:microsoft:internet_explorer:8.0.6001:beta:*:sp2:*:*:*:*:*:*:*:*",
			wantWFN: common.WellFormedName{
				"part":     "a",
				"vendor":   "microsoft",
				"product":  "internet_explorer",
				"version":  "8\\.0\\.6001",
				"update":   "beta",
				"language": "sp2",
			},
			expectedError: "",
		},
		{
			name:          "Invalid CPE",
			cpe:           "invalid_cpe",
			wantWFN:       common.WellFormedName{},
			expectedError: "Error: Formatted String must start with \"cpe:2.3\".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCPE, err := NewCPE(tt.cpe)

			if err != nil {
				if tt.expectedError != "" && !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error log containing %q, got %q", tt.expectedError, err.Error())
				} else if tt.expectedError == "" {
					t.Errorf("Expected no error, got %q", err.Error())
				}
				return
			}

			if gotCPE.WellFormedName().String() != tt.wantWFN.String() {
				t.Errorf("NewCPE() = %v, want %v", gotCPE.WellFormedName(), tt.wantWFN)
			}
		})
	}
}

func TestCPE_SearchQuery(t *testing.T) {
	tests := []struct {
		name string
		cpe  string
		want string
	}{
		{
			name: "Full CPE",
			cpe:  "cpe:2.3:a:microsoft:internet_explorer:8.0.6001:beta:*:sp2:*:*:*:*",
			want: "cpe:2.3:a:microsoft:internet_explorer:-",
		},
		{
			name: "CPE without version",
			cpe:  "cpe:2.3:a:microsoft:internet_explorer:*:*:*:*:*:*:*:*",
			want: "cpe:2.3:a:microsoft:internet_explorer:-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCPE(tt.cpe)
			if err != nil {
				t.Errorf("NewCPE(%q) returned error: %v", tt.cpe, err)
				return
			}
			if got.SearchQuery() != tt.want {
				t.Errorf("NewCPE(%q).SearchQuery() = %q, want %q", tt.cpe, got.SearchQuery(), tt.want)
			}
		})
	}
}

func TestCPE_String(t *testing.T) {
	tests := []struct {
		name string
		cpe  string
		want string
	}{
		{
			name: "Full CPE",
			cpe:  "cpe:2.3:a:microsoft:internet_explorer:8.0.6001:beta:*:sp2:*:*:*:*",
			want: "cpe:2.3:a:microsoft:internet_explorer:8.0.6001:beta:*:sp2:*:*:*:*",
		},
		{
			name: "CPE without version",
			cpe:  "cpe:2.3:a:microsoft:internet_explorer:*:*:*:*:*:*:*:*",
			want: "cpe:2.3:a:microsoft:internet_explorer:*:*:*:*:*:*:*:*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCPE(tt.cpe)
			if err != nil {
				t.Errorf("NewCPE(%q) returned error: %v", tt.cpe, err)
				return
			}
			if got.String() != tt.want {
				t.Errorf("NewCPE(%q).String() = %q, want %q", tt.cpe, got.String(), tt.want)
			}
		})
	}
}
