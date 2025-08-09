package model

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLabel_Creation verifies that Label type preserves exact casing
func TestLabel_Creation(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple capitalized label",
			input:    "Asset",
			expected: "Asset",
		},
		{
			name:     "Mixed case label - Addomain",
			input:    "Addomain",
			expected: "Addomain",
		},
		{
			name:     "CamelCase label",
			input:    "ForceChangePassword",
			expected: "ForceChangePassword",
		},
		{
			name:     "All uppercase label",
			input:    "TTL",
			expected: "TTL",
		},
		{
			name:     "All lowercase label",
			input:    "webpage",
			expected: "webpage",
		},
		{
			name:     "Snake case with uppercase",
			input:    "HAS_VULNERABILITY",
			expected: "HAS_VULNERABILITY",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Single character",
			input:    "A",
			expected: "A",
		},
		{
			name:     "Label with numbers",
			input:    "Asset123",
			expected: "Asset123",
		},
		{
			name:     "Label with special characters",
			input:    "Asset-Type_2",
			expected: "Asset-Type_2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			label := NewLabel(tc.input)

			// Assert
			assert.Equal(t, Label(tc.expected), label, "Label should preserve exact casing")
			assert.Equal(t, tc.expected, string(label), "Label should convert to string with exact casing")
		})
	}
}

// TestLabel_StringConversion verifies Label's string conversion methods
func TestLabel_StringConversion(t *testing.T) {
	t.Run("String method returns exact value", func(t *testing.T) {
		// Arrange
		expectedValue := "ForceChangePassword"

		// Act
		label := NewLabel(expectedValue)
		result := label.String()

		// Assert
		assert.Equal(t, expectedValue, result, "String() method should return exact label value")
	})

	t.Run("Type casting to string preserves value", func(t *testing.T) {
		// Arrange
		expectedValue := "Addomain"

		// Act
		label := NewLabel(expectedValue)
		result := string(label)

		// Assert
		assert.Equal(t, expectedValue, result, "Type casting to string should preserve exact value")
	})
}

// TestLabel_Equality verifies Label equality comparisons
func TestLabel_Equality(t *testing.T) {
	t.Run("Labels with same value are equal", func(t *testing.T) {
		// Arrange & Act
		label1 := NewLabel("Asset")
		label2 := NewLabel("Asset")

		// Assert
		assert.Equal(t, label1, label2, "Labels with same value should be equal")
	})

	t.Run("Labels with different casing are not equal", func(t *testing.T) {
		// Arrange & Act
		label1 := NewLabel("Asset")
		label2 := NewLabel("asset")

		// Assert
		assert.NotEqual(t, label1, label2, "Labels with different casing should not be equal")
	})

	t.Run("Labels with different values are not equal", func(t *testing.T) {
		// Arrange & Act
		label1 := NewLabel("Asset")
		label2 := NewLabel("Credential")

		// Assert
		assert.NotEqual(t, label1, label2, "Labels with different values should not be equal")
	})
}

// TestLabelRegistry_Initialization verifies registry is properly initialized
func TestLabelRegistry_Initialization(t *testing.T) {
	t.Run("Global registry is initialized", func(t *testing.T) {
		// Assert
		require.NotNil(t, GetLabelRegistry(), "Global label registry should be initialized")
	})

	t.Run("Registry starts empty or with predefined labels", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()

		// Act
		registry.Clear() // Clear for testing

		// Assert - after clearing, registry should be empty
		allLabels := registry.List()
		assert.Empty(t, allLabels, "Registry should be empty after clearing")
	})
}

// TestLabelRegistry_Registration verifies label registration functionality
func TestLabelRegistry_Registration(t *testing.T) {
	t.Run("Register label with lowercase key", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.Clear()

		// Act
		label := NewLabel("ForceChangePassword")

		// Assert
		retrieved := registry.Get("forcechangepassword")
		require.NotNil(t, retrieved, "Should retrieve registered label")
		assert.Equal(t, label, *retrieved, "Retrieved label should match registered label")
	})

	t.Run("Multiple registrations of same label", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.Clear()

		// Act
		label1 := NewLabel("Asset")
		label2 := NewLabel("Asset")

		// Assert
		retrieved := registry.Get("asset")
		require.NotNil(t, retrieved, "Should retrieve registered label")
		assert.Equal(t, label2, *retrieved, "Should return the most recently registered label")
		assert.Equal(t, label1, label2, "Both labels should be equal")
	})

	t.Run("Registration with different casings", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.Clear()

		// Act - Register labels with different casings but same lowercase key
		_ = NewLabel("Asset")
		_ = NewLabel("ASSET") // This will overwrite the previous registration

		// Assert
		retrieved := registry.Get("asset")
		require.NotNil(t, retrieved, "Should retrieve registered label")
		assert.Equal(t, Label("ASSET"), *retrieved, "Should return the most recently registered label")
	})
}

