package model

import (
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/model/beta"
	"github.com/stretchr/testify/assert"
)

func TestRisk_StateSeverity(t *testing.T) {
	tests := []struct {
		status   string
		state    string
		severity string
	}{
		{"TI", "T", "I"},
		{"D", "D", ""},
		{"", "", ""},
	}

	for _, test := range tests {
		t.Run(test.status, func(t *testing.T) {
			risk := NewRisk(&Asset{DNS: "test", Name: "test"}, "test", test.status)
			assert.Equal(t, test.state, risk.State())
			assert.Equal(t, test.severity, risk.Severity())
			assert.True(t, risk.Is(test.state), "expected Is(%s) to return true for %s", test.state, test.status)
		})
	}
}

func TestRisk_Set(t *testing.T) {
	tests := []struct {
		initial          string
		state            string
		expected         string
		expectedPriority int
	}{
		{OpenCritical, Remediated, RemediatedCritical, 0},
		{DeletedCriticalDuplicate, Open, OpenCritical, 0},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			risk := NewRisk(&Asset{DNS: "test", Name: "test"}, "test", test.initial)
			risk.Set(test.state)
			assert.Equal(t, test.expected, risk.Status)
			assert.Equal(t, test.expectedPriority, risk.Priority)
		})
	}
}

func TestRisk_MergePriority(t *testing.T) {
	tests := []struct {
		initial          string
		update           string
		expected         string
		expectedPriority int
	}{
		{OpenCritical, OpenLow, OpenLow, 30},
		{DeletedLowDuplicate, OpenHigh, OpenHigh, 10},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			risk := NewRisk(&Asset{DNS: "test", Name: "test"}, "test", test.initial)
			update := Risk{Status: test.update}
			risk.Merge(update)
			assert.Equal(t, test.expected, risk.Status)
			assert.Equal(t, test.expectedPriority, risk.Priority)
		})
	}
}

func TestRiskConstructors(t *testing.T) {
	testAsset := NewAsset("example.com", "Example Asset")
	testWebpage := NewWebpageFromString("https://gladiator.systems", nil)
	tests := []struct {
		name         string
		target       Target
		riskName     string
		expectedName string
		dns          string
	}{
		{
			name:     "Same DNS",
			target:   &testAsset,
			riskName: "test-risk",
			dns:      "example.com",
		},
		{
			name:     "Different DNS",
			target:   &testAsset,
			riskName: "test-risk",
			dns:      "subdomain.example.com",
		},
		{
			name:     "Same DNS",
			target:   &testAsset,
			riskName: "test-risk",
			dns:      "example.com",
		},
		{
			name:     "Different DNS",
			target:   &testAsset,
			riskName: "test-risk",
			dns:      "subdomain.example.com",
		},
		{
			name:     "Same DNS",
			target:   &testWebpage,
			riskName: "test-risk",
			dns:      "example.com",
		},
		{
			name:     "Different DNS",
			target:   &testWebpage,
			riskName: "test-risk",
			dns:      "subdomain.example.com",
		},
		{
			name:         "Format Name",
			target:       &testAsset,
			riskName:     "Test Risk",
			expectedName: "test-risk",
			dns:          "example.com",
		},
		{
			name:         "Format Name (CVE)",
			target:       &testAsset,
			riskName:     "CVE-2023-12345",
			expectedName: "CVE-2023-12345",
			dns:          "example.com",
		},
		{
			name:         "Format Name (CVE should be uppercase)",
			target:       &testAsset,
			riskName:     "cve-2023-12345",
			expectedName: "CVE-2023-12345",
			dns:          "example.com",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			risk1 := NewRisk(test.target, test.riskName, TriageInfo)
			assert.Equal(t, test.target.Group(), risk1.DNS, "NewRisk: DNS should match target group")

			risk2 := NewRiskWithDNS(test.target, test.riskName, test.dns, TriageInfo)
			assert.Equal(t, test.dns, risk2.DNS, "NewRiskWithDNS: DNS should match provided DNS")
			assert.Equal(t, risk1.Name, risk2.Name, "Names should match")
			assert.Equal(t, risk1.Status, risk2.Status, "Status should match")
			assert.Equal(t, risk1.Source, risk2.Source, "Source should match")
			assert.Equal(t, risk1.Target, risk2.Target, "Target should match")
		})
	}
}

func TestRisk_PendingAsset(t *testing.T) {
	originalAsset := NewAsset("example.com", "Example Asset")
	risk := NewRisk(&originalAsset, "test-risk", TriageInfo)

	pendingAsset, ok := risk.PendingAsset()
	if !ok {
		t.Errorf("expected PendingAsset to return a valid asset")
	}

	assert.Equal(t, Pending, pendingAsset.Status, "Status should be Pending")
	assert.Equal(t, originalAsset.Key, pendingAsset.Key, "Key should not change")
	assert.Equal(t, originalAsset.DNS, pendingAsset.DNS, "DNS should not change")
	assert.Equal(t, originalAsset.Name, pendingAsset.Name, "Name should not change")
	assert.Equal(t, originalAsset.Source, pendingAsset.Source, "Source should not change")
	assert.Equal(t, originalAsset.Created, pendingAsset.Created, "Created should not change")
	assert.Equal(t, originalAsset.Visited, pendingAsset.Visited, "Visited should not change")
	assert.Equal(t, originalAsset.TTL, pendingAsset.TTL, "TTL should not change")

	port := NewPort("tcp", 80, &originalAsset)
	portRisk := NewRisk(&port, "test-risk", TriageInfo)

	// True negative
	pendingAsset, ok = portRisk.PendingAsset()
	if ok {
		t.Errorf("expected PendingAsset to return false for port-based risk")
	}
}

