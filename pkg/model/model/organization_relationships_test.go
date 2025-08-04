package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrganizationParentSubsidiary_NewOrganizationParentSubsidiary(t *testing.T) {
	parent := NewOrganization("Parent Corp")
	subsidiary := NewOrganization("Subsidiary Inc")

	rel := NewOrganizationParentSubsidiary(&parent, &subsidiary, 100.0, RelationshipTypeWhollyOwned)

	assert.Equal(t, OrganizationParentSubsidiaryLabel, rel.Label())
	assert.Equal(t, 100.0, rel.OwnershipPercentage)
	assert.Equal(t, RelationshipTypeWhollyOwned, rel.RelationshipType)
	assert.NotEmpty(t, rel.EffectiveDate)
	assert.NotEmpty(t, rel.Key)
}

func TestOrganizationParentSubsidiary_Valid(t *testing.T) {
	parent := NewOrganization("Parent Corp")
	subsidiary := NewOrganization("Subsidiary Inc")

	tests := []struct {
		name                string
		ownershipPercentage float64
		relationshipType    string
		expected            bool
	}{
		{
			name:                "valid wholly owned",
			ownershipPercentage: 100.0,
			relationshipType:    RelationshipTypeWhollyOwned,
			expected:            true,
		},
		{
			name:                "valid majority owned",
			ownershipPercentage: 75.0,
			relationshipType:    RelationshipTypeMajorityOwned,
			expected:            true,
		},
		{
			name:                "valid minority owned",
			ownershipPercentage: 25.0,
			relationshipType:    RelationshipTypeMinorityOwned,
			expected:            true,
		},
		{
			name:                "invalid ownership percentage - negative",
			ownershipPercentage: -10.0,
			relationshipType:    RelationshipTypeWhollyOwned,
			expected:            false,
		},
		{
			name:                "invalid ownership percentage - over 100",
			ownershipPercentage: 150.0,
			relationshipType:    RelationshipTypeWhollyOwned,
			expected:            false,
		},
		{
			name:                "invalid relationship type",
			ownershipPercentage: 50.0,
			relationshipType:    "invalid_type",
			expected:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel := NewOrganizationParentSubsidiary(&parent, &subsidiary, tt.ownershipPercentage, tt.relationshipType)
			assert.Equal(t, tt.expected, rel.Valid())
		})
	}
}

func TestOrganizationParentSubsidiary_OwnershipMethods(t *testing.T) {
	parent := NewOrganization("Parent Corp")
	subsidiary := NewOrganization("Subsidiary Inc")

	// Test wholly owned
	whollyOwned := NewOrganizationParentSubsidiary(&parent, &subsidiary, 100.0, RelationshipTypeWhollyOwned)
	assert.True(t, whollyOwned.IsWhollyOwned())
	assert.False(t, whollyOwned.IsMajorityOwned())

	// Test majority owned
	majorityOwned := NewOrganizationParentSubsidiary(&parent, &subsidiary, 75.0, RelationshipTypeMajorityOwned)
	assert.False(t, majorityOwned.IsWhollyOwned())
	assert.True(t, majorityOwned.IsMajorityOwned())

	// Test minority owned
	minorityOwned := NewOrganizationParentSubsidiary(&parent, &subsidiary, 25.0, RelationshipTypeMinorityOwned)
	assert.False(t, minorityOwned.IsWhollyOwned())
	assert.False(t, minorityOwned.IsMajorityOwned())

	// Test edge case - 99% ownership should be considered wholly owned
	almostWhollyOwned := NewOrganizationParentSubsidiary(&parent, &subsidiary, 99.0, RelationshipTypeMajorityOwned)
	assert.True(t, almostWhollyOwned.IsWhollyOwned())
}

func TestOrganizationNameHistory_NewOrganizationNameHistory(t *testing.T) {
	org := NewOrganization("Current Name")

	rel := NewOrganizationNameHistory(&org, "Old Name", "Current Name", "2023-01-01T00:00:00Z")

	assert.Equal(t, OrganizationNameHistoryLabel, rel.Label())
	assert.Equal(t, "Old Name", rel.OldName)
	assert.Equal(t, "Current Name", rel.NewName)
	assert.Equal(t, "2023-01-01T00:00:00Z", rel.ChangeDate)
	assert.NotEmpty(t, rel.Key)
}

func TestOrganizationNameHistory_Valid(t *testing.T) {
	org := NewOrganization("Current Name")

	tests := []struct {
		name       string
		oldName    string
		newName    string
		changeDate string
		expected   bool
	}{
		{
			name:       "valid name history",
			oldName:    "Old Name",
			newName:    "New Name",
			changeDate: "2023-01-01T00:00:00Z",
			expected:   true,
		},
		{
			name:       "empty old name",
			oldName:    "",
			newName:    "New Name",
			changeDate: "2023-01-01T00:00:00Z",
			expected:   false,
		},
		{
			name:       "empty new name",
			oldName:    "Old Name",
			newName:    "",
			changeDate: "2023-01-01T00:00:00Z",
			expected:   false,
		},
		{
			name:       "empty change date",
			oldName:    "Old Name",
			newName:    "New Name",
			changeDate: "",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel := NewOrganizationNameHistory(&org, tt.oldName, tt.newName, tt.changeDate)
			assert.Equal(t, tt.expected, rel.Valid())
		})
	}
}