// TestLabelRegistry_CaseInsensitiveRetrieval verifies case-insensitive lookup
func TestLabelRegistry_CaseInsensitiveRetrieval(t *testing.T) {
	testCases := []struct {
		name         string
		registerAs   string
		lookupKeys   []string
		expectedLabel string
	}{
		{
			name:         "Asset label retrieval",
			registerAs:   "Asset",
			lookupKeys:   []string{"asset", "ASSET", "AsSeT", "aSsEt"},
			expectedLabel: "Asset",
		},
		{
			name:         "Addomain label retrieval",
			registerAs:   "Addomain",
			lookupKeys:   []string{"addomain", "ADDOMAIN", "AdDoMaIn"},
			expectedLabel: "Addomain",
		},
		{
			name:         "ForceChangePassword label retrieval",
			registerAs:   "ForceChangePassword",
			lookupKeys:   []string{"forcechangepassword", "FORCECHANGEPASSWORD", "ForceChangePassword"},
			expectedLabel: "ForceChangePassword",
		},
		{
			name:         "HAS_VULNERABILITY label retrieval",
			registerAs:   "HAS_VULNERABILITY",
			lookupKeys:   []string{"has_vulnerability", "HAS_VULNERABILITY", "Has_Vulnerability"},
			expectedLabel: "HAS_VULNERABILITY",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			registry := GetLabelRegistry()
			registry.Clear()
			
			// Act
			registeredLabel := NewLabel(tc.registerAs)

			// Assert - all lookup keys should return the same label
			for _, key := range tc.lookupKeys {
				retrieved := registry.Get(key)
				require.NotNil(t, retrieved, "Should retrieve label for key: %s", key)
				assert.Equal(t, Label(tc.expectedLabel), *retrieved, "Retrieved label should match expected for key: %s", key)
				assert.Equal(t, registeredLabel, *retrieved, "Retrieved label should match registered label for key: %s", key)
			}
		})
	}
}

// TestLabelRegistry_Get verifies the Get method behavior
func TestLabelRegistry_Get(t *testing.T) {
	t.Run("Get returns nil for non-existent label", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.Clear()

		// Act
		result := registry.Get("nonexistent")

		// Assert
		assert.Nil(t, result, "Get should return nil for non-existent label")
	})

	t.Run("Get returns correct label after registration", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.Clear()
		label := NewLabel("Technology")

		// Act
		result := registry.Get("technology")

		// Assert
		require.NotNil(t, result, "Get should return registered label")
		assert.Equal(t, label, *result, "Retrieved label should match registered label")
	})

	t.Run("Get with empty string", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.Clear()
		_ = NewLabel("")

		// Act
		result := registry.Get("")

		// Assert
		require.NotNil(t, result, "Should retrieve empty label")
		assert.Equal(t, Label(""), *result, "Should return empty label")
	})
}

// TestLabelRegistry_GetOrCreate verifies GetOrCreate functionality
func TestLabelRegistry_GetOrCreate(t *testing.T) {
	t.Run("GetOrCreate returns existing label", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.Clear()
		existingLabel := NewLabel("Vulnerability")

		// Act
		result := registry.GetOrCreate("vulnerability", "Vulnerability")

		// Assert
		assert.Equal(t, existingLabel, result, "Should return existing label")
	})

	t.Run("GetOrCreate creates new label when not exists", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.Clear()

		// Act
		result := registry.GetOrCreate("newlabel", "NewLabel")

		// Assert
		assert.Equal(t, Label("NewLabel"), result, "Should create and return new label")
		
		// Verify it was registered
		retrieved := registry.Get("newlabel")
		require.NotNil(t, retrieved, "New label should be registered")
		assert.Equal(t, result, *retrieved, "Retrieved label should match created label")
	})

	t.Run("GetOrCreate with different casings uses existing", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.Clear()
		existingLabel := NewLabel("Asset")

		// Act - Try to create with different casing
		result := registry.GetOrCreate("asset", "ASSET")

		// Assert - Should return existing label, not create new one
		assert.Equal(t, existingLabel, result, "Should return existing label, not create new one")
	})
}

