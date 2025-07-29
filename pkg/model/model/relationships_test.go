package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test the relationship key generation
func TestDiscoveredRelationshipKeyGeneration(t *testing.T) {
	tests := []struct {
		name          string
		sourceKey     string
		targetKey     string
		expectedKey   string
		shouldContain []string
	}{
		{
			name:        "Standard asset to resource relationship",
			sourceKey:   "#asset#amazon#411435703965",
			targetKey:   "#awsresource#411435703965#arn:aws:account:ap-northeast-1:411435703965:411435703965",
			expectedKey: "#asset#amazon#411435703965#DISCOVERED#awsresource#411435703965#arn:aws:account:ap-northeast-1:411435703965:411435703965",
			shouldContain: []string{
				"#asset#amazon#411435703965",
				"#DISCOVERED#",
				"#awsresource#411435703965#arn:aws:account:ap-northeast-1:411435703965:411435703965",
			},
		},
		{
			name:        "Check for missing separator issue",
			sourceKey:   "#asset#test",
			targetKey:   "#resource#test",
			expectedKey: "#asset#test#DISCOVERED#resource#test",
			shouldContain: []string{
				"#DISCOVERED#",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := NewAsset("test-source", "test-source")
			source.Key = tt.sourceKey
			target := NewAsset("test-target", "test-target")
			target.Key = tt.targetKey

			rel := NewDiscovered(&source, &target)

			assert.Equal(t, tt.expectedKey, rel.GetKey())

			for _, substr := range tt.shouldContain {
				assert.Contains(t, rel.GetKey(), substr)
			}
		})
	}
}

// Test the Visit behavior
func TestDiscoveredRelationshipVisit(t *testing.T) {
	// Create original relationship from "database"
	dbSource := NewAsset("amazon", "123")
	dbSource.Key = "#asset#amazon#123"
	dbTarget := NewAsset("resource", "456")
	dbTarget.Key = "#resource#456"
	dbRel := NewDiscovered(&dbSource, &dbTarget)
	dbRel.Base().Capability = "original-capability"
	dbRel.Base().AttachmentPath = "/original/path"
	dbRel.Base().Visited = "2024-01-01"

	// Create new relationship
	newSource := NewAsset("amazon", "123")
	newSource.Key = "#asset#amazon#123"
	newTarget := NewAsset("resource", "456")
	newTarget.Key = "#resource#456"
	newRel := NewDiscovered(&newSource, &newTarget)
	newRel.Base().Capability = "new-capability"
	newRel.Base().Visited = "2024-01-02"

	// Store original values
	originalDbKey := dbRel.GetKey()
	originalNewKey := newRel.GetKey()

	// Simulate the Visit call
	dbRel.Base().Visit(newRel)

	// Test expectations after Visit
	t.Run("Visit updates visited time", func(t *testing.T) {
		assert.Equal(t, "2024-01-02", dbRel.Base().Visited)
	})

	t.Run("Visit updates capability", func(t *testing.T) {
		assert.Equal(t, "new-capability", dbRel.Base().Capability)
	})

	t.Run("Visit preserves attachment path when new one is empty", func(t *testing.T) {
		assert.Equal(t, "/original/path", dbRel.Base().AttachmentPath)
	})

	t.Run("Visit replaces source and target", func(t *testing.T) {
		assert.Equal(t, &newSource, dbRel.Base().Source)
		assert.Equal(t, &newTarget, dbRel.Base().Target)
	})

	t.Run("Keys remain unchanged after Visit", func(t *testing.T) {
		assert.Equal(t, originalDbKey, dbRel.GetKey())
		assert.Equal(t, originalNewKey, newRel.GetKey())
	})
}

