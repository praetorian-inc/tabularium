package model

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBaseAsset_Visit(t *testing.T) {
	tests := []struct {
		name       string
		baseAsset  BaseAsset
		otherAsset BaseAsset
		wantStatus string
		wantOrigin string
		wantTags   []string
		wantTTL    bool
	}{
		{
			name: "existing: frozen, other: active",
			baseAsset: BaseAsset{
				Status: Frozen,
				TTL:    0,
			},
			otherAsset: BaseAsset{
				Status:  Active,
				Secret:  &[]string{"secret"}[0],
				Visited: "other",
			},
			wantStatus: Frozen,
			wantTTL:    false,
		},
		{
			name: "existing: active zero TTL, other: active",
			baseAsset: BaseAsset{
				Status: Active,
				TTL:    0,
			},
			otherAsset: BaseAsset{
				Status:  Active,
				Secret:  &[]string{"secret"}[0],
				Visited: "other",
			},
			wantStatus: Active,
			wantTTL:    false,
		},
		{
			name: "existing: pending self source, other: active",
			baseAsset: BaseAsset{
				Status: Pending,
				TTL:    Future(24),
				Source: SelfSource,
			},
			otherAsset: BaseAsset{
				Status:  Active,
				Secret:  &[]string{"secret"}[0],
				Visited: "other",
				TTL:     22,
			},
			wantStatus: Active,
			wantTTL:    true,
		},
		{
			name: "existing: pending seed source, other: active",
			baseAsset: BaseAsset{
				Status: Pending,
				TTL:    Future(24),
				Source: SeedSource,
			},
			otherAsset: BaseAsset{
				Status:  Active,
				Secret:  &[]string{"secret"}[0],
				Visited: "other",
			},
			wantStatus: Pending,
			wantTTL:    true,
		},
		{
			name: "existing: pending account source, other: active",
			baseAsset: BaseAsset{
				Status: Pending,
				TTL:    Future(24),
				Source: AccountSource,
			},
			otherAsset: BaseAsset{
				Status:  Active,
				Secret:  &[]string{"secret"}[0],
				Visited: "other",
			},
			wantStatus: Pending,
			wantTTL:    true,
		},
		{
			name: "existing: active high self source, other: active",
			baseAsset: BaseAsset{
				Status: ActiveHigh,
				TTL:    Future(24),
				Source: SelfSource,
			},
			otherAsset: BaseAsset{
				Status:  Active,
				Secret:  &[]string{"newsecret"}[0],
				Visited: "other",
				TTL:     22,
			},
			wantStatus: ActiveHigh,
			wantTTL:    true,
		},
		{
			name: "existing: origin empty, other: origin set",
			baseAsset: BaseAsset{
				Origin: "",
			},
			otherAsset: BaseAsset{
				Origin: "other",
			},
			wantOrigin: "other",
		},
		{
			name: "existing: origin set, other: origin set",
			baseAsset: BaseAsset{
				Origin: "existing",
			},
			otherAsset: BaseAsset{
				Origin: "other",
			},
			wantOrigin: "existing",
		},
		{
			name: "tags become unique set",
			baseAsset: BaseAsset{
				Tags: Tags{Tags: []string{"tag1", "tag2"}},
			},
			otherAsset: BaseAsset{
				Tags: Tags{Tags: []string{"tag1", "tag3"}},
			},
			wantTags: []string{"tag1", "tag2", "tag3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startTTL := tt.baseAsset.TTL
			result := tt.baseAsset
			wrapper := &Asset{BaseAsset: tt.otherAsset}
			result.Visit(wrapper)

			if result.Status != tt.wantStatus {
				t.Errorf("Visit() status = %v, want %v", result.Status, tt.wantStatus)
			}

			if tt.wantTTL && result.TTL != tt.otherAsset.TTL {
				t.Errorf("Visit() TTL = %v, want other TTL %v", result.TTL, tt.otherAsset.TTL)
			}

			if !tt.wantTTL && result.TTL != startTTL {
				t.Errorf("Visit() TTL = %v, want %v", result.TTL, startTTL)
			}

			if result.Secret != tt.otherAsset.Secret {
				t.Errorf("Visit() secret = %v, want %v", result.Secret, tt.otherAsset.Secret)
			}

			if result.Visited != tt.otherAsset.Visited {
				t.Error("Visit() visited timestamp not set to other")
			}

			if tt.wantOrigin != "" && result.Origin != tt.wantOrigin {
				t.Errorf("Visit() origin = %v, want %v", result.Origin, tt.wantOrigin)
			}

			assert.Equal(t, tt.wantTags, result.Tags.Tags)
		})
	}
}