// TestLabelRegistry_List verifies List method functionality
func TestLabelRegistry_List(t *testing.T) {
	t.Run("List returns all registered labels", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.Clear()
		
		labels := []string{"Asset", "Credential", "Technology", "Vulnerability"}
		for _, l := range labels {
			NewLabel(l)
		}

		// Act
		allLabels := registry.List()

		// Assert
		assert.Len(t, allLabels, len(labels), "Should return all registered labels")
		
		// Verify all labels are present
		labelMap := make(map[Label]bool)
		for _, l := range allLabels {
			labelMap[l] = true
		}
		
		for _, expectedLabel := range labels {
			assert.True(t, labelMap[Label(expectedLabel)], "List should contain label: %s", expectedLabel)
		}
	})

	t.Run("List returns empty slice when registry is empty", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.Clear()

		// Act
		allLabels := registry.List()

		// Assert
		assert.Empty(t, allLabels, "List should return empty slice for empty registry")
	})
}

// TestLabelRegistry_ThreadSafety verifies concurrent access safety
func TestLabelRegistry_ThreadSafety(t *testing.T) {
	t.Run("Concurrent registrations", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.Clear()
		
		numGoroutines := 100
		numLabelsPerGoroutine := 10
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		// Act - Concurrent registrations
		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				defer wg.Done()
				for j := 0; j < numLabelsPerGoroutine; j++ {
					labelName := fmt.Sprintf("Label_%d_%d", goroutineID, j)
					NewLabel(labelName)
				}
			}(i)
		}
		wg.Wait()

		// Assert - Verify all labels were registered
		allLabels := registry.List()
		assert.GreaterOrEqual(t, len(allLabels), numGoroutines*numLabelsPerGoroutine/2, 
			"Should have registered many labels (some may have same lowercase key)")
	})

	t.Run("Concurrent reads and writes", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.Clear()
		
		// Pre-register some labels
		initialLabels := []string{"Asset", "Credential", "Technology"}
		for _, l := range initialLabels {
			NewLabel(l)
		}

		numReaders := 50
		numWriters := 50
		iterations := 100
		var wg sync.WaitGroup
		wg.Add(numReaders + numWriters)

		// Act - Concurrent reads
		for i := 0; i < numReaders; i++ {
			go func(readerID int) {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					// Read existing labels
					for _, label := range initialLabels {
						result := registry.Get(strings.ToLower(label))
						assert.NotNil(t, result, "Should retrieve label during concurrent access")
					}
					// List all labels
					_ = registry.List()
				}
			}(i)
		}

		// Act - Concurrent writes
		for i := 0; i < numWriters; i++ {
			go func(writerID int) {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					labelName := fmt.Sprintf("ConcurrentLabel_%d_%d", writerID, j)
					NewLabel(labelName)
				}
			}(i)
		}

		// Assert - No panics or race conditions
		wg.Wait()
		
		// Verify initial labels are still present
		for _, label := range initialLabels {
			result := registry.Get(strings.ToLower(label))
			require.NotNil(t, result, "Initial label should still be present: %s", label)
			assert.Equal(t, Label(label), *result, "Initial label should have correct value: %s", label)
		}
	})

	t.Run("GetOrCreate concurrent access", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.Clear()
		
		numGoroutines := 100
		var wg sync.WaitGroup
		wg.Add(numGoroutines)
		
		results := make([]Label, numGoroutines)

		// Act - Multiple goroutines try to GetOrCreate the same label
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				results[id] = registry.GetOrCreate("concurrent", "Concurrent")
			}(i)
		}
		wg.Wait()

		// Assert - All goroutines should get the same label
		firstLabel := results[0]
		for i, label := range results {
			assert.Equal(t, firstLabel, label, "All goroutines should receive the same label instance (goroutine %d)", i)
		}
		
		// Verify only one label exists in registry
		retrieved := registry.Get("concurrent")
		require.NotNil(t, retrieved, "Label should exist in registry")
		assert.Equal(t, firstLabel, *retrieved, "Registry should contain the same label")
	})
}

