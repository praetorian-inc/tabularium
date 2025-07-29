package plural

import "testing"

func TestPlural(t *testing.T) {

	tests := []struct {
		in       string
		expected string
	}{
		{
			in:       "a",
			expected: "a",
		},
		{
			in:       "asset",
			expected: "assets",
		},
		{
			in:       "vulnerability",
			expected: "vulnerabilities",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := Plural(tt.in)
			if got != tt.expected {
				t.Errorf("Plural() = %v, want %v", got, tt.expected)
			}
		})
	}
}
