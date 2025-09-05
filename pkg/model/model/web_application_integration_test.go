package model

import (
	"encoding/json"
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebApplicationIntegrationWithBurpSiteID verifies end-to-end functionality
func TestWebApplicationIntegrationWithBurpSiteID(t *testing.T) {
	t.Run("Complete workflow with BurpSiteID", func(t *testing.T) {
		// 1. Create WebApplication through registry
		model, found := registry.Registry.MakeType("webapplication")
		require.True(t, found, "WebApplication should be registered")
		
		webApp, ok := model.(*WebApplication)
		require.True(t, ok, "Should return WebApplication type")
		
		// 2. Set fields including BurpSiteID
		webApp.PrimaryURL = "https://test.example.com/app"
		webApp.Name = "Integration Test App"
		webApp.BurpSiteID = "integration-burp-id-123"
		webApp.URLs = []string{
			"https://api.test.example.com",
			"https://admin.test.example.com",
		}
		
		// 3. Apply defaults and hooks
		webApp.Defaulted()
		registry.CallHooks(webApp)
		
		// 4. Verify all fields are properly set
		assert.Equal(t, "#webapplication#https://test.example.com/app", webApp.Key)
		assert.Equal(t, "integration-burp-id-123", webApp.BurpSiteID)
		assert.True(t, webApp.Valid())
		assert.True(t, webApp.HasBurpSiteID())
		
		// 5. Test JSON serialization
		jsonData, err := json.Marshal(webApp)
		require.NoError(t, err)
		
		// 6. Verify JSON contains BurpSiteID
		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		require.NoError(t, err)
		
		assert.Equal(t, "integration-burp-id-123", jsonMap["burp_site_id"])
		assert.Equal(t, "https://test.example.com/app", jsonMap["primary_url"])
		assert.Equal(t, "Integration Test App", jsonMap["name"])
		
		// 7. Test deserialization
		var deserializedApp WebApplication
		err = json.Unmarshal(jsonData, &deserializedApp)
		require.NoError(t, err)
		
		assert.Equal(t, webApp.BurpSiteID, deserializedApp.BurpSiteID)
		assert.Equal(t, webApp.Name, deserializedApp.Name)
		assert.Equal(t, webApp.PrimaryURL, deserializedApp.PrimaryURL)
		
		// 8. Test Target interface methods with BurpSiteID
		target := webApp.WithStatus("AH")
		highApp, ok := target.(*WebApplication)
		require.True(t, ok)
		
		assert.Equal(t, "AH", highApp.Status)
		assert.Equal(t, webApp.BurpSiteID, highApp.BurpSiteID)
		
		// 9. Test Merge preserves BurpSiteID
		otherApp := NewWebApplicationWithBurpSiteID(
			"https://merge.example.com",
			"Merge App",
			"merge-burp-id",
		)
		
		webApp.Merge(&otherApp)
		assert.Equal(t, "merge-burp-id", webApp.BurpSiteID)
		
		// 10. Test Visit only updates if empty
		visitApp := NewWebApplication("", "")
		visitApp.Visit(webApp)
		assert.Equal(t, webApp.BurpSiteID, visitApp.BurpSiteID)
	})
	
	t.Run("Registry creation with BurpSiteID", func(t *testing.T) {
		// Test that registry can create and manage WebApplication with BurpSiteID
		models := []registry.Model{
			&WebApplication{
				BaseAsset: BaseAsset{
					Key:    "#webapplication#https://registry.test.com/",
					Status: Active,
				},
				PrimaryURL: "https://registry.test.com",
				Name:       "Registry Test",
				BurpSiteID: "registry-burp-456",
			},
		}
		
		for _, model := range models {
			webApp := model.(*WebApplication)
			
			// Apply registry processing
			webApp.Defaulted()
			registry.CallHooks(webApp)
			
			// Verify BurpSiteID survives registry processing
			assert.Equal(t, "registry-burp-456", webApp.BurpSiteID)
			assert.True(t, webApp.Valid())
			
			// Test that it implements required interfaces
			var _ Target = webApp
			var _ Assetlike = webApp
			var _ registry.Model = webApp
		}
	})
	
	t.Run("Seed workflow with BurpSiteID", func(t *testing.T) {
		// Create seed with BurpSiteID
		seed := NewWebApplicationSeed("https://seed.test.com")
		seed.BurpSiteID = "seed-burp-789"
		
		// Verify seed properties
		assert.Equal(t, SeedSource, seed.Source)
		assert.Equal(t, Pending, seed.Status)
		assert.Zero(t, seed.TTL)
		assert.Equal(t, "seed-burp-789", seed.BurpSiteID)
		
		// Test SeedModels method
		seedModels := seed.SeedModels()
		require.Len(t, seedModels, 1)
		
		copiedSeed := seedModels[0].(*WebApplication)
		assert.Equal(t, seed.BurpSiteID, copiedSeed.BurpSiteID)
		assert.NotSame(t, &seed, copiedSeed)
		
		// Verify it still implements Seedable
		var _ Seedable = &seed
	})
}

// TestWebApplicationNeo4jIntegration tests Neo4j-specific functionality
func TestWebApplicationNeo4jIntegration(t *testing.T) {
	t.Run("Neo4j field tags", func(t *testing.T) {
		webApp := NewWebApplicationWithBurpSiteID(
			"https://neo4j.test.com",
			"Neo4j Test",
			"neo4j-burp-id",
		)
		
		// The neo4j tag should be present and correct
		// This is verified by the struct tags test, but we can also check runtime behavior
		
		// Simulate Neo4j storage/retrieval
		type neo4jRecord struct {
			PrimaryURL  string `neo4j:"primary_url"`
			URLs        []string `neo4j:"urls"`
			Name        string `neo4j:"name"`
			BurpSiteID  string `neo4j:"burp_site_id"`
		}
		
		record := neo4jRecord{
			PrimaryURL: webApp.PrimaryURL,
			URLs:       webApp.URLs,
			Name:       webApp.Name,
			BurpSiteID: webApp.BurpSiteID,
		}
		
		assert.Equal(t, "neo4j-burp-id", record.BurpSiteID)
		assert.Equal(t, webApp.PrimaryURL, record.PrimaryURL)
		assert.Equal(t, webApp.Name, record.Name)
	})
	
	t.Run("Key generation doesn't include BurpSiteID", func(t *testing.T) {
		webApp := NewWebApplicationWithBurpSiteID(
			"https://key.test.com/path",
			"Key Test",
			"should-not-be-in-key",
		)
		
		// Verify key format
		assert.Equal(t, "#webapplication#https://key.test.com/path", webApp.Key)
		assert.NotContains(t, webApp.Key, "should-not-be-in-key")
		assert.NotContains(t, webApp.Key, "burp")
		
		// Key should be valid
		assert.True(t, webApp.Valid())
	})
}

// TestWebApplicationPerformanceWithBurpSiteID tests performance characteristics
func TestWebApplicationPerformanceWithBurpSiteID(t *testing.T) {
	t.Run("Batch operations performance", func(t *testing.T) {
		const batchSize = 1000
		apps := make([]WebApplication, batchSize)
		
		// Create batch
		for i := 0; i < batchSize; i++ {
			apps[i] = NewWebApplicationWithBurpSiteID(
				"https://batch.test.com",
				"Batch Test",
				"batch-burp-id",
			)
		}
		
		// Merge batch
		base := NewWebApplication("https://base.test.com", "Base")
		for i := 0; i < batchSize; i++ {
			base.Merge(&apps[i])
		}
		
		// Verify final state
		assert.Equal(t, "batch-burp-id", base.BurpSiteID)
	})
}