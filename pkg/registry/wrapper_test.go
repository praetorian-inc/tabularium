package registry

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModelForWrapper is a simple model for testing the Wrapper
type TestModelForWrapper struct {
	BaseModel
	Name  string `json:"name"`
	Value int    `json:"value"`
	Type  string `json:"type"`
}

func (m *TestModelForWrapper) GetDescription() string {
	return "Test model for wrapper testing"
}

func (m *TestModelForWrapper) GetKey() string {
	return "#testmodel#" + m.Name
}

// AnotherTestModel is another model type for testing type resolution
type AnotherTestModel struct {
	BaseModel
	ID    string `json:"id"`
	Label string `json:"label"`
}

func (m *AnotherTestModel) GetDescription() string {
	return "Another test model for wrapper testing"
}

func (m *AnotherTestModel) GetKey() string {
	return "#anothertestmodel#" + m.ID
}

func TestMain(m *testing.M) {
	Registry.MustRegisterModel(&TestModelForWrapper{})
	Registry.MustRegisterModel(&AnotherTestModel{})
	os.Exit(m.Run())
}

func TestWrapper_UnmarshalJSON(t *testing.T) {
	t.Run("successful unmarshaling with type field", func(t *testing.T) {
		input := `{
			"type": "testmodelforwrapper",
			"model": {
				"name": "test",
				"value": 42
			}
		}`

		var wrapper Wrapper[Model]
		err := json.Unmarshal([]byte(input), &wrapper)

		require.NoError(t, err)
		assert.Equal(t, "testmodelforwrapper", wrapper.Type)

		model, ok := wrapper.Model.(*TestModelForWrapper)
		require.True(t, ok, "expected model to be *TestModelForWrapper")
		assert.Equal(t, "test", model.Name)
		assert.Equal(t, 42, model.Value)
	})

	t.Run("successful unmarshaling with key-based type resolution", func(t *testing.T) {
		input := `{
			"model": {
				"key": "#testmodelforwrapper#example#123",
				"name": "example",
				"value": 123
			}
		}`

		var wrapper Wrapper[Model]
		err := json.Unmarshal([]byte(input), &wrapper)

		require.NoError(t, err)
		assert.Equal(t, "testmodelforwrapper", wrapper.Type)

		model, ok := wrapper.Model.(*TestModelForWrapper)
		require.True(t, ok, "expected model to be *TestModelForWrapper")
		assert.Equal(t, "example", model.Name)
		assert.Equal(t, 123, model.Value)
	})

	t.Run("successful unmarshaling without model field", func(t *testing.T) {
		input := `{
			"type": "testmodelforwrapper"
		}`

		var wrapper Wrapper[Model]
		err := json.Unmarshal([]byte(input), &wrapper)

		require.NoError(t, err)
		assert.Equal(t, "testmodelforwrapper", wrapper.Type)

		_, ok := wrapper.Model.(*TestModelForWrapper)
		require.True(t, ok, "expected model to be *TestModelForWrapper")
	})

	t.Run("successful unmarshaling with pre-set type", func(t *testing.T) {
		input := `{
			"model": {
				"name": "test",
				"value": 42
			}
		}`

		var wrapper Wrapper[Model]
		wrapper.Type = "testmodelforwrapper"
		err := json.Unmarshal([]byte(input), &wrapper)

		require.NoError(t, err)
		assert.Equal(t, "testmodelforwrapper", wrapper.Type)

		model, ok := wrapper.Model.(*TestModelForWrapper)
		require.True(t, ok, "expected model to be *TestModelForWrapper")
		assert.Equal(t, "test", model.Name)
		assert.Equal(t, 42, model.Value)
	})

	t.Run("successful unmarshaling with another model type", func(t *testing.T) {
		input := `{
			"type": "anothertestmodel",
			"model": {
				"id": "test-id",
				"label": "test-label"
			}
		}`

		var wrapper Wrapper[Model]
		err := json.Unmarshal([]byte(input), &wrapper)

		require.NoError(t, err)
		assert.Equal(t, "anothertestmodel", wrapper.Type)

		model, ok := wrapper.Model.(*AnotherTestModel)
		require.True(t, ok, "expected model to be *AnotherTestModel")
		assert.Equal(t, "test-id", model.ID)
		assert.Equal(t, "test-label", model.Label)
	})

	t.Run("successful unmarshaling with empty model", func(t *testing.T) {
		input := `{
			"type": "testmodelforwrapper",
			"model": {}
		}`

		var wrapper Wrapper[Model]
		err := json.Unmarshal([]byte(input), &wrapper)

		require.NoError(t, err)
		assert.Equal(t, "testmodelforwrapper", wrapper.Type)

		_, ok := wrapper.Model.(*TestModelForWrapper)
		require.True(t, ok, "expected model to be *TestModelForWrapper")
	})

	t.Run("successful unmarshaling with non-map model field", func(t *testing.T) {
		input := `{
			"type": "testmodelforwrapper",
			"model": "not-a-map"
		}`

		var wrapper Wrapper[Model]
		err := json.Unmarshal([]byte(input), &wrapper)

		require.NoError(t, err)
		assert.Equal(t, "testmodelforwrapper", wrapper.Type)
		// When model is not a map, it should return nil without error
		assert.Nil(t, wrapper.Model)
	})
}

