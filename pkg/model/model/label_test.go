package model

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/praetorian-inc/tabularium/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLabels_Registered(t *testing.T) {
	modelRegistry := registry.Registry
	labelRegistry := GetLabelRegistry()

	for name, modelType := range modelRegistry.GetAllTypes() {
		instance := reflect.New(modelType.Elem()).Interface()

		graphModel, ok := instance.(GraphModel)
		if !ok {
			continue
		}

		labels := graphModel.GetLabels()

		for _, label := range labels {
			if label == "" {
				continue
			}

			registeredLabel, exists := labelRegistry.Get(label)
			require.True(t, exists, "Label %q from model %q should be registered", label, name)

			assert.Equal(t, label, registeredLabel, "Registered label should match exactly for model %q", name)
		}
	}
}

func TestLabel_Creation(t *testing.T) {
	tests := []struct {
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
			input:    "HAS_VULNERABILITY",
			expected: "HAS_VULNERABILITY",
		},
		{
			name:     "All lowercase label",
			input:    "ttl",
			expected: "ttl",
		},
		{
			name:     "Snake case with uppercase",
			input:    "INSTANCE_OF",
			expected: "INSTANCE_OF",
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
			input:    "Asset-Test_123",
			expected: "Asset-Test_123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GetLabelRegistry().mu.Lock()
			GetLabelRegistry().labels = make(map[string]string)
			GetLabelRegistry().mu.Unlock()

			label := NewLabel(tt.input)
			assert.Equal(t, tt.expected, label, "Label should preserve exact casing")
		})
	}
}

func TestLabelRegistry_Initialization(t *testing.T) {
	t.Run("Global registry is initialized", func(t *testing.T) {
		require.NotNil(t, GetLabelRegistry(), "Global label registry should be initialized")
	})

	t.Run("Registry starts empty or with predefined labels", func(t *testing.T) {
		registry := GetLabelRegistry()
		allLabels := registry.List()
		assert.NotNil(t, allLabels, "List should return a non-nil slice")
	})
}

func TestLabelRegistry_MustRegister(t *testing.T) {
	t.Run("Register label with lowercase key", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.mu.Lock()
		registry.labels = make(map[string]string)
		registry.mu.Unlock()

		// Act
		registry.MustRegister("ForceChangePassword")

		// Assert
		retrieved, exists := registry.Get("forcechangepassword")
		require.True(t, exists, "Should retrieve registered label")
		assert.Equal(t, "ForceChangePassword", retrieved, "Retrieved label should match registered label")
	})

	t.Run("Multiple registrations of same label", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.mu.Lock()
		registry.labels = make(map[string]string)
		registry.mu.Unlock()

		registry.MustRegister("Asset")
		registry.MustRegister("Asset")
		registry.MustRegister("Asset")

		// Assert
		retrieved, exists := registry.Get("asset")
		require.True(t, exists, "Should retrieve registered label")
		assert.Equal(t, "Asset", retrieved, "Label should be registered once")
	})

	t.Run("Registration panics on casing collision", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.mu.Lock()
		registry.labels = make(map[string]string)
		registry.mu.Unlock()

		registry.MustRegister("Asset")

		assert.Panics(t, func() {
			registry.MustRegister("ASSET")
		}, "Should panic when registering different casing with same lowercase key")
	})
}

func TestLabelRegistry_CaseInsensitiveRetrieval(t *testing.T) {
	tests := []struct {
		name           string
		registerValue  string
		retrievalKeys  []string
		expectedResult string
	}{
		{
			name:           "Asset label retrieval",
			registerValue:  "Asset",
			retrievalKeys:  []string{"asset", "Asset", "ASSET", "aSsEt"},
			expectedResult: "Asset",
		},
		{
			name:           "Addomain label retrieval",
			registerValue:  "Addomain",
			retrievalKeys:  []string{"addomain", "Addomain", "ADDOMAIN", "aDdOmAiN"},
			expectedResult: "Addomain",
		},
		{
			name:           "ForceChangePassword label retrieval",
			registerValue:  "ForceChangePassword",
			retrievalKeys:  []string{"forcechangepassword", "ForceChangePassword", "FORCECHANGEPASSWORD"},
			expectedResult: "ForceChangePassword",
		},
		{
			name:           "HAS_VULNERABILITY label retrieval",
			registerValue:  "HAS_VULNERABILITY",
			retrievalKeys:  []string{"has_vulnerability", "HAS_VULNERABILITY", "Has_Vulnerability"},
			expectedResult: "HAS_VULNERABILITY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			registry := GetLabelRegistry()
			registry.mu.Lock()
			registry.labels = make(map[string]string)
			registry.mu.Unlock()

			registry.MustRegister(tt.registerValue)

			for _, key := range tt.retrievalKeys {
				retrieved, exists := registry.Get(key)
				require.True(t, exists, "Should retrieve label with key %q", key)
				assert.Equal(t, tt.expectedResult, retrieved, "Retrieved label should match for key %q", key)
			}
		})
	}
}