func TestBaseAsset_TagsMerge(t *testing.T) {
	t.Run("when specified, tags are overwritten", func(t *testing.T) {
		original := Asset{BaseAsset: BaseAsset{Tags: Tags{Tags: []string{"tag1", "tag2"}}}}
		update := Asset{BaseAsset: BaseAsset{Tags: Tags{Tags: []string{"tag2", "tag3"}}}}
		original.Merge(&update)
		assert.Equal(t, update.Tags, original.Tags)
	})

	t.Run("when specified empty, tags are empty", func(t *testing.T) {
		original := Asset{BaseAsset: BaseAsset{Tags: Tags{Tags: []string{"tag1", "tag2"}}}}
		update := Asset{BaseAsset: BaseAsset{Tags: Tags{Tags: []string{}}}}
		original.Merge(&update)
		assert.Equal(t, update.Tags, original.Tags)
	})

	t.Run("when unspecified, tags are preserved", func(t *testing.T) {
		tags := Tags{Tags: []string{"tag1", "tag2"}}
		original := Asset{BaseAsset: BaseAsset{Tags: tags}}
		update := Asset{BaseAsset: BaseAsset{}}
		original.Merge(&update)
		assert.Equal(t, tags, original.Tags)
	})
}

func TestMetadata_Merge(t *testing.T) {
	tests := []struct {
		name   string
		base   Metadata
		other  Metadata
		expect Metadata
	}{
		{
			name:  "merge empty base with empty other",
			base:  Metadata{},
			other: Metadata{},
		},
		{
			name: "merge empty base with populated other",
			base: Metadata{},
			other: Metadata{
				ASNumber: "AS12345",
			},
			expect: Metadata{
				ASNumber: "AS12345",
			},
		},
		{
			name: "merge populated base with empty other",
			base: Metadata{
				ASNumber: "AS12345",
			},
			other: Metadata{},
			expect: Metadata{
				ASNumber: "AS12345",
			},
		},
		{
			name: "merge populated base with populated slices",
			base: Metadata{
				ASNumber: "AS12345",
			},
			other: Metadata{
				OriginationData: OriginationData{
					AttackSurface: []string{"b", "c"},
				},
			},
			expect: Metadata{
				ASNumber: "AS12345",
				OriginationData: OriginationData{
					AttackSurface: []string{"b", "c"},
				},
			},
		},
		{
			name: "merge populated base with populated other slices",
			base: Metadata{
				ASNumber: "AS12345",
				OriginationData: OriginationData{
					AttackSurface: []string{"a", "b"},
				},
			},
			other: Metadata{},
			expect: Metadata{
				ASNumber: "AS12345",
				OriginationData: OriginationData{
					AttackSurface: []string{"a", "b"},
				},
			},
		},
		{
			name: "merge populated base with both populated slices",
			base: Metadata{
				ASNumber: "AS12345",
				OriginationData: OriginationData{
					AttackSurface: []string{"a", "b"},
				},
			},
			other: Metadata{
				OriginationData: OriginationData{
					AttackSurface: []string{"b", "c"},
				},
			},
			expect: Metadata{
				ASNumber: "AS12345",
				OriginationData: OriginationData{
					AttackSurface: []string{"b", "c"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.base.Merge(tt.other)
			assert.Equal(t, tt.expect, tt.base)
		})
	}
}

// Modify the TestAsset_MergeWithComments test to work with the new History model
func TestAsset_MergeWithComments(t *testing.T) {
	tests := []struct {
		name           string
		original       BaseAsset
		update         BaseAsset
		expectedStatus string
		expectedHist   []HistoryRecord
	}{
		{
			name: "Status change with comment",
			original: BaseAsset{
				Status: "A",
				History: History{
					History: []HistoryRecord{},
				},
			},
			update: BaseAsset{
				Status:  "D",
				Comment: "Deleting asset",
				Source:  "test",
			},
			expectedStatus: "D",
			expectedHist: []HistoryRecord{
				{
					From:    "A",
					To:      "D",
					By:      "test",
					Comment: "Deleting asset",
					Updated: Now(),
				},
			},
		},
		{
			name: "Comment only update",
			original: BaseAsset{
				Status: "A",
				History: History{
					History: []HistoryRecord{},
				},
			},
			update: BaseAsset{
				Comment: "Adding note",
				Source:  "test",
			},
			expectedStatus: "A",
			expectedHist: []HistoryRecord{
				{
					By:      "test",
					Comment: "Adding note",
					Updated: Now(),
				},
			},
		},
		{
			name: "Status update without comment",
			original: BaseAsset{
				Status: "A",
				History: History{
					History: []HistoryRecord{},
				},
			},
			update: BaseAsset{
				Status: "D",
				Source: "test",
			},
			expectedStatus: "D",
			expectedHist: []HistoryRecord{
				{
					From:    "A",
					To:      "D",
					By:      "test",
					Updated: Now(),
				},
			},
		},
		{
			name: "Remove history entry",
			original: BaseAsset{
				Status: "A",
				History: History{
					History: []HistoryRecord{
						{
							From:    "", // Changed: needs to NOT have From/To set to trigger removal
							To:      "",
							By:      "test",
							Comment: "First change",
							Updated: Now(),
						},
					},
				},
			},
			update: BaseAsset{
				Status:  "A",
				Source:  "test",
				History: History{Remove: &[]int{0}[0]},
			},
			expectedStatus: "A",
			expectedHist:   []HistoryRecord{},
		},

		{
			name: "Clear comment but keep status change",
			original: BaseAsset{
				Status: "A",
				History: History{
					History: []HistoryRecord{
						{
							From:    "A",
							To:      "B",
							By:      "test",
							Comment: "Test comment",
							Updated: Now(),
						},
					},
				},
			},
			update: BaseAsset{
				History: History{
					Remove: &[]int{0}[0],
				},
			},
			expectedStatus: "A",
			expectedHist: []HistoryRecord{
				{
					From:    "A",
					To:      "B",
					By:      "test",
					Comment: "",
					Updated: Now(),
				},
			},
		},
		{
			name: "Empty update",
			original: BaseAsset{
				Status: "A",
				History: History{
					History: []HistoryRecord{},
				},
			},
			update:         BaseAsset{},
			expectedStatus: "A",
			expectedHist:   []HistoryRecord{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapper := &Asset{BaseAsset: tt.update}
			tt.original.Merge(wrapper)

			assert.Equal(t, tt.original.Status, tt.expectedStatus, "Status = %v, want %v", tt.original.Status, tt.expectedStatus)
			assert.Equal(t, tt.original.History.History, tt.expectedHist, "History = %v, want %v", tt.original.History.History, tt.expectedHist)

			for i := range tt.expectedHist {
				actual := tt.original.History.History[i]
				expected := tt.expectedHist[i]

				assert.Equal(t, actual.From, expected.From, "History[%d].From = %v, want %v", i, actual.From, expected.From)
				assert.Equal(t, actual.To, expected.To, "History[%d].To = %v, want %v", i, actual.To, expected.To)
				assert.Equal(t, actual.Comment, expected.Comment, "History[%d].Comment = %v, want %v", i, actual.Comment, expected.Comment)
				assert.Equal(t, actual.By, expected.By, "History[%d].By = %v, want %v", i, actual.By, expected.By)
			}
		})
	}
}

func TestAsset_SetStatusFromLastSeen(t *testing.T) {
	tests := []struct {
		name       string
		lastSeen   string
		layout     string
		wantStatus string
	}{
		{
			name:       "RFC3339 within 24h",
			lastSeen:   time.Now().Add(-12 * time.Hour).Format(time.RFC3339),
			layout:     time.RFC3339,
			wantStatus: Active,
		},
		{
			name:       "RFC3339 older than 24h",
			lastSeen:   time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
			layout:     time.RFC3339,
			wantStatus: Pending,
		},
		{
			name:       "RFC1123 within 24h",
			lastSeen:   time.Now().Add(-12 * time.Hour).Format(time.RFC1123),
			layout:     time.RFC1123,
			wantStatus: Active,
		},
		{
			name:       "RFC1123 older than 24h",
			lastSeen:   time.Now().Add(-48 * time.Hour).Format(time.RFC1123),
			layout:     time.RFC1123,
			wantStatus: Pending,
		},
		{
			name:       "Invalid date format",
			lastSeen:   "invalid-date",
			layout:     time.RFC3339,
			wantStatus: Pending,
		},
		{
			name:       "Empty date string",
			lastSeen:   "",
			layout:     time.RFC3339,
			wantStatus: Pending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := BaseAsset{}
			a.SetStatusFromLastSeen(tt.lastSeen, tt.layout)

			assert.Equal(t, a.Status, tt.wantStatus, "SetStatusFromLastSeen() status = %v, want %v", a.Status, tt.wantStatus)
		})
	}
}

func TestAsset_IsStatus(t *testing.T) {
	tests := []struct {
		name  string
		asset BaseAsset
		value string
		want  bool
	}{
		{
			name: "matches status prefix",
			asset: BaseAsset{
				Status: ActiveHigh,
			},
			value: Active,
			want:  true,
		},
		{
			name: "matches exact status",
			asset: BaseAsset{
				Status: Pending,
			},
			value: Pending,
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.asset.IsStatus(tt.value)
			assert.Equal(t, tt.want, actual, "Asset.Is(%v) = %v, want %v", tt.value, actual, tt.want)
		})
	}
}

