package model

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

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
				return w.normalize()
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
	otherApp, ok := other.(*WebApplication)
	if !ok {
		return
	}

	// Update fields with non-empty values from other
	if otherApp.PrimaryURL != "" {
		w.PrimaryURL = otherApp.PrimaryURL
	}
	if otherApp.Name != "" {
		w.Name = otherApp.Name
	}
	if otherApp.BurpSiteID != "" {
		w.BurpSiteID = otherApp.BurpSiteID
	}

	// Merge URLs without duplicates
	w.mergeURLs(otherApp.URLs)
}

// Visit updates empty fields from another Assetlike without overwriting existing values
func (w *WebApplication) Visit(other Assetlike) {
	w.BaseAsset.Visit(other)
	otherApp, ok := other.(*WebApplication)
	if !ok {
		return
	}

	// Only update if our fields are empty
	if w.PrimaryURL == "" && otherApp.PrimaryURL != "" {
		w.PrimaryURL = otherApp.PrimaryURL
	}
	if w.Name == "" && otherApp.Name != "" {
		w.Name = otherApp.Name
	}
	if w.BurpSiteID == "" && otherApp.BurpSiteID != "" {
		w.BurpSiteID = otherApp.BurpSiteID
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

// IsPublic returns true if the web application is publicly accessible
func (w *WebApplication) IsPublic() bool {
	return !w.IsPrivate()
}

// HasBurpSiteID returns true if the web application has a Burp Suite site ID configured
func (w *WebApplication) HasBurpSiteID() bool {
	return w.BurpSiteID != ""
}

// normalizeURL normalizes a URL for consistent storage and comparison.
// It converts scheme and host to lowercase, removes default ports,
// ensures a path exists, and strips query/fragment components.
func normalizeURL(rawURL string) (string, error) {
	if rawURL == "" {
		return "", fmt.Errorf("empty URL")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	if parsed.Scheme == "" {
		return "", fmt.Errorf("URL missing scheme")
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("URL missing host")
	}

	// Normalize scheme and host to lowercase
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)

	// Remove default ports for cleaner URLs
	parsed.Host = removeDefaultPort(parsed.Scheme, parsed.Host)

	// Ensure path exists and is normalized
	if parsed.Path == "" {
		parsed.Path = "/"
	} else {
		// Normalize path to lowercase for consistency
		parsed.Path = strings.ToLower(parsed.Path)
	}

	// Remove query and fragment for canonical URL
	parsed.RawQuery = ""
	parsed.Fragment = ""

	return parsed.String(), nil
}

// removeDefaultPort removes default HTTP/HTTPS ports from the host string
func removeDefaultPort(scheme, host string) string {
	switch scheme {
	case "http":
		if strings.HasSuffix(host, ":80") {
			return strings.TrimSuffix(host, ":80")
		}
	case "https":
		if strings.HasSuffix(host, ":443") {
			return strings.TrimSuffix(host, ":443")
		}
	}
	return host
}

// normalize performs URL normalization and key generation for the WebApplication
func (w *WebApplication) normalize() error {
	if w.PrimaryURL != "" {
		normalizedURL, err := normalizeURL(w.PrimaryURL)
		if err != nil {
			return fmt.Errorf("failed to normalize PrimaryURL: %w", err)
		}
		w.PrimaryURL = normalizedURL
		w.Key = w.generateKey()
	}

	// Normalize additional URLs, filtering out invalid ones
	w.URLs = w.normalizeURLList(w.URLs)
	return nil
}

// generateKey creates the Neo4j key for this WebApplication
func (w *WebApplication) generateKey() string {
	key := fmt.Sprintf("#webapplication#%s", w.PrimaryURL)
	if len(key) > MaxKeyLength {
		key = key[:MaxKeyLength]
	}
	return key
}

// normalizeURLList normalizes a list of URLs, filtering out invalid ones
func (w *WebApplication) normalizeURLList(urls []string) []string {
	if len(urls) == 0 {
		return urls
	}

	normalizedURLs := make([]string, 0, len(urls))
	seen := make(map[string]bool)

	for _, u := range urls {
		if normalized, err := normalizeURL(u); err == nil && !seen[normalized] {
			normalizedURLs = append(normalizedURLs, normalized)
			seen[normalized] = true
		}
	}
	return normalizedURLs
}

// mergeURLs merges additional URLs without creating duplicates
func (w *WebApplication) mergeURLs(otherURLs []string) {
	if len(otherURLs) == 0 {
		return
	}

	urlSet := make(map[string]bool, len(w.URLs)+len(otherURLs))
	for _, u := range w.URLs {
		urlSet[u] = true
	}

	for _, u := range otherURLs {
		if !urlSet[u] {
			w.URLs = append(w.URLs, u)
			urlSet[u] = true
		}
	}
}

// NewWebApplication creates a new WebApplication with the specified primary URL and name.
// It initializes default values and runs registry hooks for proper setup.
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

func (w *WebApplication) GetDescription() string {
	return "Represents a web application with a primary URL and associated URLs, designed for security testing and attack surface management."
}
