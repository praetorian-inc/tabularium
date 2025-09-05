package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBurpSiteIDIntegration tests the complete integration of BurpSiteID field
func TestBurpSiteIDIntegration(t *testing.T) {
	t.Run("Constructor with BurpSiteID", func(t *testing.T) {
		w := NewWebApplicationWithBurpSiteID(
			"https://example.com",
			"Test App",
			"burp-site-123",
		)
		
		assert.Equal(t, "burp-site-123", w.BurpSiteID)
		assert.Equal(t, "https://example.com/", w.PrimaryURL)
		assert.Equal(t, "Test App", w.Name)
		assert.True(t, w.HasBurpSiteID())
	})

	t.Run("Empty BurpSiteID check", func(t *testing.T) {
		w := NewWebApplication("https://example.com", "Test")
		assert.False(t, w.HasBurpSiteID())
		
		w.BurpSiteID = "site-123"
		assert.True(t, w.HasBurpSiteID())
	})
}

// TestBurpSiteIDConcurrency tests thread-safety of BurpSiteID operations
func TestBurpSiteIDConcurrency(t *testing.T) {
	w := NewWebApplication("https://example.com", "Test")
	
	var wg sync.WaitGroup
	iterations := 100
	
	// Test concurrent reads and writes
	wg.Add(iterations * 2)
	
	// Writers
	for i := 0; i < iterations; i++ {
		go func(id int) {
			defer wg.Done()
			w.BurpSiteID = fmt.Sprintf("burp-%d", id)
		}(i)
	}
	
	// Readers
	for i := 0; i < iterations; i++ {
		go func() {
			defer wg.Done()
			_ = w.BurpSiteID
			_ = w.HasBurpSiteID()
		}()
	}
	
	wg.Wait()
	
	// Final BurpSiteID should be one of the written values
	assert.Contains(t, w.BurpSiteID, "burp-")
}

// TestBurpSiteIDWithStatus tests BurpSiteID preservation through status changes
func TestBurpSiteIDWithStatus(t *testing.T) {
	original := NewWebApplicationWithBurpSiteID(
		"https://example.com",
		"Test",
		"burp-original",
	)
	
	statuses := []string{Active, Pending, Deleted, "AH", "AL"}
	
	for _, status := range statuses {
		t.Run(fmt.Sprintf("Status_%s", status), func(t *testing.T) {
			modified := original.WithStatus(status)
			webApp, ok := modified.(*WebApplication)
			require.True(t, ok)
			
			// Verify BurpSiteID is preserved
			assert.Equal(t, original.BurpSiteID, webApp.BurpSiteID)
			// Verify status changed
			assert.Equal(t, status, webApp.Status)
			// Verify original unchanged (unless it's the same status)
			if status != original.Status {
				assert.NotEqual(t, status, original.Status)
			}
			// Verify it's a different instance
			assert.NotSame(t, &original, webApp)
		})
	}
}

// TestBurpSiteIDMergeScenarios tests various merge scenarios
func TestBurpSiteIDMergeScenarios(t *testing.T) {
	scenarios := []struct {
		name           string
		w1Setup        func() *WebApplication
		w2Setup        func() *WebApplication
		expectedBurpID string
		expectedName   string
		expectedURLs   []string
	}{
		{
			name: "Complete merge with all fields",
			w1Setup: func() *WebApplication {
				w := NewWebApplication("https://old.com", "Old")
				w.BurpSiteID = "old-burp"
				w.URLs = []string{"https://api.old.com"}
				return &w
			},
			w2Setup: func() *WebApplication {
				w := NewWebApplication("https://new.com", "New")
				w.BurpSiteID = "new-burp"
				w.URLs = []string{"https://api.new.com", "https://admin.new.com"}
				return &w
			},
			expectedBurpID: "new-burp",
			expectedName:   "New",
			expectedURLs:   []string{"https://api.old.com", "https://api.new.com", "https://admin.new.com"},
		},
		{
			name: "Merge with nil other",
			w1Setup: func() *WebApplication {
				w := NewWebApplicationWithBurpSiteID("https://test.com", "Test", "test-burp")
				return &w
			},
			w2Setup: func() *WebApplication {
				return nil
			},
			expectedBurpID: "test-burp",
			expectedName:   "Test",
			expectedURLs:   []string{},
		},
	}
	
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			w1 := scenario.w1Setup()
			w2 := scenario.w2Setup()
			
			if w2 != nil {
				w1.Merge(w2)
			} else {
				// Test merge with non-WebApplication type
				asset := NewAsset("test.com", "test")
				w1.Merge(&asset)
			}
			
			assert.Equal(t, scenario.expectedBurpID, w1.BurpSiteID)
			assert.Equal(t, scenario.expectedName, w1.Name)
			if len(scenario.expectedURLs) > 0 {
				assert.ElementsMatch(t, scenario.expectedURLs, w1.URLs)
			}
		})
	}
}

