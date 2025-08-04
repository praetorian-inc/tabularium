package model

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func init() {
	registry.Registry.MustRegisterModel(&OrganizationParentSubsidiary{})
	registry.Registry.MustRegisterModel(&OrganizationNameHistory{})
	registry.Registry.MustRegisterModel(&OrganizationMerger{})
}

// Relationship labels
const (
	OrganizationParentSubsidiaryLabel = "PARENT_SUBSIDIARY"
	OrganizationNameHistoryLabel      = "NAME_HISTORY"
	OrganizationMergerLabel           = "MERGED_INTO"
)

// OrganizationParentSubsidiary represents a parent-subsidiary relationship between organizations
type OrganizationParentSubsidiary struct {
	*BaseRelationship
	// Ownership percentage (0-100)
	OwnershipPercentage float64 `neo4j:"ownershipPercentage" json:"ownershipPercentage,omitempty" desc:"Percentage of ownership (0-100)." example:"100"`
	// When the relationship became effective
	EffectiveDate string `neo4j:"effectiveDate" json:"effectiveDate,omitempty" desc:"When the relationship became effective (RFC3339)." example:"2020-01-01T00:00:00Z"`
	// When the relationship ended (if applicable)
	EndDate string `neo4j:"endDate" json:"endDate,omitempty" desc:"When the relationship ended (RFC3339)." example:"2023-12-31T23:59:59Z"`
	// Type of relationship (wholly_owned, majority_owned, minority_owned, joint_venture)
	RelationshipType string `neo4j:"relationshipType" json:"relationshipType" desc:"Type of parent-subsidiary relationship." example:"wholly_owned"`
	// Legal jurisdiction
	Jurisdiction string `neo4j:"jurisdiction" json:"jurisdiction,omitempty" desc:"Legal jurisdiction of the relationship." example:"Delaware, USA"`
	// Source of the relationship information
	Source string `neo4j:"source" json:"source,omitempty" desc:"Source of relationship information." example:"SEC filings"`
}

// OrganizationNameHistory represents historical name changes for organizations
type OrganizationNameHistory struct {
	*BaseRelationship
	// The old name
	OldName string `neo4j:"oldName" json:"oldName" desc:"Previous organization name." example:"Walmart Stores Inc"`
	// The new name
	NewName string `neo4j:"newName" json:"newName" desc:"New organization name." example:"Walmart Inc"`
	// Date of name change
	ChangeDate string `neo4j:"changeDate" json:"changeDate" desc:"Date when the name change occurred (RFC3339)." example:"2018-02-01T00:00:00Z"`
	// Reason for name change
	ChangeReason string `neo4j:"changeReason" json:"changeReason,omitempty" desc:"Reason for the name change." example:"Corporate restructuring"`
	// Legal filing reference
	FilingReference string `neo4j:"filingReference" json:"filingReference,omitempty" desc:"Legal filing reference for the name change." example:"SEC Form 8-K filed 2018-01-11"`
}

// OrganizationMerger represents merger and acquisition relationships
type OrganizationMerger struct {
	*BaseRelationship
	// Date of merger/acquisition
	MergerDate string `neo4j:"mergerDate" json:"mergerDate" desc:"Date when the merger/acquisition occurred (RFC3339)." example:"2020-06-15T00:00:00Z"`
	// Transaction value
	TransactionValue float64 `neo4j:"transactionValue" json:"transactionValue,omitempty" desc:"Transaction value in USD." example:"16000000000"`
	// Currency of transaction
	Currency string `neo4j:"currency" json:"currency,omitempty" desc:"Currency of the transaction." example:"USD"`
	// Type of transaction (merger, acquisition, spin_off)
	TransactionType string `neo4j:"transactionType" json:"transactionType" desc:"Type of transaction." example:"acquisition"`
	// Status of the transaction (pending, completed, cancelled)
	Status string `neo4j:"status" json:"status" desc:"Status of the transaction." example:"completed"`
	// Regulatory approval status
	RegulatoryApproval string `neo4j:"regulatoryApproval" json:"regulatoryApproval,omitempty" desc:"Regulatory approval status." example:"FTC approved"`
}

// Relationship type constants
const (
	RelationshipTypeWhollyOwned   = "wholly_owned"
	RelationshipTypeMajorityOwned = "majority_owned"
	RelationshipTypeMinorityOwned = "minority_owned"
	RelationshipTypeJointVenture  = "joint_venture"
)

// Transaction type constants
const (
	TransactionTypeMerger      = "merger"
	TransactionTypeAcquisition = "acquisition"
	TransactionTypeSpinOff     = "spin_off"
)