func TestOrganizationMerger_NewOrganizationMerger(t *testing.T) {
	acquirer := NewOrganization("Acquirer Corp")
	target := NewOrganization("Target Inc")

	rel := NewOrganizationMerger(&acquirer, &target, "2023-06-15T00:00:00Z", TransactionTypeAcquisition)

	assert.Equal(t, OrganizationMergerLabel, rel.Label())
	assert.Equal(t, "2023-06-15T00:00:00Z", rel.MergerDate)
	assert.Equal(t, TransactionTypeAcquisition, rel.TransactionType)
	assert.Equal(t, TransactionStatusPending, rel.Status)
	assert.NotEmpty(t, rel.Key)
}

func TestOrganizationMerger_Valid(t *testing.T) {
	acquirer := NewOrganization("Acquirer Corp")
	target := NewOrganization("Target Inc")

	tests := []struct {
		name            string
		mergerDate      string
		transactionType string
		status          string
		expected        bool
	}{
		{
			name:            "valid acquisition",
			mergerDate:      "2023-06-15T00:00:00Z",
			transactionType: TransactionTypeAcquisition,
			status:          TransactionStatusCompleted,
			expected:        true,
		},
		{
			name:            "valid merger",
			mergerDate:      "2023-06-15T00:00:00Z",
			transactionType: TransactionTypeMerger,
			status:          TransactionStatusPending,
			expected:        true,
		},
		{
			name:            "empty merger date",
			mergerDate:      "",
			transactionType: TransactionTypeAcquisition,
			status:          TransactionStatusCompleted,
			expected:        false,
		},
		{
			name:            "invalid transaction type",
			mergerDate:      "2023-06-15T00:00:00Z",
			transactionType: "invalid_type",
			status:          TransactionStatusCompleted,
			expected:        false,
		},
		{
			name:            "invalid status",
			mergerDate:      "2023-06-15T00:00:00Z",
			transactionType: TransactionTypeAcquisition,
			status:          "invalid_status",
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel := NewOrganizationMerger(&acquirer, &target, tt.mergerDate, tt.transactionType)
			rel.Status = tt.status
			assert.Equal(t, tt.expected, rel.Valid())
		})
	}
}

func TestOrganizationMerger_StatusMethods(t *testing.T) {
	acquirer := NewOrganization("Acquirer Corp")
	target := NewOrganization("Target Inc")

	// Test pending status
	pendingRel := NewOrganizationMerger(&acquirer, &target, "2023-06-15T00:00:00Z", TransactionTypeAcquisition)
	assert.True(t, pendingRel.IsPending())
	assert.False(t, pendingRel.IsCompleted())

	// Test completed status
	completedRel := NewOrganizationMerger(&acquirer, &target, "2023-06-15T00:00:00Z", TransactionTypeAcquisition)
	completedRel.Status = TransactionStatusCompleted
	assert.False(t, completedRel.IsPending())
	assert.True(t, completedRel.IsCompleted())
}

func TestOrganizationMerger_GetTransactionValueFormatted(t *testing.T) {
	acquirer := NewOrganization("Acquirer Corp")
	target := NewOrganization("Target Inc")

	tests := []struct {
		name     string
		value    float64
		currency string
		expected string
	}{
		{
			name:     "billions",
			value:    16000000000,
			currency: "USD",
			expected: "16.0B USD",
		},
		{
			name:     "millions",
			value:    500000000,
			currency: "USD",
			expected: "500.0M USD",
		},
		{
			name:     "thousands",
			value:    50000,
			currency: "USD",
			expected: "50.0K USD",
		},
		{
			name:     "no currency specified",
			value:    1000000000,
			currency: "",
			expected: "1.0B USD",
		},
		{
			name:     "zero value",
			value:    0,
			currency: "USD",
			expected: "Not disclosed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel := NewOrganizationMerger(&acquirer, &target, "2023-06-15T00:00:00Z", TransactionTypeAcquisition)
			rel.TransactionValue = tt.value
			rel.Currency = tt.currency
			assert.Equal(t, tt.expected, rel.GetTransactionValueFormatted())
		})
	}
}