func TestRisk_Valid(t *testing.T) {
	validRisk := NewRisk(&Asset{DNS: "test", Name: "test"}, "test", TriageInfo)
	missingKey := NewRisk(&Asset{DNS: "test", Name: "test"}, "test", TriageInfo)
	missingKey.Key = ""
	missingStatus := NewRisk(&Asset{DNS: "test", Name: "test"}, "test", "")
	missingName := NewRisk(&Asset{DNS: "test", Name: "test"}, "", TriageInfo)
	missingDNS := NewRiskWithDNS(&Asset{DNS: "test", Name: "test"}, "test", "", TriageInfo)

	assert.True(t, validRisk.Valid())
	assert.False(t, missingKey.Valid())
	assert.False(t, missingStatus.Valid())
	assert.False(t, missingName.Valid())
	assert.False(t, missingDNS.Valid())
}

type betaAsset struct {
	beta.Beta
	Asset
}

func NewBetaAsset(dns, name string) betaAsset {
	return betaAsset{
		Asset: NewAsset(dns, name),
	}
}

func TestRisk_IsBeta(t *testing.T) {
	normalAsset := NewAsset("example.com", "example.com")
	betaAsset := NewBetaAsset("example.com", "example.com")

	risk := NewRisk(&normalAsset, "test", TriageInfo)
	assert.False(t, risk.Beta)

	risk = NewRisk(&betaAsset, "test", TriageInfo)
	assert.True(t, risk.Beta)
}

func TestRisk_TagsVist(t *testing.T) {
	t.Run("tags become a unique set", func(t *testing.T) {
		original := Risk{Tags: Tags{Tags: []string{"tag1", "tag2"}}}
		update := Risk{Tags: Tags{Tags: []string{"tag2", "tag3"}}}
		original.Visit(update)
		assert.Equal(t, []string{"tag1", "tag2", "tag3"}, original.Tags.Tags)
	})

	t.Run("when specified empty, original tags are preserved", func(t *testing.T) {
		tags := []string{"tag1", "tag2"}
		original := Risk{Tags: Tags{Tags: tags}}
		update := Risk{Tags: Tags{Tags: []string{}}}
		original.Visit(update)
		assert.Equal(t, tags, original.Tags.Tags)
	})
}

func TestRisk_TagsMerge(t *testing.T) {
	t.Run("when specified, tags are overwritten", func(t *testing.T) {
		original := Risk{Tags: Tags{Tags: []string{"tag1", "tag2"}}}
		update := Risk{Tags: Tags{Tags: []string{"tag2", "tag3"}}}
		original.Merge(update)
		assert.Equal(t, update.Tags, original.Tags)
	})

	t.Run("when specified empty, tags are empty", func(t *testing.T) {
		original := Risk{Tags: Tags{Tags: []string{"tag1", "tag2"}}}
		update := Risk{Tags: Tags{Tags: []string{}}}
		original.Merge(update)
		assert.Equal(t, update.Tags, original.Tags)
	})

	t.Run("when unspecified, tags are preserved", func(t *testing.T) {
		tags := Tags{Tags: []string{"tag1", "tag2"}}
		original := Risk{Tags: tags}
		update := Risk{}
		original.Merge(update)
		assert.Equal(t, tags, original.Tags)
	})
}

