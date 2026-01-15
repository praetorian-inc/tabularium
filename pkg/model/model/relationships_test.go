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
		source, target := dbRel.Nodes()
		assert.Equal(t, newSource.GetKey(), source.GetKey())
		assert.Equal(t, newTarget.GetKey(), target.GetKey())
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
		source, target := existingRel.Nodes()
		assert.Equal(t, newSource.GetKey(), source.GetKey())
		assert.Equal(t, newTarget.GetKey(), target.GetKey())
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

// Test HasWebpage relationship functionality
func TestHasWebpageRelationship(t *testing.T) {
	t.Run("Label returns correct value", func(t *testing.T) {
		source := NewWebApplication("https://example.com", "Example App")
		target := NewWebpageFromString("https://example.com/page", &source)

		rel := NewHasWebpage(&source, &target)
		assert.Equal(t, HasWebpageLabel, rel.Label())
		assert.Equal(t, "HAS_WEBPAGE", rel.Label())
	})

	t.Run("Key generation follows pattern", func(t *testing.T) {
		source := NewWebApplication("https://example.com", "Example App")
		source.Key = "#webapplication#https://example.com"
		target := NewWebpageFromString("https://example.com/page", &source)
		target.Key = "#webpage#https://example.com/page#webapplication#https://example.com"

		rel := NewHasWebpage(&source, &target)
		expectedKey := "#webapplication#https://example.com#HAS_WEBPAGE#webpage#https://example.com/page#webapplication#https://example.com"
		assert.Equal(t, expectedKey, rel.GetKey())
		assert.Contains(t, rel.GetKey(), "#HAS_WEBPAGE#")
		assert.Contains(t, rel.GetKey(), source.GetKey())
		assert.Contains(t, rel.GetKey(), target.GetKey())
	})

	t.Run("Relationship implements GraphRelationship interface", func(t *testing.T) {
		source := NewWebApplication("https://example.com", "Example App")
		target := NewWebpageFromString("https://example.com/page", &source)

		rel := NewHasWebpage(&source, &target)

		// Test that it implements GraphRelationship interface
		var _ GraphRelationship = rel
		assert.NotNil(t, rel.Base())
		assert.Equal(t, HasWebpageLabel, rel.Label())
		assert.True(t, rel.Valid())
	})

	t.Run("Visit functionality works correctly", func(t *testing.T) {
		// Create original relationship
		dbSource := NewWebApplication("https://example.com", "Example App")
		dbSource.Key = "#webapplication#https://example.com"
		dbTarget := NewWebpageFromString("https://example.com/page", &dbSource)
		dbTarget.Key = "#webpage#https://example.com/page#webapplication#https://example.com"
		dbRel := NewHasWebpage(&dbSource, &dbTarget)
		dbRel.Base().Capability = "original-crawler"
		dbRel.Base().Visited = "2024-01-01"

		// Create new relationship with updated data
		newSource := NewWebApplication("https://example.com", "Example App")
		newSource.Key = "#webapplication#https://example.com"
		newTarget := NewWebpageFromString("https://example.com/page", &newSource)
		newTarget.Key = "#webpage#https://example.com/page#webapplication#https://example.com"
		newRel := NewHasWebpage(&newSource, &newTarget)
		newRel.Base().Capability = "new-crawler"
		newRel.Base().Visited = "2024-01-02"

		// Perform Visit
		dbRel.Base().Visit(newRel)

		// Verify updates
		assert.Equal(t, "2024-01-02", dbRel.Base().Visited)
		assert.Equal(t, "new-crawler", dbRel.Base().Capability)
		source, target := dbRel.Nodes()
		assert.Equal(t, newSource.GetKey(), source.GetKey())
		assert.Equal(t, newTarget.GetKey(), target.GetKey())
	})

	t.Run("Nodes returns correct source and target", func(t *testing.T) {
		source := NewWebApplication("https://example.com", "Example App")
		target := NewWebpageFromString("https://example.com/page", &source)

		rel := NewHasWebpage(&source, &target)
		sourceNode, targetNode := rel.Nodes()

		assert.Equal(t, source.GetKey(), sourceNode.GetKey())
		assert.Equal(t, target.GetKey(), targetNode.GetKey())
	})
}

