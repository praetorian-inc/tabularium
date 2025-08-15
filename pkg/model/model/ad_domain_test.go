package model

import (
	"encoding/json"
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestADDomain_NewADDomain(t *testing.T) {
	ad := NewADDomain("acme.local")
	assert.Equal(t, "acme.local", ad.Name)
	assert.Equal(t, "#addomain#acme.local#acme.local", ad.Key)
}

func TestADDomain_Valid(t *testing.T) {
	ad := NewADDomain("acme.local")
	assert.True(t, ad.Valid())

	// non domain
	ad = NewADDomain("acme")
	assert.False(t, ad.Valid())

	// internal IP
	ad = NewADDomain("192.168.1.1")
	assert.False(t, ad.Valid())

	// internal CIDR
	ad = NewADDomain("192.168.1.1/24")
	assert.False(t, ad.Valid())

	// empty
	ad = NewADDomain("")
	assert.False(t, ad.Valid())

	// mangled key
	ad = NewADDomain("acme.local")
	ad.Key = "#addomain#"
	assert.False(t, ad.Valid())

	// empty key
	ad = NewADDomain("acme.local")
	ad.Key = ""
	assert.False(t, ad.Valid())
}

func TestADDomain_IsClass(t *testing.T) {
	ad := NewADDomain("acme.local")
	assert.True(t, ad.IsClass("addomain"))
	assert.False(t, ad.IsClass("somethingelse"))
}

func TestADDomain_Unmarshall(t *testing.T) {
	tests := []struct {
		name  string
		data  string
		valid bool
	}{
		{
			name:  "valid addomain - name",
			data:  `{"type": "addomain", "name": "acme.local"}`,
			valid: true,
		},
		{
			name:  "valid addomain - group",
			data:  `{"type": "addomain", "group": "acme.local"}`,
			valid: true,
		},
		{
			name:  "valid addomain - identifier",
			data:  `{"type": "addomain", "identifier": "acme.local"}`,
			valid: true,
		},
		{
			name:  "invalid addomain - missing name, identifier, and group",
			data:  `{"type": "addomain"}`,
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a registry.Wrapper[Assetlike]
			err := json.Unmarshal([]byte(tt.data), &a)
			require.NoError(t, err)

			registry.CallHooks(a.Model)
			assert.Equal(t, tt.valid, a.Model.Valid())
		})
	}
}

func TestADDomain_SeedModels(t *testing.T) {
	seed := NewADDomainSeed("example.local")
	seedModels := seed.SeedModels()

	assert.Equal(t, 1, len(seedModels))
	assert.Equal(t, &seed, seedModels[0])
	assert.Contains(t, seed.GetLabels(), SeedLabel)
}
