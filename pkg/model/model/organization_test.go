package model

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganization_NewOrganization(t *testing.T) {
	org := NewOrganization("Walmart")

	assert.Equal(t, "Walmart", org.PrimaryName)
	assert.Equal(t, "organization", org.Class)
	assert.True(t, len(org.Names) >= 1)
	assert.Equal(t, "Walmart", org.Names[0].Name)
	assert.Equal(t, NameTypePrimary, org.Names[0].Type)
	assert.Equal(t, NameStatusActive, org.Names[0].Status)
	assert.NotEmpty(t, org.Names[0].DateAdded)
	assert.Contains(t, org.Key, "#organization#walmart#Walmart")
}

func TestOrganization_Valid(t *testing.T) {
	tests := []struct {
		name     string
		org      Organization
		expected bool
	}{
		{
			name:     "valid organization",
			org:      NewOrganization("Walmart"),
			expected: true,
		},
		{
			name: "missing primary name",
			org: Organization{
				Names: []OrganizationName{
					{Name: "Test", Type: NameTypePrimary, Status: NameStatusActive},
				},
			},
			expected: false,
		},
		{
			name: "invalid key format",
			org: func() Organization {
				org := NewOrganization("Walmart")
				org.BaseAsset.Key = "invalid-key"
				return org
			}(),
			expected: false,
		},
		{
			name: "missing primary name in names list",
			org: Organization{
				BaseAsset:   BaseAsset{Key: "#organization#walmart#Walmart"},
				PrimaryName: "Walmart",
				Names: []OrganizationName{
					{Name: "Walmart Inc", Type: NameTypeLegal, Status: NameStatusActive},
				},
			},
			expected: false,
		},
		{
			name: "invalid name in names list",
			org: Organization{
				BaseAsset:   BaseAsset{Key: "#organization#walmart#Walmart"},
				PrimaryName: "Walmart",
				Names: []OrganizationName{
					{Name: "Walmart", Type: NameTypePrimary, Status: NameStatusActive},
					{Name: "", Type: NameTypeLegal, Status: NameStatusActive}, // Invalid: empty name
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.org.Valid())
		})
	}
}

func TestOrganizationName_Valid(t *testing.T) {
	tests := []struct {
		name     string
		orgName  OrganizationName
		expected bool
	}{
		{
			name: "valid primary name",
			orgName: OrganizationName{
				Name:   "Walmart",
				Type:   NameTypePrimary,
				Status: NameStatusActive,
			},
			expected: true,
		},
		{
			name: "valid legal name",
			orgName: OrganizationName{
				Name:   "Walmart Inc",
				Type:   NameTypeLegal,
				Status: NameStatusActive,
			},
			expected: true,
		},
		{
			name: "empty name",
			orgName: OrganizationName{
				Name:   "",
				Type:   NameTypePrimary,
				Status: NameStatusActive,
			},
			expected: false,
		},
		{
			name: "invalid type",
			orgName: OrganizationName{
				Name:   "Walmart",
				Type:   "invalid",
				Status: NameStatusActive,
			},
			expected: false,
		},
		{
			name: "invalid status",
			orgName: OrganizationName{
				Name:   "Walmart",
				Type:   NameTypePrimary,
				Status: "invalid",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.orgName.Valid())
		})
	}
}

func TestOrganization_AddName(t *testing.T) {
	org := NewOrganization("Walmart")

	// Test adding valid name
	err := org.AddName("Walmart Inc", NameTypeLegal, "manual")
	assert.NoError(t, err)
	assert.Len(t, org.Names, 2)

	// Verify the added name
	found := false
	for _, name := range org.Names {
		if name.Name == "Walmart Inc" && name.Type == NameTypeLegal {
			assert.Equal(t, NameStatusActive, name.Status)
			assert.Equal(t, "manual", name.Source)
			assert.NotEmpty(t, name.DateAdded)
			found = true
			break
		}
	}
	assert.True(t, found, "Added name not found")

	// Test adding duplicate name
	err = org.AddName("Walmart Inc", NameTypeLegal, "manual")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name already exists")

	// Test adding empty name
	err = org.AddName("", NameTypeDBA, "manual")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name cannot be empty")

	// Test adding invalid type
	err = org.AddName("Test Name", "invalid_type", "manual")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid name type")
}

