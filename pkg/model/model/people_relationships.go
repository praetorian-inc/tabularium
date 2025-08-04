package model

import (
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func init() {
	registry.Registry.MustRegisterModel(&PersonWorksFor{})
	registry.Registry.MustRegisterModel(&PersonReportsTo{})
}

const (
	// Person-Organization relationship types
	WorksForLabel = "PersonWorksFor"
	WorksForType  = "WORKS_FOR"

	// Person-Person relationship types
	ReportsToLabel = "PersonReportsTo"
	ReportsToType  = "REPORTS_TO"

	// Employment statuses
	EmploymentActive     = "active"
	EmploymentInactive   = "inactive"
	EmploymentSuspended  = "suspended"
	EmploymentTerminated = "terminated"

	// Employment types
	EmploymentFullTime   = "full-time"
	EmploymentPartTime   = "part-time"
	EmploymentContract   = "contract"
	EmploymentIntern     = "intern"
	EmploymentConsultant = "consultant"

	// Reporting relationship types
	ReportingDirect     = "direct"
	ReportingIndirect   = "indirect"
	ReportingFunctional = "functional"
	ReportingMatrix     = "matrix"
	ReportingMentor     = "mentor"
)

var (
	validEmploymentStatuses = map[string]bool{
		EmploymentActive:     true,
		EmploymentInactive:   true,
		EmploymentSuspended:  true,
		EmploymentTerminated: true,
	}

	validEmploymentTypes = map[string]bool{
		EmploymentFullTime:   true,
		EmploymentPartTime:   true,
		EmploymentContract:   true,
		EmploymentIntern:     true,
		EmploymentConsultant: true,
	}

	validReportingTypes = map[string]bool{
		ReportingDirect:     true,
		ReportingIndirect:   true,
		ReportingFunctional: true,
		ReportingMatrix:     true,
		ReportingMentor:     true,
	}
)

// PersonWorksFor represents an employment relationship between a person and an organization
type PersonWorksFor struct {
	*BaseRelationship

	// Employment details
	JobTitle         string `neo4j:"jobTitle" json:"jobTitle,omitempty" desc:"Job title or position at the organization." example:"Senior Software Engineer"`
	Department       string `neo4j:"department" json:"department,omitempty" desc:"Department or division within the organization." example:"Engineering"`
	EmploymentType   string `neo4j:"employmentType" json:"employmentType,omitempty" desc:"Type of employment." example:"full-time"`
	EmploymentStatus string `neo4j:"employmentStatus" json:"employmentStatus,omitempty" desc:"Current employment status." example:"active"`

	// Dates
	StartDate string `neo4j:"startDate" json:"startDate,omitempty" desc:"Employment start date (YYYY-MM-DD)." example:"2022-01-15"`
	EndDate   string `neo4j:"endDate" json:"endDate,omitempty" desc:"Employment end date (YYYY-MM-DD)." example:"2023-12-31"`

	// Compensation (optional, for security assessment context)
	SalaryRange string `neo4j:"salaryRange" json:"salaryRange,omitempty" desc:"Salary range or band." example:"$80,000-$120,000"`
	EquityGrant bool   `neo4j:"equityGrant" json:"equityGrant,omitempty" desc:"Whether the person has equity grants." example:"true"`

	// Security context
	HasBudgetAuthority bool   `neo4j:"hasBudgetAuthority" json:"hasBudgetAuthority,omitempty" desc:"Has budget or purchasing authority." example:"true"`
	AccessLevel        string `neo4j:"accessLevel" json:"accessLevel,omitempty" desc:"Access level within the organization." example:"admin"`
	SecurityClearance  string `neo4j:"securityClearance" json:"securityClearance,omitempty" desc:"Security clearance level." example:"secret"`

	// Relationship metadata
	IsCurrentRole bool   `neo4j:"isCurrentRole" json:"isCurrentRole" desc:"Whether this is the person's current role." example:"true"`
	Source        string `neo4j:"source" json:"source,omitempty" desc:"Source of this employment information." example:"linkedin"`
	Confidence    int    `neo4j:"confidence" json:"confidence" desc:"Confidence level in this information (1-100)." example:"95"`
}

// PersonReportsTo represents a reporting relationship between two people (manager-subordinate)
type PersonReportsTo struct {
	*BaseRelationship

	// Relationship details
	ReportingType string `neo4j:"reportingType" json:"reportingType,omitempty" desc:"Type of reporting relationship." example:"direct"`
	IsActive      bool   `neo4j:"isActive" json:"isActive" desc:"Whether this reporting relationship is currently active." example:"true"`

	// Dates
	StartDate string `neo4j:"startDate" json:"startDate,omitempty" desc:"When this reporting relationship started (YYYY-MM-DD)." example:"2022-06-01"`
	EndDate   string `neo4j:"endDate" json:"endDate,omitempty" desc:"When this reporting relationship ended (YYYY-MM-DD)." example:"2023-12-31"`

	// Context
	Organization string `neo4j:"organization" json:"organization,omitempty" desc:"Organization where this reporting relationship exists." example:"Acme Corporation"`
	Department   string `neo4j:"department" json:"department,omitempty" desc:"Department where this reporting relationship exists." example:"Engineering"`

	// Relationship strength/frequency
	MeetingFrequency string `neo4j:"meetingFrequency" json:"meetingFrequency,omitempty" desc:"How often they meet." example:"weekly"`
	InfluenceLevel   int    `neo4j:"influenceLevel" json:"influenceLevel" desc:"Manager's influence over subordinate (1-10)." example:"8"`

	// Metadata
	Source     string `neo4j:"source" json:"source,omitempty" desc:"Source of this relationship information." example:"org-chart"`
	Confidence int    `neo4j:"confidence" json:"confidence" desc:"Confidence level in this information (1-100)." example:"90"`
}

// Implement GraphRelationship interface for PersonWorksFor

func (pf *PersonWorksFor) Label() string {
	return WorksForType
}

func (pf *PersonWorksFor) Valid() bool {
	if !pf.BaseRelationship.Valid() {
		return false
	}

	// Validate employment type
	if pf.EmploymentType != "" && !validEmploymentTypes[pf.EmploymentType] {
		return false
	}

	// Validate employment status
	if pf.EmploymentStatus != "" && !validEmploymentStatuses[pf.EmploymentStatus] {
		return false
	}

	// Validate confidence range
	if pf.Confidence < 0 || pf.Confidence > 100 {
		return false
	}

	return true
}

func (pf *PersonWorksFor) GetDescription() string {
	return "Represents an employment relationship between a person and an organization, including job details and security context."
}

func (pf *PersonWorksFor) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				// Set defaults
				if pf.EmploymentStatus == "" {
					pf.EmploymentStatus = EmploymentActive
				}
				if pf.EmploymentType == "" {
					pf.EmploymentType = EmploymentFullTime
				}
				if pf.Confidence == 0 {
					pf.Confidence = 50 // Default medium confidence
				}
				if pf.IsCurrentRole == false && pf.EndDate == "" {
					pf.IsCurrentRole = true // Assume current if no end date
				}
				return nil
			},
		},
	}
}

