package model

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBreachIntelligenceAttribute_ValidKeyAndValid(t *testing.T) {
	parentKey := "#person#john@example.com#John Doe"
	entryID := "12345"
	databaseName := "LinkedIn"
	breachStatus := "BREACHED"
	riskLevel := "HIGH"

	b := NewBreachIntelligenceAttribute(parentKey, entryID, databaseName, breachStatus, riskLevel, nil)

	assert.True(t, b.Valid(), "Valid() should return true for a fully-populated breach intelligence attribute")
	assert.Contains(t, b.Key, databaseName, "Key should contain the databaseName")
	assert.Contains(t, b.Key, entryID, "Key should contain the entryID")
	assert.True(t, strings.HasPrefix(b.Key, "#breach_intelligence#"), "Key should start with '#breach_intelligence#'")
	assert.Equal(t, entryID, *b.EntryID, "EntryID pointer should equal the provided entryID")
	assert.Equal(t, databaseName, *b.DatabaseName, "DatabaseName pointer should equal the provided databaseName")
	assert.Equal(t, parentKey, b.PersonKey, "PersonKey should be set when parentKey starts with #person#")
}

func TestNewBreachIntelligenceAttribute_AssetParent(t *testing.T) {
	parentKey := "#asset#example.com#example.com"
	entryID := "67890"
	databaseName := "Adobe"
	breachStatus := "BREACHED"
	riskLevel := "MEDIUM"

	b := NewBreachIntelligenceAttribute(parentKey, entryID, databaseName, breachStatus, riskLevel, nil)

	assert.True(t, b.Valid(), "Valid() should return true for an asset-parented breach intelligence attribute")
	assert.Equal(t, parentKey, b.AssetID, "AssetID should be set when parentKey starts with #asset#")
	assert.Empty(t, b.PersonKey, "PersonKey should be empty when parentKey is an asset key")
}

func TestBreachIntelligenceAttribute_Valid_EmptyEntryID(t *testing.T) {
	emptyEntryID := ""
	databaseName := "LinkedIn"
	b := BreachIntelligenceAttribute{
		Key:          "#breach_intelligence#some_key",
		AssetID:      "#asset#example.com#example.com",
		EntryID:      &emptyEntryID,
		DatabaseName: &databaseName,
		BreachStatus: "BREACHED",
		RiskLevel:    "HIGH",
		CheckedAt:    "2024-02-04T10:00:00Z",
	}

	assert.False(t, b.Valid(), "Valid() should return false when EntryID is an empty string")
}

func TestBreachIntelligenceAttribute_Valid_EmptyDatabaseName(t *testing.T) {
	entryID := "12345"
	emptyDatabaseName := ""
	b := BreachIntelligenceAttribute{
		Key:          "#breach_intelligence#some_key",
		AssetID:      "#asset#example.com#example.com",
		EntryID:      &entryID,
		DatabaseName: &emptyDatabaseName,
		BreachStatus: "BREACHED",
		RiskLevel:    "HIGH",
		CheckedAt:    "2024-02-04T10:00:00Z",
	}

	assert.False(t, b.Valid(), "Valid() should return false when DatabaseName is an empty string")
}

func TestBreachIntelligenceAttribute_Valid_NilEntryID(t *testing.T) {
	databaseName := "LinkedIn"
	b := BreachIntelligenceAttribute{
		Key:          "#breach_intelligence#some_key",
		AssetID:      "#asset#example.com#example.com",
		EntryID:      nil,
		DatabaseName: &databaseName,
		BreachStatus: "BREACHED",
		RiskLevel:    "HIGH",
		CheckedAt:    "2024-02-04T10:00:00Z",
	}

	assert.False(t, b.Valid(), "Valid() should return false when EntryID is nil")
}