// TestScannedByRelationship tests the ScannedBy relationship functionality
func TestScannedByRelationship(t *testing.T) {
	t.Run("Label returns correct value", func(t *testing.T) {
		asset := NewAsset("10.0.0.5", "10.0.0.5")
		agent := &AegisAgent{
			ClientID: "test-agent-123",
		}
		agent.Key = "#aegisagent#test-agent-123"

		rel := NewScannedBy(&asset, agent, "nmap")
		assert.Equal(t, ScannedByLabel, rel.Label())
		assert.Equal(t, "SCANNED_BY", rel.Label())
	})

	t.Run("Key generation follows pattern", func(t *testing.T) {
		asset := NewAsset("10.0.0.5", "10.0.0.5")
		asset.Key = "#asset#10.0.0.5#10.0.0.5"
		agent := &AegisAgent{
			ClientID: "test-agent-123",
		}
		agent.Key = "#aegisagent#test-agent-123"

		rel := NewScannedBy(&asset, agent, "nmap")
		expectedKey := "#asset#10.0.0.5#10.0.0.5#SCANNED_BY#aegisagent#test-agent-123"
		assert.Equal(t, expectedKey, rel.GetKey())
		assert.Contains(t, rel.GetKey(), "#SCANNED_BY#")
		assert.Contains(t, rel.GetKey(), asset.GetKey())
		assert.Contains(t, rel.GetKey(), agent.GetKey())
	})

	t.Run("Scan type is configurable", func(t *testing.T) {
		asset := NewAsset("10.0.0.5", "10.0.0.5")
		agent := &AegisAgent{ClientID: "test-agent-123"}
		agent.Key = "#aegisagent#test-agent-123"

		// Test different scan types
		scanTypes := []string{"nmap", "ping", "http", "custom-scanner"}
		for _, scanType := range scanTypes {
			rel := NewScannedBy(&asset, agent, scanType).(*ScannedBy)
			assert.Equal(t, scanType, rel.ScanType, "ScanType should be configurable")
		}
	})

	t.Run("Scan type is set as provided by caller", func(t *testing.T) {
		asset := NewAsset("10.0.0.5", "10.0.0.5")
		agent := &AegisAgent{ClientID: "test-agent-123"}
		agent.Key = "#aegisagent#test-agent-123"

		// Empty string is accepted if caller provides it
		rel := NewScannedBy(&asset, agent, "").(*ScannedBy)
		assert.Equal(t, "", rel.ScanType, "Empty scan type should be accepted if caller provides it")
		
		// Caller should provide the scan type explicitly
		relWithType := NewScannedBy(&asset, agent, "nmap").(*ScannedBy)
		assert.Equal(t, "nmap", relWithType.ScanType, "Scan type should be set as provided by caller")
	})

	t.Run("ScanTime is set", func(t *testing.T) {
		asset := NewAsset("10.0.0.5", "10.0.0.5")
		agent := &AegisAgent{ClientID: "test-agent-123"}
		agent.Key = "#aegisagent#test-agent-123"

		rel := NewScannedBy(&asset, agent, "nmap").(*ScannedBy)
		assert.NotEmpty(t, rel.ScanTime, "ScanTime should be set")
	})

	t.Run("Relationship implements GraphRelationship interface", func(t *testing.T) {
		asset := NewAsset("10.0.0.5", "10.0.0.5")
		agent := &AegisAgent{ClientID: "test-agent-123"}
		agent.Key = "#aegisagent#test-agent-123"

		rel := NewScannedBy(&asset, agent, "nmap")

		// Test that it implements GraphRelationship interface
		var _ GraphRelationship = rel
		assert.NotNil(t, rel.Base())
		assert.Equal(t, ScannedByLabel, rel.Label())
		assert.True(t, rel.Valid())
	})

	t.Run("Visit functionality works correctly", func(t *testing.T) {
		// Create original relationship
		dbAsset := NewAsset("10.0.0.5", "10.0.0.5")
		dbAsset.Key = "#asset#10.0.0.5#10.0.0.5"
		dbAgent := &AegisAgent{ClientID: "test-agent-123"}
		dbAgent.Key = "#aegisagent#test-agent-123"
		dbRel := NewScannedBy(&dbAsset, dbAgent, "nmap").(*ScannedBy)
		dbRel.Base().Capability = "original-scanner"
		dbRel.Base().Visited = "2024-01-01"
		dbRel.ScanType = "nmap"

		// Create new relationship with updated data
		newAsset := NewAsset("10.0.0.5", "10.0.0.5")
		newAsset.Key = "#asset#10.0.0.5#10.0.0.5"
		newAgent := &AegisAgent{ClientID: "test-agent-123"}
		newAgent.Key = "#aegisagent#test-agent-123"
		newRel := NewScannedBy(&newAsset, newAgent, "ping").(*ScannedBy)
		newRel.Base().Capability = "new-scanner"
		newRel.Base().Visited = "2024-01-02"

		// Perform Visit
		dbRel.Base().Visit(newRel)

		// Verify updates
		assert.Equal(t, "2024-01-02", dbRel.Base().Visited)
		assert.Equal(t, "new-scanner", dbRel.Base().Capability)
		source, target := dbRel.Nodes()
		assert.Equal(t, newAsset.GetKey(), source.GetKey())
		assert.Equal(t, newAgent.GetKey(), target.GetKey())
	})

	t.Run("Nodes returns correct source and target", func(t *testing.T) {
		asset := NewAsset("10.0.0.5", "10.0.0.5")
		agent := &AegisAgent{ClientID: "test-agent-123"}
		agent.Key = "#aegisagent#test-agent-123"

		rel := NewScannedBy(&asset, agent, "nmap")
		sourceNode, targetNode := rel.Nodes()

		assert.Equal(t, asset.GetKey(), sourceNode.GetKey())
		assert.Equal(t, agent.GetKey(), targetNode.GetKey())
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
		source, target := existingRel.Nodes()
		assert.Equal(t, newSource.GetKey(), source.GetKey())
		assert.Equal(t, newTarget.GetKey(), target.GetKey())

		// The key should still match the source/target keys (note: no # between label and target)
		expectedKeyFromNodes := fmt.Sprintf("%s#%s%s",
			newSource.GetKey(),
			"DISCOVERED",
			newTarget.GetKey())
		assert.Equal(t, expectedKeyFromNodes, existingRel.GetKey())
	})
}
