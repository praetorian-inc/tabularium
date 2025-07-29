package model

import (
	"encoding/json"
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_NewRepository(t *testing.T) {
	// full github URL
	repo := NewRepository("https://github.com/praetorian-inc/tabularium")
	assert.Equal(t, "praetorian-inc", repo.Org)
	assert.Equal(t, "tabularium", repo.Name)
	assert.Equal(t, "https://github.com/praetorian-inc/tabularium", repo.URL)
	assert.Equal(t, "repository", repo.Class)
	assert.Equal(t, "#repository#https://github.com/praetorian-inc/tabularium#tabularium", repo.Key)

	// Check BaseAsset defaults for just first repo asset
	assert.Equal(t, Active, repo.Status)
	assert.Equal(t, SelfSource, repo.Source)
	assert.Equal(t, "tabularium", repo.Identifier())
	assert.Equal(t, "https://github.com/praetorian-inc/tabularium", repo.Group())

	// full gitlab URL
	repo = NewRepository("https://gitlab.com/praetorian-inc/tabularium")
	assert.Equal(t, "praetorian-inc", repo.Org)
	assert.Equal(t, "tabularium", repo.Name)
	assert.Equal(t, "https://gitlab.com/praetorian-inc/tabularium", repo.URL)
	assert.Equal(t, "repository", repo.Class)
	assert.Equal(t, "#repository#https://gitlab.com/praetorian-inc/tabularium#tabularium", repo.Key)

	// full bitbucket URL
	repo = NewRepository("https://bitbucket.org/praetorian-inc/tabularium")
	assert.Equal(t, "praetorian-inc", repo.Org)
	assert.Equal(t, "tabularium", repo.Name)
	assert.Equal(t, "https://bitbucket.org/praetorian-inc/tabularium", repo.URL)
	assert.Equal(t, "repository", repo.Class)
	assert.Equal(t, "#repository#https://bitbucket.org/praetorian-inc/tabularium#tabularium", repo.Key)

	// partial URL - missing schema
	repo = NewRepository("github.com/praetorian-inc/tabularium")
	assert.Equal(t, "praetorian-inc", repo.Org)
	assert.Equal(t, "tabularium", repo.Name)
	assert.Equal(t, "https://github.com/praetorian-inc/tabularium", repo.URL)
	assert.Equal(t, "repository", repo.Class)
	assert.Equal(t, "#repository#https://github.com/praetorian-inc/tabularium#tabularium", repo.Key)

	// partial URL - missing schema with trailing slash
	repo = NewRepository("github.com/praetorian-inc/tabularium/")
	assert.Equal(t, "praetorian-inc", repo.Org)
	assert.Equal(t, "tabularium", repo.Name)
	assert.Equal(t, "https://github.com/praetorian-inc/tabularium", repo.URL)
	assert.Equal(t, "repository", repo.Class)
	assert.Equal(t, "#repository#https://github.com/praetorian-inc/tabularium#tabularium", repo.Key)

}

func TestRepository_Valid(t *testing.T) {
	repo := NewRepository("https://github.com/praetorian-inc/tabularium")
	assert.True(t, repo.Valid())

	repo = NewRepository("https://gitlab.com/praetorian-inc/tabularium")
	assert.True(t, repo.Valid())

	repo = NewRepository("https://gitlab.com/praetorian-inc/tabularium/tabularium")
	assert.True(t, repo.Valid())

	repo = NewRepository("https://bitbucket.org/praetorian-inc/tabularium")
	assert.True(t, repo.Valid())

	repo = NewRepository("github.com/praetorian-inc/tabularium")
	assert.True(t, repo.Valid())

	repo = NewRepository("github.com/praetorian-inc/tabularium/")
	assert.True(t, repo.Valid())

	repo = NewRepository("https://github.com/praetorian-inc/tabularium/")
	assert.True(t, repo.Valid())

	repo = NewRepository("praetorian-inc/tabularium/")
	assert.False(t, repo.Valid())

	repo = NewRepository("https://github.com/praetorian-inc/")
	assert.False(t, repo.Valid())

	repo = NewRepository("github.com")
	assert.False(t, repo.Valid())

	repo = NewRepository("")
	assert.False(t, repo.Valid())

	repo = NewRepository("https://github.com/praetorian-inc/tabularium/tabularium")
	repo.Key = "#repository##tabularium"
	assert.False(t, repo.Valid())

	repo = NewRepository("https://github.com/praetorian-inc/tabularium")
	repo.Key = "#repository#https://github.com/praetorian-inc/tabularium#"
	assert.False(t, repo.Valid())

	repo = NewRepository("https://github.com/praetorian-inc/tabularium")
	repo.Key = "#repository##"
	assert.False(t, repo.Valid())
}

func TestRepository_IsClass(t *testing.T) {
	repo := NewRepository("https://github.com/praetorian-inc/tabularium")
	assert.True(t, repo.IsClass("repository"))
	assert.False(t, repo.IsClass("somethingelse"))
}

func TestRepository_Unmarshall(t *testing.T) {
	tests := []struct {
		name  string
		data  string
		valid bool
	}{
		{
			name:  "valid repository",
			data:  `{"type": "repository", "url": "https://github.com/praetorian-inc/tabularium"}`,
			valid: true,
		},
		{
			name:  "valid repository - group and identifier",
			data:  `{"type": "repository", "group": "https://github.com/praetorian-inc/tabularium", "identifier": "tabularium"}`,
			valid: true,
		},
		{
			name:  "invalid repository - missing url",
			data:  `{"type": "repository"}`,
			valid: false,
		},
		{
			name:  "invalid repository - missing group or identifier",
			data:  `{"type": "repository", "group": "example.com"}`,
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
