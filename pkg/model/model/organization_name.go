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
	NameTypePrimary      = "primary"
	NameTypeLegal        = "legal"
	NameTypeDBA          = "dba"
	NameTypeAbbreviation = "abbreviation"
	NameTypeCommon       = "common"
	NameTypeFormer       = "former"
	NameTypeRegional     = "regional"

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

const (
	OrganizationNameLabel = "OrganizationName"
)

type OrganizationName struct {
	registry.BaseModel
	Name          string                 `json:"name" desc:"The organization name." example:"Walmart Inc"`
	Type          string                 `json:"type" desc:"Type of organization name." example:"legal"`
	State         string                 `json:"state" desc:"State of the organization name variation." example:"active"`
	DateAdded     string                 `json:"dateAdded" desc:"When this name was added (RFC3339)." example:"2023-10-27T10:00:00Z"`
	EffectiveDate string                 `json:"effectiveDate,omitempty" desc:"When this name became effective (RFC3339)." example:"2020-01-01T00:00:00Z"`
	EndDate       string                 `json:"endDate,omitempty" desc:"When this name was discontinued (RFC3339)." example:"2021-12-31T23:59:59Z"`
	Source        string                 `json:"source,omitempty" desc:"Source where this name was discovered." example:"github"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" desc:"Additional metadata about this name variation."`
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
	normalized := strings.ToLower(strings.TrimSpace(on.Name))
	keyNormalized := regexp.MustCompile(`[^a-z0-9]`).ReplaceAllString(normalized, "")
	return fmt.Sprintf("#organizationname#%s#%s", keyNormalized, on.Type)
}

func (on *OrganizationName) GetLabels() []string {
	return []string{OrganizationNameLabel}
}

func NewOrganizationName(name, nameType, source string) OrganizationName {
	orgName := OrganizationName{
		Name:   name,
		Type:   nameType,
		Source: source,
	}
	orgName.Defaulted()
	registry.CallHooks(&orgName)
	return orgName
}
