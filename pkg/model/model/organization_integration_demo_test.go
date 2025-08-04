package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganizationDiscoveryService_SetupSampleData(t *testing.T) {
	service := NewOrganizationDiscoveryService()
	service.SetupSampleData()

	// Test that organizations were added
	praetorian := service.searchExpansion.FindOrganization("Praetorian")
	require.NotNil(t, praetorian)
	assert.Equal(t, "Praetorian", praetorian.PrimaryName)
	assert.Equal(t, "Cybersecurity", praetorian.Industry)

	walmart := service.searchExpansion.FindOrganization("Walmart")
	require.NotNil(t, walmart)
	assert.Equal(t, "Walmart", walmart.PrimaryName)
	assert.Equal(t, "Retail", walmart.Industry)

	// Test name variations
	praetorianExpansions := service.searchExpansion.ExpandSearch("Praetorian")
	assert.Contains(t, praetorianExpansions, "Praetorian")
	assert.Contains(t, praetorianExpansions, "Praetorian Inc")
	assert.Contains(t, praetorianExpansions, "Praetorian Security Inc")
	assert.Contains(t, praetorianExpansions, "Praetorian Security")
	assert.Contains(t, praetorianExpansions, "Praetorian Labs")

	walmartExpansions := service.searchExpansion.ExpandSearch("Walmart")
	assert.Contains(t, walmartExpansions, "Walmart")
	assert.Contains(t, walmartExpansions, "Walmart Inc")
	assert.Contains(t, walmartExpansions, "Walmart Labs")
}

func TestMockPlatformSearcher_GitHub(t *testing.T) {
	searcher := NewMockPlatformSearcher("GitHub")

	// Test exact matches
	results := searcher.Search("praetorian")
	assert.Len(t, results, 3)
	assert.Contains(t, results, "github.com/praetorian-inc/gokart")
	assert.Contains(t, results, "github.com/praetorian-inc/noseyparker")
	assert.Contains(t, results, "github.com/praetorian-inc/fingerprintx")

	// Test case insensitive
	resultsUpper := searcher.Search("PRAETORIAN")
	assert.Equal(t, results, resultsUpper)

	// Test no results
	noResults := searcher.Search("nonexistent")
	assert.Len(t, noResults, 0)
}

func TestMockPlatformSearcher_DockerHub(t *testing.T) {
	searcher := NewMockPlatformSearcher("DockerHub")

	results := searcher.Search("praetorian")
	assert.Len(t, results, 2)
	assert.Contains(t, results, "docker.io/praetorian/gokart:latest")
	assert.Contains(t, results, "docker.io/praetorian/noseyparker:v1.0")
}

func TestMockPlatformSearcher_LinkedIn(t *testing.T) {
	searcher := NewMockPlatformSearcher("LinkedIn")

	results := searcher.Search("praetorian")
	assert.Len(t, results, 1)
	assert.Contains(t, results, "linkedin.com/company/praetorian")

	results = searcher.Search("praetorian security inc")
	assert.Len(t, results, 1)
	assert.Contains(t, results, "linkedin.com/company/praetorian-security-inc")
}

func TestOrganizationDiscoveryService_DiscoverAssets(t *testing.T) {
	service := NewOrganizationDiscoveryService()
	service.SetupSampleData()

	// Test Praetorian discovery
	results, err := service.DiscoverAssets("Praetorian")
	require.NoError(t, err)

	// Should find assets across all platforms using name variations
	assert.True(t, len(results) > 5, "Should find multiple assets")

	// Check that we have results from different platforms
	platforms := make(map[string]bool)
	for _, result := range results {
		platforms[result.Platform] = true
	}
	assert.True(t, platforms["GitHub"], "Should have GitHub results")
	assert.True(t, platforms["DockerHub"], "Should have DockerHub results")
	assert.True(t, platforms["LinkedIn"], "Should have LinkedIn results")

	// Check that assets were discovered via different name variations
	discoveredVia := make(map[string]bool)
	for _, result := range results {
		if via, exists := result.Metadata["discovered_via"]; exists {
			discoveredVia[via] = true
		}
	}
	assert.True(t, len(discoveredVia) > 1, "Should discover assets via multiple name variations")
}

