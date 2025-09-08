package model

import (
	"strings"
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebApplicationStruct(t *testing.T) {
	tests := []struct {
		name        string
		primaryURL  string
		appName     string
		urls        []string
		expectedKey string
	}{
		{
			name:        "Basic HTTPS URL",
			primaryURL:  "https://app.example.com",
			appName:     "Example App",
			urls:        []string{},
			expectedKey: "#webapplication#https://app.example.com/",
		},
		{
			name:        "HTTP URL with default port",
			primaryURL:  "http://app.example.com:80",
			appName:     "Example App",
			urls:        []string{},
			expectedKey: "#webapplication#http://app.example.com/",
		},
		{
			name:        "HTTPS URL with default port",
			primaryURL:  "https://app.example.com:443",
			appName:     "Example App",
			urls:        []string{},
			expectedKey: "#webapplication#https://app.example.com/",
		},
		{
			name:        "URL with custom port",
			primaryURL:  "https://app.example.com:8443",
			appName:     "Example App",
			urls:        []string{},
			expectedKey: "#webapplication#https://app.example.com:8443/",
		},
		{
			name:        "URL with path",
			primaryURL:  "https://app.example.com/admin",
			appName:     "Admin Panel",
			urls:        []string{},
			expectedKey: "#webapplication#https://app.example.com/admin",
		},
		{
			name:        "URL with query and fragment",
			primaryURL:  "https://app.example.com/path?param=value#fragment",
			appName:     "Example App",
			urls:        []string{},
			expectedKey: "#webapplication#https://app.example.com/path",
		},
		{
			name:        "Mixed case URL",
			primaryURL:  "HTTPS://APP.EXAMPLE.COM/Path",
			appName:     "Example App",
			urls:        []string{},
			expectedKey: "#webapplication#https://app.example.com/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWebApplication(tt.primaryURL, tt.appName)
			
			assert.Equal(t, tt.appName, w.Name)
			assert.Equal(t, tt.expectedKey, w.Key)
			assert.Equal(t, Active, w.Status)
			assert.Equal(t, SelfSource, w.Source)
			assert.NotZero(t, w.TTL)
			assert.NotEmpty(t, w.Created)
			assert.NotEmpty(t, w.Visited)
		})
	}
}

func TestWebApplicationSeed(t *testing.T) {
	primaryURL := "https://seed.example.com"
	w := NewWebApplicationSeed(primaryURL)
	
	assert.Equal(t, primaryURL, w.Name)
	assert.Equal(t, "#webapplication#https://seed.example.com/", w.Key)
	assert.Equal(t, Pending, w.Status)
	assert.Equal(t, SeedSource, w.Source)
	assert.Equal(t, int64(0), w.TTL) // TTL should be 0 for seeds
}

func TestWebApplicationURLNormalization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "Basic HTTPS URL",
			input:    "https://example.com",
			expected: "https://example.com/",
			hasError: false,
		},
		{
			name:     "HTTP with default port",
			input:    "http://example.com:80",
			expected: "http://example.com/",
			hasError: false,
		},
		{
			name:     "HTTPS with default port",
			input:    "https://example.com:443",
			expected: "https://example.com/",
			hasError: false,
		},
		{
			name:     "Custom port preserved",
			input:    "https://example.com:8080",
			expected: "https://example.com:8080/",
			hasError: false,
		},
		{
			name:     "Query and fragment removed",
			input:    "https://example.com/path?query=value&other=param#fragment",
			expected: "https://example.com/path",
			hasError: false,
		},
		{
			name:     "Mixed case normalized",
			input:    "HTTPS://EXAMPLE.COM/PATH",
			expected: "https://example.com/path",
			hasError: false,
		},
		{
			name:     "Empty URL",
			input:    "",
			expected: "",
			hasError: true,
		},
		{
			name:     "No scheme",
			input:    "example.com",
			expected: "",
			hasError: true,
		},
		{
			name:     "No host",
			input:    "https://",
			expected: "",
			hasError: true,
		},
		{
			name:     "Path with trailing slash preserved",
			input:    "https://example.com/path/",
			expected: "https://example.com/path/",
			hasError: false,
		},
		{
			name:     "Deep path preserved",
			input:    "https://example.com/api/v1/users/123",
			expected: "https://example.com/api/v1/users/123",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized, err := normalizeURL(tt.input)
			
			if tt.hasError {
				assert.Error(t, err)
				assert.Empty(t, normalized)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, normalized)
			}
		})
	}
}