func TestMetadataVisit_EmptyBasePopulatedOther(t *testing.T) {
	base := Metadata{}
	other := Metadata{
		ASNumber: "AS12345",
		Country:  "USA",
		City:     "New York",
	}
	expected := Metadata{
		ASNumber: "AS12345",
		Country:  "USA",
		City:     "New York",
	}

	base.Visit(other)

	assert.Equal(t, expected, base)
}

func TestMetadataVisit_PopulatedBaseEmptyOther(t *testing.T) {
	base := Metadata{
		ASNumber:   "AS12345",
		Country:    "USA",
		City:       "New York",
		Registrant: "John Doe",
	}
	other := Metadata{}
	expected := Metadata{
		ASNumber:   "AS12345",
		Country:    "USA",
		City:       "New York",
		Registrant: "John Doe",
	}

	base.Visit(other)

	assert.Equal(t, expected, base)
}

func TestMetadataVisit_MergeOverlappingFields(t *testing.T) {
	base := Metadata{
		ASNumber:   "AS12345",
		Country:    "USA",
		City:       "New York",
		Registrant: "John Doe",
	}
	other := Metadata{
		ASNumber:  "AS67890",         // This should override
		ASName:    "Example Network", // This should be added
		Country:   "USA",             // Same value, should remain
		Province:  "NY",              // This should be added
		Purchased: "2023-01-01",      // This should be added
	}
	expected := Metadata{
		ASNumber:   "AS67890",         // Updated
		ASName:     "Example Network", // Added
		Country:    "USA",             // Unchanged
		Province:   "NY",              // Added
		City:       "New York",        // Unchanged
		Registrant: "John Doe",        // Unchanged
		Purchased:  "2023-01-01",      // Added
	}

	base.Visit(other)

	assert.Equal(t, expected, base)
}

