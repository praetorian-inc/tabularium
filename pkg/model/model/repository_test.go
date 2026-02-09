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

	// full dockerhub URL
	repo = NewRepository("https://hub.docker.com/r/praetorian-inc/tabularium")
	assert.Equal(t, "praetorian-inc", repo.Org)
	assert.Equal(t, "tabularium", repo.Name)
	assert.Equal(t, "https://hub.docker.com/r/praetorian-inc/tabularium", repo.URL)
	assert.Equal(t, "repository", repo.Class)
	assert.Equal(t, "#repository#https://hub.docker.com/r/praetorian-inc/tabularium#tabularium", repo.Key)

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

func TestRepository_NewRepository_AzureDevOps(t *testing.T) {
	// Azure DevOps URL format: dev.azure.com/{org}/{project}/_git/{repo}
	repo := NewRepository("https://dev.azure.com/myorg/myproject/_git/myrepo")
	assert.Equal(t, "myorg", repo.Org)
	assert.Equal(t, "myrepo", repo.Name)
	assert.Equal(t, "https://dev.azure.com/myorg/myproject/_git/myrepo", repo.URL)
	assert.Equal(t, "repository", repo.Class)
	assert.Equal(t, "#repository#https://dev.azure.com/myorg/myproject/_git/myrepo#myrepo", repo.Key)

	// Azure DevOps without https prefix
	repo = NewRepository("dev.azure.com/myorg/myproject/_git/myrepo")
	assert.Equal(t, "myorg", repo.Org)
	assert.Equal(t, "myrepo", repo.Name)
	assert.Equal(t, "https://dev.azure.com/myorg/myproject/_git/myrepo", repo.URL)
	assert.Equal(t, "repository", repo.Class)
	assert.Equal(t, "#repository#https://dev.azure.com/myorg/myproject/_git/myrepo#myrepo", repo.Key)

	// Azure DevOps with trailing slash
	repo = NewRepository("https://dev.azure.com/myorg/myproject/_git/myrepo/")
	assert.Equal(t, "myorg", repo.Org)
	assert.Equal(t, "myrepo", repo.Name)
	assert.Equal(t, "https://dev.azure.com/myorg/myproject/_git/myrepo", repo.URL)
	assert.Equal(t, "repository", repo.Class)
	assert.Equal(t, "#repository#https://dev.azure.com/myorg/myproject/_git/myrepo#myrepo", repo.Key)

	// Azure DevOps with hyphenated names
	repo = NewRepository("https://dev.azure.com/my-org/my-project/_git/my-repo")
	assert.Equal(t, "my-org", repo.Org)
	assert.Equal(t, "my-repo", repo.Name)
	assert.Equal(t, "https://dev.azure.com/my-org/my-project/_git/my-repo", repo.URL)
	assert.Equal(t, "repository", repo.Class)
	assert.Equal(t, "#repository#https://dev.azure.com/my-org/my-project/_git/my-repo#my-repo", repo.Key)
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

	repo = NewRepository("https://hub.docker.com/r/praetorian-inc/tabularium")
	assert.True(t, repo.Valid())

	repo = NewRepository("github.com/praetorian-inc/tabularium")
	assert.True(t, repo.Valid())

	repo = NewRepository("github.com/praetorian-inc/tabularium/")
	assert.True(t, repo.Valid())

	repo = NewRepository("https://github.com/praetorian-inc/tabularium/")
	assert.True(t, repo.Valid())

	// Azure DevOps valid URLs
	repo = NewRepository("https://dev.azure.com/myorg/myproject/_git/myrepo")
	assert.True(t, repo.Valid())

	repo = NewRepository("dev.azure.com/myorg/myproject/_git/myrepo")
	assert.True(t, repo.Valid())

	repo = NewRepository("https://dev.azure.com/myorg/myproject/_git/myrepo/")
	assert.True(t, repo.Valid())

	// Invalid URLs
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