func TestLabelRegistry_Get(t *testing.T) {
	t.Run("Get returns nil for non-existent label", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.mu.Lock()
		registry.labels = make(map[string]string)
		registry.mu.Unlock()

		// Act
		result, exists := registry.Get("nonexistent")

		// Assert
		assert.False(t, exists, "Should return false for non-existent label")
		assert.Empty(t, result, "Should return empty string for non-existent label")
	})

	t.Run("Get returns correct label after registration", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.mu.Lock()
		registry.labels = make(map[string]string)
		registry.mu.Unlock()
		registry.MustRegister("Technology")

		// Act
		result, exists := registry.Get("technology")

		// Assert
		require.True(t, exists, "Should return true for existing label")
		assert.Equal(t, "Technology", result, "Should return correct label")
	})

	t.Run("Get with empty string", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.mu.Lock()
		registry.labels = make(map[string]string)
		registry.mu.Unlock()
		registry.MustRegister("")

		// Act
		result, exists := registry.Get("")

		// Assert
		require.True(t, exists, "Should return true for empty string label")
		assert.Equal(t, "", result, "Should return empty string")
	})
}

func TestLabelRegistry_List(t *testing.T) {
	t.Run("List returns all registered labels", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.mu.Lock()
		registry.labels = make(map[string]string)
		registry.mu.Unlock()

		labels := []string{"Asset", "Risk", "Vulnerability", "Technology"}
		for _, label := range labels {
			registry.MustRegister(label)
		}

		// Act
		allLabels := registry.List()

		// Assert
		assert.Len(t, allLabels, len(labels), "Should return all registered labels")
		for _, label := range labels {
			assert.Contains(t, allLabels, label, "Should contain label %q", label)
		}
	})

	t.Run("List returns empty slice when registry is empty", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.mu.Lock()
		registry.labels = make(map[string]string)
		registry.mu.Unlock()

		// Act
		allLabels := registry.List()

		// Assert
		assert.Empty(t, allLabels, "Should return empty slice when no labels are registered")
	})
}

func TestLabelRegistry_ThreadSafety(t *testing.T) {
	t.Run("Concurrent registrations", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.mu.Lock()
		registry.labels = make(map[string]string)
		registry.mu.Unlock()

		labels := []string{"Asset", "Risk", "Vulnerability", "Technology", "Attribute", "Webpage"}
		var wg sync.WaitGroup

		// Act
		for _, label := range labels {
			wg.Add(1)
			go func(l string) {
				defer wg.Done()
				registry.MustRegister(l)
			}(label)
		}
		wg.Wait()

		// Assert
		allLabels := registry.List()
		assert.Len(t, allLabels, len(labels), "All labels should be registered")
		for _, label := range labels {
			assert.Contains(t, allLabels, label, "Should contain label %q", label)
		}
	})

	t.Run("Concurrent reads and writes", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.mu.Lock()
		registry.labels = make(map[string]string)
		registry.mu.Unlock()

		var wg sync.WaitGroup
		stopCh := make(chan struct{})

		// Start writers
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					label := fmt.Sprintf("Label_%d_%d", id, j)
					registry.MustRegister(label)
				}
			}(i)
		}

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-stopCh:
						return
					default:
						_ = registry.List()
						for _, label := range []string{"Asset", "Risk", "Vulnerability"} {
							result, exists := registry.Get(strings.ToLower(label))
							if exists {
								_ = result
							}
						}
					}
				}
			}()
		}

		time.Sleep(100 * time.Millisecond)
		close(stopCh)
		wg.Wait()

		assert.True(t, true, "Concurrent operations completed without issues")
	})
}

