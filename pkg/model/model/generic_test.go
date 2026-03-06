package model

import (
	"encoding/json"
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneric_Class(t *testing.T) {
	tests := []struct {
		name      string
		dns       string
		assetName string
		want      string
	}{
		{
			name:      "always returns generic",
			dns:       "my-group",
			assetName: "my-name",
			want:      "generic",
		},
		{
			name:      "generic class for unicode",
			dns:       "日本語",
			assetName: "テスト",
			want:      "generic",
		},
		{
			name:      "generic class for IP-like strings",
			dns:       "10.0.0.1",
			assetName: "10.0.0.2",
			want:      "generic",
		},
		{
			name:      "generic class for domain-like strings",
			dns:       "example.com",
			assetName: "sub.example.com",
			want:      "generic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGeneric(tt.dns, tt.assetName)
			assert.Equal(t, tt.want, g.GetClass())
		})
	}
}

func TestGeneric_Valid(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want bool
	}{
		{
			name: "valid generic key",
			key:  "#generic#my-group#my-name",
			want: true,
		},
		{
			name: "valid generic key with spaces",
			key:  "#generic#hello world#foo bar",
			want: true,
		},
		{
			name: "valid generic key with special chars",
			key:  "#generic#special!@$%^&*()#also-special",
			want: true,
		},
		{
			name: "invalid - missing generic prefix",
			key:  "#asset#my-group#my-name",
			want: false,
		},
		{
			name: "invalid - only one segment after #generic",
			key:  "#generic#my-group",
			want: false,
		},
		{
			name: "invalid - empty key",
			key:  "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := Generic{}
			g.Key = tt.key
			assert.Equal(t, tt.want, g.Valid())
		})
	}
}

func TestGeneric_Hooks(t *testing.T) {
	t.Run("generates key from DNS and Name", func(t *testing.T) {
		g := NewGeneric("my-group", "my-name")
		assert.Equal(t, "#generic#my-group#my-name", g.Key)
		assert.Equal(t, "generic", g.Class)
		assert.True(t, g.Valid())
	})

	t.Run("rejects empty DNS", func(t *testing.T) {
		g := Generic{Name: "my-name"}
		g.Defaulted()
		err := registry.CallHooks(&g)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "generic asset requires non-empty dns")
	})

	t.Run("rejects empty Name", func(t *testing.T) {
		g := Generic{DNS: "my-group"}
		g.Defaulted()
		err := registry.CallHooks(&g)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "generic asset requires non-empty name")
	})

	t.Run("rejects DNS containing hash", func(t *testing.T) {
		g := Generic{DNS: "my#group", Name: "my-name"}
		g.Defaulted()
		err := registry.CallHooks(&g)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "generic asset dns must not contain '#'")
	})

	t.Run("rejects Name containing hash", func(t *testing.T) {
		g := Generic{DNS: "my-group", Name: "my#name"}
		g.Defaulted()
		err := registry.CallHooks(&g)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "generic asset name must not contain '#'")
	})
}

func TestGeneric_ArbitraryStrings(t *testing.T) {
	tests := []struct {
		name        string
		dns         string
		assetName   string
		expectedKey string
	}{
		{
			name:        "unicode strings",
			dns:         "日本語",
			assetName:   "テスト",
			expectedKey: "#generic#日本語#テスト",
		},
		{
			name:        "strings with spaces",
			dns:         "hello world",
			assetName:   "foo bar",
			expectedKey: "#generic#hello world#foo bar",
		},
		{
			name:        "special characters",
			dns:         "test@example",
			assetName:   "val!ue",
			expectedKey: "#generic#test@example#val!ue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGeneric(tt.dns, tt.assetName)
			assert.True(t, g.Valid(), "expected valid generic asset")
			assert.Equal(t, tt.expectedKey, g.Key)
			assert.Equal(t, tt.dns, g.DNS)
			assert.Equal(t, tt.assetName, g.Name)
		})
	}
}

func TestGeneric_Unmarshall(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		valid   bool
		wantErr bool
	}{
		{
			name:  "valid generic with dns and name",
			data:  `{"type": "generic", "dns": "my-group", "name": "my-id"}`,
			valid: true,
		},
		{
			name:  "valid generic with group and identifier",
			data:  `{"type": "generic", "group": "my-group", "identifier": "my-id"}`,
			valid: true,
		},
		{
			name:    "invalid generic - empty dns and name",
			data:    `{"type": "generic"}`,
			valid:   false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a registry.Wrapper[Assetlike]
			err := json.Unmarshal([]byte(tt.data), &a)
			require.NoError(t, err)

			err = registry.CallHooks(a.Model)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.valid, a.Model.Valid())
		})
	}
}