func TestOrganization_GetActiveNames(t *testing.T) {
	org := NewOrganization("Walmart")
	org.AddName("Walmart Inc", NameTypeLegal, "manual")
	org.AddName("WMT", NameTypeAbbreviation, "manual")

	// Add an inactive name
	org.Names = append(org.Names, OrganizationName{
		Name:      "Old Name",
		Type:      NameTypeFormer,
		Status:    NameStatusHistoric,
		DateAdded: Now(),
	})

	activeNames := org.GetActiveNames()
	assert.Len(t, activeNames, 3) // Should only include active names
	assert.Contains(t, activeNames, "Walmart")
	assert.Contains(t, activeNames, "Walmart Inc")
	assert.Contains(t, activeNames, "WMT")
	assert.NotContains(t, activeNames, "Old Name")
}

func TestOrganization_GetNamesByType(t *testing.T) {
	org := NewOrganization("Walmart")
	org.AddName("Walmart Inc", NameTypeLegal, "manual")
	org.AddName("Walmart Corporation", NameTypeLegal, "manual")
	org.AddName("WMT", NameTypeAbbreviation, "manual")

	legalNames := org.GetNamesByType(NameTypeLegal)
	assert.Len(t, legalNames, 2)
	assert.Contains(t, legalNames, "Walmart Inc")
	assert.Contains(t, legalNames, "Walmart Corporation")

	abbrevNames := org.GetNamesByType(NameTypeAbbreviation)
	assert.Len(t, abbrevNames, 1)
	assert.Contains(t, abbrevNames, "WMT")

	primaryNames := org.GetNamesByType(NameTypePrimary)
	assert.Len(t, primaryNames, 1)
	assert.Contains(t, primaryNames, "Walmart")
}

func TestNormalizeOrganizationName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Walmart", "walmart"},
		{"Walmart Inc", "walmart"},
		{"Walmart Incorporated", "walmart"},
		{"Walmart Corp", "walmart"},
		{"Walmart Corporation", "walmart"},
		{"Walmart LLC", "walmart"},
		{"Walmart Ltd", "walmart"},
		{"Walmart Limited", "walmart"},
		{"Walmart Co", "walmart"},
		{"Walmart Company", "walmart"},
		{"Praetorian Security Inc", "praetoriansecurity"},
		{"Praetorian-Security Inc.", "praetoriansecurity"},
		{"WALMART INC.", "walmart"},
		{"  Walmart  Inc  ", "walmart"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeOrganizationName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOrganizationSearchExpansion(t *testing.T) {
	ose := NewOrganizationSearchExpansion()

	// Create test organizations
	walmart := NewOrganization("Walmart")
	walmart.AddName("Walmart Inc", NameTypeLegal, "manual")
	walmart.AddName("WMT", NameTypeAbbreviation, "manual")

	praetorian := NewOrganization("Praetorian")
	praetorian.AddName("Praetorian Inc", NameTypeLegal, "manual")
	praetorian.AddName("Praetorian Security Inc", NameTypeLegal, "manual")
	praetorian.AddName("Praetorian Security", NameTypeDBA, "manual")

	// Add to search expansion
	ose.AddOrganization(&walmart)
	ose.AddOrganization(&praetorian)

	// Test search expansion for Walmart
	walmartExpansions := ose.ExpandSearch("Walmart")
	assert.Contains(t, walmartExpansions, "Walmart")
	assert.Contains(t, walmartExpansions, "Walmart Inc")
	assert.Contains(t, walmartExpansions, "WMT")

	// Test search expansion for Praetorian
	praetorianExpansions := ose.ExpandSearch("Praetorian")
	assert.Contains(t, praetorianExpansions, "Praetorian")
	assert.Contains(t, praetorianExpansions, "Praetorian Inc")
	assert.Contains(t, praetorianExpansions, "Praetorian Security Inc")
	assert.Contains(t, praetorianExpansions, "Praetorian Security")

	// Test search by alternative name
	walmartExpansionsByLegal := ose.ExpandSearch("Walmart Inc")
	assert.Equal(t, walmartExpansions, walmartExpansionsByLegal)

	// Test search by abbreviation
	walmartExpansionsByAbbrev := ose.ExpandSearch("WMT")
	assert.Equal(t, walmartExpansions, walmartExpansionsByAbbrev)

	// Test unknown organization
	unknownExpansions := ose.ExpandSearch("Unknown Org")
	assert.Len(t, unknownExpansions, 1)
	assert.Equal(t, "Unknown Org", unknownExpansions[0])
}

func TestOrganizationSearchExpansion_FindOrganization(t *testing.T) {
	ose := NewOrganizationSearchExpansion()

	walmart := NewOrganization("Walmart")
	walmart.AddName("Walmart Inc", NameTypeLegal, "manual")

	ose.AddOrganization(&walmart)

	// Test finding by primary name
	found := ose.FindOrganization("Walmart")
	assert.NotNil(t, found)
	assert.Equal(t, "Walmart", found.PrimaryName)

	// Test finding by legal name
	found = ose.FindOrganization("Walmart Inc")
	assert.NotNil(t, found)
	assert.Equal(t, "Walmart", found.PrimaryName)

	// Test finding unknown organization
	found = ose.FindOrganization("Unknown")
	assert.Nil(t, found)
}

func TestOrganization_IsClass(t *testing.T) {
	org := NewOrganization("Walmart")
	assert.True(t, org.IsClass("organization"))
	assert.False(t, org.IsClass("asset"))
}

func TestOrganization_Unmarshall(t *testing.T) {
	tests := []struct {
		name  string
		data  string
		valid bool
	}{
		{
			name:  "valid organization - primary name",
			data:  `{"type": "organization", "primaryName": "Walmart"}`,
			valid: true,
		},
		{
			name:  "valid organization - with names",
			data:  `{"type": "organization", "primaryName": "Walmart", "names": [{"name": "Walmart", "type": "primary", "status": "active"}, {"name": "Walmart Inc", "type": "legal", "status": "active"}]}`,
			valid: true,
		},
		{
			name:  "invalid organization - missing primary name",
			data:  `{"type": "organization"}`,
			valid: false,
		},
		{
			name:  "invalid organization - empty primary name",
			data:  `{"type": "organization", "primaryName": ""}`,
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a registry.Wrapper[Assetlike]
			err := json.Unmarshal([]byte(tt.data), &a)
			require.NoError(t, err)

			registry.CallHooks(a.Model)
			assert.Equal(t, tt.valid, a.Model.Valid())
		})
	}
}

