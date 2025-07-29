package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type getTypesTestModelA struct {
	BaseModel
	Name string
}

func (m *getTypesTestModelA) GetDescription() string {
	return "Test model A for GetTypes testing"
}

type getTypesTestModelB struct {
	BaseModel
	Value int
}

func (m *getTypesTestModelB) GetDescription() string {
	return "Test model B for GetTypes testing"
}

type getTypesTestModelC struct {
	BaseModel
	Data string
}

func (m *getTypesTestModelC) GetDescription() string {
	return "Test model C for GetTypes testing"
}

type getTypesTestInterface interface {
	Model
	GetTestValue() string
}

func (m *getTypesTestModelA) GetTestValue() string {
	return m.Name
}

func init() {
	Registry.MustRegisterModel(&getTypesTestModelA{})
	Registry.MustRegisterModel(&getTypesTestModelB{})
	Registry.MustRegisterModel(&getTypesTestModelC{})
}

func TestGetTypes_ConcreteTypes(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func() []string
		expected []string
	}{
		{
			name: "get specific concrete type A",
			testFunc: func() []string {
				return GetTypes[*getTypesTestModelA](Registry)
			},
			expected: []string{"gettypestestmodela"},
		},
		{
			name: "get specific concrete type B",
			testFunc: func() []string {
				return GetTypes[*getTypesTestModelB](Registry)
			},
			expected: []string{"gettypestestmodelb"},
		},
		{
			name: "get specific concrete type C",
			testFunc: func() []string {
				return GetTypes[*getTypesTestModelC](Registry)
			},
			expected: []string{"gettypestestmodelc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.testFunc()
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestGetTypes_InterfaceImplementation(t *testing.T) {
	t.Run("get types implementing test interface", func(t *testing.T) {
		result := GetTypes[getTypesTestInterface](Registry)
		expected := []string{"gettypestestmodela"}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("get types implementing Model interface", func(t *testing.T) {
		result := GetTypes[Model](Registry)
		assert.Contains(t, result, "gettypestestmodela")
		assert.Contains(t, result, "gettypestestmodelb")
		assert.Contains(t, result, "gettypestestmodelc")
		assert.GreaterOrEqual(t, len(result), 3)
	})
}

func TestGetTypes_EdgeCases(t *testing.T) {
	t.Run("empty registry", func(t *testing.T) {
		emptyRegistry := NewTypeRegistry()
		result := GetTypes[*getTypesTestModelA](emptyRegistry)
		assert.Empty(t, result)
	})

	t.Run("interface with no implementations", func(t *testing.T) {
		type nonExistentInterface interface {
			Model
			NonExistentMethod() string
		}

		result := GetTypes[nonExistentInterface](Registry)
		assert.Empty(t, result)
	})

	t.Run("registry with single type", func(t *testing.T) {
		singleRegistry := NewTypeRegistry()
		singleRegistry.MustRegisterModel(&getTypesTestModelA{})

		result := GetTypes[*getTypesTestModelA](singleRegistry)
		expected := []string{"gettypestestmodela"}
		assert.ElementsMatch(t, expected, result)
	})
}
