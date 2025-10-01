package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSubscriptionValidator(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{
			name: "valid subscription with all fields",
			input: map[string]any{
				"endDate":       "2025-12-31",
				"estimatedCost": 0,
				"scanSchedule":  "daily",
				"selectedTier":  "ultimate",
				"startDate":     "2025-01-01",
			},
			wantErr: false,
		},
		{
			name: "valid subscription with only required date fields",
			input: map[string]any{
				"startDate": "2024-01-01",
				"endDate":   "2024-12-31",
			},
			wantErr: false,
		},
		{
			name: "valid subscription with no date fields",
			input: map[string]any{
				"msrp":         100000,
				"selectedTier": "basic",
			},
			wantErr: false,
		},
		{
			name:    "invalid subscription - not a map",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "invalid subscription - nil",
			input:   nil,
			wantErr: true,
		},
		{
			name: "invalid subscription - startDate not a string",
			input: map[string]any{
				"startDate": 123,
			},
			wantErr: true,
		},
		{
			name: "invalid subscription - endDate not a string",
			input: map[string]any{
				"endDate": 123,
			},
			wantErr: true,
		},
		{
			name: "invalid subscription - invalid startDate format",
			input: map[string]any{
				"startDate": "2024-13-01",
			},
			wantErr: true,
		},
		{
			name: "invalid subscription - invalid endDate format",
			input: map[string]any{
				"endDate": "invalid-date",
			},
			wantErr: true,
		},
		{
			name: "invalid subscription - startDate with time",
			input: map[string]any{
				"startDate": "2024-01-01T00:00:00Z",
			},
			wantErr: true,
		},
		{
			name: "valid subscription - empty string dates should fail",
			input: map[string]any{
				"startDate": "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := subscriptionValidator(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("subscriptionValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSubscriptionJSON(t *testing.T) {
	subscriptionJSON := `{
		"endDate": "2025-12-31",
		"estimatedCost": 0,
		"numberOfAssets": 15000,
		"scanSchedule": "daily",
		"selectedTier": "ultimate",
		"startDate": "2025-01-01"
	}`

	var sub Subscription
	err := json.Unmarshal([]byte(subscriptionJSON), &sub)
	if err != nil {
		t.Fatalf("Failed to unmarshal subscription JSON: %v", err)
	}

	// Verify the subscription data
	if sub["endDate"] != "2025-12-31" {
		t.Errorf("Expected endDate to be '2025-12-31', got %v", sub["endDate"])
	}
	if sub["startDate"] != "2025-01-01" {
		t.Errorf("Expected startDate to be '2025-01-01', got %v", sub["startDate"])
	}
	if sub["selectedTier"] != "ultimate" {
		t.Errorf("Expected selectedTier to be 'ultimate', got %v", sub["selectedTier"])
	}
	if sub["scanSchedule"] != "daily" {
		t.Errorf("Expected scanSchedule to be 'daily', got %v", sub["scanSchedule"])
	}

	if assets, ok := sub["numberOfAssets"].(float64); !ok || assets != 15000 {
		t.Errorf("Expected numberOfAssets to be 15000, got %v", sub["numberOfAssets"])
	}

}

func TestSubscriptionWithin(t *testing.T) {
	tests := []struct {
		name     string
		sub      Subscription
		testTime time.Time
		want     bool
	}{
		{
			name: "time within subscription period",
			sub: Subscription{
				"startDate": "2025-01-01",
				"endDate":   "2025-12-31",
			},
			testTime: time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC),
			want:     true,
		},
		{
			name: "time before subscription start",
			sub: Subscription{
				"startDate": "2025-01-01",
				"endDate":   "2025-12-31",
			},
			testTime: time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			want:     false,
		},
		{
			name: "time after subscription end",
			sub: Subscription{
				"startDate": "2025-01-01",
				"endDate":   "2025-12-31",
			},
			testTime: time.Date(2026, 1, 1, 0, 0, 1, 0, time.UTC),
			want:     false,
		},
		{
			name: "time at subscription start",
			sub: Subscription{
				"startDate": "2025-01-01",
				"endDate":   "2025-12-31",
			},
			testTime: time.Date(2025, 1, 1, 0, 0, 1, 0, time.UTC),
			want:     true,
		},
		{
			name: "time at subscription end",
			sub: Subscription{
				"startDate": "2025-01-01",
				"endDate":   "2025-12-31",
			},
			testTime: time.Date(2025, 12, 30, 23, 59, 59, 0, time.UTC),
			want:     true,
		},
		{
			name: "subscription without start date",
			sub: Subscription{
				"endDate": "2025-12-31",
			},
			testTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			want:     true,
		},
		{
			name: "subscription without end date",
			sub: Subscription{
				"startDate": "2025-01-01",
			},
			testTime: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
			want:     true,
		},
		{
			name: "subscription without any dates",
			sub: Subscription{
				"msrp": 100000,
			},
			testTime: time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC),
			want:     true,
		},
		{
			name: "subscription with invalid start date",
			sub: Subscription{
				"startDate": "invalid-date",
				"endDate":   "2025-12-31",
			},
			testTime: time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC),
			want:     true, // Should default to no start restriction
		},
		{
			name: "subscription with invalid end date",
			sub: Subscription{
				"startDate": "2025-01-01",
				"endDate":   "invalid-date",
			},
			testTime: time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC),
			want:     true, // Should default to no end restriction
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.sub.Within(tt.testTime)
			if got != tt.want {
				t.Errorf("Subscription.Within() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigurationValidation(t *testing.T) {
	tests := []struct {
		name       string
		configName string
		value      any
		wantErr    bool
	}{
		{
			name:       "valid subscription configuration",
			configName: "subscription",
			value: map[string]any{
				"startDate": "2025-01-01",
				"endDate":   "2025-12-31",
				"msrp":      200000,
			},
			wantErr: false,
		},
		{
			name:       "invalid subscription configuration",
			configName: "subscription",
			value: map[string]any{
				"startDate": "invalid-date",
			},
			wantErr: true,
		},
		{
			name:       "unknown configuration type",
			configName: "unknown",
			value:      "any-value",
			wantErr:    false, // Should not error for unknown config types
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfiguration(tt.configName, tt.value)
			err := config.Valid()
			if (err != nil) != tt.wantErr {
				t.Errorf("Configuration.Valid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSettingValidationWithSubscription(t *testing.T) {
	// Test that settings don't validate subscription data (only configurations do)
	setting := NewSetting("subscription", map[string]any{
		"startDate": "invalid-date", // This should not cause validation error for settings
	})

	err := setting.Valid()
	if err != nil {
		t.Errorf("Setting.Valid() should not validate subscription data, got error: %v", err)
	}
}

func TestCompleteSubscriptionWorkflow(t *testing.T) {
	// Test the complete workflow with the provided example data
	subscriptionData := map[string]any{
		"endDate":       "2025-12-31",
		"estimatedCost": 0,
		"msrp":          200000,
		"msrpBreakdown": map[string]any{
			"assetCost":   100000,
			"fqdnCost":    0,
			"platformFee": 25000,
			"supportTier": 75000,
		},
		"numberOfAssets": 15000,
		"scanSchedule":   "daily",
		"selectedTier":   "ultimate",
		"startDate":      "2025-01-01",
	}

	// Test subscription validation
	err := subscriptionValidator(subscriptionData)
	if err != nil {
		t.Errorf("subscriptionValidator() failed for valid data: %v", err)
	}

	// Test configuration creation and validation
	config := NewConfiguration("subscription", subscriptionData)
	if config.Name != "subscription" {
		t.Errorf("Expected configuration name to be 'subscription', got %s", config.Name)
	}
	if config.Key != "#configuration#subscription" {
		t.Errorf("Expected configuration key to be '#configuration#subscription', got %s", config.Key)
	}

	err = config.Valid()
	if err != nil {
		t.Errorf("Configuration.Valid() failed: %v", err)
	}

	// Test subscription type casting and usage
	sub := Subscription(subscriptionData)

	// Test date checking
	testTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC) // June 15, 2025
	if !sub.Within(testTime) {
		t.Error("Expected test time to be within subscription period")
	}

	beforeStart := time.Date(2024, 12, 31, 12, 0, 0, 0, time.UTC)
	if sub.Within(beforeStart) {
		t.Error("Expected time before start to be outside subscription period")
	}

	afterEnd := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	if sub.Within(afterEnd) {
		t.Error("Expected time after end to be outside subscription period")
	}
}
