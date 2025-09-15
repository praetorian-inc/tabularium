package model

import (
	"fmt"
	"net/url"
	"regexp"
	"slices"

	"github.com/praetorian-inc/tabularium/pkg/lib/normalize"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// WebApplication represents a web application as a security testing target.

type BurpSite struct {
	ID   string `neo4j:"id" json:"id" desc:"Burp Suite site ID for integration with Burp Suite Enterprise" example:"313"`
	Name string `neo4j:"name" json:"name" desc:"Name of the Burp Suite site usually hostname" example:"app.example.com"`
}

func (b *BurpSite) Override(other BurpSite) {
	if other.ID != "" {
		b.ID = other.ID
	}
	if other.Name != "" {
		b.Name = other.Name
	}
}

type WebApplication struct {
	BaseAsset
	PrimaryURL string   `neo4j:"primary_url" json:"primary_url" desc:"The primary/canonical URL of the web application" example:"https://app.example.com"`
	URLs       []string `neo4j:"urls" json:"urls" desc:"Additional URLs associated with this web application" example:"[\"https://api.example.com\", \"https://admin.example.com\"]"`
	Name       string   `neo4j:"name" json:"name" desc:"Name of the web application" example:"Example App"`
	Burp       BurpSite `neo4j:"burp_site" json:"burp_site" desc:"Burp Suite site for integration with Burp Suite Enterprise" example:"{id: 313, name: \"app.example.com\"}"`
}

const WebApplicationLabel = "WebApplication"

var webAppKeyRegex = regexp.MustCompile(`^#webapplication#https?://[^?#]+$`)

func init() {
	MustRegisterLabel(WebApplicationLabel)
	registry.Registry.MustRegisterModel(&WebApplication{})
}

func (w *WebApplication) GetDescription() string {
	return "Represents a web application with a primary URL and associated URLs, designed for security testing and attack surface management."
}

func (w *WebApplication) GetLabels() []string {
	labels := []string{WebApplicationLabel, AssetLabel, TTLLabel}
	if w.Source == SeedSource {
		labels = append(labels, SeedLabel)
	}
	return labels
}

func (w *WebApplication) GetHooks() []registry.Hook {
	return []registry.Hook{
		useGroupAndIdentifier(w, &w.Name, &w.PrimaryURL),
		{
			Call: func() error {
				if w.PrimaryURL == "" {
					return fmt.Errorf("WebApplication requires non-empty PrimaryURL")
				}

				normalizedURL, err := normalize.Normalize(w.PrimaryURL)
				if err != nil {
					return fmt.Errorf("failed to normalize PrimaryURL: %w", err)
				}
				w.PrimaryURL = normalizedURL

				key := fmt.Sprintf("#webapplication#%s", w.PrimaryURL)
				if len(key) > 2048 {
					key = key[:2048]
				}
				w.Key = key

				normalizedURLs := make([]string, 0, len(w.URLs))
				for _, u := range w.URLs {
					if normalized, err := normalize.Normalize(u); err == nil {
						normalizedURLs = append(normalizedURLs, normalized)
					}
				}
				w.URLs = normalizedURLs

				return nil
			},
		},
		setGroupAndIdentifier(w, &w.Name, &w.PrimaryURL),
	}
}

func (w *WebApplication) Defaulted() {
	w.BaseAsset.Defaulted()
	w.Class = "webapplication"
	if w.URLs == nil {
		w.URLs = []string{}
	}
}

func (w *WebApplication) Valid() bool {
	return webAppKeyRegex.MatchString(w.Key)
}

func (w *WebApplication) WithStatus(status string) Target {
	ret := *w
	ret.Status = status
	return &ret
}

func (w *WebApplication) GetPrimaryURL() url.URL {
	parsed, err := url.Parse(w.PrimaryURL)
	if err != nil {
		return url.URL{}
	}
	return *parsed
}

func (w *WebApplication) Group() string {
	return w.Name
}

func (w *WebApplication) Identifier() string {
	return w.PrimaryURL
}

func (w *WebApplication) Merge(other Assetlike) {
	w.BaseAsset.Merge(other)
	otherApp, ok := other.(*WebApplication)
	if !ok {
		return
	}
	if otherApp.Name != "" {
		w.Name = otherApp.Name
	}
	for _, u := range otherApp.URLs {
		if !slices.Contains(w.URLs, u) {
			w.URLs = append(w.URLs, u)
		}
		w.Burp.Override(otherApp.Burp)
	}
}

func (w *WebApplication) Visit(other Assetlike) {
	w.BaseAsset.Visit(other)
	otherApp, ok := other.(*WebApplication)
	if !ok {
		return
	}
	if otherApp.Name != "" && w.Name == "" {
		w.Name = otherApp.Name
	}
	w.Burp.Override(otherApp.Burp)
}

func (w *WebApplication) Attribute(name, value string) Attribute {
	return NewAttribute(name, value, w)
}

func NewWebApplication(primaryURL, name string) WebApplication {
	w := WebApplication{
		PrimaryURL: primaryURL,
		Name:       name,
		URLs:       []string{},
		Burp:       BurpSite{},
	}

	w.Defaulted()
	registry.CallHooks(&w)

	return w
}

func NewWebApplicationSeed(primaryURL string) WebApplication {
	w := NewWebApplication(primaryURL, primaryURL)
	w.Source = SeedSource
	w.Status = Pending
	w.TTL = 0
	return w
}

func (w *WebApplication) SeedModels() []Seedable {
	copy := *w
	return []Seedable{&copy}
}
