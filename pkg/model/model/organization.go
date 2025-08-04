package model

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func init() {
	registry.Registry.MustRegisterModel(&Organization{})
	registry.Registry.MustRegisterModel(&OrganizationName{})
}

const (
	OrganizationLabel = "Organization"

	// Name types
	NameTypePrimary      = "primary"
	NameTypeLegal        = "legal"
	NameTypeDBA          = "dba"
	NameTypeAbbreviation = "abbreviation"
	NameTypeCommon       = "common"
	NameTypeFormer       = "former"
	NameTypeRegional     = "regional"

	// Name statuses
	NameStatusActive   = "active"
	NameStatusInactive = "inactive"
	NameStatusHistoric = "historic"
)

var (
	organizationKey = regexp.MustCompile(`^#organization#([^#]+)#([^#]+)$`)
	validNameTypes  = map[string]bool{
		NameTypePrimary:      true,
		NameTypeLegal:        true,
		NameTypeDBA:          true,
		NameTypeAbbreviation: true,
		NameTypeCommon:       true,
		NameTypeFormer:       true,
		NameTypeRegional:     true,
	}
	validNameStatuses = map[string]bool{
		NameStatusActive:   true,
		NameStatusInactive: true,
		NameStatusHistoric: true,
	}
)

// Organization represents an organization entity with multiple name variations
type Organization struct {
	BaseAsset
	// Primary name is the canonical/preferred name for the organization
	PrimaryName string `neo4j:"primaryName" json:"primaryName" desc:"Primary canonical name of the organization." example:"Walmart"`
	// Names contains all name variations for this organization
	Names []OrganizationName `neo4j:"-" json:"names" desc:"All name variations and aliases for this organization."`
	// Industry classification
	Industry string `neo4j:"industry" json:"industry,omitempty" desc:"Industry classification of the organization." example:"Retail"`
	// Geographic information
	Country string `neo4j:"country" json:"country,omitempty" desc:"Primary country of operation." example:"United States"`
	Region  string `neo4j:"region" json:"region,omitempty" desc:"Primary region of operation." example:"North America"`
	// Stock ticker if publicly traded
	StockTicker string `neo4j:"stockTicker" json:"stockTicker,omitempty" desc:"Stock ticker symbol if publicly traded." example:"WMT"`
	// Website URL
	Website string `neo4j:"website" json:"website,omitempty" desc:"Primary website URL." example:"https://www.walmart.com"`
	// Description
	Description string `neo4j:"description" json:"description,omitempty" desc:"Brief description of the organization."`
}

