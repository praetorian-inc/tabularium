package model

import (
	"encoding/json"
	"fmt"
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
			expectedKey: "#webapplication#https://app.example.com/Path",
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
	longHost := strings.Repeat("verylongsubdomain", 100) + ".example.com"
	longURL := "https://" + longHost + "/very/long/path/that/keeps/going"

	w := NewWebApplication(longURL, "Long URL Test")

	assert.LessOrEqual(t, len(w.Key), 2048)
	assert.True(t, strings.HasPrefix(w.Key, "#webapplication#"))
}

func TestWebApplicationLabels(t *testing.T) {
	w := NewWebApplication("https://example.com", "Example")
	labels := w.GetLabels()
	expectedLabels := []string{WebApplicationLabel, AssetLabel, TTLLabel}
	assert.ElementsMatch(t, expectedLabels, labels)

	seedApp := NewWebApplicationSeed("https://seed.example.com")
	seedLabels := seedApp.GetLabels()
	expectedSeedLabels := []string{WebApplicationLabel, AssetLabel, TTLLabel, SeedLabel}
	assert.ElementsMatch(t, expectedSeedLabels, seedLabels)
	assert.True(t, seedApp.IsSeed())

	assert.Empty(t, seedApp.BurpSiteID)
	assert.Empty(t, seedApp.BurpFolderID)
	assert.Empty(t, seedApp.BurpScheduleID)
}

func TestWebApplicationTargetInterface(t *testing.T) {
	w := NewWebApplication("https://app.example.com/admin", "Admin Panel")

	assert.Equal(t, Active, w.GetStatus())
	assert.True(t, w.IsStatus("A"))
	assert.False(t, w.IsStatus("P"))

	newStatus := w.WithStatus(Pending)
	assert.Equal(t, Pending, newStatus.GetStatus())
	assert.Equal(t, Active, w.GetStatus())

	assert.Equal(t, "https://app.example.com/admin", w.Identifier())
	assert.Equal(t, "Admin Panel", w.Group())

	rootApp := NewWebApplication("https://example.com", "Root")
	assert.Equal(t, "Root", rootApp.Group())
	assert.Equal(t, "https://example.com/", rootApp.Identifier())
}

func TestWebApplicationMergeURLs(t *testing.T) {
	w1 := NewWebApplication("https://app.example.com", "App 1")
	w1.URLs = []string{"https://api.example.com"}

	w2 := NewWebApplication("https://app.example.com", "App 2")
	w2.URLs = []string{"https://admin.example.com", "https://api.example.com"}

	w1.Merge(&w2)
	assert.Equal(t, "https://app.example.com/", w1.PrimaryURL)
	assert.Contains(t, w1.URLs, "https://admin.example.com")
	assert.Contains(t, w1.URLs, "https://api.example.com")
	assert.Len(t, w1.URLs, 2)
}

func TestWebApplicationMergeBurpMetadata(t *testing.T) {
	w1 := NewWebApplication("https://app.example.com", "App 1")
	w1.BurpSiteID = "old-site"
	w1.BurpFolderID = "old-folder"
	w1.BurpScheduleID = "old-schedule"

	w2 := NewWebApplication("https://app.example.com", "App 1")
	w2.BurpSiteID = "new-site"
	w2.BurpScheduleID = "new-schedule"

	w1.Merge(&w2)

	assert.Equal(t, "new-site", w1.BurpSiteID)
	assert.Equal(t, "old-folder", w1.BurpFolderID)
	assert.Equal(t, "new-schedule", w1.BurpScheduleID)
}

func TestWebApplicationVisitBurpMetadata(t *testing.T) {
	w1 := NewWebApplication("https://existing.example.com", "Existing")
	w1.BurpSiteID = "current-site"
	w1.BurpFolderID = "current-folder"
	w1.BurpScheduleID = "current-schedule"

	incoming := NewWebApplication("https://incoming.example.com", "Incoming")
	incoming.BurpSiteID = "incoming-site"
	incoming.BurpFolderID = "incoming-folder"

	w1.Visit(&incoming)

	assert.Equal(t, "incoming-site", w1.BurpSiteID)
	assert.Equal(t, "incoming-folder", w1.BurpFolderID)
	assert.Equal(t, "current-schedule", w1.BurpScheduleID)
}