func TestOrganizationRelationshipService_SubsidiaryOperations(t *testing.T) {
	service := NewOrganizationRelationshipService()

	// Create test organizations
	parent := NewOrganization("Parent Corp")
	subsidiary1 := NewOrganization("Subsidiary One")
	subsidiary2 := NewOrganization("Subsidiary Two")

	// Add organizations to service
	service.AddOrganization(&parent)
	service.AddOrganization(&subsidiary1)
	service.AddOrganization(&subsidiary2)

	// Create relationships
	rel1 := NewOrganizationParentSubsidiary(&parent, &subsidiary1, 100.0, RelationshipTypeWhollyOwned)
	rel2 := NewOrganizationParentSubsidiary(&parent, &subsidiary2, 75.0, RelationshipTypeMajorityOwned)

	service.AddRelationship(rel1)
	service.AddRelationship(rel2)

	// Test getting subsidiaries
	subsidiaries := service.GetSubsidiaries(parent.GetKey())
	assert.Len(t, subsidiaries, 2)

	subNames := make([]string, len(subsidiaries))
	for i, sub := range subsidiaries {
		subNames[i] = sub.PrimaryName
	}
	assert.Contains(t, subNames, "Subsidiary One")
	assert.Contains(t, subNames, "Subsidiary Two")

	// Test getting parent organizations
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

	// Add name history
	history1 := NewOrganizationNameHistory(&org, "Original Name", "Intermediate Name", "2020-01-01T00:00:00Z")
	history2 := NewOrganizationNameHistory(&org, "Intermediate Name", "Current Name", "2022-01-01T00:00:00Z")

	service.AddRelationship(history1)
	service.AddRelationship(history2)

	// Test getting name history
	history := service.GetNameHistory(org.GetKey())
	assert.Len(t, history, 2)

	// Check that we have both history entries
	oldNames := make([]string, len(history))
	newNames := make([]string, len(history))
	for i, h := range history {
		oldNames[i] = h.OldName
		newNames[i] = h.NewName
	}

	assert.Contains(t, oldNames, "Original Name")
	assert.Contains(t, oldNames, "Intermediate Name")
	assert.Contains(t, newNames, "Intermediate Name")
	assert.Contains(t, newNames, "Current Name")
}

func TestOrganizationRelationshipService_OrganizationFamily(t *testing.T) {
	service := NewOrganizationRelationshipService()

	// Create a complex organization structure
	grandparent := NewOrganization("Grandparent Corp")
	parent1 := NewOrganization("Parent One")
	parent2 := NewOrganization("Parent Two")
	child1 := NewOrganization("Child One")
	child2 := NewOrganization("Child Two")
	sibling := NewOrganization("Sibling Corp")

	// Add all organizations
	orgs := []*Organization{&grandparent, &parent1, &parent2, &child1, &child2, &sibling}
	for _, org := range orgs {
		service.AddOrganization(org)
	}

	// Create relationships
	service.AddRelationship(NewOrganizationParentSubsidiary(&grandparent, &parent1, 100.0, RelationshipTypeWhollyOwned))
	service.AddRelationship(NewOrganizationParentSubsidiary(&grandparent, &parent2, 100.0, RelationshipTypeWhollyOwned))
	service.AddRelationship(NewOrganizationParentSubsidiary(&parent1, &child1, 100.0, RelationshipTypeWhollyOwned))
	service.AddRelationship(NewOrganizationParentSubsidiary(&parent1, &child2, 100.0, RelationshipTypeWhollyOwned))
	// sibling is not connected to the family

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

	// Create parent-subsidiary relationships
	service.AddRelationship(NewOrganizationParentSubsidiary(&walmartInc, &walmartStores, 100.0, RelationshipTypeWhollyOwned))
	service.AddRelationship(NewOrganizationParentSubsidiary(&walmartInc, &samsClub, 100.0, RelationshipTypeWhollyOwned))
	service.AddRelationship(NewOrganizationParentSubsidiary(&walmartInc, &walmartEcommerce, 100.0, RelationshipTypeWhollyOwned))

	// Create name history
	nameHistory1 := NewOrganizationNameHistory(&walmartInc, "Wal-Mart Stores Inc", "Walmart Inc", "2018-02-01T00:00:00Z")
	nameHistory1.ChangeReason = "Corporate rebranding"
	service.AddRelationship(nameHistory1)

	// Create a merger (e.g., acquisition of an e-commerce company)
	merger := NewOrganizationMerger(&walmartInc, &walmartEcommerce, "2016-08-08T00:00:00Z", TransactionTypeAcquisition)
	merger.TransactionValue = 3300000000 // $3.3B
	merger.Status = TransactionStatusCompleted
	service.AddRelationship(merger)

	// Test the complete corporate structure
	subsidiaries := service.GetSubsidiaries(walmartInc.GetKey())
	assert.Len(t, subsidiaries, 3)

	nameHistory := service.GetNameHistory(walmartInc.GetKey())
	assert.Len(t, nameHistory, 1)
	assert.Equal(t, "Wal-Mart Stores Inc", nameHistory[0].OldName)
	assert.Equal(t, "Walmart Inc", nameHistory[0].NewName)
	assert.Equal(t, "Corporate rebranding", nameHistory[0].ChangeReason)

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
		service.AddOrganization(&subsidiary)
		rel := NewOrganizationParentSubsidiary(&parent, &subsidiary, 100.0, RelationshipTypeWhollyOwned)
		service.AddRelationship(rel)
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
		service.AddOrganization(&parent)
		service.AddRelationship(NewOrganizationParentSubsidiary(&grandparent, &parent, 100.0, RelationshipTypeWhollyOwned))

		for j := 0; j < 100; j++ {
			child := NewOrganization(fmt.Sprintf("Child%d_%d", i, j))
			service.AddOrganization(&child)
			service.AddRelationship(NewOrganizationParentSubsidiary(&parent, &child, 100.0, RelationshipTypeWhollyOwned))
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
