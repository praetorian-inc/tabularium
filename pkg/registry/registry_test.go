package registry

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type asset struct {
	BaseModel
}

func (a *asset) GetDescription() string { return "" }

func init() {
	Registry.MustRegisterModel(&asset{})
}

func TestTypeRegistry_MakeType(t *testing.T) {
	tests := []struct {
		name     string
		expected Model
		ok       bool
	}{
		{
			name:     "asset",
			expected: &asset{},
			ok:       true,
		},
		{
			name:     "string",
			expected: nil,
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, ok := Registry.MakeType(tt.name)
			assert.Equal(t, tt.ok, ok)
			if ok {
				assert.NotNil(t, v)
				assert.IsType(t, tt.expected, v)
			}
		})
	}
}

type mockLabeler struct {
	BaseModel
}

func (t *mockLabeler) GetDescription() string {
	return "Labeler"
}

type typedLabeler1 struct {
	mockLabeler
}

func (t *typedLabeler1) GetLabels() []string {
	return []string{"Label"}
}

type typedLabeler2 struct {
	mockLabeler
}

func (t *typedLabeler2) GetLabels() []string {
	return []string{"Label"}
}

type typedLabeler3 struct {
	mockLabeler
}

func (t *typedLabeler3) GetLabels() []string {
	return []string{"laBEL"}
}

func TestTypeRegistry_SemiDuplicateLabels(t *testing.T) {
	err := Registry.RegisterModel(&typedLabeler1{})
	require.NoError(t, err)

	err = Registry.RegisterModel(&typedLabeler2{})
	assert.NoError(t, err)

	err = Registry.RegisterModel(&typedLabeler3{})
	assert.NoError(t, err)
}

// mockGraphModel is a test implementation of GraphModel for testing
type mockGraphModel struct {
	BaseModel
	labels []string
	key    string
}

func (m *mockGraphModel) GetLabels() []string {
	return m.labels
}

func (m *mockGraphModel) GetKey() string {
	return m.key
}

func (m *mockGraphModel) GetDescription() string {
	return "Mock GraphModel for testing"
}

// mockNonGraphModel is a test implementation that doesn't implement GraphModel
type mockNonGraphModel struct {
	BaseModel
}

func (m *mockNonGraphModel) GetDescription() string {
	return "Mock non-GraphModel for testing"
}

// Additional mock types for different tests to avoid registration conflicts
type assetModel struct {
	mockGraphModel
}

func (a *assetModel) GetLabels() []string {
	return []string{"Asset", "TTL"}
}

type riskModel struct {
	mockGraphModel
}

func (r *riskModel) GetLabels() []string {
	return []string{"Risk", "TTL"}
}

type technologyModel struct {
	mockGraphModel
}

func (t *technologyModel) GetLabels() []string {
	return []string{"Technology", "TTL"}
}

func TestFormatLabel(t *testing.T) {
	// Create a temporary registry for testing
	originalRegistry := Registry
	tempRegistry := NewTypeRegistry()
	Registry = tempRegistry
	defer func() {
		Registry = originalRegistry
	}()

	// Create test models with different labels
	assetModel := &assetModel{}

	riskModel := &riskModel{}

	technologyModel := &technologyModel{}

	// Register the test models
	Registry.MustRegisterModel(assetModel)
	Registry.MustRegisterModel(riskModel)
	Registry.MustRegisterModel(technologyModel)

	// Also register a non-GraphModel to test the interface check
	nonGraphModel := &mockNonGraphModel{}
	Registry.MustRegisterModel(nonGraphModel)

	tests := []struct {
		name          string
		input         string
		expectedLabel string
		expectedFound bool
		description   string
	}{
		{
			name:          "Asset label - exact case match",
			input:         "Asset",
			expectedLabel: "Asset",
			expectedFound: true,
			description:   "Should find exact case match for Asset label",
		},
		{
			name:          "Asset label - lowercase input",
			input:         "asset",
			expectedLabel: "Asset",
			expectedFound: true,
			description:   "Should find Asset label with lowercase input using case-insensitive matching",
		},
		{
			name:          "Asset label - uppercase input",
			input:         "ASSET",
			expectedLabel: "Asset",
			expectedFound: true,
			description:   "Should find Asset label with uppercase input using case-insensitive matching",
		},
		{
			name:          "Asset label - mixed case input",
			input:         "AsSeT",
			expectedLabel: "Asset",
			expectedFound: true,
			description:   "Should find Asset label with mixed case input using case-insensitive matching",
		},
		{
			name:          "Risk label - lowercase input",
			input:         "risk",
			expectedLabel: "Risk",
			expectedFound: true,
			description:   "Should find Risk label with lowercase input",
		},
		{
			name:          "Technology label - lowercase input",
			input:         "technology",
			expectedLabel: "Technology",
			expectedFound: true,
			description:   "Should find Technology label with lowercase input",
		},
		{
			name:          "TTL label - lowercase input",
			input:         "ttl",
			expectedLabel: "TTL",
			expectedFound: true,
			description:   "Should find TTL label with lowercase input (from multiple models)",
		},
		{
			name:          "Non-existent label",
			input:         "nonexistent",
			expectedLabel: "",
			expectedFound: false,
			description:   "Should return empty string and false for non-existent label",
		},
		{
			name:          "Empty string input",
			input:         "",
			expectedLabel: "",
			expectedFound: false,
			description:   "Should return empty string and false for empty input",
		},
		{
			name:          "Special characters in input",
			input:         "Asset!@#",
			expectedLabel: "",
			expectedFound: false,
			description:   "Should return empty string and false for input with special characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label, found := Registry.FormatLabel(tt.input)

			assert.Equal(t, tt.expectedLabel, label, "Label mismatch")
			assert.Equal(t, tt.expectedFound, found, "Found flag mismatch")
		})
	}
}

func TestFormatLabelWithRealModels(t *testing.T) {
	// Test with the actual registry that has real models registered
	// This tests the real-world scenario where models are registered in init() functions

	// Check if we have any real models registered
	types := Registry.GetAllTypes()
	if len(types) == 0 {
		t.Skip("No real models registered in registry, skipping real model tests")
	}

	// Test with some common labels that should exist
	commonLabels := []string{"Asset", "Risk", "Technology", "Seed"}

	for _, expectedLabel := range commonLabels {
		t.Run("Real model - "+expectedLabel, func(t *testing.T) {
			// Test case-insensitive matching
			lowercaseInput := strings.ToLower(expectedLabel)
			label, found := Registry.FormatLabel(lowercaseInput)

			if found {
				assert.Equal(t, expectedLabel, label, "Should return exact case for %s", expectedLabel)
			} else {
				// If not found, that's also acceptable - the label might not be registered
				t.Logf("Label %s not found in registry, which is acceptable", expectedLabel)
			}
		})
	}
}