func TestMetadataVisit_AllFields(t *testing.T) {
	base := Metadata{}
	other := Metadata{
		ASNumber:   "AS12345",
		ASName:     "Example Network",
		ASRange:    "192.168.0.0/16",
		Country:    "USA",
		Province:   "NY",
		City:       "New York",
		Purchased:  "2023-01-01",
		Updated:    "2023-06-01",
		Expiration: "2024-01-01",
		Registrant: "John Doe",
		Registrar:  "Example Registrar",
	}
	expected := Metadata{
		ASNumber:   "AS12345",
		ASName:     "Example Network",
		ASRange:    "192.168.0.0/16",
		Country:    "USA",
		Province:   "NY",
		City:       "New York",
		Purchased:  "2023-01-01",
		Updated:    "2023-06-01",
		Expiration: "2024-01-01",
		Registrant: "John Doe",
		Registrar:  "Example Registrar",
	}

	base.Visit(other)

	assert.Equal(t, expected, base)
}

func TestMetadataVisit_EmptyStringsDoNotOverride(t *testing.T) {
	base := Metadata{
		ASNumber: "AS12345",
		Country:  "USA",
	}
	other := Metadata{
		ASNumber: "",
		Country:  "",
	}
	expected := Metadata{
		ASNumber: "AS12345",
		Country:  "USA",
	}

	base.Visit(other)

	assert.Equal(t, expected, base)
}