func (pf *PersonWorksFor) Defaulted() {
	// This will be handled by hooks
}

// Implement GraphRelationship interface for PersonReportsTo

func (rt *PersonReportsTo) Label() string {
	return ReportsToType
}

func (rt *PersonReportsTo) Valid() bool {
	if !rt.BaseRelationship.Valid() {
		return false
	}

	// Validate reporting type
	if rt.ReportingType != "" && !validReportingTypes[rt.ReportingType] {
		return false
	}

	// Validate influence level range
	if rt.InfluenceLevel < 0 || rt.InfluenceLevel > 10 {
		return false
	}

	// Validate confidence range
	if rt.Confidence < 0 || rt.Confidence > 100 {
		return false
	}

	return true
}

func (rt *PersonReportsTo) GetDescription() string {
	return "Represents a reporting relationship between two people, capturing organizational hierarchy and management structure."
}

func (rt *PersonReportsTo) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				// Set defaults
				if rt.ReportingType == "" {
					rt.ReportingType = ReportingDirect
				}
				if rt.Confidence == 0 {
					rt.Confidence = 50 // Default medium confidence
				}
				if rt.InfluenceLevel == 0 {
					rt.InfluenceLevel = 5 // Default medium influence
				}
				if rt.IsActive == false && rt.EndDate == "" {
					rt.IsActive = true // Assume active if no end date
				}
				return nil
			},
		},
	}
}

