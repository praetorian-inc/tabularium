package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrganization_SubsidiaryMethods(t *testing.T) {
	parent := NewOrganization("Parent Corp")
	subsidiary := NewOrganization("Subsidiary Inc")

	subsidiary.SetParentOrganization(parent.GetKey(), 100.0, SubsidiaryTypeWhollyOwned)

	assert.Equal(t, parent.GetKey(), subsidiary.ParentOrganization)
	assert.Equal(t, 100.0, subsidiary.OwnershipPercentage)
	assert.Equal(t, SubsidiaryTypeWhollyOwned, subsidiary.SubsidiaryType)
	assert.True(t, subsidiary.IsSubsidiary())
	assert.True(t, subsidiary.IsWhollyOwned())
}

func TestOrganization_SubsidiaryValidation(t *testing.T) {
	parent := NewOrganization("Parent Corp")
	subsidiary := NewOrganization("Subsidiary Inc")

	tests := []struct {
		name                string
		ownershipPercentage float64
		subsidiaryType      string
		expectedWhollyOwned bool
		expectedSubsidiary  bool
	}{
		{
			name:                "wholly owned subsidiary",
			ownershipPercentage: 100.0,
			subsidiaryType:      SubsidiaryTypeWhollyOwned,
			expectedWhollyOwned: true,
			expectedSubsidiary:  true,
		},
		{
			name:                "majority owned subsidiary",
			ownershipPercentage: 75.0,
			subsidiaryType:      SubsidiaryTypeMajorityOwned,
			expectedWhollyOwned: false,
			expectedSubsidiary:  true,
		},
		{
			name:                "minority owned subsidiary",
			ownershipPercentage: 25.0,
			subsidiaryType:      SubsidiaryTypeMinorityOwned,
			expectedWhollyOwned: false,
			expectedSubsidiary:  true,
		},
		{
			name:                "joint venture",
			ownershipPercentage: 50.0,
			subsidiaryType:      SubsidiaryTypeJointVenture,
			expectedWhollyOwned: false,
			expectedSubsidiary:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subsidiary.SetParentOrganization(parent.GetKey(), tt.ownershipPercentage, tt.subsidiaryType)
			assert.Equal(t, tt.expectedSubsidiary, subsidiary.IsSubsidiary())
			assert.Equal(t, tt.expectedWhollyOwned, subsidiary.IsWhollyOwned())
			assert.Equal(t, tt.ownershipPercentage, subsidiary.OwnershipPercentage)
			assert.Equal(t, tt.subsidiaryType, subsidiary.SubsidiaryType)
		})
	}
}

func TestOrganization_HistoricalMethods(t *testing.T) {
	org := NewOrganization("Test Corp")

	// Test merger history
	mergedOrg1 := NewOrganization("Merged Corp 1")
	mergedOrg2 := NewOrganization("Merged Corp 2")

	org.AddMergedOrganization(mergedOrg1.GetKey())
	org.AddMergedOrganization(mergedOrg2.GetKey())
	org.AddMergedOrganization(mergedOrg1.GetKey())

	assert.True(t, org.HasMergerHistory())
	assert.Len(t, org.MergedOrganizations, 2)
	assert.Contains(t, org.MergedOrganizations, mergedOrg1.GetKey())
	assert.Contains(t, org.MergedOrganizations, mergedOrg2.GetKey())
}

func TestOrganizationRelationshipService_SubsidiaryOperations(t *testing.T) {
	service := NewOrganizationRelationshipService()

	parent := NewOrganization("Parent Corp")
	subsidiary1 := NewOrganization("Subsidiary One")
	subsidiary2 := NewOrganization("Subsidiary Two")

	service.AddOrganization(&parent)
	service.AddOrganization(&subsidiary1)
	service.AddOrganization(&subsidiary2)

	subsidiary1.SetParentOrganization(parent.GetKey(), 100.0, SubsidiaryTypeWhollyOwned)
	subsidiary2.SetParentOrganization(parent.GetKey(), 75.0, SubsidiaryTypeMajorityOwned)

	subsidiaries := service.GetSubsidiaries(parent.GetKey())
	assert.Len(t, subsidiaries, 2)

	subNames := make([]string, len(subsidiaries))
	for i, sub := range subsidiaries {
		subNames[i] = sub.PrimaryName
	}
	assert.Contains(t, subNames, "Subsidiary One")
	assert.Contains(t, subNames, "Subsidiary Two")

	parents1 := service.GetParentOrganizations(subsidiary1.GetKey())
	assert.Len(t, parents1, 1)
	assert.Equal(t, "Parent Corp", parents1[0].PrimaryName)

	parents2 := service.GetParentOrganizations(subsidiary2.GetKey())
	assert.Len(t, parents2, 1)
	assert.Equal(t, "Parent Corp", parents2[0].PrimaryName)
}

func TestOrganizationRelationshipService_NameHistory(t *testing.T) {
	service := NewOrganizationRelationshipService()

	org := NewOrganization("Current Name")

	service.AddOrganization(&org)
}