func TestMetadata_VisitOrigin(t *testing.T) {
	tests := []struct {
		name   string
		base   Metadata
		other  Metadata
		expect Metadata
	}{
		{
			name: "append new origin values",
			base: Metadata{
				ASNumber: "1234",
				OriginationData: OriginationData{
					Origins: []string{"a", "b"},
				},
			},
			other: Metadata{
				OriginationData: OriginationData{
					Origins: []string{"b", "c"},
				},
			},
			expect: Metadata{
				ASNumber: "1234",
				OriginationData: OriginationData{
					Origins: []string{"a", "b", "c"},
				},
			},
		},
		{
			name: "empty origin values",
			base: Metadata{
				ASNumber: "1234",
				OriginationData: OriginationData{
					Origins: []string{},
				},
			},
			other: Metadata{
				OriginationData: OriginationData{
					Origins: []string{},
				},
			},
			expect: Metadata{
				ASNumber: "1234",
				OriginationData: OriginationData{
					Origins: []string{},
				},
			},
		},
		{
			name: "other origin is nil",
			base: Metadata{
				ASNumber: "1234",
				OriginationData: OriginationData{
					Origins: []string{"a", "b"},
				},
			},
			other: Metadata{
				ASNumber: "5678",
			},
			expect: Metadata{
				ASNumber: "5678",
				OriginationData: OriginationData{
					Origins: []string{"a", "b"},
				},
			},
		},
		{
			name: "base origin is nil",
			base: Metadata{
				ASNumber: "1234",
			},
			other: Metadata{
				OriginationData: OriginationData{
					Origins: []string{"a", "b"},
				},
			},
			expect: Metadata{
				ASNumber: "1234",
				OriginationData: OriginationData{
					Origins: []string{"a", "b"},
				},
			},
		},
		{
			name: "regular string fields update",
			base: Metadata{
				ASNumber:  "1234",
				ASName:    "old name",
				Registrar: "old registrar",
				OriginationData: OriginationData{
					Origins: []string{"a"},
				},
			},
			other: Metadata{
				ASName:   "new name",
				Province: "new province",
				OriginationData: OriginationData{
					Origins: []string{"b"},
				},
			},
			expect: Metadata{
				ASNumber:  "1234",
				ASName:    "new name",
				Province:  "new province",
				Registrar: "old registrar",
				OriginationData: OriginationData{
					Origins: []string{"a", "b"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.base.Visit(tt.other)

			// Test regular string fields
			v := reflect.ValueOf(&tt.base).Elem()
			expectV := reflect.ValueOf(&tt.expect).Elem()

			for i := 0; i < v.NumField(); i++ {
				field := v.Field(i)
				expectField := expectV.Field(i)

				if field.Kind() == reflect.String {
					if field.String() != expectField.String() {
						t.Errorf("field %s = %v, want %v",
							v.Type().Field(i).Name,
							field.String(),
							expectField.String())
					}
				}
			}

			// Test origin slice specifically
			// Convert slices to maps for comparison since order doesn't matter
			gotOrigin := make(map[string]bool)
			for _, s := range tt.base.OriginationData.Origins {
				gotOrigin[s] = true
			}
			expectOrigin := make(map[string]bool)
			for _, s := range tt.expect.OriginationData.Origins {
				expectOrigin[s] = true
			}

			assert.Equal(t, expectOrigin, gotOrigin)
		})
	}
}

func TestBaseAsset_GetPartitionKey(t *testing.T) {
	tests := []struct {
		name      string
		baseAsset BaseAsset
		wantKey   string
	}{
		{
			name: "returns identifier when set",
			baseAsset: BaseAsset{
				Identifier: "192.168.1.1",
				Group:      "example.com",
			},
			wantKey: "192.168.1.1",
		},
		{
			name: "returns identifier for different group, same identifier",
			baseAsset: BaseAsset{
				Identifier: "192.168.1.1",
				Group:      "different.com",
			},
			wantKey: "192.168.1.1",
		},
		{
			name: "returns empty string when identifier not set",
			baseAsset: BaseAsset{
				Group: "example.com",
			},
			wantKey: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.baseAsset.GetPartitionKey()
			assert.Equal(t, tt.wantKey, got)
		})
	}
}

func TestAsset_GetPartitionKey_SameIP(t *testing.T) {
	// Test that different DNS names on the same IP resolve to the same partition
	asset1 := NewAsset("example.com", "192.168.1.1")
	asset2 := NewAsset("test.com", "192.168.1.1")
	asset3 := NewAsset("another.com", "192.168.1.1")

	partition1 := asset1.GetPartitionKey()
	partition2 := asset2.GetPartitionKey()
	partition3 := asset3.GetPartitionKey()

	// All should have the same partition key (the IP address)
	assert.Equal(t, "192.168.1.1", partition1, "asset1 should use IP as partition key")
	assert.Equal(t, "192.168.1.1", partition2, "asset2 should use IP as partition key")
	assert.Equal(t, "192.168.1.1", partition3, "asset3 should use IP as partition key")
	assert.Equal(t, partition1, partition2, "asset1 and asset2 should have same partition")
	assert.Equal(t, partition1, partition3, "asset1 and asset3 should have same partition")
}

func TestAsset_GetPartitionKey_DifferentAssetTypes(t *testing.T) {
	// Test various asset types
	ipv4Asset := NewAsset("example.com", "192.168.1.1")
	domainAsset := NewAsset("example.com", "example.com")
	cidrAsset := NewAsset("10.0.0.0/8", "10.0.0.0/8")

	assert.Equal(t, "192.168.1.1", ipv4Asset.GetPartitionKey(), "IPv4 asset should use IP")
	assert.Equal(t, "example.com", domainAsset.GetPartitionKey(), "Domain asset should use domain")
	assert.Equal(t, "10.0.0.0/8", cidrAsset.GetPartitionKey(), "CIDR asset should use CIDR")

	// Verify different IPs get different partitions
	ipv4Asset2 := NewAsset("different.com", "10.0.0.1")
	assert.NotEqual(t, ipv4Asset.GetPartitionKey(), ipv4Asset2.GetPartitionKey(),
		"Different IPs should have different partitions")
}

func TestPartitioned_IntegrationSameIPDifferentDNSAndPorts(t *testing.T) {
	// Integration test: Verify that the Partitioned interface works correctly
	// across multiple assets and ports on the same IP address.
	// This simulates a real-world scenario where multiple domains resolve to
	// the same IP and have various ports open.

	sharedIP := "192.168.1.1"

	// Create three different DNS names pointing to the same IP
	asset1 := NewAsset("example.com", sharedIP)
	asset2 := NewAsset("test.com", sharedIP)
	asset3 := NewAsset("api.example.com", sharedIP)

	// Create multiple ports on each asset
	port1_80 := NewPort("tcp", 80, &asset1)
	port1_443 := NewPort("tcp", 443, &asset1)

	port2_80 := NewPort("tcp", 80, &asset2)
	port2_443 := NewPort("tcp", 443, &asset2)
	port2_8080 := NewPort("tcp", 8080, &asset2)

	port3_22 := NewPort("tcp", 22, &asset3)
	port3_443 := NewPort("tcp", 443, &asset3)

	// Collect all partition keys
	assetPartitions := []string{
		asset1.GetPartitionKey(),
		asset2.GetPartitionKey(),
		asset3.GetPartitionKey(),
	}

	portPartitions := []string{
		port1_80.GetPartitionKey(),
		port1_443.GetPartitionKey(),
		port2_80.GetPartitionKey(),
		port2_443.GetPartitionKey(),
		port2_8080.GetPartitionKey(),
		port3_22.GetPartitionKey(),
		port3_443.GetPartitionKey(),
	}

	// Verify all assets partition by the shared IP
	for i, partition := range assetPartitions {
		assert.Equal(t, sharedIP, partition,
			"Asset %d should partition by shared IP %s", i+1, sharedIP)
	}

	// Verify all ports partition by the shared IP
	for i, partition := range portPartitions {
		assert.Equal(t, sharedIP, partition,
			"Port %d should partition by shared IP %s", i+1, sharedIP)
	}

	// Verify all partition keys are identical
	allPartitions := append(assetPartitions, portPartitions...)
	for i := 1; i < len(allPartitions); i++ {
		assert.Equal(t, allPartitions[0], allPartitions[i],
			"All assets and ports on the same IP should have identical partition keys")
	}

	// Verify interface compliance
	var _ Partitioned = &asset1
	var _ Partitioned = &port1_80

	// Test that a different IP produces a different partition
	differentAsset := NewAsset("different.com", "10.0.0.1")
	differentPort := NewPort("tcp", 80, &differentAsset)

	assert.NotEqual(t, sharedIP, differentAsset.GetPartitionKey(),
		"Different IP should have different partition")
	assert.NotEqual(t, sharedIP, differentPort.GetPartitionKey(),
		"Port on different IP should have different partition")
}

func TestPartitioned_InterfaceTypeAssertion(t *testing.T) {
	// Test that the Partitioned interface can be used with type assertions
	asset := NewAsset("example.com", "192.168.1.1")

	// Test as interface{}
	var model interface{} = &asset
	partitioned, ok := model.(Partitioned)
	assert.True(t, ok, "Asset should implement Partitioned interface")
	assert.Equal(t, "192.168.1.1", partitioned.GetPartitionKey())

	// Test with Port
	port := NewPort("tcp", 80, &asset)
	var portModel interface{} = &port
	portPartitioned, ok := portModel.(Partitioned)
	assert.True(t, ok, "Port should implement Partitioned interface")
	assert.Equal(t, "192.168.1.1", portPartitioned.GetPartitionKey())

	// Test fallback behavior for types without Partitioned
	// (this would be tested if we had a type that implements HasKey but not Partitioned)
}