// Transaction status constants
const (
	TransactionStatusPending   = "pending"
	TransactionStatusCompleted = "completed"
	TransactionStatusCancelled = "cancelled"
)

// NewOrganizationParentSubsidiary creates a new parent-subsidiary relationship
func NewOrganizationParentSubsidiary(parent, subsidiary *Organization, ownershipPercentage float64, relationshipType string) *OrganizationParentSubsidiary {
	return &OrganizationParentSubsidiary{
		BaseRelationship:    NewBaseRelationship(parent, subsidiary, OrganizationParentSubsidiaryLabel),
		OwnershipPercentage: ownershipPercentage,
		RelationshipType:    relationshipType,
		EffectiveDate:       Now(),
	}
}

// NewOrganizationNameHistory creates a new name history relationship
func NewOrganizationNameHistory(organization *Organization, oldName, newName string, changeDate string) *OrganizationNameHistory {
	// Create a dummy target for the relationship (we could create a specific NameChange entity if needed)
	return &OrganizationNameHistory{
		BaseRelationship: NewBaseRelationship(organization, organization, OrganizationNameHistoryLabel),
		OldName:          oldName,
		NewName:          newName,
		ChangeDate:       changeDate,
	}
}

// NewOrganizationMerger creates a new merger relationship
func NewOrganizationMerger(acquirer, target *Organization, mergerDate string, transactionType string) *OrganizationMerger {
	return &OrganizationMerger{
		BaseRelationship: NewBaseRelationship(target, acquirer, OrganizationMergerLabel),
		MergerDate:       mergerDate,
		TransactionType:  transactionType,
		Status:           TransactionStatusPending,
	}
}

// Label implementations
func (ops *OrganizationParentSubsidiary) Label() string {
	return OrganizationParentSubsidiaryLabel
}

func (onh *OrganizationNameHistory) Label() string {
	return OrganizationNameHistoryLabel
}

func (om *OrganizationMerger) Label() string {
	return OrganizationMergerLabel
}

// Validation methods
func (ops *OrganizationParentSubsidiary) Valid() bool {
	if !ops.BaseRelationship.Valid() {
		return false
	}
	if ops.OwnershipPercentage < 0 || ops.OwnershipPercentage > 100 {
		return false
	}
	validTypes := map[string]bool{
		RelationshipTypeWhollyOwned:   true,
		RelationshipTypeMajorityOwned: true,
		RelationshipTypeMinorityOwned: true,
		RelationshipTypeJointVenture:  true,
	}
	if !validTypes[ops.RelationshipType] {
		return false
	}
	return true
}

func (onh *OrganizationNameHistory) Valid() bool {
	if !onh.BaseRelationship.Valid() {
		return false
	}
	if onh.OldName == "" || onh.NewName == "" {
		return false
	}
	if onh.ChangeDate == "" {
		return false
	}
	return true
}

func (om *OrganizationMerger) Valid() bool {
	if !om.BaseRelationship.Valid() {
		return false
	}
	if om.MergerDate == "" {
		return false
	}
	validTransactionTypes := map[string]bool{
		TransactionTypeMerger:      true,
		TransactionTypeAcquisition: true,
		TransactionTypeSpinOff:     true,
	}
	if !validTransactionTypes[om.TransactionType] {
		return false
	}
	validStatuses := map[string]bool{
		TransactionStatusPending:   true,
		TransactionStatusCompleted: true,
		TransactionStatusCancelled: true,
	}
	if !validStatuses[om.Status] {
		return false
	}
	return true
}

// Description methods
func (ops *OrganizationParentSubsidiary) GetDescription() string {
	return "Represents a parent-subsidiary relationship between organizations, including ownership percentage and relationship details."
}

func (onh *OrganizationNameHistory) GetDescription() string {
	return "Represents historical name changes for organizations, tracking the evolution of organization names over time."
}

func (om *OrganizationMerger) GetDescription() string {
	return "Represents merger and acquisition relationships between organizations, including transaction details and status."
}

// Helper methods for working with relationships

// IsWhollyOwned checks if this is a wholly owned subsidiary relationship
func (ops *OrganizationParentSubsidiary) IsWhollyOwned() bool {
	return ops.RelationshipType == RelationshipTypeWhollyOwned || ops.OwnershipPercentage >= 99.0
}

// IsMajorityOwned checks if this is a majority owned subsidiary relationship
func (ops *OrganizationParentSubsidiary) IsMajorityOwned() bool {
	return ops.RelationshipType == RelationshipTypeMajorityOwned || (ops.OwnershipPercentage > 50.0 && ops.OwnershipPercentage < 99.0)
}