// TestExistingTabulariumLabels verifies all existing Tabularium labels work correctly
func TestExistingTabulariumLabels(t *testing.T) {
	// Define all existing node and relationship labels from Tabularium
	nodeLabels := []struct {
		name     string
		expected string
	}{
		{"Asset", "Asset"},
		{"Addomain", "Addomain"},
		{"Attribute", "Attribute"},
		{"Cloud", "Cloud"},
		{"Credential", "Credential"},
		{"Integration", "Integration"},
		{"Preseed", "Preseed"},
		{"Repository", "Repository"},
		{"Risk", "Risk"},
		{"Seed", "Seed"},
		{"Technology", "Technology"},
		{"Threat", "Threat"},
		{"TTL", "TTL"},
		{"Vulnerability", "Vulnerability"},
		{"Webpage", "Webpage"},
	}

	relationshipLabels := []struct {
		name     string
		expected string
	}{
		{"DISCOVERED", "DISCOVERED"},
		{"HAS_VULNERABILITY", "HAS_VULNERABILITY"},
		{"INSTANCE_OF", "INSTANCE_OF"},
		{"HAS_ATTRIBUTE", "HAS_ATTRIBUTE"},
		{"HAS_TECHNOLOGY", "HAS_TECHNOLOGY"},
		{"HAS_CREDENTIAL", "HAS_CREDENTIAL"},
	}

	t.Run("Node labels registration and retrieval", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.Clear()

		// Act - Register all node labels
		registeredLabels := make(map[string]Label)
		for _, labelDef := range nodeLabels {
			label := NewLabel(labelDef.expected)
			registeredLabels[labelDef.name] = label
		}

		// Assert - Verify all can be retrieved case-insensitively
		for _, labelDef := range nodeLabels {
			// Test various case variations
			lowerKey := strings.ToLower(labelDef.name)
			upperKey := strings.ToUpper(labelDef.name)
			
			lowerResult := registry.Get(lowerKey)
			require.NotNil(t, lowerResult, "Should retrieve label with lowercase key: %s", lowerKey)
			assert.Equal(t, Label(labelDef.expected), *lowerResult, "Label should have correct casing for key: %s", lowerKey)
			
			upperResult := registry.Get(upperKey)
			require.NotNil(t, upperResult, "Should retrieve label with uppercase key: %s", upperKey)
			assert.Equal(t, Label(labelDef.expected), *upperResult, "Label should have correct casing for key: %s", upperKey)
		}
	})

	t.Run("Relationship labels registration and retrieval", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.Clear()

		// Act - Register all relationship labels
		for _, labelDef := range relationshipLabels {
			NewLabel(labelDef.expected)
		}

		// Assert - Verify all can be retrieved case-insensitively
		for _, labelDef := range relationshipLabels {
			lowerKey := strings.ToLower(labelDef.name)
			result := registry.Get(lowerKey)
			require.NotNil(t, result, "Should retrieve relationship label: %s", labelDef.name)
			assert.Equal(t, Label(labelDef.expected), *result, "Relationship label should have correct casing: %s", labelDef.name)
		}
	})

	t.Run("All labels can coexist in registry", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.Clear()

		// Act - Register all labels
		allLabels := append(nodeLabels, relationshipLabels...)
		for _, labelDef := range allLabels {
			NewLabel(labelDef.expected)
		}

		// Assert - Verify total count
		registeredLabels := registry.List()
		assert.Len(t, registeredLabels, len(allLabels), "All labels should be registered")

		// Verify each label can be retrieved
		for _, labelDef := range allLabels {
			result := registry.Get(strings.ToLower(labelDef.name))
			require.NotNil(t, result, "Should retrieve label: %s", labelDef.name)
			assert.Equal(t, Label(labelDef.expected), *result, "Label should have correct value: %s", labelDef.name)
		}
	})
}

