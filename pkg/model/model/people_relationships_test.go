package model

import (
	"encoding/json"
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersonWorksFor_NewPersonWorksFor(t *testing.T) {
	person := NewPerson("John Smith")
	person.SetCurrentCompany("Acme Corp")

	org := Asset{
		BaseAsset: BaseAsset{
			Key: "#organization#acmecorp",
		},
		DNS: "acme.com",
	}
	org.Defaulted()

	rel := NewPersonWorksFor(&person, &org).(*PersonWorksFor)

	source, target := rel.Nodes()
	assert.Equal(t, person.GetKey(), source.GetKey())
	assert.Equal(t, org.GetKey(), target.GetKey())
	assert.Equal(t, WorksForType, rel.Label())
}

func TestPersonWorksFor_Valid(t *testing.T) {
	person := NewPerson("John Smith")
	org := Asset{BaseAsset: BaseAsset{Key: "#org#acme"}, DNS: "acme.com"}
	org.Defaulted()

	tests := []struct {
		name     string
		rel      *PersonWorksFor
		expected bool
	}{
		{
			name:     "valid relationship",
			rel:      NewPersonWorksFor(&person, &org).(*PersonWorksFor),
			expected: true,
		},
		{
			name: "invalid employment type",
			rel: func() *PersonWorksFor {
				rel := NewPersonWorksFor(&person, &org).(*PersonWorksFor)
				rel.EmploymentType = "invalid"
				return rel
			}(),
			expected: false,
		},
		{
			name: "invalid employment status",
			rel: func() *PersonWorksFor {
				rel := NewPersonWorksFor(&person, &org).(*PersonWorksFor)
				rel.EmploymentStatus = "invalid"
				return rel
			}(),
			expected: false,
		},
		{
			name: "invalid confidence - too low",
			rel: func() *PersonWorksFor {
				rel := NewPersonWorksFor(&person, &org).(*PersonWorksFor)
				rel.Confidence = -1
				return rel
			}(),
			expected: false,
		},
		{
			name: "invalid confidence - too high",
			rel: func() *PersonWorksFor {
				rel := NewPersonWorksFor(&person, &org).(*PersonWorksFor)
				rel.Confidence = 101
				return rel
			}(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.rel.Valid())
		})
	}
}