func TestOrganizationRelationshipService_OrganizationFamily(t *testing.T) {
	service := NewOrganizationRelationshipService()

	grandparent := NewOrganization("Grandparent Corp")
	parent1 := NewOrganization("Parent One")
	parent2 := NewOrganization("Parent Two")
	child1 := NewOrganization("Child One")
	child2 := NewOrganization("Child Two")
	sibling := NewOrganization("Sibling Corp") // Not connected to family

	orgs := []*Organization{&grandparent, &parent1, &parent2, &child1, &child2, &sibling}
	for _, org := range orgs {
		service.AddOrganization(org)
	}

	parent1.SetParentOrganization(grandparent.GetKey(), 100.0, SubsidiaryTypeWhollyOwned)
	parent2.SetParentOrganization(grandparent.GetKey(), 100.0, SubsidiaryTypeWhollyOwned)
	child1.SetParentOrganization(parent1.GetKey(), 100.0, SubsidiaryTypeWhollyOwned)
	child2.SetParentOrganization(parent1.GetKey(), 100.0, SubsidiaryTypeWhollyOwned)

	// Test getting organization family starting from grandparent
	family := service.GetOrganizationFamily(grandparent.GetKey())
	assert.Len(t, family, 5) // Should include grandparent, parent1, parent2, child1, child2

	familyNames := make([]string, len(family))
	for i, org := range family {
		familyNames[i] = org.PrimaryName
	}

	assert.Contains(t, familyNames, "Grandparent Corp")
	assert.Contains(t, familyNames, "Parent One")
	assert.Contains(t, familyNames, "Parent Two")
	assert.Contains(t, familyNames, "Child One")
	assert.Contains(t, familyNames, "Child Two")
	assert.NotContains(t, familyNames, "Sibling Corp")

	// Test getting family from a child perspective
	childFamily := service.GetOrganizationFamily(child1.GetKey())
	assert.Len(t, childFamily, 5) // Should be the same family

	// Test getting family for unconnected organization
	siblingFamily := service.GetOrganizationFamily(sibling.GetKey())
	assert.Len(t, siblingFamily, 1) // Should only include itself
	assert.Equal(t, "Sibling Corp", siblingFamily[0].PrimaryName)
}

func TestOrganizationRelationshipService_ComplexScenario(t *testing.T) {
	// Test a complex real-world-like scenario
	service := NewOrganizationRelationshipService()

	// Create organizations representing a corporate structure like Walmart
	walmart := NewOrganization("Walmart")
	walmartInc := NewOrganization("Walmart Inc")
	walmartStores := NewOrganization("Walmart Stores Inc")
	samsClub := NewOrganization("Sam's Club")
	walmartEcommerce := NewOrganization("Walmart eCommerce")

	// Add all organizations
	orgs := []*Organization{&walmart, &walmartInc, &walmartStores, &samsClub, &walmartEcommerce}
	for _, org := range orgs {
		service.AddOrganization(org)
	}

	// Create parent-subsidiary relationships using simplified approach
	walmartStores.SetParentOrganization(walmartInc.GetKey(), 100.0, SubsidiaryTypeWhollyOwned)
	samsClub.SetParentOrganization(walmartInc.GetKey(), 100.0, SubsidiaryTypeWhollyOwned)
	walmartEcommerce.SetParentOrganization(walmartInc.GetKey(), 100.0, SubsidiaryTypeWhollyOwned)

	// Create DISCOVERED relationships to connect them

	// Add name history using simplified approach
	// Note: Historical name tracking now uses OrganizationName relationships

	// Add merger history using simplified approach
	walmartInc.AddMergedOrganization(walmartEcommerce.GetKey())
	walmartInc.LastAcquisitionDate = "2016-08-08T00:00:00Z"

	// Test the complete corporate structure
	subsidiaries := service.GetSubsidiaries(walmartInc.GetKey())
	assert.Len(t, subsidiaries, 3)

	// Note: Name history functionality removed - use OrganizationName relationships

	// Test that the org has the expected historical information
	// Note: Historical dates now tracked in OrganizationName EffectiveDate/EndDate
	assert.Contains(t, walmartInc.MergedOrganizations, walmartEcommerce.GetKey())
	assert.Equal(t, "2016-08-08T00:00:00Z", walmartInc.LastAcquisitionDate)

	family := service.GetOrganizationFamily(walmartInc.GetKey())
	assert.Len(t, family, 4) // Walmart Inc + 3 subsidiaries
}

// Benchmark for performance testing with 1000+ organizations as required
func BenchmarkOrganizationRelationshipService_GetSubsidiaries(b *testing.B) {
	service := NewOrganizationRelationshipService()

	// Create 1000 parent-subsidiary relationships
	parent := NewOrganization("MegaCorp")
	service.AddOrganization(&parent)

	for i := 0; i < 1000; i++ {
		subsidiary := NewOrganization(fmt.Sprintf("Subsidiary%d", i))
		subsidiary.SetParentOrganization(parent.GetKey(), 100.0, SubsidiaryTypeWhollyOwned)
		service.AddOrganization(&subsidiary)

	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		subsidiaries := service.GetSubsidiaries(parent.GetKey())
		if len(subsidiaries) != 1000 {
			b.Fatalf("Expected 1000 subsidiaries, got %d", len(subsidiaries))
		}
	}
}

func BenchmarkOrganizationRelationshipService_GetOrganizationFamily(b *testing.B) {
	service := NewOrganizationRelationshipService()

	// Create a deep hierarchy: 1 grandparent -> 10 parents -> 100 children each
	grandparent := NewOrganization("GrandparentCorp")
	service.AddOrganization(&grandparent)

	for i := 0; i < 10; i++ {
		parent := NewOrganization(fmt.Sprintf("Parent%d", i))
		parent.SetParentOrganization(grandparent.GetKey(), 100.0, SubsidiaryTypeWhollyOwned)
		service.AddOrganization(&parent)

		for j := 0; j < 100; j++ {
			child := NewOrganization(fmt.Sprintf("Child%d_%d", i, j))
			child.SetParentOrganization(parent.GetKey(), 100.0, SubsidiaryTypeWhollyOwned)
			service.AddOrganization(&child)

		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		family := service.GetOrganizationFamily(grandparent.GetKey())
		if len(family) != 1011 { // 1 grandparent + 10 parents + 1000 children
			b.Fatalf("Expected 1011 family members, got %d", len(family))
		}
	}
}
