package registry

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
