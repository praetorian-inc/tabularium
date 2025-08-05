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
}

const (
	OrganizationLabel = "Organization"

	// Subsidiary relationship types
	SubsidiaryTypeWhollyOwned   = "wholly_owned"
	SubsidiaryTypeMajorityOwned = "majority_owned"
	SubsidiaryTypeMinorityOwned = "minority_owned"
	SubsidiaryTypeJointVenture  = "joint_venture"
)

var (
	organizationKey = regexp.MustCompile(`^#organization#([^#]+)#([^#]+)$`)
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

	// Organizational relationships (stored as properties, connected via DISCOVERED relationships)
	// Parent organization key (if this is a subsidiary)
	ParentOrganization string `neo4j:"parentOrganization" json:"parentOrganization,omitempty" desc:"Key of parent organization if this is a subsidiary." example:"#organization#walmart#Walmart"`
	// Ownership percentage by parent (0-100)
	OwnershipPercentage float64 `neo4j:"ownershipPercentage" json:"ownershipPercentage,omitempty" desc:"Percentage owned by parent organization." example:"100"`
	// Subsidiary relationship type
	SubsidiaryType string `neo4j:"subsidiaryType" json:"subsidiaryType,omitempty" desc:"Type of subsidiary relationship." example:"wholly_owned"`

	// Historical information
	// Former names of this organization
	FormerNames []string `neo4j:"formerNames" json:"formerNames,omitempty" desc:"Previous names of this organization." example:"['Walmart Stores Inc', 'Wal-Mart Stores']"`
	// Date of last name change
	LastNameChange string `neo4j:"lastNameChange" json:"lastNameChange,omitempty" desc:"Date of most recent name change (RFC3339)." example:"2018-02-01T00:00:00Z"`

	// Merger/acquisition information
	// Organizations that were merged into this one
	MergedOrganizations []string `neo4j:"mergedOrganizations" json:"mergedOrganizations,omitempty" desc:"Keys of organizations that were merged into this one." example:"['#organization#samsclub#Sams Club']"`
	// Date of last major acquisition/merger
	LastAcquisitionDate string `neo4j:"lastAcquisitionDate" json:"lastAcquisitionDate,omitempty" desc:"Date of most recent major acquisition (RFC3339)." example:"2020-06-15T00:00:00Z"`
}

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

func (o *Organization) GetLabels() []string {
	return []string{OrganizationLabel, AssetLabel, TTLLabel}
}

func (o *Organization) GetClass() string {
	return "organization"
}

func (o *Organization) Group() string {
	return o.PrimaryName
}

func (o *Organization) Identifier() string {
	return o.PrimaryName
}

func (o *Organization) WithStatus(status string) Target {
	ret := *o
	ret.Status = status
	return &ret
}

func (o *Organization) Attribute(name, value string) Attribute {
	attr := NewAttribute(name, value, o)
	return attr
}

func (o *Organization) Seed() Seed {
	s := NewSeed(o.PrimaryName)
	s.SetStatus(o.Status)
	return s
}

func (o *Organization) Defaulted() {
	o.BaseAsset.Defaulted()
	o.Class = "organization"
}

func (o *Organization) GetHooks() []registry.Hook {
	return []registry.Hook{
		useGroupAndIdentifier(o, &o.PrimaryName, &o.PrimaryName),
		{
			Call: func() error {
				// Normalize the primary name for consistent storage and key generation
				// Convert to lowercase and trim spaces
				normalized := strings.ToLower(strings.TrimSpace(o.PrimaryName))

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

				// Remove special characters for key generation (keep only alphanumeric)
				keyNormalized := regexp.MustCompile(`[^a-z0-9]`).ReplaceAllString(normalized, "")

				// Generate key based on normalized primary name
				o.Key = fmt.Sprintf("#organization#%s#%s", keyNormalized, o.PrimaryName)
				o.BaseAsset.Identifier = o.PrimaryName
				o.BaseAsset.Group = o.PrimaryName

				// Ensure primary name is in names list
				hasPrimary := false
				for i, name := range o.Names {
					if name.Type == NameTypePrimary {
						// Update existing primary name
						o.Names[i].Name = o.PrimaryName
						o.Names[i].State = NameStateActive
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
						State:     NameStateActive,
						DateAdded: Now(),
					})
				}

				return nil
			},
		},
		setGroupAndIdentifier(o, &o.PrimaryName, &o.PrimaryName),
	}
}

func (o *Organization) GetDescription() string {
	return "Represents an organization with multiple name variations and associated meta data."
}

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
		State:     NameStateActive,
		DateAdded: Now(),
		Source:    source,
	})

	return nil
}

// GetNames returns organization names filtered by optional state and type
// Pass nil for no filtering on that field
func (o *Organization) GetNames(state *string, nameType *string) []string {
	var names []string
	for _, name := range o.Names {
		// Apply state filter if provided
		if state != nil && name.State != *state {
			continue
		}

		// Apply type filter if provided
		if nameType != nil && name.Type != *nameType {
			continue
		}

		names = append(names, name.Name)
	}
	sort.Strings(names)
	return names
}

func (o *Organization) GetActiveNames() []string {
	activeState := NameStateActive
	return o.GetNames(&activeState, nil)
}

func (o *Organization) GetNamesByType(nameType string) []string {
	activeState := NameStateActive
	return o.GetNames(&activeState, &nameType)
}

func (o *Organization) GetAllNameVariations() []string {
	return o.GetNames(nil, nil)
}

func (o *Organization) IsSubsidiary() bool {
	return o.ParentOrganization != ""
}

func (o *Organization) IsWhollyOwned() bool {
	return o.IsSubsidiary() && o.OwnershipPercentage >= 100.0
}

func (o *Organization) AddFormerName(formerName string) {
	// Check if already exists
	for _, existing := range o.FormerNames {
		if existing == formerName {
			return
		}
	}
	o.FormerNames = append(o.FormerNames, formerName)
}

func (o *Organization) HasMergerHistory() bool {
	return len(o.MergedOrganizations) > 0
}

func (o *Organization) AddMergedOrganization(organizationKey string) {
	// Check if already exists
	for _, existing := range o.MergedOrganizations {
		if existing == organizationKey {
			return
		}
	}
	o.MergedOrganizations = append(o.MergedOrganizations, organizationKey)
}

// SetParentOrganization establishes a parent-subsidiary relationship
func (o *Organization) SetParentOrganization(parentKey string, ownershipPercentage float64, subsidiaryType string) {
	o.ParentOrganization = parentKey
	o.OwnershipPercentage = ownershipPercentage
	o.SubsidiaryType = subsidiaryType
}

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

type OrganizationSearchExpansion struct {
	Organizations map[string]*Organization // keyed by normalized name
}

func NewOrganizationSearchExpansion() *OrganizationSearchExpansion {
	return &OrganizationSearchExpansion{
		Organizations: make(map[string]*Organization),
	}
}

func (ose *OrganizationSearchExpansion) AddOrganization(org *Organization) {
	// Index by normalized primary name
	normalizedPrimary := NormalizeOrganizationName(org.PrimaryName)
	ose.Organizations[normalizedPrimary] = org

	// Also index by all name variations
	for _, name := range org.Names {
		if name.State == NameStateActive {
			normalized := NormalizeOrganizationName(name.Name)
			if _, exists := ose.Organizations[normalized]; !exists {
				ose.Organizations[normalized] = org
			}
		}
	}
}

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

func (ose *OrganizationSearchExpansion) FindOrganization(name string) *Organization {
	normalized := NormalizeOrganizationName(name)
	return ose.Organizations[normalized]
}