func TestWebApplicationValidation(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		valid bool
	}{
		{
			name:  "Valid HTTPS webapp key",
				key:   "#webapplication#https://example.com/",
			valid: true,
		},
		{
			name:  "Valid HTTP webapp key",
				key:   "#webapplication#http://example.com/",
			valid: true,
		},
		{
			name:  "Valid webapp key with path",
				key:   "#webapplication#https://example.com/api/v1",
			valid: true,
		},
		{
			name:  "Valid webapp key with port",
				key:   "#webapplication#https://example.com:8080/",
			valid: true,
		},
		{
			name:  "Invalid key - missing prefix",
			key:   "https://example.com/",
			valid: false,
		},
		{
			name:  "Invalid key - wrong prefix",
			key:   "#webpage#https://example.com/",
			valid: false,
		},
		{
			name:  "Invalid key - no protocol",
				key:   "#webapplication#example.com/",
			valid: false,
		},
		{
			name:  "Invalid key - query parameters",
				key:   "#webapplication#https://example.com/?param=value",
			valid: false,
		},
		{
			name:  "Invalid key - fragment",
				key:   "#webapplication#https://example.com/#section",
			valid: false,
		},
		{
			name:  "Invalid key - ftp protocol",
				key:   "#webapplication#ftp://example.com/",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := WebApplication{
				BaseAsset: BaseAsset{Key: tt.key},
			}
			assert.Equal(t, tt.valid, w.Valid(), "Key validation for: %s", tt.key)
		})
	}
}

func TestWebApplicationKeyLength(t *testing.T) {
	// Create a very long URL that exceeds 2048 characters
	longHost := strings.Repeat("verylongsubdomain.", 100) + "example.com"
	longURL := "https://" + longHost + "/very/long/path/that/keeps/going"
	
	w := NewWebApplication(longURL, "Long URL Test")
	
	// Key should be truncated to exactly 2048 characters
	assert.LessOrEqual(t, len(w.Key), 2048)
	assert.True(t, strings.HasPrefix(w.Key, "#webapplication#"))
}

func TestWebApplicationLabels(t *testing.T) {
	// Test normal webapp labels
	w := NewWebApplication("https://example.com", "Example")
	labels := w.GetLabels()
	
	expectedLabels := []string{WebApplicationLabel, AssetLabel, TTLLabel}
	assert.ElementsMatch(t, expectedLabels, labels)
	
	// Test seed webapp labels
	seedApp := NewWebApplicationSeed("https://seed.example.com")
	seedLabels := seedApp.GetLabels()
	
	expectedSeedLabels := []string{WebApplicationLabel, AssetLabel, TTLLabel, SeedLabel}
	assert.ElementsMatch(t, expectedSeedLabels, seedLabels)
}

func TestWebApplicationTargetInterface(t *testing.T) {
	w := NewWebApplication("https://app.example.com/admin", "Admin Panel")
	
	// Test Target interface methods
	assert.Equal(t, Active, w.GetStatus())
	assert.True(t, w.IsStatus("A"))
	assert.False(t, w.IsStatus("P"))
	
	// Test WithStatus
	newStatus := w.WithStatus(Pending)
	assert.Equal(t, Pending, newStatus.GetStatus())
	assert.Equal(t, Active, w.GetStatus()) // Original should be unchanged
	
	// Test Group and Identifier
	assert.Equal(t, "https://app.example.com", w.Group())
	assert.Equal(t, "/admin", w.Identifier())
	
	// Test root path identifier
	rootApp := NewWebApplication("https://example.com", "Root")
	assert.Equal(t, "/", rootApp.Identifier())
}

