package model

import (
	"github.com/praetorian-inc/tabularium/pkg/registry/shared"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type asset struct {
	BaseModel
}

func (a *asset) GetDescription() string { return "dummy asset" }

func init() {
	shared.Registry.MustRegisterModel(&asset{}, "alias1", "alias2")
}

func TestTypeRegistry_GetModel(t *testing.T) {
	asset, ok := shared.Registry.MakeType("asset")
	require.True(t, ok)
	assert.NotNil(t, asset)
	assert.Equal(t, "dummy asset", asset.GetDescription())

	alias1, ok := shared.Registry.MakeType("alias1")
	require.True(t, ok)
	assert.NotNil(t, alias1)
	assert.Equal(t, "dummy asset", alias1.GetDescription())

	alias2, ok := shared.Registry.MakeType("alias2")
	require.True(t, ok)
	assert.NotNil(t, alias2)
	assert.Equal(t, "dummy asset", alias2.GetDescription())

	alias3, ok := shared.Registry.MakeType("notAnAlias")
	assert.False(t, ok)
	assert.Nil(t, alias3)
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
			v, ok := shared.Registry.MakeType(tt.name)
			assert.Equal(t, tt.ok, ok)
			if ok {
				assert.NotNil(t, v)
				assert.IsType(t, tt.expected, v)
			}
		})
	}
}
