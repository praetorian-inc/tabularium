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

type Repository struct {
	BaseAsset
	URL  string `neo4j:"url,omitempty" json:"url,omitempty" desc:"Repository URL." example:"https://github.com/praetorian-inc/tabularium"`
	Org  string `neo4j:"org,omitempty" json:"org,omitempty" desc:"Organization name." example:"praetorian-inc"`
	Name string `neo4j:"name,omitempty" json:"name,omitempty" desc:"Repository name." example:"praetorian-inc/tabularium"`
}

var (
	RepositoryLabel = NewLabel("Repository")
)

var (
	repository    = regexp.MustCompile(`^(https://)?(github\.com|gitlab\.com|bitbucket\.(com|org))/([^/]+)/(([^/]+/)*[^/]+)$`)
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
	return false
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
	r.Org = parts[3]
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