func TestRisk_MergePreservesCreated(t *testing.T) {
	t.Run("Created is preserved when existing risk has Created and update has Created", func(t *testing.T) {
		existingRisk := NewRisk(&Asset{DNS: "test", Name: "test"}, "test-vuln", TriageInfo)
		originalCreated := "2023-01-01T00:00:00Z"
		existingRisk.Created = originalCreated

		update := Risk{
			Status:  OpenHigh,
			Created: "2023-12-31T23:59:59Z",
		}

		existingRisk.Merge(update)

		assert.Equal(t, originalCreated, existingRisk.Created)
		assert.NotEqual(t, update.Created, existingRisk.Created)
	})

	t.Run("Created is set when existing risk has no Created but update has Created", func(t *testing.T) {
		existingRisk := NewRisk(&Asset{DNS: "test", Name: "test"}, "test-vuln", TriageInfo)
		existingRisk.Created = ""

		updateCreated := "2023-06-15T12:00:00Z"
		update := Risk{
			Status:  OpenHigh,
			Created: updateCreated,
		}

		existingRisk.Merge(update)

		assert.Equal(t, updateCreated, existingRisk.Created)
	})

	t.Run("Created is preserved when existing risk has Created but update has empty Created", func(t *testing.T) {
		existingRisk := NewRisk(&Asset{DNS: "test", Name: "test"}, "test-vuln", TriageInfo)
		originalCreated := "2023-01-01T00:00:00Z"
		existingRisk.Created = originalCreated

		update := Risk{
			Status:  OpenHigh,
			Created: "",
		}

		existingRisk.Merge(update)

		assert.Equal(t, originalCreated, existingRisk.Created)
	})

	t.Run("Created is preserved when both existing and update have empty Created", func(t *testing.T) {
		existingRisk := NewRisk(&Asset{DNS: "test", Name: "test"}, "test-vuln", TriageInfo)
		existingRisk.Created = ""

		update := Risk{
			Status:  OpenHigh,
			Created: "",
		}

		existingRisk.Merge(update)

		assert.Empty(t, existingRisk.Created)
	})

	t.Run("Updated is set correctly when status changes", func(t *testing.T) {
		existingRisk := NewRisk(&Asset{DNS: "test", Name: "test"}, "test-vuln", TriageInfo)
		originalCreated := "2023-01-01T00:00:00Z"
		originalUpdated := "2023-01-02T00:00:00Z"
		existingRisk.Created = originalCreated
		existingRisk.Updated = originalUpdated

		update := Risk{
			Status:  OpenHigh,
			Created: "2023-12-31T23:59:59Z",
		}

		existingRisk.Merge(update)

		assert.Equal(t, originalCreated, existingRisk.Created)
		assert.NotEqual(t, originalUpdated, existingRisk.Updated)
		assert.NotEmpty(t, existingRisk.Updated)
	})

	t.Run("Created preservation does not affect other merge behavior", func(t *testing.T) {
		existingRisk := NewRisk(&Asset{DNS: "test", Name: "test"}, "test-vuln", TriageInfo)
		originalCreated := "2023-01-01T00:00:00Z"
		existingRisk.Created = originalCreated
		existingRisk.Status = TriageInfo

		update := Risk{
			Status:  OpenHigh,
			Created: "2023-12-31T23:59:59Z",
			Tags:    Tags{Tags: []string{"new-tag"}},
		}

		existingRisk.Merge(update)

		assert.Equal(t, originalCreated, existingRisk.Created)
		assert.Equal(t, OpenHigh, existingRisk.Status)
		assert.Equal(t, update.Tags, existingRisk.Tags)
		assert.NotEmpty(t, existingRisk.Updated)
	})

	t.Run("Merge with full Risk object from API", func(t *testing.T) {
		existingRisk := NewRisk(&Asset{DNS: "test", Name: "test"}, "test-vuln", TriageInfo)
		originalCreated := "2023-01-01T10:00:00Z"
		existingRisk.Created = originalCreated
		existingRisk.Updated = "2023-01-02T10:00:00Z"

		fullRiskUpdate := Risk{
			Key:     existingRisk.Key,
			Name:    existingRisk.Name,
			Status:  OpenHigh,
			Comment: "Updated comment",
			Created: "2023-12-31T23:59:59Z",
			Updated: "2023-12-31T23:59:59Z",
			Tags:    Tags{Tags: []string{"api-tag"}},
		}

		existingRisk.Merge(fullRiskUpdate)

		assert.Equal(t, originalCreated, existingRisk.Created)
		assert.Equal(t, OpenHigh, existingRisk.Status)
		assert.Equal(t, fullRiskUpdate.Tags, existingRisk.Tags)
		assert.NotEqual(t, fullRiskUpdate.Updated, existingRisk.Updated)
		assert.NotEmpty(t, existingRisk.Updated)
	})

	t.Run("Visit with Remediated risk triggers Set which calls Merge", func(t *testing.T) {
		existingRisk := NewRisk(&Asset{DNS: "test", Name: "test"}, "test-vuln", RemediatedHigh)
		originalCreated := "2023-01-01T00:00:00Z"
		existingRisk.Created = originalCreated
		existingRisk.Updated = "2023-01-02T00:00:00Z"
		existingRisk.Visited = "2023-01-02T00:00:00Z"

		newRisk := NewRisk(&Asset{DNS: "test", Name: "test"}, "test-vuln", TriageHigh)
		newRiskCreated := "2023-12-31T23:59:59Z"
		newRisk.Created = newRiskCreated
		newRisk.Visited = "2023-12-31T23:59:59Z"

		existingRisk.Visit(newRisk)

		assert.Equal(t, originalCreated, existingRisk.Created, "Created should be preserved when Visit triggers Set->Merge")
		assert.Equal(t, OpenHigh, existingRisk.Status, "Status should change from Remediated to Open")
		assert.Equal(t, newRisk.Visited, existingRisk.Visited, "Visited should be updated from new risk")
		assert.NotEmpty(t, existingRisk.Updated, "Updated should be set when status changes")
	})
}
