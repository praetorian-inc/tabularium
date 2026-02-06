package model

import (
	"testing"
)

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		// Valid statuses
		{"Empty string is valid", "", true},
		{"Deleted is valid", Deleted, true},
		{"Pending is valid", Pending, true},
		{"Active is valid", Active, true},
		{"Frozen is valid", Frozen, true},
		{"FrozenRejected is valid", FrozenRejected, true},
		{"ActiveLow is valid", ActiveLow, true},
		{"ActivePassive is valid", ActivePassive, true},
		{"ActiveHigh is valid", ActiveHigh, true},

		// Invalid statuses
		{"Invalid status 'X' is invalid", "X", false},
		{"Invalid status 'INVALID' is invalid", "INVALID", false},
		{"Invalid status 'active' (lowercase) is invalid", "active", false},
		{"Invalid status 'pending' (lowercase) is invalid", "pending", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidStatus(tt.status)
			if result != tt.expected {
				t.Errorf("IsValidStatus(%q) = %v, expected %v", tt.status, result, tt.expected)
			}
		})
	}
}

func TestAssetValidWithStatus(t *testing.T) {
	tests := []struct {
		name     string
		asset    *Asset
		expected bool
	}{
		{
			name: "Valid asset with Active status",
			asset: &Asset{
				BaseAsset: BaseAsset{
					Key:    "#asset#dns#example.com",
					Status: Active,
				},
			},
			expected: true,
		},
		{
			name: "Valid asset with empty status",
			asset: &Asset{
				BaseAsset: BaseAsset{
					Key:    "#asset#dns#example.com",
					Status: "",
				},
			},
			expected: true,
		},
		{
			name: "Invalid asset with invalid status",
			asset: &Asset{
				BaseAsset: BaseAsset{
					Key:    "#asset#dns#example.com",
					Status: "INVALID",
				},
			},
			expected: false,
		},
		{
			name: "Invalid asset with invalid key",
			asset: &Asset{
				BaseAsset: BaseAsset{
					Key:    "invalid-key",
					Status: Active,
				},
			},
			expected: false,
		},
		{
			name: "Valid asset with Pending status",
			asset: &Asset{
				BaseAsset: BaseAsset{
					Key:    "#asset#dns#example.com",
					Status: Pending,
				},
			},
			expected: true,
		},
		{
			name: "Valid asset with Deleted status",
			asset: &Asset{
				BaseAsset: BaseAsset{
					Key:    "#asset#dns#example.com",
					Status: Deleted,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.asset.Valid()
			if result != tt.expected {
				t.Errorf("Asset.Valid() = %v, expected %v for status %q", result, tt.expected, tt.asset.Status)
			}
		})
	}
}

func TestPreseedValidWithStatus(t *testing.T) {
	tests := []struct {
		name     string
		preseed  *Preseed
		expected bool
	}{
		{
			name: "Valid preseed with Active status",
			preseed: &Preseed{
				Key:    "#preseed#whois#registrant_email#test@example.com",
				Status: Active,
			},
			expected: true,
		},
		{
			name: "Valid preseed with empty status",
			preseed: &Preseed{
				Key:    "#preseed#whois#registrant_email#test@example.com",
				Status: "",
			},
			expected: true,
		},
		{
			name: "Invalid preseed with invalid status",
			preseed: &Preseed{
				Key:    "#preseed#whois#registrant_email#test@example.com",
				Status: "INVALID",
			},
			expected: false,
		},
		{
			name: "Invalid preseed with empty key",
			preseed: &Preseed{
				Key:    "",
				Status: Active,
			},
			expected: false,
		},
		{
			name: "Valid preseed with Pending status",
			preseed: &Preseed{
				Key:    "#preseed#whois#registrant_email#test@example.com",
				Status: Pending,
			},
			expected: true,
		},
		{
			name: "Valid preseed with FrozenRejected status",
			preseed: &Preseed{
				Key:    "#preseed#whois#registrant_email#test@example.com",
				Status: FrozenRejected,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.preseed.Valid()
			if result != tt.expected {
				t.Errorf("Preseed.Valid() = %v, expected %v for status %q", result, tt.expected, tt.preseed.Status)
			}
		})
	}
}