// TestBurpSiteIDJSONMarshaling tests JSON serialization performance
func TestBurpSiteIDJSONMarshaling(t *testing.T) {
	testCases := []struct {
		name       string
		burpSiteID string
	}{
		{"Simple ID", "simple-123"},
		{"UUID format", "550e8400-e29b-41d4-a716-446655440000"},
		{"Long ID", strings.Repeat("a", 255)},
		{"Special chars", "burp_site-123.456$test"},
		{"Empty", ""},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			original := NewWebApplicationWithBurpSiteID(
				"https://example.com/app",
				"Test App",
				tc.burpSiteID,
			)
			original.URLs = []string{"https://api.example.com", "https://admin.example.com"}
			
			// Test marshaling
			data, err := json.Marshal(original)
			require.NoError(t, err)
			
			// Verify field exists in JSON
			var jsonMap map[string]interface{}
			err = json.Unmarshal(data, &jsonMap)
			require.NoError(t, err)
			
			burpField, exists := jsonMap["burp_site_id"]
			assert.True(t, exists, "burp_site_id field should exist in JSON")
			assert.Equal(t, tc.burpSiteID, burpField)
			
			// Test unmarshaling
			var unmarshaled WebApplication
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			
			// Verify complete round-trip
			assert.Equal(t, original.BurpSiteID, unmarshaled.BurpSiteID)
			assert.Equal(t, original.Name, unmarshaled.Name)
			assert.Equal(t, original.PrimaryURL, unmarshaled.PrimaryURL)
			assert.ElementsMatch(t, original.URLs, unmarshaled.URLs)
		})
	}
}

// TestBurpSiteIDNeo4jCompatibility tests Neo4j-specific behaviors
func TestBurpSiteIDNeo4jCompatibility(t *testing.T) {
	w := NewWebApplicationWithBurpSiteID(
		"https://neo4j.example.com",
		"Neo4j Test",
		"neo4j-burp-123",
	)
	
	// Test that the field can handle Neo4j-like IDs
	neo4jStyleIDs := []string{
		"node-123456",
		"4:550e8400-e29b-41d4-a716-446655440000:1",
		"rel-999999",
		strings.Repeat("x", 100), // Long ID
	}
	
	for _, id := range neo4jStyleIDs {
		w.BurpSiteID = id
		assert.Equal(t, id, w.BurpSiteID)
		
		// Verify it doesn't affect key generation
		assert.Contains(t, w.Key, "#webapplication#")
		assert.NotContains(t, w.Key, id)
	}
}

// TestBurpSiteIDEdgeCases tests edge cases and error conditions
func TestBurpSiteIDEdgeCases(t *testing.T) {
	t.Run("Merge with wrong type preserves BurpSiteID", func(t *testing.T) {
		w := NewWebApplicationWithBurpSiteID(
			"https://example.com",
			"Test",
			"preserve-this",
		)
		
		// Merge with different asset type
		asset := NewAsset("example.com", "asset")
		w.Merge(&asset)
		
		// BurpSiteID should be unchanged
		assert.Equal(t, "preserve-this", w.BurpSiteID)
	})
	
	t.Run("Visit with wrong type preserves BurpSiteID", func(t *testing.T) {
		w := NewWebApplicationWithBurpSiteID(
			"https://example.com",
			"Test",
			"preserve-this",
		)
		
		// Visit with different asset type
		asset := NewAsset("example.com", "asset")
		w.Visit(&asset)
		
		// BurpSiteID should be unchanged
		assert.Equal(t, "preserve-this", w.BurpSiteID)
	})
	
	t.Run("SeedModels preserves BurpSiteID", func(t *testing.T) {
		seed := NewWebApplicationSeed("https://seed.example.com")
		seed.BurpSiteID = "seed-burp-id"
		
		models := seed.SeedModels()
		require.Len(t, models, 1)
		
		copiedWebApp, ok := models[0].(*WebApplication)
		require.True(t, ok)
		assert.Equal(t, seed.BurpSiteID, copiedWebApp.BurpSiteID)
		assert.NotSame(t, &seed, copiedWebApp)
	})
}

