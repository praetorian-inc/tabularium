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
	assert.Equal(t, 0, len(org.Names))
	assert.Contains(t, org.Key, "#organization#walmart#Walmart")

	primaryName, relationship := org.CreatePrimaryNameRelationship()
	assert.Equal(t, "Walmart", primaryName.Name)
	assert.Equal(t, NameTypePrimary, primaryName.Type)
	assert.Equal(t, NameStateActive, primaryName.State)
	assert.NotEmpty(t, primaryName.DateAdded)
	assert.NotNil(t, relationship)
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
				Name:  "Walmart",
				Type:  NameTypePrimary,
				State: NameStateActive,
			},
			expected: true,
		},
		{
			name: "valid legal name",
			orgName: OrganizationName{
				Name:  "Walmart Inc",
				Type:  NameTypeLegal,
				State: NameStateActive,
			},
			expected: true,
		},
		{
			name: "empty name",
			orgName: OrganizationName{
				Name:  "",
				Type:  NameTypePrimary,
				State: NameStateActive,
			},
			expected: false,
		},
		{
			name: "invalid type",
			orgName: OrganizationName{
				Name:  "Walmart",
				Type:  "invalid",
				State: NameStateActive,
			},
			expected: false,
		},
		{
			name: "invalid state",
			orgName: OrganizationName{
				Name:  "Walmart",
				Type:  NameTypePrimary,
				State: "invalid",
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
			data:  `{"type": "organization", "primaryName": "Walmart", "names": [{"name": "Walmart", "type": "primary", "state": "active"}, {"name": "Walmart Inc", "type": "legal", "state": "active"}]}`,
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

// Benchmark tests for performance requirements
func BenchmarkOrganizationSearchExpansion_AddOrganization(b *testing.B) {
	ose := NewOrganizationSearchExpansion()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		org := NewOrganization(fmt.Sprintf("TestOrg%d", i))
		ose.AddOrganization(&org)
	}
}

func BenchmarkOrganizationSearchExpansion_ExpandSearch(b *testing.B) {
	ose := NewOrganizationSearchExpansion()

	// Setup 1000 organizations as per performance requirement
	for i := 0; i < 1000; i++ {
		org := NewOrganization(fmt.Sprintf("TestOrg%d", i))
		ose.AddOrganization(&org)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ose.ExpandSearch(fmt.Sprintf("TestOrg%d", i%1000))
	}
}

func TestOrganizationName_GetKey(t *testing.T) {
	tests := []struct {
		name        string
		orgName     OrganizationName
		expectedKey string
	}{
		{
			name: "primary name with special characters",
			orgName: OrganizationName{
				Name: "Walmart Inc.",
				Type: NameTypePrimary,
			},
			expectedKey: "#organizationname#walmartinc#primary",
		},
		{
			name: "legal name with spaces and symbols",
			orgName: OrganizationName{
				Name: "Praetorian Security, Inc.",
				Type: NameTypeLegal,
			},
			expectedKey: "#organizationname#praetoriansecurityinc#legal",
		},
		{
			name: "dba name with numbers and hyphens",
			orgName: OrganizationName{
				Name: "Test-Corp 123",
				Type: NameTypeDBA,
			},
			expectedKey: "#organizationname#testcorp123#dba",
		},
		{
			name: "abbreviation with uppercase",
			orgName: OrganizationName{
				Name: "IBM",
				Type: NameTypeAbbreviation,
			},
			expectedKey: "#organizationname#ibm#abbreviation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := tt.orgName.GetKey()
			assert.Equal(t, tt.expectedKey, key)
			assert.NotEmpty(t, key, "GetKey should not return empty string")
		})
	}
}
