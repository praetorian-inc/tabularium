package model

import (
	"testing"
)

func TestGetFieldNameForType(t *testing.T) {
	tests := []struct {
		name     string
		typename string
		want     string
	}{
		{
			name:     "basic auth without credentials",
			typename: AuthTypeBasic,
			want:     FieldBasicAuth,
		},
		{
			name:     "basic auth credentials",
			typename: CredentialTypeBasic,
			want:     FieldBasicAuth,
		},
		{
			name:     "api key without credentials",
			typename: AuthTypeAPIKey,
			want:     FieldAPIKeyAuth,
		},
		{
			name:     "api key credentials",
			typename: CredentialTypeAPIKey,
			want:     FieldAPIKeyAuth,
		},
		{
			name:     "bearer token without credentials",
			typename: AuthTypeBearerToken,
			want:     FieldBearerTokenAuth,
		},
		{
			name:     "bearer token credentials",
			typename: CredentialTypeBearerToken,
			want:     FieldBearerTokenAuth,
		},
		{
			name:     "unsupported type",
			typename: AuthTypeUnsupported,
			want:     "",
		},
		{
			name:     "unknown type",
			typename: "SomeUnknownType",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFieldNameForType(tt.typename)
			if got != tt.want {
				t.Errorf("GetFieldNameForType(%q) = %q, want %q", tt.typename, got, tt.want)
			}
		})
	}
}

func TestIsCredentialType(t *testing.T) {
	tests := []struct {
		name     string
		typename string
		want     bool
	}{
		{
			name:     "basic auth credentials",
			typename: CredentialTypeBasic,
			want:     true,
		},
		{
			name:     "api key credentials",
			typename: CredentialTypeAPIKey,
			want:     true,
		},
		{
			name:     "bearer token credentials",
			typename: CredentialTypeBearerToken,
			want:     true,
		},
		{
			name:     "basic auth without credentials",
			typename: AuthTypeBasic,
			want:     false,
		},
		{
			name:     "api key without credentials",
			typename: AuthTypeAPIKey,
			want:     false,
		},
		{
			name:     "bearer token without credentials",
			typename: AuthTypeBearerToken,
			want:     false,
		},
		{
			name:     "unsupported type",
			typename: AuthTypeUnsupported,
			want:     false,
		},
		{
			name:     "unknown type",
			typename: "SomeUnknownType",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCredentialType(tt.typename)
			if got != tt.want {
				t.Errorf("IsCredentialType(%q) = %v, want %v", tt.typename, got, tt.want)
			}
		})
	}
}

func TestConstantsAreLowercase(t *testing.T) {
	// Verify that API constants that must be lowercase for Burp API are indeed lowercase
	tests := []struct {
		name     string
		constant string
	}{
		{"DestinationHeader", DestinationHeader},
		{"DestinationQuery", DestinationQuery},
		{"DestinationCookie", DestinationCookie},
		{"RequestMethodGet", RequestMethodGet},
		{"RequestMethodPost", RequestMethodPost},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != toLower(tt.constant) {
				t.Errorf("%s = %q, expected lowercase", tt.name, tt.constant)
			}
		})
	}
}

// Helper function to check if string is lowercase
func toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}
