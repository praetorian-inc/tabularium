package normalize

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "basic HTTPS URL",
			input:    "https://example.com/path",
			expected: "https://example.com/path",
			wantErr:  false,
		},
		{
			name:     "HTTPS URL with default port",
			input:    "https://example.com:443/path",
			expected: "https://example.com/path",
			wantErr:  false,
		},
		{
			name:     "HTTP URL with default port",
			input:    "http://example.com:80/path",
			expected: "http://example.com/path",
			wantErr:  false,
		},
		{
			name:     "URL with query parameters (removed)",
			input:    "https://example.com/path?query=value",
			expected: "https://example.com/path",
			wantErr:  false,
		},
		{
			name:     "URL with fragment (removed)",
			input:    "https://example.com/path#fragment",
			expected: "https://example.com/path",
			wantErr:  false,
		},
		{
			name:     "empty URL",
			input:    "",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "URL without scheme",
			input:    "example.com/path",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "URL without host",
			input:    "https:///path",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "URL without path gets default path",
			input:    "https://example.com",
			expected: "https://example.com/",
			wantErr:  false,
		},
		{
			name:     "mixed case normalization",
			input:    "HTTPS://EXAMPLE.COM/Path",
			expected: "https://example.com/Path",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Normalize(tt.input)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveDefaultPorts(t *testing.T) {
	tests := []struct {
		name     string
		input    url.URL
		expected url.URL
	}{
		{
			name:     "HTTP with default port",
			input:    url.URL{Scheme: "http", Host: "example.com:80", Path: "/path"},
			expected: url.URL{Scheme: "http", Host: "example.com", Path: "/path"},
		},
		{
			name:     "HTTPS with default port",
			input:    url.URL{Scheme: "https", Host: "example.com:443", Path: "/path"},
			expected: url.URL{Scheme: "https", Host: "example.com", Path: "/path"},
		},
		{
			name:     "HTTP with custom port",
			input:    url.URL{Scheme: "http", Host: "example.com:8080", Path: "/path"},
			expected: url.URL{Scheme: "http", Host: "example.com:8080", Path: "/path"},
		},
		{
			name:     "HTTPS with custom port",
			input:    url.URL{Scheme: "https", Host: "example.com:8443", Path: "/path"},
			expected: url.URL{Scheme: "https", Host: "example.com:8443", Path: "/path"},
		},
		{
			name:     "no port specified",
			input:    url.URL{Scheme: "https", Host: "example.com", Path: "/path"},
			expected: url.URL{Scheme: "https", Host: "example.com", Path: "/path"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveDefaultPorts(tt.input)
			assert.Equal(t, tt.expected.String(), result.String())
		})
	}
}