func (rt *PersonReportsTo) Defaulted() {
	// This will be handled by hooks
}

// Constructor functions

func NewPersonWorksFor(person, organization GraphModel) GraphRelationship {
	return &PersonWorksFor{
		BaseRelationship: NewBaseRelationship(person, organization, WorksForType),
	}
}

func NewPersonReportsTo(subordinate, manager GraphModel) GraphRelationship {
	return &PersonReportsTo{
		BaseRelationship: NewBaseRelationship(subordinate, manager, ReportsToType),
	}
}

// Helper functions

func (pf *PersonWorksFor) IsCurrentEmployment() bool {
	return pf.IsCurrentRole && pf.EmploymentStatus == EmploymentActive
}

func (pf *PersonWorksFor) HasDecisionMakingPower() bool {
	return pf.HasBudgetAuthority ||
		strings.Contains(strings.ToLower(pf.JobTitle), "manager") ||
		strings.Contains(strings.ToLower(pf.JobTitle), "director") ||
		strings.Contains(strings.ToLower(pf.JobTitle), "vp") ||
		strings.Contains(strings.ToLower(pf.JobTitle), "chief") ||
		strings.Contains(strings.ToLower(pf.JobTitle), "head") ||
		strings.Contains(strings.ToLower(pf.JobTitle), "lead")
}

func (pf *PersonWorksFor) GetSeniorityLevel() string {
	title := strings.ToLower(pf.JobTitle)

	if strings.Contains(title, "intern") || strings.Contains(title, "trainee") {
		return SeniorityEntry
	}
	if strings.Contains(title, "junior") || strings.Contains(title, "associate") {
		return SeniorityEntry
	}
	if strings.Contains(title, "senior") || strings.Contains(title, "lead") {
		return SenioritSenior
	}
	if strings.Contains(title, "principal") || strings.Contains(title, "staff") {
		return SenioritSenior
	}
	if strings.Contains(title, "manager") || strings.Contains(title, "director") {
		return SeniorityExecutive
	}
	if strings.Contains(title, "vp") || strings.Contains(title, "vice") {
		return SeniorityExecutive
	}
	if strings.Contains(title, "chief") || strings.Contains(title, "ceo") ||
		strings.Contains(title, "cto") || strings.Contains(title, "cfo") ||
		strings.Contains(title, "cmo") || strings.Contains(title, "cso") {
		return SeniorityCLevel
	}

	// Default to mid-level if can't determine
	return SeniorityMid
}

func (rt *PersonReportsTo) IsDirectReport() bool {
	return rt.ReportingType == ReportingDirect && rt.IsActive
}

func (rt *PersonReportsTo) GetRelationshipStrength() string {
	if rt.InfluenceLevel >= 8 {
		return "high"
	}
	if rt.InfluenceLevel >= 5 {
		return "medium"
	}
	return "low"
}

func (rt *PersonReportsTo) HasRegularInteraction() bool {
	frequency := strings.ToLower(rt.MeetingFrequency)
	return strings.Contains(frequency, "daily") ||
		(frequency == "weekly") ||
		(strings.Contains(frequency, "biweekly") && rt.InfluenceLevel >= 6) ||
		(strings.Contains(frequency, "bi-weekly") && rt.InfluenceLevel >= 6)
}

// Validation helpers

func IsValidEmploymentStatus(status string) bool {
	return validEmploymentStatuses[status]
}

func IsValidEmploymentType(empType string) bool {
	return validEmploymentTypes[empType]
}

func IsValidReportingType(reportingType string) bool {
	return validReportingTypes[reportingType]
}