func TestGeneric_Merge(t *testing.T) {
	tests := []struct {
		name     string
		existing Generic
		update   Generic
		expected Generic
	}{
		{
			name:     "basic merge",
			existing: Generic{DNS: "my-group", Name: "my-name"},
			update:   Generic{DNS: "my-group", Name: "my-name"},
			expected: Generic{DNS: "my-group", Name: "my-name"},
		},
		{
			name:     "promote to seed",
			existing: Generic{DNS: "my-group", Name: "my-name"},
			update:   Generic{DNS: "my-group", Name: "my-name", BaseAsset: BaseAsset{Source: SeedSource}},
			expected: Generic{DNS: "my-group", Name: "my-name", BaseAsset: BaseAsset{Source: SeedSource}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.existing.Merge(&tt.update)

			assert.Equal(t, tt.expected.Created, tt.existing.Created)
			assert.Equal(t, tt.expected.Visited, tt.existing.Visited)
			assert.Equal(t, tt.expected.Source, tt.existing.Source)
			assert.Equal(t, tt.expected.GetLabels(), tt.existing.GetLabels())
		})
	}
}

func TestGeneric_Visit(t *testing.T) {
	tests := []struct {
		name     string
		existing Generic
		update   Generic
		expected Generic
	}{
		{
			name:     "basic visit",
			existing: Generic{DNS: "my-group", Name: "my-name"},
			update:   Generic{DNS: "my-group", Name: "my-name"},
			expected: Generic{DNS: "my-group", Name: "my-name"},
		},
		{
			name:     "promote to seed",
			existing: Generic{DNS: "my-group", Name: "my-name"},
			update:   Generic{DNS: "my-group", Name: "my-name", BaseAsset: BaseAsset{Source: SeedSource}},
			expected: Generic{DNS: "my-group", Name: "my-name", BaseAsset: BaseAsset{Source: SeedSource}},
		},
		{
			name:     "visit propagates tags",
			existing: Generic{DNS: "my-group", Name: "my-name", BaseAsset: BaseAsset{Tags: Tags{Tags: []string{"production"}}}},
			update:   Generic{DNS: "my-group", Name: "my-name", BaseAsset: BaseAsset{Tags: Tags{Tags: []string{"critical"}}}},
			expected: Generic{DNS: "my-group", Name: "my-name", BaseAsset: BaseAsset{Tags: Tags{Tags: []string{"production", "critical"}}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.existing.Visit(&tt.update)

			assert.Equal(t, tt.expected.Created, tt.existing.Created)
			assert.Equal(t, tt.expected.Visited, tt.existing.Visited)
			assert.Equal(t, tt.expected.Source, tt.existing.Source)
			assert.Equal(t, tt.expected.GetLabels(), tt.existing.GetLabels())
			assert.Equal(t, tt.expected.Tags.Tags, tt.existing.Tags.Tags)
		})
	}
}

func TestGeneric_GroupAndIdentifier(t *testing.T) {
	g := NewGeneric("my-dns", "my-name")
	assert.Equal(t, "my-dns", g.Group())
	assert.Equal(t, "my-name", g.Identifier())
	assert.Equal(t, "my-name", g.GetPartitionKey())
}

func TestGeneric_IsPrivate(t *testing.T) {
	g := NewGeneric("10.0.0.1", "10.0.0.1")
	assert.False(t, g.IsPrivate(), "generic assets are never private")
}

func TestGeneric_GetLabels(t *testing.T) {
	t.Run("non-seed labels", func(t *testing.T) {
		g := NewGeneric("my-group", "my-name")
		labels := g.GetLabels()
		assert.Contains(t, labels, GenericLabel)
		assert.Contains(t, labels, AssetLabel)
		assert.Contains(t, labels, TTLLabel)
		assert.NotContains(t, labels, SeedLabel)
	})

	t.Run("seed labels", func(t *testing.T) {
		g := NewGenericSeed("my-name")
		labels := g.GetLabels()
		assert.Contains(t, labels, GenericLabel)
		assert.Contains(t, labels, AssetLabel)
		assert.Contains(t, labels, TTLLabel)
		assert.Contains(t, labels, SeedLabel)
	})
}

func TestGeneric_SeedModels(t *testing.T) {
	g := NewGenericSeed("my-name")
	seedModels := g.SeedModels()

	assert.Equal(t, 1, len(seedModels))
	assert.Equal(t, &g, seedModels[0])
	assert.Contains(t, g.GetLabels(), SeedLabel)
}

func TestGeneric_WithStatus(t *testing.T) {
	g := NewGeneric("my-group", "my-name")
	target := g.WithStatus(Pending)
	updated, ok := target.(*Generic)
	require.True(t, ok)
	assert.Equal(t, Pending, updated.Status)
	assert.Equal(t, Active, g.Status, "original should be unchanged")
}

func TestGeneric_SetSource(t *testing.T) {
	g := NewGeneric("my-group", "my-name")
	g.SetSource(SeedSource)
	assert.Equal(t, SeedSource, g.Source)
	assert.Equal(t, "generic", g.Class)
}