func TestOrganization_JSONSerialization(t *testing.T) {
	org := NewOrganization("Walmart")
	org.AddName("Walmart Inc", NameTypeLegal, "manual")
	org.Industry = "Retail"
	org.Country = "United States"
	org.StockTicker = "WMT"
	org.Website = "https://www.walmart.com"
	org.Description = "Multinational retail corporation"

	// Marshal to JSON
	data, err := json.Marshal(org)
	require.NoError(t, err)

	// Unmarshal back
	var unmarshaled Organization
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify key fields
	assert.Equal(t, org.PrimaryName, unmarshaled.PrimaryName)
	assert.Equal(t, org.Industry, unmarshaled.Industry)
	assert.Equal(t, org.Country, unmarshaled.Country)
	assert.Equal(t, org.StockTicker, unmarshaled.StockTicker)
	assert.Equal(t, org.Website, unmarshaled.Website)
	assert.Equal(t, org.Description, unmarshaled.Description)
	assert.Len(t, unmarshaled.Names, len(org.Names))
}

func TestOrganization_ExampleUseCases(t *testing.T) {
	// Example from Jira story
	t.Run("Praetorian example", func(t *testing.T) {
		org := NewOrganization("Praetorian")
		org.AddName("Praetorian Inc", NameTypeLegal, "legal_docs")
		org.AddName("Praetorian Security Inc", NameTypeDBA, "github")
		org.AddName("Praetorian Security", NameTypeCommon, "linkedin")
		org.AddName("Praetorian Labs", NameTypeDBA, "dockerhub")

		ose := NewOrganizationSearchExpansion()
		ose.AddOrganization(&org)

		// Test search expansion
		expansions := ose.ExpandSearch("Praetorian")

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
	})
}

// Benchmark tests for performance requirements
func BenchmarkOrganizationSearchExpansion_AddOrganization(b *testing.B) {
	ose := NewOrganizationSearchExpansion()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		org := NewOrganization(fmt.Sprintf("TestOrg%d", i))
		org.AddName(fmt.Sprintf("TestOrg%d Inc", i), NameTypeLegal, "test")
		ose.AddOrganization(&org)
	}
}

func BenchmarkOrganizationSearchExpansion_ExpandSearch(b *testing.B) {
	ose := NewOrganizationSearchExpansion()

	// Setup 1000 organizations as per performance requirement
	for i := 0; i < 1000; i++ {
		org := NewOrganization(fmt.Sprintf("TestOrg%d", i))
		org.AddName(fmt.Sprintf("TestOrg%d Inc", i), NameTypeLegal, "test")
		ose.AddOrganization(&org)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ose.ExpandSearch(fmt.Sprintf("TestOrg%d", i%1000))
	}
}