func TestWebApplicationAssetlikeInterface(t *testing.T) {
	w1 := NewWebApplication("https://app.example.com", "App 1")
	w1.URLs = []string{"https://api.example.com"}
	
	w2 := NewWebApplication("https://app.example.com", "App 2") 
	w2.URLs = []string{"https://admin.example.com", "https://api.example.com"}
	w2.PrimaryURL = "https://updated.example.com"
	
	// Test Merge
	w1.Merge(&w2)
	assert.Equal(t, "https://updated.example.com", w1.PrimaryURL)
	assert.Contains(t, w1.URLs, "https://admin.example.com")
	assert.Contains(t, w1.URLs, "https://api.example.com")
	// Should not have duplicates
	apiCount := 0
	for _, url := range w1.URLs {
		if url == "https://api.example.com" {
			apiCount++
		}
	}
	assert.Equal(t, 1, apiCount)
	
	// Test Visit
	w3 := NewWebApplication("", "")
	w4 := NewWebApplication("https://visit.example.com", "Visit Test")
	w3.Visit(&w4)
	assert.Equal(t, "https://visit.example.com/", w3.PrimaryURL) // URL normalization adds trailing slash
	assert.Equal(t, "Visit Test", w3.Name)
	
	// Test Attribute creation
	attr := w1.Attribute("test", "value")
	assert.Equal(t, "test", attr.Name)
	assert.Equal(t, "value", attr.Value)
}

func TestWebApplicationRegistryIntegration(t *testing.T) {
	// Test that WebApplication is properly registered
	model, found := registry.Registry.MakeType("webapplication")
	assert.True(t, found)
	assert.IsType(t, &WebApplication{}, model)
	
	// Test that it implements Target interface
	if target, ok := model.(Target); ok {
		assert.NotNil(t, target)
	} else {
		t.Fatal("WebApplication should implement Target interface")
	}
	
	// Test that it implements Assetlike interface
	if assetlike, ok := model.(Assetlike); ok {
		assert.NotNil(t, assetlike)
	} else {
		t.Fatal("WebApplication should implement Assetlike interface")
	}
}

func TestWebApplicationEdgeCases(t *testing.T) {
	// Test empty URLs slice handling
	w := NewWebApplication("https://example.com", "Test")
	assert.NotNil(t, w.URLs)
	assert.Empty(t, w.URLs)
	
	// Test URL normalization failure handling
	w2 := WebApplication{
		PrimaryURL: "invalid-url",
	}
	hooks := w2.GetHooks()
	require.NotEmpty(t, hooks)
	
	// The hook should handle invalid URLs gracefully
	err := hooks[1].Call()
	assert.Error(t, err) // Should return error for invalid URL
	
	// Test empty primary_url validation
	w3 := WebApplication{
		PrimaryURL: "",
	}
	hooks3 := w3.GetHooks()
	require.NotEmpty(t, hooks3)
	
	// The hook should reject empty PrimaryURL
	err3 := hooks3[1].Call()
	assert.Error(t, err3) // Should return error for empty PrimaryURL
	assert.Contains(t, err3.Error(), "requires non-empty PrimaryURL")
	
	// Test Group() and Identifier() methods with empty primary_url
	// When PrimaryURL is empty, url.Parse("") succeeds but returns empty scheme/host
	// Group() returns "://" and Identifier() returns w.PrimaryURL (empty string)
	assert.Equal(t, "://", w3.Group())
	assert.Equal(t, "/", w3.Identifier())
	
	// Test merge with non-WebApplication
	asset := NewAsset("example.com", "example.com")
	w.Merge(&asset)
	// Should not panic and BaseAsset merge should work
	
	// Test visit with non-WebApplication  
	w.Visit(&asset)
	// Should not panic and BaseAsset visit should work
}