func TestExistingTabulariumLabels(t *testing.T) {
	t.Run("Node labels registration and retrieval", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.mu.Lock()
		registry.labels = make(map[string]string)
		registry.mu.Unlock()

		nodeLabels := []struct {
			name  string
			value string
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

		// Act
		for _, labelDef := range nodeLabels {
			label := NewLabel(labelDef.value)
			assert.Equal(t, labelDef.value, label, "Label %q should preserve casing", labelDef.name)
		}

		// Assert
		for _, labelDef := range nodeLabels {
			lowerKey := strings.ToLower(labelDef.value)
			result, exists := registry.Get(lowerKey)
			require.True(t, exists, "Should retrieve label %q", labelDef.name)
			assert.Equal(t, labelDef.value, result, "Label %q should have correct casing", labelDef.name)
		}
	})

	t.Run("Relationship labels registration and retrieval", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.mu.Lock()
		registry.labels = make(map[string]string)
		registry.mu.Unlock()

		relationshipLabels := []struct {
			name  string
			value string
		}{
			{"DISCOVERED", "DISCOVERED"},
			{"HAS_VULNERABILITY", "HAS_VULNERABILITY"},
			{"INSTANCE_OF", "INSTANCE_OF"},
			{"HAS_ATTRIBUTE", "HAS_ATTRIBUTE"},
			{"HAS_TECHNOLOGY", "HAS_TECHNOLOGY"},
			{"HAS_CREDENTIAL", "HAS_CREDENTIAL"},
		}

		// Act
		for _, labelDef := range relationshipLabels {
			label := NewLabel(labelDef.value)
			assert.Equal(t, labelDef.value, label, "Label %q should preserve casing", labelDef.name)
		}

		// Assert
		for _, labelDef := range relationshipLabels {
			lowerKey := strings.ToLower(labelDef.value)
			result, exists := registry.Get(lowerKey)
			require.True(t, exists, "Should retrieve label %q", labelDef.name)
			assert.Equal(t, labelDef.value, result, "Label %q should have correct casing", labelDef.name)
		}
	})

	t.Run("All labels can coexist in registry", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.mu.Lock()
		registry.labels = make(map[string]string)
		registry.mu.Unlock()

		allLabels := []string{
			"Asset", "Addomain", "Attribute", "Cloud", "Credential",
			"Integration", "Preseed", "Repository", "Risk", "Seed",
			"Technology", "Threat", "TTL", "Vulnerability", "Webpage",
			"DISCOVERED", "HAS_VULNERABILITY", "INSTANCE_OF",
			"HAS_ATTRIBUTE", "HAS_TECHNOLOGY", "HAS_CREDENTIAL",
		}

		// Act
		for _, label := range allLabels {
			NewLabel(label)
		}

		// Assert
		registeredLabels := registry.List()
		assert.Len(t, registeredLabels, len(allLabels), "All labels should be registered")
		for _, label := range allLabels {
			assert.Contains(t, registeredLabels, label, "Registry should contain label %q", label)
			result, exists := registry.Get(strings.ToLower(label))
			require.True(t, exists, "Should retrieve label %q", label)
			assert.Equal(t, label, result, "Label %q should have correct casing", label)
		}
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("Label with only spaces", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.mu.Lock()
		registry.labels = make(map[string]string)
		registry.mu.Unlock()

		// Act
		label := NewLabel("   ")

		// Assert
		assert.Equal(t, "   ", label, "Should preserve spaces")
		result, exists := registry.Get("   ")
		require.True(t, exists)
		assert.Equal(t, "   ", result)
	})

	t.Run("Label with leading/trailing spaces", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.mu.Lock()
		registry.labels = make(map[string]string)
		registry.mu.Unlock()

		// Act
		label := NewLabel("  Asset  ")

		// Assert
		assert.Equal(t, "  Asset  ", label, "Should preserve all spaces")
		result, exists := registry.Get("  asset  ")
		require.True(t, exists)
		assert.Equal(t, "  Asset  ", result)
	})

	t.Run("Very long label name", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.mu.Lock()
		registry.labels = make(map[string]string)
		registry.mu.Unlock()

		longName := strings.Repeat("VeryLongLabel", 100)

		// Act
		label := NewLabel(longName)

		// Assert
		assert.Equal(t, longName, label, "Should handle long label names")
		result, exists := registry.Get(strings.ToLower(longName))
		require.True(t, exists)
		assert.Equal(t, longName, result)
	})

	t.Run("Unicode characters in label", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.mu.Lock()
		registry.labels = make(map[string]string)
		registry.mu.Unlock()

		// Act
		label := NewLabel("资产Asset")

		// Assert
		assert.Equal(t, "资产Asset", label, "Should handle Unicode characters")
		result, exists := registry.Get(strings.ToLower("资产Asset"))
		require.True(t, exists)
		assert.Equal(t, "资产Asset", result)
	})

	t.Run("Label with emoji", func(t *testing.T) {
		// Arrange
		registry := GetLabelRegistry()
		registry.mu.Lock()
		registry.labels = make(map[string]string)
		registry.mu.Unlock()

		// Act
		label := NewLabel("Asset🔒")

		// Assert
		assert.Equal(t, "Asset🔒", label, "Should handle emoji")
		result, exists := registry.Get(strings.ToLower("Asset🔒"))
		require.True(t, exists)
		assert.Equal(t, "Asset🔒", result)
	})
}