// Test the problematic scenario from processRelationship
func TestProcessRelationshipScenario(t *testing.T) {
	// Simulate the scenario where we have an existing relationship
	// and try to create a new one between the same nodes

	// Existing relationship (as if from database)
	existingSource := NewAsset("amazon", "411435703965")
	existingSource.Key = "#asset#amazon#411435703965"
	existingTarget := NewAsset("awsresource", "411435703965")
	existingTarget.Key = "#awsresource#411435703965#arn:aws:account:ap-northeast-1:411435703965:411435703965"
	existingRel := NewDiscovered(&existingSource, &existingTarget)
	existingRel.Base().Visited = "2024-01-01"
	existingRelKey := existingRel.GetKey()

	// New relationship (being processed)
	newSource := NewAsset("amazon", "411435703965")
	newSource.Key = "#asset#amazon#411435703965"
	newTarget := NewAsset("awsresource", "411435703965")
	newTarget.Key = "#awsresource#411435703965#arn:aws:account:ap-northeast-1:411435703965:411435703965"
	newRel := NewDiscovered(&newSource, &newTarget)
	newRelKey := newRel.GetKey()

	t.Run("Keys should be identical for same source/target", func(t *testing.T) {
		assert.Equal(t, existingRelKey, newRelKey)
	})

	// Simulate the Visit pattern from processRelationship
	existingRel.Base().Visit(newRel)

	t.Run("After Visit, existing rel has new source/target objects", func(t *testing.T) {
		assert.Equal(t, &newSource, existingRel.Base().Source)
		assert.Equal(t, &newTarget, existingRel.Base().Target)
	})

	t.Run("Key remains unchanged after Visit", func(t *testing.T) {
		assert.Equal(t, existingRelKey, existingRel.GetKey())
	})
}

// Test edge cases that might cause constraint violations
func TestConstraintViolationScenarios(t *testing.T) {
	t.Run("Different nodes with same key", func(t *testing.T) {
		// This shouldn't happen but let's test it
		source1 := NewAsset("test", "123")
		source1.Key = "#asset#123"
		source2 := NewAsset("test", "123") // Same key, different object
		source2.Key = "#asset#123"
		target := NewAsset("resource", "456")
		target.Key = "#resource#456"

		rel1 := NewDiscovered(&source1, &target)
		rel2 := NewDiscovered(&source2, &target)

		assert.Equal(t, rel1.GetKey(), rel2.GetKey())
	})

	t.Run("Empty or invalid keys", func(t *testing.T) {
		source := NewAsset("empty", "empty")
		source.Key = ""
		target := NewAsset("resource", "456")
		target.Key = "#resource#456"

		rel := NewDiscovered(&source, &target)
		assert.Contains(t, rel.GetKey(), "#DISCOVERED#")
	})

	t.Run("Key with special characters", func(t *testing.T) {
		source := NewAsset("test", "region")
		source.Key = "#asset#test:region"
		target := NewAsset("resource", "aws")
		target.Key = "#resource#arn:aws:service:region:account:resource"

		rel := NewDiscovered(&source, &target)
		expectedKey := "#asset#test:region#DISCOVERED#resource#arn:aws:service:region:account:resource"
		assert.Equal(t, expectedKey, rel.GetKey())
	})
}

// Test the actual MERGE query generation scenario
func TestMergeQueryScenario(t *testing.T) {
	// This test documents what properties should be in MERGE vs SET

	source := NewAsset("amazon", "411435703965")
	source.Key = "#asset#amazon#411435703965"
	target := NewAsset("awsresource", "411435703965")
	target.Key = "#awsresource#411435703965#arn:aws:account:ap-northeast-1:411435703965:411435703965"
	rel := NewDiscovered(&source, &target)

	username := "zach.grace+play@praetorian.com"

	t.Run("MERGE should include constraint properties", func(t *testing.T) {
		// The MERGE pattern should include:
		mergeProps := map[string]interface{}{
			"key":      rel.GetKey(),
			"username": username,
		}

		assert.NotEmpty(t, mergeProps["key"])
		assert.NotEmpty(t, mergeProps["username"])
	})

	t.Run("SET should exclude constraint properties", func(t *testing.T) {
		// After MERGE, SET should include all other properties except key and username
		// to avoid constraint violations

		// This would be done in the actual query builder
		allProps := map[string]interface{}{
			"key":            rel.GetKey(),
			"username":       username,
			"capability":     "amazon",
			"created":        "2024-01-01",
			"visited":        "2024-01-02",
			"attachmentPath": "/path",
		}

		// Remove constraint properties for SET
		setProps := make(map[string]interface{})
		for k, v := range allProps {
			if k != "key" && k != "username" {
				setProps[k] = v
			}
		}

		assert.NotContains(t, setProps, "key")
		assert.NotContains(t, setProps, "username")
		assert.Contains(t, setProps, "capability")
	})
}

