package model

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func init() {
	registry.Registry.MustRegisterModel(&OrganizationName{})
}

const (
	// Name types
	NameTypePrimary      = "primary"
	NameTypeLegal        = "legal"
	NameTypeDBA          = "dba"
	NameTypeAbbreviation = "abbreviation"
	NameTypeCommon       = "common"
	NameTypeFormer       = "former"
	NameTypeRegional     = "regional"

	// Name states (validity/usage state of the name variation)
	NameStateActive   = "active"
	NameStateInactive = "inactive"
	NameStateHistoric = "historic"
)

var (
	validNameTypes = map[string]bool{
		NameTypePrimary:      true,
		NameTypeLegal:        true,
		NameTypeDBA:          true,
		NameTypeAbbreviation: true,
		NameTypeCommon:       true,
		NameTypeFormer:       true,
		NameTypeRegional:     true,
	}
	validNameStates = map[string]bool{
		NameStateActive:   true,
		NameStateInactive: true,
		NameStateHistoric: true,
	}
)

// OrganizationName represents a single name variation for an organization
type OrganizationName struct {
	registry.BaseModel
	// The actual name
	Name string `json:"name" desc:"The organization name." example:"Walmart Inc"`
	// Type of name (primary, legal, dba, abbreviation, etc.)
	Type string `json:"type" desc:"Type of organization name." example:"legal"`
	// State of the name (active, inactive, historic)
	State string `json:"state" desc:"State of the organization name variation." example:"active"`
	// When this name was added/discovered
	DateAdded string `json:"dateAdded" desc:"When this name was added (RFC3339)." example:"2023-10-27T10:00:00Z"`
	// When this name became effective (for historical tracking)
	EffectiveDate string `json:"effectiveDate,omitempty" desc:"When this name became effective (RFC3339)." example:"2020-01-01T00:00:00Z"`
	// When this name was discontinued (for historical tracking)
	EndDate string `json:"endDate,omitempty" desc:"When this name was discontinued (RFC3339)." example:"2021-12-31T23:59:59Z"`
	// Source of the name (where it was discovered)
	Source string `json:"source,omitempty" desc:"Source where this name was discovered." example:"github"`
	// Additional metadata
	Metadata map[string]interface{} `json:"metadata,omitempty" desc:"Additional metadata about this name variation."`
}

func (on *OrganizationName) Valid() bool {
	if on.Name == "" {
		return false
	}
	if !validNameTypes[on.Type] {
		return false
	}
	if !validNameStates[on.State] {
		return false
	}
	return true
}

func (on *OrganizationName) GetDescription() string {
	return "Represents a single name variation for an organization, including type, status, and historical tracking information."
}

func (on *OrganizationName) GetHooks() []registry.Hook {
	return []registry.Hook{}
}

func (on *OrganizationName) Defaulted() {
	on.State = NameStateActive
	on.DateAdded = Now()
}

func (on *OrganizationName) GetKey() string {
	// Normalize the name for key generation (lowercase, alphanumeric only)
	normalized := strings.ToLower(strings.TrimSpace(on.Name))
	keyNormalized := regexp.MustCompile(`[^a-z0-9]`).ReplaceAllString(normalized, "")

	// Generate key based on normalized name and type for uniqueness
	return fmt.Sprintf("#organizationname#%s#%s", keyNormalized, on.Type)
}