func TestWebApplicationTTLBehavior(t *testing.T) {
	// Regular webapp should have non-zero TTL
	w1 := NewWebApplication("https://example.com", "Regular")
	assert.NotZero(t, w1.TTL)
	assert.Equal(t, SelfSource, w1.Source)
	
	// Seed webapp should have zero TTL
	w2 := NewWebApplicationSeed("https://seed.example.com")
	assert.Zero(t, w2.TTL)
	assert.Equal(t, SeedSource, w2.Source)
	
	// Test that changing source after creation doesn't automatically change TTL
	// TTL=0 is only applied during seed creation in NewWebApplicationSeed
	w3 := NewWebApplication("https://test.example.com", "Test")
	originalTTL := w3.TTL
	assert.NotZero(t, originalTTL) // Should have non-zero TTL initially
	
	// Change to seed source - TTL should remain the same unless explicitly set
	w3.Source = SeedSource
	w3.Defaulted() // This doesn't change TTL for existing instances
	assert.NotZero(t, w3.TTL) // TTL remains unchanged
	assert.Equal(t, originalTTL, w3.TTL) // TTL should be the same
}

func TestWebApplicationDescription(t *testing.T) {
	w := WebApplication{}
	description := w.GetDescription()
	assert.NotEmpty(t, description)
	assert.Contains(t, strings.ToLower(description), "web application")
}

func TestWebApplicationURLsNormalization(t *testing.T) {
	w := WebApplication{
		PrimaryURL: "https://example.com",
		URLs: []string{
			"https://api.example.com:443",
			"http://admin.example.com:80",
			"https://MIXED.EXAMPLE.COM/Path?query=1#frag",
			"invalid-url", // This should be filtered out
			"https://valid.example.com",
		},
	}
	
	hooks := w.GetHooks()
	require.NotEmpty(t, hooks)
	
	// Call the normalization hook
	err := hooks[1].Call()
	assert.NoError(t, err)
	
	// Check that URLs were normalized and invalid ones filtered
	expectedURLs := []string{
		"https://api.example.com/",
		"http://admin.example.com/",
		"https://mixed.example.com/path",
		"https://valid.example.com/",
	}
	
	assert.ElementsMatch(t, expectedURLs, w.URLs)
}

func TestWebApplicationSeedModels(t *testing.T) {
	webApp := NewWebApplicationSeed("https://app.example.com/dashboard")
	
	// Verify the Seedable interface is implemented
	var seedable Seedable = &webApp
	assert.NotNil(t, seedable)
	
	// Test SeedModels method
	seedModels := webApp.SeedModels()
	
	assert.Len(t, seedModels, 1)
	
	// Verify it returns a copy, not the original
	returnedWebApp := seedModels[0].(*WebApplication)
	assert.NotSame(t, &webApp, returnedWebApp)
	
	// Verify the copy has the same data
	assert.Equal(t, webApp.PrimaryURL, returnedWebApp.PrimaryURL)
	assert.Equal(t, webApp.Name, returnedWebApp.Name)
	assert.Equal(t, webApp.Status, returnedWebApp.Status)
	assert.Equal(t, webApp.Source, returnedWebApp.Source)
	assert.Equal(t, webApp.Key, returnedWebApp.Key)
}

func TestWebApplicationSeedableInterface(t *testing.T) {
	webApp := NewWebApplicationSeed("https://example.com")
	
	// Test GetSource/SetSource methods (inherited from BaseAsset)
	assert.Equal(t, SeedSource, webApp.GetSource())
	
	webApp.SetSource("test-source")
	// Should remain SeedSource because seed source always wins
	assert.Equal(t, SeedSource, webApp.GetSource())
	
	// Test SetOrigin (inherited from BaseAsset)
	webApp.SetOrigin("test-origin")
	assert.Equal(t, "test-origin", webApp.Origin)
	
	// Test Target interface methods
	assert.True(t, webApp.IsStatus(Pending))
	assert.Equal(t, "https://example.com", webApp.Group())
	assert.Equal(t, "/", webApp.Identifier())
}