// Test to reproduce the actual error scenario
func TestReproduceConstraintViolation(t *testing.T) {
	// This test reproduces the exact scenario that causes the constraint violation:
	// processRelationship finds an existing relationship, calls Visit, then tries to re-insert it

	// Step 1: Create a relationship that represents one "found in database"
	dbSource := NewAsset("amazon", "411435703965")
	dbSource.Key = "#asset#amazon#411435703965"
	dbTarget := NewAsset("awsresource", "411435703965")
	dbTarget.Key = "#awsresource#411435703965#arn:aws:account:ap-northeast-1:411435703965:411435703965"
	existingRel := NewDiscovered(&dbSource, &dbTarget)
	existingRel.Base().Capability = "old-capability"
	existingRel.Base().Visited = "2024-01-01"

	// Step 2: Create a "new" relationship with the SAME source/target (but different metadata)
	// This simulates what happens when processModel creates NewDiscovered(parent, m)
	// where parent and m have the same keys as an existing relationship
	newSource := NewAsset("amazon", "411435703965")
	newSource.Key = "#asset#amazon#411435703965" // Same key as existing
	newTarget := NewAsset("awsresource", "411435703965")
	newTarget.Key = "#awsresource#411435703965#arn:aws:account:ap-northeast-1:411435703965:411435703965" // Same key as existing
	newRel := NewDiscovered(&newSource, &newTarget)
	newRel.Base().Capability = "new-capability"
	newRel.Base().Visited = "2024-01-02"

	t.Run("Same key relationships cause constraint violation scenario", func(t *testing.T) {
		// Both relationships should have identical keys (the constraint violation key)
		expectedKey := "#asset#amazon#411435703965#DISCOVERED#awsresource#411435703965#arn:aws:account:ap-northeast-1:411435703965:411435703965"
		assert.Equal(t, expectedKey, existingRel.GetKey())
		assert.Equal(t, expectedKey, newRel.GetKey())

		// Simulate the processRelationship flow:
		// 1. Find existing relationship (existingRel)
		// 2. Call Visit to update it with new data
		existingRel.Base().Visit(newRel)

		// 3. After Visit, we have the existing relationship with updated metadata
		assert.Equal(t, "new-capability", existingRel.Base().Capability)
		assert.Equal(t, "2024-01-02", existingRel.Base().Visited)

		// 4. The key remains the same (this is the constraint violation key)
		assert.Equal(t, expectedKey, existingRel.GetKey())

		// This is the problem: processRelationship tries to insert existingRel again
		// even though it already exists in the database with this exact key
		t.Logf("Relationship key that causes constraint violation: %s", existingRel.GetKey())
	})

	t.Run("Visit updates source and target references correctly", func(t *testing.T) {
		// After Visit, the existing relationship should have the new source/target objects
		assert.Equal(t, &newSource, existingRel.Base().Source)
		assert.Equal(t, &newTarget, existingRel.Base().Target)

		// The key should still match the source/target keys (note: no # between label and target)
		expectedKeyFromNodes := fmt.Sprintf("%s#%s%s",
			newSource.GetKey(),
			"DISCOVERED",
			newTarget.GetKey())
		assert.Equal(t, expectedKeyFromNodes, existingRel.GetKey())
	})
}