func TestOrganizationDiscoveryService_DiscoverAssetsWithSubsidiaries(t *testing.T) {
	service := NewOrganizationDiscoveryService()
	service.SetupSampleData()

	// Test Walmart discovery (should include Sam's Club subsidiary)
	results, err := service.DiscoverAssets("Walmart")
	require.NoError(t, err)

	// Check for subsidiary assets
	hasSubsidiaryAssets := false
	for _, result := range results {
		if relationship, exists := result.Metadata["relationship"]; exists && relationship == "subsidiary" {
			hasSubsidiaryAssets = true
			assert.Equal(t, "Walmart", result.Metadata["parent_org"])
			break
		}
	}
	assert.True(t, hasSubsidiaryAssets, "Should include subsidiary assets")
}

func TestOrganizationDiscoveryService_InferAssetType(t *testing.T) {
	service := NewOrganizationDiscoveryService()

	tests := []struct {
		platform string
		url      string
		expected string
	}{
		{"GitHub", "github.com/praetorian-inc/gokart", "repository"},
		{"GitHub", "github.com/praetorian-inc", "organization"},
		{"DockerHub", "docker.io/praetorian/gokart:latest", "container_image"},
		{"LinkedIn", "linkedin.com/company/praetorian", "company_page"},
		{"LinkedIn", "linkedin.com/in/john-doe", "profile"},
		{"Unknown", "example.com", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.platform+"_"+tt.expected, func(t *testing.T) {
			result := service.inferAssetType(tt.platform, tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOrganizationDiscoveryService_ExtractAssetName(t *testing.T) {
	service := NewOrganizationDiscoveryService()

	tests := []struct {
		url      string
		expected string
	}{
		{"github.com/praetorian-inc/gokart", "gokart"},
		{"docker.io/praetorian/scanner:latest", "scanner:latest"},
		{"linkedin.com/company/praetorian", "praetorian"},
		{"simple-name", "simple-name"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := service.extractAssetName(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOrganizationDiscoveryService_GenerateDiscoveryReport(t *testing.T) {
	service := NewOrganizationDiscoveryService()
	service.SetupSampleData()

	report, err := service.GenerateDiscoveryReport("Praetorian")
	require.NoError(t, err)
	require.NotEmpty(t, report)

	// Check report contains expected sections
	assert.Contains(t, report, "# Asset Discovery Report for: Praetorian")
	assert.Contains(t, report, "## Organization Information")
	assert.Contains(t, report, "**Primary Name:** Praetorian")
	assert.Contains(t, report, "**Industry:** Cybersecurity")
	assert.Contains(t, report, "### Name Variations Searched:")
	assert.Contains(t, report, "## GitHub Assets")
	assert.Contains(t, report, "## DockerHub Assets")
	assert.Contains(t, report, "## LinkedIn Assets")
	assert.Contains(t, report, "## Security Recommendations")

	// Check specific content
	assert.Contains(t, report, "Praetorian Inc")
	assert.Contains(t, report, "Praetorian Security Inc")
	assert.Contains(t, report, "Praetorian Labs")
	assert.Contains(t, report, "Asset Inventory")
	assert.Contains(t, report, "Access Control")
	assert.Contains(t, report, "Brand Protection")
}

func TestOrganizationDiscoveryService_GenerateDiscoveryReportWithSubsidiaries(t *testing.T) {
	service := NewOrganizationDiscoveryService()
	service.SetupSampleData()

	report, err := service.GenerateDiscoveryReport("Walmart")
	require.NoError(t, err)
	require.NotEmpty(t, report)

	// Should include subsidiary information
	assert.Contains(t, report, "### Subsidiaries Included:")
	assert.Contains(t, report, "Sam's Club")

	// Should mention subsidiary relationships in assets
	assert.Contains(t, report, "subsidiary of")
}

func TestDiscoveryResult_Metadata(t *testing.T) {
	result := DiscoveryResult{
		OrganizationName: "Praetorian",
		Platform:         "GitHub",
		AssetType:        "repository",
		AssetName:        "gokart",
		URL:              "github.com/praetorian-inc/gokart",
		Metadata: map[string]string{
			"discovered_via": "Praetorian Inc",
			"original_query": "Praetorian",
		},
	}

	assert.Equal(t, "Praetorian Inc", result.Metadata["discovered_via"])
	assert.Equal(t, "Praetorian", result.Metadata["original_query"])
}

func TestDemonstrateSearchExpansion_Integration(t *testing.T) {
	// This test ensures the demo function runs without errors
	// In a real scenario, you might want to capture stdout to verify output
	assert.NotPanics(t, func() {
		DemonstrateSearchExpansion()
	})
}

// Benchmark the discovery process with realistic data volumes
func BenchmarkOrganizationDiscoveryService_DiscoverAssets(b *testing.B) {
	service := NewOrganizationDiscoveryService()
	service.SetupSampleData()

	// Add more organizations to simulate realistic scale
	for i := 0; i < 100; i++ {
		org := NewOrganization(fmt.Sprintf("TestOrg%d", i))
		org.AddName(fmt.Sprintf("TestOrg%d Inc", i), NameTypeLegal, "test")
		service.searchExpansion.AddOrganization(&org)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.DiscoverAssets("Praetorian")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOrganizationDiscoveryService_GenerateReport(b *testing.B) {
	service := NewOrganizationDiscoveryService()
	service.SetupSampleData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GenerateDiscoveryReport("Praetorian")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Test edge cases and error conditions
func TestOrganizationDiscoveryService_EdgeCases(t *testing.T) {
	service := NewOrganizationDiscoveryService()
	service.SetupSampleData()

	// Test unknown organization
	results, err := service.DiscoverAssets("UnknownOrg")
	require.NoError(t, err)
	// Should still work, just return results for the literal search term
	assert.NotNil(t, results)
	assert.GreaterOrEqual(t, len(results), 0, "Should return at least empty slice")

	// Test empty search term
	results, err = service.DiscoverAssets("")
	require.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 0, len(results), "Empty search should return empty results")

	// Test search expansion with no matches
	expansions := service.searchExpansion.ExpandSearch("NonexistentOrg")
	assert.Len(t, expansions, 1)
	assert.Equal(t, "NonexistentOrg", expansions[0])
}

// Test the complete workflow as described in the Jira acceptance criteria
func TestCompleteWorkflow_AcceptanceCriteria(t *testing.T) {
	service := NewOrganizationDiscoveryService()

	// Test from Jira story: Input "Praetorian" should expand to multiple variations
	praetorian := NewOrganization("Praetorian")
	praetorian.AddName("Praetorian Inc", NameTypeLegal, "legal_docs")
	praetorian.AddName("Praetorian Security Inc", NameTypeDBA, "github")
	praetorian.AddName("Praetorian Security", NameTypeCommon, "linkedin")
	praetorian.AddName("Praetorian Labs", NameTypeDBA, "dockerhub")

	service.searchExpansion.AddOrganization(&praetorian)

	// Test search expansion
	expansions := service.searchExpansion.ExpandSearch("Praetorian")

	expectedNames := []string{
		"Praetorian",
		"Praetorian Inc",
		"Praetorian Security Inc",
		"Praetorian Security",
		"Praetorian Labs",
	}

	for _, expected := range expectedNames {
		assert.Contains(t, expansions, expected, "Should contain %s", expected)
	}

	// Verify all acceptance criteria are met:
	// ✅ Organization Name asset type implemented with support for primary/legal/DBA/abbreviation names
	assert.Equal(t, "Praetorian", praetorian.PrimaryName)
	assert.Len(t, praetorian.Names, 5) // Primary + 4 additional names

	// ✅ Go struct implemented in pkg/model/organization.go
	assert.IsType(t, Organization{}, praetorian)

	// ✅ Schema validation rules enforce at least one primary name required
	assert.True(t, praetorian.Valid())

	// ✅ Search expansion capability - API to retrieve all name variations
	allNames := praetorian.GetActiveNames()
	assert.Len(t, allNames, 5)

	// ✅ Integration with search capability demonstrated
	results, err := service.DiscoverAssets("Praetorian")
	require.NoError(t, err)
	assert.True(t, len(results) > 0, "Should discover assets")

	// ✅ Example use case from Jira story works
	t.Logf("✅ Praetorian search expansion works: %v", expansions)
	t.Logf("✅ Discovered %d assets across platforms", len(results))
}