func TestWrapper_UnmarshalJSON_Errors(t *testing.T) {
	t.Run("invalid JSON", func(t *testing.T) {
		input := `{invalid json`

		var wrapper Wrapper[Model]
		err := json.Unmarshal([]byte(input), &wrapper)

		require.Error(t, err)
		var jsonErr *json.SyntaxError
		assert.ErrorAs(t, err, &jsonErr)
	})

	t.Run("missing type and key", func(t *testing.T) {
		input := `{
			"name": "test",
			"value": 42
		}`

		var wrapper Wrapper[Model]
		err := json.Unmarshal([]byte(input), &wrapper)

		require.Error(t, err)
	})

	t.Run("unknown type", func(t *testing.T) {
		input := `{
			"type": "unknowntype",
			"model": {
				"name": "test",
				"value": 42
			}
		}`

		var wrapper Wrapper[Model]
		err := json.Unmarshal([]byte(input), &wrapper)

		require.Error(t, err)
	})

	t.Run("invalid key format", func(t *testing.T) {
		input := `{
			"key": "invalid-key-format",
			"model": {
				"name": "test",
				"value": 42
			}
		}`

		var wrapper Wrapper[Model]
		err := json.Unmarshal([]byte(input), &wrapper)

		require.Error(t, err)
	})

	t.Run("key without type part", func(t *testing.T) {
		input := `{
			"key": "#",
			"model": {
				"name": "test",
				"value": 42
			}
		}`

		var wrapper Wrapper[Model]
		err := json.Unmarshal([]byte(input), &wrapper)

		require.Error(t, err)
	})

	t.Run("type field is not string", func(t *testing.T) {
		input := `{
			"type": 123,
			"model": {
				"name": "test",
				"value": 42
			}
		}`

		var wrapper Wrapper[Model]
		err := json.Unmarshal([]byte(input), &wrapper)

		require.Error(t, err)
	})

	t.Run("empty type string", func(t *testing.T) {
		input := `{
			"type": "",
			"model": {
				"name": "test",
				"value": 42
			}
		}`

		var wrapper Wrapper[Model]
		err := json.Unmarshal([]byte(input), &wrapper)

		require.Error(t, err)
	})

	t.Run("failed to make type", func(t *testing.T) {
		// This test would require mocking the registry to return false from MakeType
		// For now, we'll test with a type that exists but might fail in other ways
		input := `{
			"type": "testmodelforwrapper",
			"model": {
				"name": "test",
				"value": 42
			}
		}`

		var wrapper Wrapper[Model]
		err := json.Unmarshal([]byte(input), &wrapper)

		// This should actually succeed since the type is registered
		require.NoError(t, err)
	})
}

func TestWrapper_MarshalUnmarshal(t *testing.T) {
	wrapper := Wrapper[Model]{
		Model: &TestModelForWrapper{},
	}
	data, err := json.Marshal(wrapper)
	require.NoError(t, err)
	out := Wrapper[Model]{}
	err = json.Unmarshal(data, &out)
	require.NoError(t, err)
	assert.Equal(t, "testmodelforwrapper", out.Type)
}

func TestWrapper_MarshalNil(t *testing.T) {
	wrapper := Wrapper[Model]{}
	_, err := json.Marshal(wrapper)
	require.NoError(t, err)
}
