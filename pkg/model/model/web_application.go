package model

import (
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"

	uu "github.com/praetorian-inc/tabularium/pkg/lib/url"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// WebApplication represents a web application as a security testing target.
// It extends BaseAsset with web-specific properties including URLs and Burp Suite integration.
type WebApplication struct {
	BaseAsset
	PrimaryURL string   `neo4j:"primary_url" json:"primary_url" desc:"The primary/canonical URL of the web application" example:"https://app.example.com"`
	URLs       []string `neo4j:"urls" json:"urls" desc:"Additional URLs associated with this web application" example:"[\"https://api.example.com\", \"https://admin.example.com\"]"`
	Name       string   `neo4j:"name" json:"name" desc:"Name of the web application" example:"Example App"`
	BurpSiteID string   `neo4j:"burp_site_id" json:"burp_site_id" desc:"Burp Suite site ID for integration with Burp Suite Enterprise" example:"abc123-def456-ghi789"`
}

const (
	WebApplicationLabel = "WebApplication"
	// MaxKeyLength defines the maximum length for Neo4j keys
	MaxKeyLength = 2048
)

var (
	// webAppKeyRegex validates the WebApplication key format
	webAppKeyRegex = regexp.MustCompile(`^#webapplication#https?://[^?#]+$`)
)

func init() {
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

				normalizedURL, err := uu.Normalize(w.PrimaryURL)
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
					if normalized, err := uu.Normalize(u); err == nil {
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

// Defaulted initializes default values for WebApplication fields
func (w *WebApplication) Defaulted() {
	w.BaseAsset.Defaulted()
	if w.URLs == nil {
		w.URLs = make([]string, 0)
	}
	// BurpSiteID defaults to empty string, which is handled by Go's zero value
}

// Valid checks if the WebApplication has a properly formatted key
func (w *WebApplication) Valid() bool {
	return w.Key != "" && webAppKeyRegex.MatchString(w.Key)
}

// WithStatus creates a copy with the specified status, preserving all fields including BurpSiteID
func (w *WebApplication) WithStatus(status string) Target {
	ret := *w
	ret.Status = status
	// Deep copy URLs to avoid shared slice references
	ret.URLs = make([]string, len(w.URLs))
	copy(ret.URLs, w.URLs)
	return &ret
}

func (w *WebApplication) Group() string {
	if parsed, err := url.Parse(w.PrimaryURL); err == nil {
		return fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)
	}
	return w.PrimaryURL
}

func (w *WebApplication) Identifier() string {
	if parsed, err := url.Parse(w.PrimaryURL); err == nil {
		if parsed.Path == "" || parsed.Path == "/" {
			return "/"
		}
		return parsed.Path
	}
	return w.PrimaryURL
}

// Merge combines data from another Assetlike, preferring non-empty values from other
func (w *WebApplication) Merge(other Assetlike) {
	w.BaseAsset.Merge(other)
	if otherApp, ok := other.(*WebApplication); ok {
		if otherApp.Name != "" {
			w.Name = otherApp.Name
		}
		if otherApp.PrimaryURL != "" {
			w.PrimaryURL = otherApp.PrimaryURL
		}
		for _, u := range otherApp.URLs {
			if !slices.Contains(w.URLs, u) {
				w.URLs = append(w.URLs, u)
			}
		}
		if otherApp.BurpSiteID != "" {
			w.BurpSiteID = otherApp.BurpSiteID
		}
	}
}

// Visit updates empty fields from another Assetlike without overwriting existing values
func (w *WebApplication) Visit(other Assetlike) {
	w.BaseAsset.Visit(other)
	if otherApp, ok := other.(*WebApplication); ok {
		if otherApp.Name != "" && w.Name == "" {
			w.Name = otherApp.Name
		}
		if otherApp.PrimaryURL != "" && w.PrimaryURL == "" {
			w.PrimaryURL = otherApp.PrimaryURL
		}
		if otherApp.BurpSiteID != "" && w.BurpSiteID == "" {
			w.BurpSiteID = otherApp.BurpSiteID
		}
	}
}

func (w *WebApplication) Attribute(name, value string) Attribute {
	return NewAttribute(name, value, w)
}

// IsHTTP returns true if the PrimaryURL uses HTTP or HTTPS protocol
func (w *WebApplication) IsHTTP() bool {
	return strings.HasPrefix(w.PrimaryURL, "http://") || strings.HasPrefix(w.PrimaryURL, "https://")
}

// IsHTTPS returns true if the PrimaryURL uses HTTPS protocol
func (w *WebApplication) IsHTTPS() bool {
	return strings.HasPrefix(w.PrimaryURL, "https://")
}

// HasBurpSiteID returns true if the WebApplication has a non-empty BurpSiteID
func (w *WebApplication) HasBurpSiteID() bool {
	return w.BurpSiteID != ""
}

func NewWebApplication(primaryURL, name string) WebApplication {
	w := WebApplication{
		PrimaryURL: primaryURL,
		Name:       name,
		URLs:       make([]string, 0),
		// BurpSiteID is intentionally left empty (zero value)
	}

	w.Defaulted()
	registry.CallHooks(&w)

	return w
}

// NewWebApplicationWithBurpSiteID creates a new WebApplication with a Burp Suite site ID
func NewWebApplicationWithBurpSiteID(primaryURL, name, burpSiteID string) WebApplication {
	w := NewWebApplication(primaryURL, name)
	w.BurpSiteID = burpSiteID
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
