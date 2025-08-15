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

type Organization struct {
	BaseAsset

	PrimaryName string             `neo4j:"primaryName" json:"primaryName" desc:"Primary canonical name of the organization." example:"Walmart"`
	Names       []OrganizationName `neo4j:"-" json:"names" desc:"All name variations and aliases for this organization."`
	Industry    string             `neo4j:"industry" json:"industry,omitempty" desc:"Industry classification of the organization." example:"Retail"`
	Country     string             `neo4j:"country" json:"country,omitempty" desc:"Primary country of operation." example:"United States"`
	Region      string             `neo4j:"region" json:"region,omitempty" desc:"Primary region of operation." example:"North America"`
	StockTicker string             `neo4j:"stockTicker" json:"stockTicker,omitempty" desc:"Stock ticker symbol if publicly traded." example:"WMT"`
	Website     string             `neo4j:"website" json:"website,omitempty" desc:"Primary website URL." example:"https://www.walmart.com"`
	Description string             `neo4j:"description" json:"description,omitempty" desc:"Brief description of the organization."`

	// Organizational relationships (stored as properties, connected via DISCOVERED relationships)
	ParentOrganization  string  `neo4j:"parentOrganization" json:"parentOrganization,omitempty" desc:"Key of parent organization if this is a subsidiary." example:"#organization#walmart#Walmart"`
	OwnershipPercentage float64 `neo4j:"ownershipPercentage" json:"ownershipPercentage,omitempty" desc:"Percentage owned by parent organization." example:"100"`
	SubsidiaryType      string  `neo4j:"subsidiaryType" json:"subsidiaryType,omitempty" desc:"Type of subsidiary relationship." example:"wholly_owned"`

	// Merger/acquisition information
	MergedOrganizations []string `neo4j:"mergedOrganizations" json:"mergedOrganizations,omitempty" desc:"Keys of organizations that were merged into this one." example:"['#organization#samsclub#Sams Club']"`
	LastAcquisitionDate string   `neo4j:"lastAcquisitionDate" json:"lastAcquisitionDate,omitempty" desc:"Date of most recent major acquisition (RFC3339)." example:"2020-06-15T00:00:00Z"`
}

func (o *Organization) Valid() bool {
	if o.PrimaryName == "" {
		return false
	}
	if !organizationKey.MatchString(o.Key) {
		return false
	}

	return true
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
				keyNormalized := NormalizeOrganizationName(o.PrimaryName)
				o.Key = fmt.Sprintf("#organization#%s#%s", keyNormalized, o.PrimaryName)
				o.BaseAsset.Identifier = o.PrimaryName
				o.BaseAsset.Group = o.PrimaryName

				return nil
			},
		},
		setGroupAndIdentifier(o, &o.PrimaryName, &o.PrimaryName),
	}
}

func (o *Organization) GetDescription() string {
	return "Represents an organization with multiple name variations and associated meta data."
}

// CreateNameRelationship creates a new OrganizationName node and relationship
func (o *Organization) CreateNameRelationship(name, nameType, source string) (*OrganizationName, GraphRelationship, error) {
	if name == "" {
		return nil, nil, fmt.Errorf("name cannot be empty")
	}
	if !validNameTypes[nameType] {
		return nil, nil, fmt.Errorf("invalid name type: %s", nameType)
	}

	orgName := NewOrganizationName(name, nameType, source)
	relationship := NewHasOrganizationName(o, &orgName)

	return &orgName, relationship, nil
}

func (o *Organization) IsSubsidiary() bool {
	return o.ParentOrganization != ""
}

func (o *Organization) IsWhollyOwned() bool {
	return o.IsSubsidiary() && o.OwnershipPercentage >= 100.0
}

func (o *Organization) HasMergerHistory() bool {
	return len(o.MergedOrganizations) > 0
}

func (o *Organization) AddMergedOrganization(organizationKey string) {
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

// CreatePrimaryNameRelationship creates the primary OrganizationName node and relationship
func (o *Organization) CreatePrimaryNameRelationship() (*OrganizationName, GraphRelationship) {
	primaryName := NewOrganizationName(o.PrimaryName, NameTypePrimary, "")
	relationship := NewHasOrganizationName(o, &primaryName)
	return &primaryName, relationship
}

// NormalizeOrganizationName normalizes an organization name for consistent key generation
func NormalizeOrganizationName(name string) string {
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
	normalizedPrimary := NormalizeOrganizationName(org.PrimaryName)
	ose.Organizations[normalizedPrimary] = org

	for _, name := range org.Names {
		if name.State == NameStateActive {
			normalized := NormalizeOrganizationName(name.Name)
			if _, exists := ose.Organizations[normalized]; !exists {
				ose.Organizations[normalized] = org
			}
		}
	}
}

// AddOrganizationWithNames adds an organization and its associated name nodes to the search expansion
func (ose *OrganizationSearchExpansion) AddOrganizationWithNames(org *Organization, names []OrganizationName) {
	normalizedPrimary := NormalizeOrganizationName(org.PrimaryName)
	ose.Organizations[normalizedPrimary] = org

	for _, name := range names {
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
		var names []string
		names = append(names, org.PrimaryName)
		for _, name := range org.Names {
			if name.State == NameStateActive {
				names = append(names, name.Name)
			}
		}
		sort.Strings(names)
		return names
	}

	return []string{searchTerm}
}

func (ose *OrganizationSearchExpansion) FindOrganization(name string) *Organization {
	normalized := NormalizeOrganizationName(name)
	return ose.Organizations[normalized]
}