func TestWebApplicationRegistryIntegration(t *testing.T) {
	model, found := registry.Registry.MakeType("webapplication")
	assert.True(t, found)
	assert.IsType(t, &WebApplication{}, model)

	if target, ok := model.(Target); ok {
		assert.NotNil(t, target)
	} else {
		t.Fatal("WebApplication should implement Target interface")
	}

	if assetlike, ok := model.(Assetlike); ok {
		assert.NotNil(t, assetlike)
	} else {
		t.Fatal("WebApplication should implement Assetlike interface")
	}
}

func TestWebApplicationTTLBehavior(t *testing.T) {
	w1 := NewWebApplication("https://example.com", "Regular")
	assert.NotZero(t, w1.TTL)
	assert.Equal(t, SelfSource, w1.Source)

	w2 := NewWebApplicationSeed("https://seed.example.com")
	assert.Zero(t, w2.TTL)
	assert.Equal(t, SeedSource, w2.Source)
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
			"invalid-url",
			"https://valid.example.com",
		},
	}

	hooks := w.GetHooks()
	require.NotEmpty(t, hooks)

	for _, hook := range hooks {
		err := hook.Call()
		assert.NoError(t, err)
	}

	expectedURLs := []string{
		"https://api.example.com/",
		"http://admin.example.com/",
		"https://mixed.example.com/Path",
		"https://valid.example.com/",
	}

	assert.ElementsMatch(t, expectedURLs, w.URLs)
}

func TestWebApplicationSeedModels(t *testing.T) {
	webApp := NewWebApplicationSeed("https://app.example.com/dashboard")

	var seedable Seedable = &webApp
	assert.NotNil(t, seedable)

	seedModels := webApp.SeedModels()

	assert.Len(t, seedModels, 1)

	returnedWebApp := seedModels[0].(*WebApplication)
	assert.NotSame(t, &webApp, returnedWebApp)

	assert.Equal(t, webApp.PrimaryURL, returnedWebApp.PrimaryURL)
	assert.Equal(t, webApp.Name, returnedWebApp.Name)
	assert.Equal(t, webApp.Status, returnedWebApp.Status)
	assert.Equal(t, webApp.Source, returnedWebApp.Source)
	assert.Equal(t, webApp.Key, returnedWebApp.Key)
}

func TestWebApplicationSeedPromotion(t *testing.T) {
	tests := []struct {
		name              string
		existingSource    string
		incomingSource    string
		expectPromotion   bool
		expectedPromotion string
	}{
		{
			name:              "Promotion from self to seed",
			existingSource:    SelfSource,
			incomingSource:    SeedSource,
			expectPromotion:   true,
			expectedPromotion: SeedLabel,
		},
		{
			name:              "No promotion - both self source",
			existingSource:    SelfSource,
			incomingSource:    SelfSource,
			expectPromotion:   false,
			expectedPromotion: NO_PENDING_LABEL_ADDITION,
		},
		{
			name:              "No promotion - both seed source",
			existingSource:    SeedSource,
			incomingSource:    SeedSource,
			expectPromotion:   false,
			expectedPromotion: NO_PENDING_LABEL_ADDITION,
		},
		{
			name:              "No promotion - seed to self (downgrade)",
			existingSource:    SeedSource,
			incomingSource:    SelfSource,
			expectPromotion:   false,
			expectedPromotion: NO_PENDING_LABEL_ADDITION,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			existing := NewWebApplication("https://test.example.com", "Test App")
			existing.Source = tt.existingSource

			incoming := NewWebApplication("https://test.example.com", "Test App")
			incoming.Source = tt.incomingSource

			existing.Merge(&incoming)

			if tt.expectPromotion {
				pendingPromotion, required := PendingLabelAddition(&existing)
				assert.Equal(t, tt.expectedPromotion, pendingPromotion,
					"Expected pending promotion to be set")
				assert.True(t, required,
					"PendingPromotion should return true")
			} else {
				pendingPromotion, required := PendingLabelAddition(&existing)
				assert.Equal(t, tt.expectedPromotion, pendingPromotion,
					"Expected no pending promotion")
				assert.False(t, required,
					"HasPendingPromotion should return false")
			}
		})
	}
}