// OrganizationName represents a single name variation for an organization
type OrganizationName struct {
	registry.BaseModel
	// The actual name
	Name string `json:"name" desc:"The organization name." example:"Walmart Inc"`
	// Type of name (primary, legal, dba, abbreviation, etc.)
	Type string `json:"type" desc:"Type of organization name." example:"legal"`
	// Status of the name (active, inactive, historic)
	Status string `json:"status" desc:"Status of the organization name." example:"active"`
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

// IsPrivate returns false as organizations are typically public entities
func (o *Organization) IsPrivate() bool {
	return false
}

// GetKey returns the organization's key
func (o *Organization) GetKey() string {
	return o.Key
}

// Valid validates the organization model
func (o *Organization) Valid() bool {
	if o.PrimaryName == "" {
		return false
	}
	if !organizationKey.MatchString(o.Key) {
		return false
	}

	// Must have at least one name and it must be the primary name
	hasPrimary := false
	for _, name := range o.Names {
		if !name.Valid() {
			return false
		}
		if name.Type == NameTypePrimary && name.Name == o.PrimaryName {
			hasPrimary = true
		}
	}

	return hasPrimary
}

// Valid validates an organization name
func (on *OrganizationName) Valid() bool {
	if on.Name == "" {
		return false
	}
	if !validNameTypes[on.Type] {
		return false
	}
	if !validNameStatuses[on.Status] {
		return false
	}
	return true
}

// GetDescription returns the model description for OrganizationName
func (on *OrganizationName) GetDescription() string {
	return "Represents a single name variation for an organization, including type, status, and historical tracking information."
}

// GetHooks returns hooks for OrganizationName (empty for this model)
func (on *OrganizationName) GetHooks() []registry.Hook {
	return []registry.Hook{}
}

// Defaulted sets default values for OrganizationName
func (on *OrganizationName) Defaulted() {
	if on.Status == "" {
		on.Status = NameStatusActive
	}
	if on.DateAdded == "" {
		on.DateAdded = Now()
	}
}

// GetKey returns the key for OrganizationName (not typically used as standalone entity)
func (on *OrganizationName) GetKey() string {
	return ""
}

// GetLabels returns the labels for this organization
func (o *Organization) GetLabels() []string {
	return []string{OrganizationLabel, AssetLabel, TTLLabel}
}

// GetClass returns the asset class
func (o *Organization) GetClass() string {
	return "organization"
}

// GetStatus returns the current status
func (o *Organization) GetStatus() string {
	return o.Status
}

// Group returns the organization identifier
func (o *Organization) Group() string {
	return o.PrimaryName
}

// Identifier returns the organization identifier
func (o *Organization) Identifier() string {
	return o.PrimaryName
}

// IsStatus checks if the organization has the given status
func (o *Organization) IsStatus(value string) bool {
	return strings.HasPrefix(o.Status, value)
}

// WithStatus creates a copy with the given status
func (o *Organization) WithStatus(status string) Target {
	ret := *o
	ret.Status = status
	return &ret
}

// Attribute creates an attribute for this organization
func (o *Organization) Attribute(name, value string) Attribute {
	attr := NewAttribute(name, value, o)
	return attr
}

// Seed creates a seed from this organization
func (o *Organization) Seed() Seed {
	s := NewSeed(o.PrimaryName)
	s.SetStatus(o.Status)
	return s
}

// Defaulted sets default values
func (o *Organization) Defaulted() {
	o.BaseAsset.Defaulted()
	o.Class = "organization"

	// Ensure primary name is in the names list
	if o.PrimaryName != "" && len(o.Names) == 0 {
		o.Names = []OrganizationName{
			{
				Name:      o.PrimaryName,
				Type:      NameTypePrimary,
				Status:    NameStatusActive,
				DateAdded: Now(),
			},
		}
	}
}

// GetHooks returns the hooks for this organization
func (o *Organization) GetHooks() []registry.Hook {
	return []registry.Hook{
		useGroupAndIdentifier(o, &o.PrimaryName, &o.PrimaryName),
		{
			Call: func() error {
				// Generate key based on normalized primary name
				normalized := NormalizeOrganizationName(o.PrimaryName)
				o.Key = fmt.Sprintf("#organization#%s#%s", normalized, o.PrimaryName)
				o.BaseAsset.Identifier = o.PrimaryName
				o.BaseAsset.Group = o.PrimaryName

				// Ensure primary name is in names list
				o.ensurePrimaryNameInList()

				return nil
			},
		},
		setGroupAndIdentifier(o, &o.PrimaryName, &o.PrimaryName),
	}
}

// ensurePrimaryNameInList ensures the primary name exists in the names list
func (o *Organization) ensurePrimaryNameInList() {
	hasPrimary := false
	for i, name := range o.Names {
		if name.Type == NameTypePrimary {
			// Update existing primary name
			o.Names[i].Name = o.PrimaryName
			o.Names[i].Status = NameStatusActive
			if o.Names[i].DateAdded == "" {
				o.Names[i].DateAdded = Now()
			}
			hasPrimary = true
			break
		}
	}

	if !hasPrimary {
		// Add primary name to list
		o.Names = append(o.Names, OrganizationName{
			Name:      o.PrimaryName,
			Type:      NameTypePrimary,
			Status:    NameStatusActive,
			DateAdded: Now(),
		})
	}
}

// GetDescription returns the model description
func (o *Organization) GetDescription() string {
	return "Represents an organization with multiple name variations, supporting comprehensive cross-platform asset discovery."
}

// AddName adds a new name variation to the organization
func (o *Organization) AddName(name, nameType, source string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if !validNameTypes[nameType] {
		return fmt.Errorf("invalid name type: %s", nameType)
	}

	// Check for duplicates
	for _, existing := range o.Names {
		if existing.Name == name && existing.Type == nameType {
			return fmt.Errorf("name already exists: %s (%s)", name, nameType)
		}
	}

	// Add the new name
	o.Names = append(o.Names, OrganizationName{
		Name:      name,
		Type:      nameType,
		Status:    NameStatusActive,
		DateAdded: Now(),
		Source:    source,
	})

	return nil
}

// GetActiveNames returns all active names for the organization
func (o *Organization) GetActiveNames() []string {
	var names []string
	for _, name := range o.Names {
		if name.Status == NameStatusActive {
			names = append(names, name.Name)
		}
	}
	sort.Strings(names)
	return names
}

// GetNamesByType returns all names of a specific type
func (o *Organization) GetNamesByType(nameType string) []string {
	var names []string
	for _, name := range o.Names {
		if name.Type == nameType && name.Status == NameStatusActive {
			names = append(names, name.Name)
		}
	}
	sort.Strings(names)
	return names
}

// GetAllNameVariations returns all name variations (active and inactive)
func (o *Organization) GetAllNameVariations() []string {
	var names []string
	for _, name := range o.Names {
		names = append(names, name.Name)
	}
	sort.Strings(names)
	return names
}

// NewOrganization creates a new organization with the given primary name
func NewOrganization(primaryName string) Organization {
	org := Organization{
		PrimaryName: primaryName,
	}
	org.Defaulted()
	registry.CallHooks(&org)
	return org
}

// NormalizeOrganizationName normalizes an organization name for consistent key generation
func NormalizeOrganizationName(name string) string {
	// Convert to lowercase and trim spaces
	normalized := strings.ToLower(strings.TrimSpace(name))

	// Remove common suffixes/prefixes (including variations with periods)
	suffixes := []string{
		" inc.", " inc",
		" incorporated",
		" corp.", " corp",
		" corporation",
		" llc",
		" ltd.", " ltd",
		" limited",
		" co.", " co",
		" company",
	}

	for _, suffix := range suffixes {
		if strings.HasSuffix(normalized, suffix) {
			normalized = strings.TrimSuffix(normalized, suffix)
			normalized = strings.TrimSpace(normalized)
			break
		}
	}

	// Remove special characters (keep only alphanumeric)
	normalized = regexp.MustCompile(`[^a-z0-9]`).ReplaceAllString(normalized, "")

	return normalized
}

// OrganizationSearchExpansion provides search expansion capabilities
type OrganizationSearchExpansion struct {
	Organizations map[string]*Organization // keyed by normalized name
}

// NewOrganizationSearchExpansion creates a new search expansion service
func NewOrganizationSearchExpansion() *OrganizationSearchExpansion {
	return &OrganizationSearchExpansion{
		Organizations: make(map[string]*Organization),
	}
}

// AddOrganization adds an organization to the search expansion index
func (ose *OrganizationSearchExpansion) AddOrganization(org *Organization) {
	// Index by normalized primary name
	normalizedPrimary := NormalizeOrganizationName(org.PrimaryName)
	ose.Organizations[normalizedPrimary] = org

	// Also index by all name variations
	for _, name := range org.Names {
		if name.Status == NameStatusActive {
			normalized := NormalizeOrganizationName(name.Name)
			if _, exists := ose.Organizations[normalized]; !exists {
				ose.Organizations[normalized] = org
			}
		}
	}
}

// ExpandSearch returns all name variations for a given organization name
func (ose *OrganizationSearchExpansion) ExpandSearch(searchTerm string) []string {
	if searchTerm == "" {
		return []string{}
	}

	normalized := NormalizeOrganizationName(searchTerm)

	if org, exists := ose.Organizations[normalized]; exists {
		return org.GetActiveNames()
	}

	// If not found, return the original search term
	return []string{searchTerm}
}

// FindOrganization finds an organization by any of its names
func (ose *OrganizationSearchExpansion) FindOrganization(name string) *Organization {
	normalized := NormalizeOrganizationName(name)
	return ose.Organizations[normalized]
}
