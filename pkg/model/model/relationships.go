package model

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// GetDescription returns a description for the BaseRelationship model.
func (br *BaseRelationship) GetDescription() string {
	return "Represents the base structure for all graph relationships, containing source, target, and metadata."
}

type BaseRelationship struct {
	// Source and Target are used internally for graph construction, not stored directly.
	registry.BaseModel
	Source         GraphModel `neo4j:"-" json:"-"`
	Target         GraphModel `neo4j:"-" json:"-"`
	Created        string     `neo4j:"created" json:"created" desc:"Timestamp when the relationship was created (RFC3339)." example:"2023-10-27T10:00:00Z"`
	Visited        string     `neo4j:"visited" json:"visited" desc:"Timestamp when the relationship was last visited or confirmed (RFC3339)." example:"2023-10-27T11:00:00Z"`
	Capability     string     `neo4j:"capability" json:"capability" desc:"The capability or tool that discovered/created this relationship." example:"portscan"`
	Key            string     `neo4j:"key" json:"key" desc:"Unique key identifying the relationship." example:"<source_key>#DISCOVERED#<target_key>"`
	AttachmentPath string     `neo4j:"attachmentPath" json:"attachmentPath"`
	Attachment     File       `neo4j:"-" json:"attachment"`
}

func init() {
	registry.Registry.MustRegisterModel(&BaseRelationship{})
	registry.Registry.MustRegisterModel(&Discovered{})
	registry.Registry.MustRegisterModel(&HasVulnerability{})
	registry.Registry.MustRegisterModel(&InstanceOf{})
	registry.Registry.MustRegisterModel(&HasAttribute{})
	registry.Registry.MustRegisterModel(&HasTechnology{})
	registry.Registry.MustRegisterModel(&HasCredential{})
	registry.Registry.MustRegisterModel(&HasOrganizationName{})
}

func (br *BaseRelationship) GetKey() string {
	return br.Key
}

func (base *BaseRelationship) Base() *BaseRelationship {
	return base
}

func (base *BaseRelationship) Visit(other GraphRelationship) {
	base.Visited = other.Base().Visited
	if other.Base().Capability != "" {
		base.Capability = other.Base().Capability
	}
	base.Source = other.Base().Source
	base.Target = other.Base().Target
	if other.Base().AttachmentPath != "" {
		base.AttachmentPath = other.Base().AttachmentPath
		base.Attachment = other.Base().Attachment
	}
}

func (base *BaseRelationship) Valid() bool {
	return base.Key != ""
}

func (base *BaseRelationship) Nodes() (GraphModel, GraphModel) {
	return base.Source, base.Target
}

func NewBaseRelationship(source, target GraphModel, label string) *BaseRelationship {
	return &BaseRelationship{
		Source:  source,
		Target:  target,
		Created: Now(),
		Visited: Now(),
		Key:     fmt.Sprintf("%s#%s%s", source.GetKey(), label, target.GetKey()),
	}
}

// GetDescription returns a description for the Discovered relationship model.
func (d *Discovered) GetDescription() string {
	return "Represents a discovery relationship between two entities (e.g., a host discovered a service)."
}

type Discovered struct {
	*BaseRelationship
}

func NewDiscovered(source, target GraphModel) GraphRelationship {
	return &Discovered{
		BaseRelationship: NewBaseRelationship(source, target, DiscoveredLabel),
	}
}

const DiscoveredLabel = "DISCOVERED"

func (d Discovered) Label() string {
	return DiscoveredLabel
}

// GetDescription returns a description for the HasVulnerability relationship model.
func (hv *HasVulnerability) GetDescription() string {
	return "Represents the relationship indicating an asset has a specific vulnerability."
}

type HasVulnerability struct {
	*BaseRelationship
}

func NewHasVulnerability(source, target GraphModel) GraphRelationship {
	return &HasVulnerability{
		BaseRelationship: NewBaseRelationship(source, target, HasVulnerabilityLabel),
	}
}

const HasVulnerabilityLabel = "HAS_VULNERABILITY"

func (a HasVulnerability) Label() string {
	return HasVulnerabilityLabel
}

// A pointer to HasVulnerability since its a GraphRelationship
func (hv *HasVulnerability) Hydrate() (string, func([]byte) error) {
	return hv.AttachmentPath, func(data []byte) error {
		hv.Base().Attachment = NewFile(hv.AttachmentPath)
		hv.Base().Attachment.Bytes = data
		return nil
	}
}

func (hv *HasVulnerability) Dehydrate() (File, Hydratable) {
	copy := hv.Attachment
	hv.Base().Attachment = File{}
	return copy, hv
}

// GetDescription returns a description for the InstanceOf relationship model.
func (io *InstanceOf) GetDescription() string {
	return "Represents an 'instance of' relationship (e.g., a process is an instance of a software package)."
}

type InstanceOf struct {
	*BaseRelationship
}

func NewInstanceOf(source, target GraphModel) GraphRelationship {
	return &InstanceOf{
		BaseRelationship: NewBaseRelationship(source, target, InstanceOfLabel),
	}
}

const InstanceOfLabel = "INSTANCE_OF"

func (a InstanceOf) Label() string {
	return InstanceOfLabel
}

const HasAttributeLabel = "HAS_ATTRIBUTE"

// GetDescription returns a description for the HasAttribute relationship model.
func (ha *HasAttribute) GetDescription() string {
	return "Represents the relationship indicating an entity has a specific attribute."
}

type HasAttribute struct {
	*BaseRelationship
}

func NewHasAttribute(source, target GraphModel) GraphRelationship {
	return &HasAttribute{
		BaseRelationship: NewBaseRelationship(source, target, HasAttributeLabel),
	}
}

func (a HasAttribute) Label() string {
	return HasAttributeLabel
}

const HasTechnologyLabel = "HAS_TECHNOLOGY"

// GetDescription returns a description for the HasTechnology relationship model.
func (ht *HasTechnology) GetDescription() string {
	return "Represents the relationship indicating an asset uses or runs a specific technology."
}

type HasTechnology struct {
	*BaseRelationship
}

func NewHasTechnology(source, target GraphModel) GraphRelationship {
	return &HasTechnology{
		BaseRelationship: NewBaseRelationship(source, target, HasTechnologyLabel),
	}
}

func (a HasTechnology) Label() string {
	return HasTechnologyLabel
}

const HasCredentialLabel = "HAS_CREDENTIAL"

// GetDescription returns a description for the HasCredential relationship model.
func (hc *HasCredential) GetDescription() string {
	return "Represents the relationship indicating an entity has a specific credential."
}

type HasCredential struct {
	*BaseRelationship
}

func NewCredentialRelationship(asset GraphModel, credential *Credential) GraphRelationship {
	return &HasCredential{
		BaseRelationship: NewBaseRelationship(asset, credential, HasCredentialLabel),
	}
}

func (hc *HasCredential) Label() string {
	return HasCredentialLabel
}

const HasOrganizationNameLabel = "HAS_ORGANIZATION_NAME"

// GetDescription returns a description for the HasOrganizationName relationship model.
func (hon *HasOrganizationName) GetDescription() string {
	return "Represents the relationship between an organization and one of its name variations."
}

type HasOrganizationName struct {
	*BaseRelationship
}

func NewHasOrganizationName(organization *Organization, organizationName *OrganizationName) GraphRelationship {
	return &HasOrganizationName{
		BaseRelationship: NewBaseRelationship(organization, organizationName, HasOrganizationNameLabel),
	}
}

func (hon *HasOrganizationName) Label() string {
	return HasOrganizationNameLabel
}