func TestPersonWorksFor_GetSeniorityLevel(t *testing.T) {
	tests := []struct {
		jobTitle string
		expected string
	}{
		{"Software Engineering Intern", SeniorityEntry},
		{"Junior Developer", SeniorityEntry},
		{"Associate Software Engineer", SeniorityEntry},
		{"Software Engineer", SeniorityMid},
		{"Senior Software Engineer", SenioritSenior},
		{"Lead Developer", SenioritSenior},
		{"Principal Engineer", SenioritSenior},
		{"Staff Engineer", SenioritSenior},
		{"Engineering Manager", SeniorityExecutive},
		{"Director of Engineering", SeniorityExecutive},
		{"VP of Engineering", SeniorityExecutive},
		{"Chief Technology Officer", SeniorityCLevel},
		{"CEO", SeniorityCLevel},
		{"CTO", SeniorityCLevel},
	}

	person := NewPerson("Test Person")
	org := Asset{BaseAsset: BaseAsset{Key: "#org#company"}, DNS: "company.com"}
	org.Defaulted()

	for _, tt := range tests {
		t.Run(tt.jobTitle, func(t *testing.T) {
			rel := NewPersonWorksFor(&person, &org).(*PersonWorksFor)
			rel.JobTitle = tt.jobTitle
			result := rel.GetSeniorityLevel()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPersonWorksFor_HasDecisionMakingPower(t *testing.T) {
	tests := []struct {
		name               string
		jobTitle           string
		hasBudgetAuthority bool
		expected           bool
	}{
		{
			name:               "budget authority",
			jobTitle:           "Software Engineer",
			hasBudgetAuthority: true,
			expected:           true,
		},
		{
			name:               "manager title",
			jobTitle:           "Engineering Manager",
			hasBudgetAuthority: false,
			expected:           true,
		},
		{
			name:               "director title",
			jobTitle:           "Director of Product",
			hasBudgetAuthority: false,
			expected:           true,
		},
		{
			name:               "VP title",
			jobTitle:           "VP of Sales",
			hasBudgetAuthority: false,
			expected:           true,
		},
		{
			name:               "chief title",
			jobTitle:           "Chief Technology Officer",
			hasBudgetAuthority: false,
			expected:           true,
		},
		{
			name:               "head title",
			jobTitle:           "Head of Security",
			hasBudgetAuthority: false,
			expected:           true,
		},
		{
			name:               "lead title",
			jobTitle:           "Tech Lead",
			hasBudgetAuthority: false,
			expected:           true,
		},
		{
			name:               "regular employee",
			jobTitle:           "Software Engineer",
			hasBudgetAuthority: false,
			expected:           false,
		},
	}

	person := NewPerson("Test Person")
	org := Asset{BaseAsset: BaseAsset{Key: "#org#company"}, DNS: "company.com"}
	org.Defaulted()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel := NewPersonWorksFor(&person, &org).(*PersonWorksFor)
			rel.JobTitle = tt.jobTitle
			rel.HasBudgetAuthority = tt.hasBudgetAuthority
			result := rel.HasDecisionMakingPower()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPersonWorksFor_IsCurrentEmployment(t *testing.T) {
	person := NewPerson("John Smith")
	org := Asset{BaseAsset: BaseAsset{Key: "#org#acme"}, DNS: "acme.com"}
	org.Defaulted()

	rel := NewPersonWorksFor(&person, &org).(*PersonWorksFor)
	registry.CallHooks(rel) // Apply defaults

	// Should be current by default
	assert.True(t, rel.IsCurrentEmployment())

	// Not current if not active role
	rel.IsCurrentRole = false
	assert.False(t, rel.IsCurrentEmployment())

	// Not current if inactive status
	rel.IsCurrentRole = true
	rel.EmploymentStatus = EmploymentTerminated
	assert.False(t, rel.IsCurrentEmployment())
}

func TestPersonReportsTo_NewPersonReportsTo(t *testing.T) {
	subordinate := NewPerson("John Smith")
	subordinate.SetCurrentCompany("Acme Corp")

	manager := NewPerson("Jane Manager")
	manager.SetCurrentCompany("Acme Corp")

	rel := NewPersonReportsTo(&subordinate, &manager).(*PersonReportsTo)
	registry.CallHooks(rel) // Apply defaults

	source, target := rel.Nodes()
	assert.Equal(t, subordinate.GetKey(), source.GetKey())
	assert.Equal(t, manager.GetKey(), target.GetKey())
	assert.Equal(t, ReportsToType, rel.Label())
	assert.Equal(t, ReportingDirect, rel.ReportingType)
	assert.Equal(t, 50, rel.Confidence)
	assert.Equal(t, 5, rel.InfluenceLevel)
	assert.True(t, rel.IsActive)
}

func TestPersonReportsTo_Valid(t *testing.T) {
	subordinate := NewPerson("Sub Person")
	manager := NewPerson("Manager Person")

	tests := []struct {
		name     string
		rel      *PersonReportsTo
		expected bool
	}{
		{
			name:     "valid relationship",
			rel:      NewPersonReportsTo(&subordinate, &manager).(*PersonReportsTo),
			expected: true,
		},
		{
			name: "invalid reporting type",
			rel: func() *PersonReportsTo {
				rel := NewPersonReportsTo(&subordinate, &manager).(*PersonReportsTo)
				rel.ReportingType = "invalid"
				return rel
			}(),
			expected: false,
		},
		{
			name: "invalid influence level - too low",
			rel: func() *PersonReportsTo {
				rel := NewPersonReportsTo(&subordinate, &manager).(*PersonReportsTo)
				rel.InfluenceLevel = -1
				return rel
			}(),
			expected: false,
		},
		{
			name: "invalid influence level - too high",
			rel: func() *PersonReportsTo {
				rel := NewPersonReportsTo(&subordinate, &manager).(*PersonReportsTo)
				rel.InfluenceLevel = 11
				return rel
			}(),
			expected: false,
		},
		{
			name: "invalid confidence - too low",
			rel: func() *PersonReportsTo {
				rel := NewPersonReportsTo(&subordinate, &manager).(*PersonReportsTo)
				rel.Confidence = -1
				return rel
			}(),
			expected: false,
		},
		{
			name: "invalid confidence - too high",
			rel: func() *PersonReportsTo {
				rel := NewPersonReportsTo(&subordinate, &manager).(*PersonReportsTo)
				rel.Confidence = 101
				return rel
			}(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.rel.Valid())
		})
	}
}

func TestPersonReportsTo_IsDirectReport(t *testing.T) {
	subordinate := NewPerson("Sub Person")
	manager := NewPerson("Manager Person")

	rel := NewPersonReportsTo(&subordinate, &manager).(*PersonReportsTo)
	registry.CallHooks(rel) // Apply defaults

	// Should be direct report by default
	assert.True(t, rel.IsDirectReport())

	// Not direct if not active
	rel.IsActive = false
	assert.False(t, rel.IsDirectReport())

	// Not direct if indirect reporting
	rel.IsActive = true
	rel.ReportingType = ReportingIndirect
	assert.False(t, rel.IsDirectReport())
}

func TestPersonReportsTo_GetRelationshipStrength(t *testing.T) {
	tests := []struct {
		influenceLevel int
		expected       string
	}{
		{10, "high"},
		{9, "high"},
		{8, "high"},
		{7, "medium"},
		{6, "medium"},
		{5, "medium"},
		{4, "low"},
		{3, "low"},
		{2, "low"},
		{1, "low"},
	}

	subordinate := NewPerson("Sub Person")
	manager := NewPerson("Manager Person")

	for _, tt := range tests {
		t.Run("influence_"+string(rune(tt.influenceLevel+'0')), func(t *testing.T) {
			rel := NewPersonReportsTo(&subordinate, &manager).(*PersonReportsTo)
			rel.InfluenceLevel = tt.influenceLevel
			result := rel.GetRelationshipStrength()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPersonReportsTo_HasRegularInteraction(t *testing.T) {
	tests := []struct {
		name             string
		meetingFrequency string
		influenceLevel   int
		expected         bool
	}{
		{
			name:             "daily meetings",
			meetingFrequency: "daily",
			influenceLevel:   5,
			expected:         true,
		},
		{
			name:             "weekly meetings",
			meetingFrequency: "weekly",
			influenceLevel:   5,
			expected:         true,
		},
		{
			name:             "bi-weekly with high influence",
			meetingFrequency: "bi-weekly",
			influenceLevel:   7,
			expected:         true,
		},
		{
			name:             "bi-weekly with low influence",
			meetingFrequency: "bi-weekly",
			influenceLevel:   4,
			expected:         false,
		},
		{
			name:             "monthly meetings",
			meetingFrequency: "monthly",
			influenceLevel:   8,
			expected:         false,
		},
	}

	subordinate := NewPerson("Sub Person")
	manager := NewPerson("Manager Person")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel := NewPersonReportsTo(&subordinate, &manager).(*PersonReportsTo)
			rel.MeetingFrequency = tt.meetingFrequency
			rel.InfluenceLevel = tt.influenceLevel
			result := rel.HasRegularInteraction()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPersonWorksFor_JSONSerialization(t *testing.T) {
	person := NewPerson("John Smith")
	org := Asset{BaseAsset: BaseAsset{Key: "#organization#acmecorp"}, DNS: "acme.com"}
	org.Defaulted()

	rel := NewPersonWorksFor(&person, &org).(*PersonWorksFor)
	rel.JobTitle = "Senior Software Engineer"
	rel.Department = "Engineering"
	rel.EmploymentType = EmploymentFullTime
	rel.EmploymentStatus = EmploymentActive
	rel.StartDate = "2022-01-15"
	rel.SalaryRange = "$100,000-$130,000"
	rel.EquityGrant = true
	rel.HasBudgetAuthority = true
	rel.AccessLevel = AccessLevelAdmin
	rel.SecurityClearance = "secret"
	rel.Source = "linkedin"
	rel.Confidence = 85

	// Marshal to JSON
	data, err := json.Marshal(rel)
	require.NoError(t, err)

	// Unmarshal back
	var unmarshaled PersonWorksFor
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify key fields
	assert.Equal(t, rel.JobTitle, unmarshaled.JobTitle)
	assert.Equal(t, rel.Department, unmarshaled.Department)
	assert.Equal(t, rel.EmploymentType, unmarshaled.EmploymentType)
	assert.Equal(t, rel.EmploymentStatus, unmarshaled.EmploymentStatus)
	assert.Equal(t, rel.StartDate, unmarshaled.StartDate)
	assert.Equal(t, rel.SalaryRange, unmarshaled.SalaryRange)
	assert.Equal(t, rel.EquityGrant, unmarshaled.EquityGrant)
	assert.Equal(t, rel.HasBudgetAuthority, unmarshaled.HasBudgetAuthority)
	assert.Equal(t, rel.AccessLevel, unmarshaled.AccessLevel)
	assert.Equal(t, rel.SecurityClearance, unmarshaled.SecurityClearance)
	assert.Equal(t, rel.Confidence, unmarshaled.Confidence)
}

func TestPersonReportsTo_JSONSerialization(t *testing.T) {
	subordinate := NewPerson("John Smith")
	manager := NewPerson("Jane Manager")

	rel := NewPersonReportsTo(&subordinate, &manager).(*PersonReportsTo)
	rel.ReportingType = ReportingDirect
	rel.StartDate = "2022-06-01"
	rel.Organization = "Acme Corporation"
	rel.Department = "Engineering"
	rel.MeetingFrequency = "weekly"
	rel.InfluenceLevel = 8
	rel.Source = "org-chart"
	rel.Confidence = 95

	// Marshal to JSON
	data, err := json.Marshal(rel)
	require.NoError(t, err)

	// Unmarshal back
	var unmarshaled PersonReportsTo
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify key fields
	assert.Equal(t, rel.ReportingType, unmarshaled.ReportingType)
	assert.Equal(t, rel.StartDate, unmarshaled.StartDate)
	assert.Equal(t, rel.Organization, unmarshaled.Organization)
	assert.Equal(t, rel.Department, unmarshaled.Department)
	assert.Equal(t, rel.MeetingFrequency, unmarshaled.MeetingFrequency)
	assert.Equal(t, rel.InfluenceLevel, unmarshaled.InfluenceLevel)
	assert.Equal(t, rel.Confidence, unmarshaled.Confidence)
}

// Note: Relationship unmarshalling tests are commented out because the registry system
// doesn't currently support GraphRelationship unmarshalling in the same way as GraphModel.
// The core relationship functionality is tested through direct instantiation.

/*
func TestPersonWorksFor_Unmarshall(t *testing.T) {
	// Skipped: Registry system doesn't support GraphRelationship unmarshalling
}

func TestPersonReportsTo_Unmarshall(t *testing.T) {
	// Skipped: Registry system doesn't support GraphRelationship unmarshalling
}
*/

func TestValidationHelpers(t *testing.T) {
	// Test employment status validation
	assert.True(t, IsValidEmploymentStatus(EmploymentActive))
	assert.True(t, IsValidEmploymentStatus(EmploymentInactive))
	assert.True(t, IsValidEmploymentStatus(EmploymentTerminated))
	assert.False(t, IsValidEmploymentStatus("invalid"))

	// Test employment type validation
	assert.True(t, IsValidEmploymentType(EmploymentFullTime))
	assert.True(t, IsValidEmploymentType(EmploymentPartTime))
	assert.True(t, IsValidEmploymentType(EmploymentContract))
	assert.False(t, IsValidEmploymentType("invalid"))

	// Test reporting type validation
	assert.True(t, IsValidReportingType(ReportingDirect))
	assert.True(t, IsValidReportingType(ReportingIndirect))
	assert.True(t, IsValidReportingType(ReportingFunctional))
	assert.False(t, IsValidReportingType("invalid"))
}

func TestPersonWorksFor_PrivacyAndSecurity(t *testing.T) {
	person := NewPerson("John Smith")
	org := Asset{BaseAsset: BaseAsset{Key: "#org#acme"}, DNS: "acme.com"}
	org.Defaulted()

	rel := NewPersonWorksFor(&person, &org).(*PersonWorksFor)

	// Test security-related fields
	rel.SecurityClearance = "top-secret"
	rel.AccessLevel = AccessLevelPrivileged
	rel.HasBudgetAuthority = true

	assert.True(t, rel.Valid())
	assert.Equal(t, "top-secret", rel.SecurityClearance)
	assert.Equal(t, AccessLevelPrivileged, rel.AccessLevel)
	assert.True(t, rel.HasBudgetAuthority)
	assert.True(t, rel.HasDecisionMakingPower())
}

func TestPersonReportsTo_PrivacyAndSecurity(t *testing.T) {
	subordinate := NewPerson("Sub Person")
	manager := NewPerson("Manager Person")

	rel := NewPersonReportsTo(&subordinate, &manager).(*PersonReportsTo)

	// Test relationship strength assessment
	rel.InfluenceLevel = 9
	rel.MeetingFrequency = "daily"

	assert.True(t, rel.Valid())
	assert.Equal(t, "high", rel.GetRelationshipStrength())
	assert.True(t, rel.HasRegularInteraction())
}

func TestPersonWorksFor_EdgeCases(t *testing.T) {
	// Test with minimal data
	person := NewPerson("Minimal Person")
	org := Asset{BaseAsset: BaseAsset{Key: "#org#test"}, DNS: "test.com"}
	org.Defaulted()

	rel := NewPersonWorksFor(&person, &org).(*PersonWorksFor)
	registry.CallHooks(rel)

	assert.True(t, rel.Valid())
	assert.Equal(t, EmploymentActive, rel.EmploymentStatus)
	assert.Equal(t, EmploymentFullTime, rel.EmploymentType)
	assert.Equal(t, 50, rel.Confidence)

	// Test boundary values for confidence
	rel.Confidence = 0
	assert.True(t, rel.Valid())
	rel.Confidence = 100
	assert.True(t, rel.Valid())
}

func TestPersonReportsTo_EdgeCases(t *testing.T) {
	// Test with minimal data
	subordinate := NewPerson("Sub Person")
	manager := NewPerson("Manager Person")

	rel := NewPersonReportsTo(&subordinate, &manager).(*PersonReportsTo)
	registry.CallHooks(rel)

	assert.True(t, rel.Valid())
	assert.Equal(t, ReportingDirect, rel.ReportingType)
	assert.Equal(t, 50, rel.Confidence)
	assert.Equal(t, 5, rel.InfluenceLevel)

	// Test boundary values
	rel.InfluenceLevel = 0
	assert.True(t, rel.Valid())
	rel.InfluenceLevel = 10
	assert.True(t, rel.Valid())

	rel.Confidence = 0
	assert.True(t, rel.Valid())
	rel.Confidence = 100
	assert.True(t, rel.Valid())
}

func TestComplexEmploymentScenario(t *testing.T) {
	// Create a complex employment scenario
	person := NewPerson("Sarah Johnson")
	person.SetCurrentCompany("Acme Corporation")

	org := Asset{BaseAsset: BaseAsset{Key: "#organization#acmecorp"}, DNS: "acme.com"}
	org.Defaulted()

	manager := NewPerson("Mike CTO")
	manager.SetCurrentCompany("Acme Corporation")

	// Employment relationship
	employment := NewPersonWorksFor(&person, &org).(*PersonWorksFor)
	employment.JobTitle = "VP of Engineering"
	employment.Department = "Engineering"
	employment.StartDate = "2020-03-15"
	employment.SalaryRange = "$180,000-$220,000"
	employment.EquityGrant = true
	employment.HasBudgetAuthority = true
	employment.AccessLevel = AccessLevelAdmin
	employment.IsCurrentRole = true
	employment.EmploymentStatus = EmploymentActive
	employment.Source = "hrms"
	employment.Confidence = 95

	// Reporting relationship
	reporting := NewPersonReportsTo(&person, &manager).(*PersonReportsTo)
	reporting.ReportingType = ReportingDirect
	reporting.IsActive = true
	reporting.Organization = "Acme Corporation"
	reporting.Department = "Engineering"
	reporting.MeetingFrequency = "weekly"
	reporting.InfluenceLevel = 8
	reporting.Source = "org-chart"
	reporting.Confidence = 90

	// Validate both relationships
	assert.True(t, employment.Valid())
	assert.True(t, reporting.Valid())

	// Test employment analysis
	assert.True(t, employment.IsCurrentEmployment())
	assert.True(t, employment.HasDecisionMakingPower())
	assert.Equal(t, SeniorityExecutive, employment.GetSeniorityLevel())

	// Test reporting analysis
	assert.True(t, reporting.IsDirectReport())
	assert.Equal(t, "high", reporting.GetRelationshipStrength())
	assert.True(t, reporting.HasRegularInteraction())

	// Get nodes for logging
	empSource, empTarget := employment.Nodes()
	repSource, repTarget := reporting.Nodes()

	t.Logf("✅ Complex employment scenario successful")
	t.Logf("   - Employment: %s -> %s (%s)", empSource.GetKey(), empTarget.GetKey(), employment.JobTitle)
	t.Logf("   - Reporting: %s -> %s (%s)", repSource.GetKey(), repTarget.GetKey(), reporting.ReportingType)
	t.Logf("   - Seniority: %s", employment.GetSeniorityLevel())
	t.Logf("   - Decision Maker: %v", employment.HasDecisionMakingPower())
	t.Logf("   - Relationship Strength: %s", reporting.GetRelationshipStrength())
}