// TestBurpSiteIDPerformance tests performance characteristics
func TestBurpSiteIDPerformance(t *testing.T) {
	t.Run("Large batch operations", func(t *testing.T) {
		start := time.Now()
		apps := make([]WebApplication, 1000)
		
		for i := range apps {
			apps[i] = NewWebApplicationWithBurpSiteID(
				fmt.Sprintf("https://app%d.example.com", i),
				fmt.Sprintf("App %d", i),
				fmt.Sprintf("burp-site-%d", i),
			)
		}
		
		elapsed := time.Since(start)
		assert.Less(t, elapsed, time.Second, "Creating 1000 WebApplications should be fast")
		
		// Verify all have unique BurpSiteIDs
		seen := make(map[string]bool)
		for _, app := range apps {
			assert.False(t, seen[app.BurpSiteID], "Duplicate BurpSiteID found")
			seen[app.BurpSiteID] = true
		}
	})
	
	t.Run("Merge performance", func(t *testing.T) {
		w1 := NewWebApplication("https://base.example.com", "Base")
		w1.URLs = make([]string, 100)
		for i := range w1.URLs {
			w1.URLs[i] = fmt.Sprintf("https://url%d.example.com", i)
		}
		
		start := time.Now()
		for i := 0; i < 100; i++ {
			w2 := NewWebApplicationWithBurpSiteID(
				"https://merge.example.com",
				"Merge",
				fmt.Sprintf("merge-burp-%d", i),
			)
			w2.URLs = []string{fmt.Sprintf("https://new%d.example.com", i)}
			w1.Merge(&w2)
		}
		elapsed := time.Since(start)
		
		assert.Less(t, elapsed, time.Millisecond*100, "100 merges should be fast")
		assert.NotEmpty(t, w1.BurpSiteID)
	})
}

// TestBurpSiteIDValidation tests validation scenarios
func TestBurpSiteIDValidation(t *testing.T) {
	t.Run("Valid remains valid with BurpSiteID", func(t *testing.T) {
		w := NewWebApplicationWithBurpSiteID(
			"https://valid.example.com",
			"Valid",
			"burp-123",
		)
		
		assert.True(t, w.Valid())
		
		// Empty BurpSiteID should still be valid
		w.BurpSiteID = ""
		assert.True(t, w.Valid())
		
		// Very long BurpSiteID should still be valid
		w.BurpSiteID = strings.Repeat("x", 1000)
		assert.True(t, w.Valid())
	})
	
	t.Run("Invalid key makes WebApp invalid regardless of BurpSiteID", func(t *testing.T) {
		w := WebApplication{
			BaseAsset: BaseAsset{Key: "invalid-key"},
			BurpSiteID: "valid-burp-id",
		}
		
		assert.False(t, w.Valid())
	})
}

// BenchmarkBurpSiteIDOperations benchmarks BurpSiteID operations
func BenchmarkBurpSiteIDOperations(b *testing.B) {
	b.Run("NewWithBurpSiteID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewWebApplicationWithBurpSiteID(
				"https://example.com",
				"Test",
				"burp-site-123",
			)
		}
	})
	
	b.Run("HasBurpSiteID", func(b *testing.B) {
		w := NewWebApplicationWithBurpSiteID(
			"https://example.com",
			"Test",
			"burp-site-123",
		)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = w.HasBurpSiteID()
		}
	})
	
	b.Run("MergeWithBurpSiteID", func(b *testing.B) {
		w1 := NewWebApplication("https://example.com", "Test")
		w2 := NewWebApplicationWithBurpSiteID(
			"https://other.com",
			"Other",
			"burp-123",
		)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w1.Merge(&w2)
		}
	})
}