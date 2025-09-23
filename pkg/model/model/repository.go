package model

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func init() {
	registry.Registry.MustRegisterModel(&Repository{})
}

type RepositoryTag struct {
	Name        string `neo4j:"name,omitempty" json:"name,omitempty" desc:"Tag name." example:"latest"`
	Digest      string `neo4j:"digest,omitempty" json:"digest,omitempty" desc:"Tag digest." example:"sha256:abc123"`
	LastUpdated string `neo4j:"last_updated,omitempty" json:"last_updated,omitempty" desc:"Tag last updated timestamp." example:"2023-01-01T00:00:00Z"`
}

type Repository struct {
	BaseAsset
	URL         string          `neo4j:"url,omitempty" json:"url,omitempty" desc:"Repository URL." example:"https://github.com/praetorian-inc/tabularium"`
	Org         string          `neo4j:"org,omitempty" json:"org,omitempty" desc:"Organization name." example:"praetorian-inc"`
	Name        string          `neo4j:"name,omitempty" json:"name,omitempty" desc:"Repository name." example:"praetorian-inc/tabularium"`
	PullCount   int             `neo4j:"pull_count,omitempty" json:"pull_count,omitempty" desc:"Number of pulls for Docker Hub repositories." example:"1000"`
	StarCount   int             `neo4j:"star_count,omitempty" json:"star_count,omitempty" desc:"Number of stars for Docker Hub repositories." example:"50"`
	Private     bool            `neo4j:"private,omitempty" json:"private,omitempty" desc:"Whether the repository is private." example:"false"`
	LastUpdated string          `neo4j:"last_updated,omitempty" json:"last_updated,omitempty" desc:"Last updated timestamp." example:"2023-01-01T00:00:00Z"`
	Tags        []RepositoryTag `neo4j:"tags,omitempty" json:"tags,omitempty" desc:"Tags in the repository." example:"[{\"name\":\"latest\",\"digest\":\"sha256:abc123\"}]"`
}

const (
	RepositoryLabel = "Repository"
)

var (
	repository    = regexp.MustCompile(`^(https://)?(github\.com|gitlab\.com|bitbucket\.(com|org)|hub\.docker\.com)/([^/]+)/(([^/]+/)*[^/]+)$`)
	repositoryKey = regexp.MustCompile(`^#repository(#[^#]+){2,}$`)
)

func (r *Repository) GetLabels() []string {
	return []string{RepositoryLabel, AssetLabel, TTLLabel}
}

func (r *Repository) Valid() bool {
	return repository.MatchString(r.URL) && repositoryKey.MatchString(r.Key)
}

func (r *Repository) Attribute(name, value string) Attribute {
	attr := NewAttribute(name, value, r)
	return attr
}

func (r *Repository) Identifier() string {
	return r.Name
}

func (r *Repository) Group() string {
	return r.URL
}

func (r *Repository) IsClass(value string) bool {
	return value == r.Class
}

func (r *Repository) IsPrivate() bool {
	return r.Private
}

func (r *Repository) WithStatus(status string) Target {
	ret := *r
	ret.Status = status
	return &ret
}

func (r *Repository) Defaulted() {
	r.BaseAsset.Defaulted()
	r.Class = "repository"
}

func (r *Repository) GetHooks() []registry.Hook {
	return []registry.Hook{
		useGroupAndIdentifier(r, &r.URL, &r.Name),
		{
			Call:        r.formatURL,
			Description: "Format the repository URL",
		},
		{
			Call:        r.extractOrgAndRepo,
			Description: "Extract the repository name and organization from the URL",
		},
		{
			Call:        r.constructKey,
			Description: "Construct the repository key",
		},
		{
			Call:        r.setBase,
			Description: "Set the base asset",
		},
	}
}

func (r *Repository) formatURL() error {
	repoURL := r.URL
	if !strings.HasPrefix(repoURL, "https://") {
		repoURL = "https://" + repoURL
	}
	repoURL = strings.TrimSuffix(repoURL, "/")

	if !repository.MatchString(repoURL) {
		return fmt.Errorf("invalid repository URL: %s", repoURL)
	}

	r.URL = repoURL
	return nil
}

func (r *Repository) extractOrgAndRepo() error {
	parts := strings.Split(r.URL, "/")
	if len(parts) < 4 {
		return fmt.Errorf("invalid repository URL: %s", r.URL)
	}
	r.Name = parts[len(parts)-1]
	r.Org = parts[len(parts)-2]
	return nil
}

func (r *Repository) constructKey() error {
	r.Key = fmt.Sprintf("#repository#%s#%s", r.URL, r.Name)
	return nil
}

func (r *Repository) setBase() error {
	r.BaseAsset.Identifier = r.Name
	r.BaseAsset.Group = r.URL
	return nil
}

func NewRepository(repoURL string) Repository {
	repository := Repository{
		URL: repoURL,
	}

	repository.Defaulted()
	if err := registry.CallHooks(&repository); err != nil {
		return Repository{}
	}

	return repository
}