// IsCompleted checks if the merger/acquisition is completed
func (om *OrganizationMerger) IsCompleted() bool {
	return om.Status == TransactionStatusCompleted
}

// IsPending checks if the merger/acquisition is pending
func (om *OrganizationMerger) IsPending() bool {
	return om.Status == TransactionStatusPending
}

// GetTransactionValueFormatted returns formatted transaction value
func (om *OrganizationMerger) GetTransactionValueFormatted() string {
	if om.TransactionValue == 0 {
		return "Not disclosed"
	}
	currency := om.Currency
	if currency == "" {
		currency = "USD"
	}

	if om.TransactionValue >= 1e9 {
		return fmt.Sprintf("%.1fB %s", om.TransactionValue/1e9, currency)
	} else if om.TransactionValue >= 1e6 {
		return fmt.Sprintf("%.1fM %s", om.TransactionValue/1e6, currency)
	} else if om.TransactionValue >= 1e3 {
		return fmt.Sprintf("%.1fK %s", om.TransactionValue/1e3, currency)
	}
	return fmt.Sprintf("%.0f %s", om.TransactionValue, currency)
}

// OrganizationRelationshipService provides utilities for working with organization relationships
type OrganizationRelationshipService struct {
	organizations map[string]*Organization
	relationships []GraphRelationship
}

// NewOrganizationRelationshipService creates a new relationship service
func NewOrganizationRelationshipService() *OrganizationRelationshipService {
	return &OrganizationRelationshipService{
		organizations: make(map[string]*Organization),
		relationships: make([]GraphRelationship, 0),
	}
}

// AddOrganization adds an organization to the service
func (ors *OrganizationRelationshipService) AddOrganization(org *Organization) {
	ors.organizations[org.GetKey()] = org
}

// AddRelationship adds a relationship to the service
func (ors *OrganizationRelationshipService) AddRelationship(rel GraphRelationship) {
	ors.relationships = append(ors.relationships, rel)
}

// GetSubsidiaries returns all subsidiaries of a given organization
func (ors *OrganizationRelationshipService) GetSubsidiaries(orgKey string) []*Organization {
	var subsidiaries []*Organization

	for _, rel := range ors.relationships {
		if rel.Label() == OrganizationParentSubsidiaryLabel {
			source, target := rel.Nodes()
			if source.GetKey() == orgKey {
				if subsidiary, exists := ors.organizations[target.GetKey()]; exists {
					subsidiaries = append(subsidiaries, subsidiary)
				}
			}
		}
	}

	return subsidiaries
}

// GetParentOrganizations returns all parent organizations of a given organization
func (ors *OrganizationRelationshipService) GetParentOrganizations(orgKey string) []*Organization {
	var parents []*Organization

	for _, rel := range ors.relationships {
		if rel.Label() == OrganizationParentSubsidiaryLabel {
			source, target := rel.Nodes()
			if target.GetKey() == orgKey {
				if parent, exists := ors.organizations[source.GetKey()]; exists {
					parents = append(parents, parent)
				}
			}
		}
	}

	return parents
}

// GetNameHistory returns the name history for a given organization
func (ors *OrganizationRelationshipService) GetNameHistory(orgKey string) []OrganizationNameHistory {
	var history []OrganizationNameHistory

	for _, rel := range ors.relationships {
		if rel.Label() == OrganizationNameHistoryLabel {
			source, _ := rel.Nodes()
			if source.GetKey() == orgKey {
				if nameHistoryRel, ok := rel.(*OrganizationNameHistory); ok {
					history = append(history, *nameHistoryRel)
				}
			}
		}
	}

	return history
}

// GetOrganizationFamily returns all related organizations (parents, subsidiaries, siblings)
func (ors *OrganizationRelationshipService) GetOrganizationFamily(orgKey string) []*Organization {
	visited := make(map[string]bool)
	family := make([]*Organization, 0)

	// BFS to find all connected organizations
	queue := []string{orgKey}
	visited[orgKey] = true

	for len(queue) > 0 {
		currentKey := queue[0]
		queue = queue[1:]

		if org, exists := ors.organizations[currentKey]; exists {
			family = append(family, org)
		}

		// Add parents and subsidiaries to queue
		for _, rel := range ors.relationships {
			if rel.Label() == OrganizationParentSubsidiaryLabel {
				source, target := rel.Nodes()

				if source.GetKey() == currentKey && !visited[target.GetKey()] {
					queue = append(queue, target.GetKey())
					visited[target.GetKey()] = true
				}
				if target.GetKey() == currentKey && !visited[source.GetKey()] {
					queue = append(queue, source.GetKey())
					visited[source.GetKey()] = true
				}
			}
		}
	}

	return family
}
