package model

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestModel struct {
	BaseModel
	Name       string `json:"name"`
	Value      int    `json:"value"`
	hookCalled bool
}

func (m *TestModel) GetDescription() string {
	return "Test model for unit testing"
}

func (m *TestModel) Defaulted() {
	m.Value = 42 // Default value
}

func (m *TestModel) GetHooks() []Hook {
	return []Hook{
		{
			Call: func() error {
				m.hookCalled = true
				return nil
			},
			Description: "Test hook that sets hookCalled to true",
		},
	}
}

type TestModelWithErrorHook struct {
	BaseModel
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type OuterModel struct {
	BaseModel
	Name            string     `json:"name"`
	Key             string     `json:"key"`
	DefaultedCalled bool       `json:"defaulted"`
	Inner           InnerModel `json:"inner"`
}

func (m *OuterModel) GetDescription() string { return "" }

func (m *OuterModel) Defaulted() { m.DefaultedCalled = true }

func (m *OuterModel) GetHooks() []Hook {
	return []Hook{
		{
			Call: func() error {
				m.Key = m.Name
				return nil
			},
		},
	}
}

type InnerModel struct {
	BaseModel
	Name            string `json:"name"`
	Key             string `json:"key"`
	DefaultedCalled bool   `json:"defaulted"`
}

func (m *InnerModel) GetDescription() string { return "" }

func (m *InnerModel) Defaulted() { m.DefaultedCalled = true }

func (m *InnerModel) GetHooks() []Hook {
	return []Hook{
		{
			Call: func() error {
				m.Key = m.Name
				return nil
			},
		},
	}
}

func (m *TestModelWithErrorHook) GetDescription() string {
	return "Test model that returns an error from its hook"
}

func (m *TestModelWithErrorHook) GetHooks() []Hook {
	return []Hook{
		{
			Call: func() error {
				return errors.New("hook error")
			},
			Description: "Test hook that always returns an error",
		},
	}
}

func TestUnmarshalModel_Success(t *testing.T) {
	t.Run("successful unmarshaling with defaults and hooks", func(t *testing.T) {
		input := `{"name":"test"}`
		model := &TestModel{}

		err := UnmarshalModel([]byte(input), model)

		require.NoError(t, err)
		assert.Equal(t, "test", model.Name)
		assert.Equal(t, 42, model.Value) // Default value should be set
		assert.True(t, model.hookCalled) // Hook should be called
	})

	t.Run("successful unmarshaling with provided values", func(t *testing.T) {
		input := `{"name":"test","value":100}`
		model := &TestModel{}

		err := UnmarshalModel([]byte(input), model)

		require.NoError(t, err)
		assert.Equal(t, "test", model.Name)
		assert.Equal(t, 100, model.Value) // Provided value should override default
		assert.True(t, model.hookCalled)  // Hook should be called
	})
}

func TestUnmarshalModel_Error(t *testing.T) {
	t.Run("invalid JSON", func(t *testing.T) {
		input := `{"name":"test",` // Invalid JSON
		model := &TestModel{}

		err := UnmarshalModel([]byte(input), model)

		require.Error(t, err)
		var jsonErr *json.SyntaxError
		assert.ErrorAs(t, err, &jsonErr)
	})

	t.Run("hook returns error", func(t *testing.T) {
		input := `{"name":"test","value":100}`
		model := &TestModelWithErrorHook{}

		err := UnmarshalModel([]byte(input), model)

		require.Error(t, err)
		assert.Equal(t, "hook error", err.Error())
		assert.Equal(t, "test", model.Name) // Model should still be unmarshaled
		assert.Equal(t, 100, model.Value)   // Model should still be unmarshaled
	})
}

func TestUnmarshalModel_EmptyInput(t *testing.T) {
	t.Run("empty JSON object", func(t *testing.T) {
		input := `{}`
		model := &TestModel{}

		err := UnmarshalModel([]byte(input), model)

		require.NoError(t, err)
		assert.Equal(t, "", model.Name)
		assert.Equal(t, 42, model.Value) // Default value should be set
		assert.True(t, model.hookCalled) // Hook should be called
	})

	t.Run("empty input", func(t *testing.T) {
		model := &TestModel{}

		err := UnmarshalModel([]byte(""), model)

		require.Error(t, err)
		var jsonErr *json.SyntaxError
		assert.ErrorAs(t, err, &jsonErr)
	})

	t.Run("nil input", func(t *testing.T) {
		model := &TestModel{}

		err := UnmarshalModel(nil, model)

		require.Error(t, err)
		var jsonErr *json.SyntaxError
		assert.ErrorAs(t, err, &jsonErr)
	})
}

func TestUnmarshalModel_Submodel(t *testing.T) {
	input := `{"name":"test", "inner": {"name": "test"}}`
	model := &OuterModel{}

	err := UnmarshalModel([]byte(input), model)
	require.NoError(t, err)
	assert.Equal(t, true, model.DefaultedCalled)
	assert.Equal(t, "test", model.Name)
	assert.Equal(t, "test", model.Key)
	assert.Equal(t, true, model.Inner.DefaultedCalled)
	assert.Equal(t, "test", model.Inner.Name)
	assert.Equal(t, "test", model.Inner.Key)
}
