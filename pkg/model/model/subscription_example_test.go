package model

import (
	"encoding/json"
	"testing"
	"time"
)

// TestExampleSubscriptionData tests the exact data provided by the user
func TestExampleSubscriptionData(t *testing.T) {
	// This is the exact JSON data provided by the user
	jsonData := `{
		"username": "chariot+alteryx@praetorian.com",
		"key": "#setting#subscription",
		"last_modified": "2025-09-29T14:55:36.862508315Z",
		"name": "subscription",
		"value": {
			"endDate": "2025-12-31",
			"estimatedCost": 0,
			"msrp": 200000,
			"msrpBreakdown": {
				"assetCost": 100000,
				"fqdnCost": 0,
				"platformFee": 25000,
				"supportTier": 75000
			},
			"numberOfAssets": 15000,
			"scanSchedule": "daily",
			"selectedTier": "ultimate",
			"startDate": "2025-01-01"
		}
	}`

	// Parse the JSON data into a Setting
	var setting Setting
	err := json.Unmarshal([]byte(jsonData), &setting)
	if err != nil {
		t.Fatalf("Failed to parse JSON data: %v", err)
	}

	// Verify the setting data was parsed correctly
	if setting.Username != "chariot+alteryx@praetorian.com" {
		t.Errorf("Expected Username to be 'chariot+alteryx@praetorian.com', got '%s'", setting.Username)
	}
	if setting.Key != "#setting#subscription" {
		t.Errorf("Expected Key to be '#setting#subscription', got '%s'", setting.Key)
	}
	if setting.Name != "subscription" {
		t.Errorf("Expected Name to be 'subscription', got '%s'", setting.Name)
	}

	// Test that the subscription value is valid according to our validator
	// Note: Settings don't validate subscription data - only Configurations do
	err = setting.Valid()
	if err != nil {
		t.Errorf("Setting validation failed: %v", err)
	}

	// Test the subscription value directly with the validator
	valueMap, ok := setting.Value.(map[string]any)
	if !ok {
		t.Fatal("Setting value is not a map[string]any")
	}

	err = subscriptionValidator(valueMap)
	if err != nil {
		t.Errorf("Subscription validation failed: %v", err)
	}

	// Test the Within method with the subscription data
	sub := Subscription(valueMap)
	
	// Test time within the subscription period (June 2025)
	withinTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	if !sub.Within(withinTime) {
		t.Error("Expected June 2025 to be within subscription period")
	}

	// Test time before subscription starts (December 2024)
	beforeTime := time.Date(2024, 12, 31, 12, 0, 0, 0, time.UTC)
	if sub.Within(beforeTime) {
		t.Error("Expected December 2024 to be outside subscription period")
	}

	// Test time after subscription ends (January 2026)
	afterTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	if sub.Within(afterTime) {
		t.Error("Expected January 2026 to be outside subscription period")
	}

	// Test Configuration creation and validation
	config := NewConfiguration("subscription", valueMap)
	err = config.Valid()
	if err != nil {
		t.Errorf("Configuration validation failed: %v", err)
	}

	if config.Key != "#configuration#subscription" {
		t.Errorf("Expected Configuration key to be '#configuration#subscription', got '%s'", config.Key)
	}

	// Verify all the subscription data fields
	if sub["selectedTier"] != "ultimate" {
		t.Errorf("Expected selectedTier to be 'ultimate', got %v", sub["selectedTier"])
	}
	if sub["scanSchedule"] != "daily" {
		t.Errorf("Expected scanSchedule to be 'daily', got %v", sub["scanSchedule"])
	}
	
	// Check numeric values (JSON unmarshaling makes these float64)
	if msrp, ok := sub["msrp"].(float64); !ok || msrp != 200000 {
		t.Errorf("Expected msrp to be 200000, got %v (%T)", sub["msrp"], sub["msrp"])
	}
	if assets, ok := sub["numberOfAssets"].(float64); !ok || assets != 15000 {
		t.Errorf("Expected numberOfAssets to be 15000, got %v (%T)", sub["numberOfAssets"], sub["numberOfAssets"])
	}

	// Check nested msrpBreakdown
	if breakdown, ok := sub["msrpBreakdown"].(map[string]any); ok {
		if assetCost, ok := breakdown["assetCost"].(float64); !ok || assetCost != 100000 {
			t.Errorf("Expected assetCost to be 100000, got %v", breakdown["assetCost"])
		}
		if platformFee, ok := breakdown["platformFee"].(float64); !ok || platformFee != 25000 {
			t.Errorf("Expected platformFee to be 25000, got %v", breakdown["platformFee"])
		}
		if supportTier, ok := breakdown["supportTier"].(float64); !ok || supportTier != 75000 {
			t.Errorf("Expected supportTier to be 75000, got %v", breakdown["supportTier"])
		}
		if fqdnCost, ok := breakdown["fqdnCost"].(float64); !ok || fqdnCost != 0 {
			t.Errorf("Expected fqdnCost to be 0, got %v", breakdown["fqdnCost"])
		}
	} else {
		t.Error("Expected msrpBreakdown to be a map[string]any")
	}
}