// TestLabelHelperFunctions verifies helper functions for label conversion
func TestLabelHelperFunctions(t *testing.T) {
	t.Run("LabelsToStrings converts Label slice to string slice", func(t *testing.T) {
		// Arrange
		labels := []Label{
			Label("Asset"),
			Label("TTL"),
			Label("Credential"),
		}

		// Act
		result := LabelsToStrings(labels)

		// Assert
		expected := []string{"Asset", "TTL", "Credential"}
		assert.Equal(t, expected, result, "Should convert Label slice to string slice preserving order and casing")
	})

	t.Run("LabelsToStrings handles empty slice", func(t *testing.T) {
		// Arrange
		var labels []Label

		// Act
		result := LabelsToStrings(labels)

		// Assert
		assert.Empty(t, result, "Should return empty string slice for empty Label slice")
	})

	t.Run("LabelsToStrings handles nil slice", func(t *testing.T) {
		// Act
		result := LabelsToStrings(nil)

		// Assert
		assert.Empty(t, result, "Should return empty string slice for nil Label slice")
	})

	t.Run("StringsToLabels converts string slice to Label slice", func(t *testing.T) {
		// Arrange
		strings := []string{"Asset", "TTL", "Credential"}

		// Act
		result := StringsToLabels(strings)

		// Assert
		expected := []Label{
			Label("Asset"),
			Label("TTL"),
			Label("Credential"),
		}
		assert.Equal(t, expected, result, "Should convert string slice to Label slice preserving order and casing")
	})

	t.Run("StringsToLabels handles empty slice", func(t *testing.T) {
		// Arrange
		var strings []string

		// Act
		result := StringsToLabels(strings)

		// Assert
		assert.Empty(t, result, "Should return empty Label slice for empty string slice")
	})

	t.Run("StringsToLabels handles nil slice", func(t *testing.T) {
		// Act
		result := StringsToLabels(nil)

		// Assert
		assert.Empty(t, result, "Should return empty Label slice for nil string slice")
	})
}

// TestLabelRegistry_Clear verifies Clear method functionality
func TestLabelRegistry_Clear(t *testing.T) {
	t.Run("Clear removes all labels", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.Clear()
		
		// Register some labels
		NewLabel("Asset")
		NewLabel("Credential")
		NewLabel("Technology")
		
		// Verify labels exist
		assert.NotNil(t, registry.Get("asset"), "Label should exist before clear")
		assert.NotNil(t, registry.Get("credential"), "Label should exist before clear")
		assert.NotNil(t, registry.Get("technology"), "Label should exist before clear")

		// Act
		registry.Clear()

		// Assert
		assert.Nil(t, registry.Get("asset"), "Label should not exist after clear")
		assert.Nil(t, registry.Get("credential"), "Label should not exist after clear")
		assert.Nil(t, registry.Get("technology"), "Label should not exist after clear")
		assert.Empty(t, registry.List(), "Registry should be empty after clear")
	})

	t.Run("Clear on empty registry is safe", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.Clear()

		// Act - Clear again on empty registry
		registry.Clear()

		// Assert - Should not panic
		assert.Empty(t, registry.List(), "Registry should remain empty")
	})
}

// TestEdgeCases verifies behavior with edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("Label with only spaces", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.Clear()

		// Act
		label := NewLabel("   ")

		// Assert
		assert.Equal(t, Label("   "), label, "Should preserve spaces")
		result := registry.Get("   ")
		require.NotNil(t, result, "Should retrieve label with spaces")
		assert.Equal(t, label, *result, "Retrieved label should match")
	})

	t.Run("Label with leading/trailing spaces", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.Clear()

		// Act
		label := NewLabel("  Asset  ")

		// Assert
		assert.Equal(t, Label("  Asset  "), label, "Should preserve all spaces")
		// Note: lowercase conversion will also preserve spaces
		result := registry.Get("  asset  ")
		require.NotNil(t, result, "Should retrieve label with spaces")
		assert.Equal(t, label, *result, "Retrieved label should match")
	})

	t.Run("Very long label name", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.Clear()
		longName := strings.Repeat("VeryLongLabelName", 100)

		// Act
		label := NewLabel(longName)

		// Assert
		assert.Equal(t, Label(longName), label, "Should handle long label names")
		result := registry.Get(strings.ToLower(longName))
		require.NotNil(t, result, "Should retrieve long label")
		assert.Equal(t, label, *result, "Retrieved long label should match")
	})

	t.Run("Unicode characters in label", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.Clear()

		// Act
		label := NewLabel("资产Asset")

		// Assert
		assert.Equal(t, Label("资产Asset"), label, "Should handle Unicode characters")
		result := registry.Get(strings.ToLower("资产Asset"))
		require.NotNil(t, result, "Should retrieve Unicode label")
		assert.Equal(t, label, *result, "Retrieved Unicode label should match")
	})

	t.Run("Label with emoji", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.Clear()

		// Act
		label := NewLabel("Asset🔒")

		// Assert
		assert.Equal(t, Label("Asset🔒"), label, "Should handle emoji in label")
		result := registry.Get(strings.ToLower("Asset🔒"))
		require.NotNil(t, result, "Should retrieve emoji label")
		assert.Equal(t, label, *result, "Retrieved emoji label should match")
	})
}