func TestWebApplicationLabelSettableInterface(t *testing.T) {
	webapp := NewWebApplication("https://example.com", "Test")
	var _ LabelSettable = &webapp

	pendingAddition, required := PendingLabelAddition(&webapp)
	assert.Equal(t, NO_PENDING_LABEL_ADDITION, pendingAddition)
	assert.False(t, required)

	webapp.PendingLabelAddition = SeedLabel
	pendingAddition, required = PendingLabelAddition(&webapp)
	assert.Equal(t, SeedLabel, pendingAddition)
	assert.True(t, required)
}

func TestWebApplicationMergeName(t *testing.T) {
	w1 := NewWebApplication("https://app.example.com", "Original Name")
	w2 := NewWebApplication("https://app.example.com", "Updated Name")

	w1.Merge(&w2)

	assert.Equal(t, "Updated Name", w1.Name)
}

func TestWebApplicationMergeEmptyName(t *testing.T) {
	w1 := NewWebApplication("https://app.example.com", "Original Name")
	w2 := NewWebApplication("https://app.example.com", "")

	w1.Merge(&w2)

	assert.Equal(t, "Original Name", w1.Name)
}

func TestWebApplicationHydrationLifecycle(t *testing.T) {
	primaryURL := "https://api.example.com"
	w := NewWebApplication(primaryURL, "Example App")
	w.BurpType = "webapplication"

	expectedPath := fmt.Sprintf("webapplication/%s/api-definition.json", RemoveReservedCharacters(w.PrimaryURL))
	assert.Equal(t, expectedPath, w.GetHydratableFilepath())

	assert.False(t, w.IsWebService())
	assert.Equal(t, SKIP_HYDRATION, w.HydratableFilepath())

	w.BurpType = "webservice"
	assert.True(t, w.IsWebService())
	assert.Equal(t, expectedPath, w.HydratableFilepath())

	originalContent := APIDefinitionResult{
		PrimaryURL: w.PrimaryURL,
		FileBasedDefinition: &FileBasedAPIDefinition{
			Filename:         "openapi.json",
			Contents:         "{}",
			EnabledEndpoints: []EnabledEndpoint{{ID: "endpoint-1"}},
		},
	}

	w.WebApplicationDetails.ApiDefinitionContent = originalContent

	file := w.HydratedFile()
	assert.Equal(t, expectedPath, file.Name)
	assert.Equal(t, expectedPath, w.ApiDefinitionContentPath)

	var parsed APIDefinitionResult
	require.NoError(t, json.Unmarshal(file.Bytes, &parsed))
	assert.Equal(t, originalContent, parsed)

	dehydrated := w.Dehydrate()
	dehydratedApp, ok := dehydrated.(*WebApplication)
	require.True(t, ok)
	assert.Equal(t, expectedPath, dehydratedApp.ApiDefinitionContentPath)
	assert.Empty(t, dehydratedApp.WebApplicationDetails.ApiDefinitionContent)

	rehydrated := NewWebApplication(primaryURL, "Rehydrated")
	rehydrated.ApiDefinitionContentPath = dehydratedApp.ApiDefinitionContentPath
	require.NoError(t, rehydrated.Hydrate(file.Bytes))
	assert.Equal(t, originalContent, rehydrated.WebApplicationDetails.ApiDefinitionContent)
}

func TestWebApplicationGobEncodeDecode(t *testing.T) {
	original := NewWebApplication("https://gob.example.com", "Gob Example")
	original.URLs = []string{"https://gob.example.com/api"}
	original.ApiDefinitionContentPath = original.HydratableFilepath()
	original.WebApplicationDetails.ApiDefinitionContent = APIDefinitionResult{PrimaryURL: original.PrimaryURL}

	data, err := original.GobEncode()
	require.NoError(t, err)

	assert.NotEmpty(t, original.WebApplicationDetails.ApiDefinitionContent)

	var decoded WebApplication
	require.NoError(t, decoded.GobDecode(data))

	assert.Equal(t, original.PrimaryURL, decoded.PrimaryURL)
	assert.Equal(t, original.URLs, decoded.URLs)
	assert.Equal(t, original.ApiDefinitionContentPath, decoded.ApiDefinitionContentPath)
	assert.Empty(t, decoded.WebApplicationDetails.ApiDefinitionContent)
}
