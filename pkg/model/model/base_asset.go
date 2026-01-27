package model

import (
	"reflect"
	"strings"
	"time"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

type BaseAsset struct {
	registry.BaseModel
	Username string `neo4j:"username" json:"username" desc:"The username associated with this asset." example:"user@example.com"`
	Key      string `neo4j:"key" json:"key" desc:"Unique key identifying the asset." example:"#asset#dns#name"`
	// Attributes
	Origin     string  `neo4j:"origin" json:"origin" desc:"The first user/capability that added this to the system." example:"whois"`
	Source     string  `neo4j:"source" json:"source" desc:"Source that added this to the system (one of self, account, seed)" example:"seed"`
	Status     string  `neo4j:"status" json:"status" desc:"Current status of the asset." example:"A"`
	Created    string  `neo4j:"created" json:"created" desc:"Timestamp when the asset was first created (RFC3339)." example:"2023-10-27T10:00:00Z"`
	Visited    string  `neo4j:"visited" json:"visited" desc:"Timestamp when the asset was last visited (RFC3339)." example:"2023-10-27T11:00:00Z"`
	TTL        int64   `neo4j:"ttl" json:"ttl" desc:"Time-to-live for the asset record (in hours)." example:"168"`
	Secret     *string `neo4j:"secret" json:"secret,omitempty" desc:"Key of the secret to be used with this asset." example:"#asset#amazon#0123456789012"`
	Comment    string  `neo4j:"-" json:"comment,omitempty" desc:"User-provided comment about the asset." example:"Initial asset discovery"`
	Identifier string  `neo4j:"identifier" json:"identifier" desc:"Unique identifier for the asset." example:"name"`
	Group      string  `neo4j:"group" json:"group" desc:"Group of the asset." example:"dns"`
	Class      string  `neo4j:"class" json:"class" desc:"Classification of the asset type." example:"repository"`
	History
	MLProperties
	Metadata
	Tags
}

func init() {
	registry.Registry.MustRegisterModel(&Metadata{})
}

func (a *BaseAsset) GetKey() string {
	return a.Key
}

func (a *BaseAsset) GetBase() *BaseAsset {
	return a
}

func (a *BaseAsset) GetClass() string {
	return a.Class
}

func (a *BaseAsset) SetStatus(status string) {
	a.Status = status
}

func (a *BaseAsset) GetStatus() string {
	return a.Status
}

func (a *BaseAsset) GetMetadata() *Metadata {
	return &a.Metadata
}

func (a *BaseAsset) GetSource() string {
	return a.Source
}

func (a *BaseAsset) SetSource(source string) {
	// a seed or account source should always win over other sources
	if a.Source == SeedSource || a.Source == AccountSource {
		return
	}
	a.Source = source
}

func (a *BaseAsset) GetOrigin() string {
	return a.Origin
}

func (a *BaseAsset) SetOrigin(origin string) {
	a.Origin = origin
}

func (a *BaseAsset) IsStatus(value string) bool {
	return strings.HasPrefix(a.Status, value)
}

// IsValidStatus checks if a status value is one of the valid status constants.
// Empty string is considered valid (for new items without status set yet).
func IsValidStatus(status string) bool {
	if status == "" {
		return true
	}
	validStatuses := []string{
		Deleted,
		Pending,
		Active,
		Frozen,
		FrozenRejected,
		ActiveLow,
		ActivePassive,
		ActiveHigh,
	}
	for _, valid := range validStatuses {
		if status == valid {
			return true
		}
	}
	return false
}

func (a *BaseAsset) IsClass(value string) bool {
	return strings.HasPrefix(a.Class, value)
}

func (a *BaseAsset) IsPrivate() bool {
	return false
}

func (a *BaseAsset) Merge(u Assetlike) {
	update := u.GetBase()
	if a.History.Update(a.Status, update.Status, update.Source, update.Comment, update.History) {
		a.Status = update.Status
	}
	if !a.IsStatus(Active) {
		a.TTL = 0
	}
	if a.Origin == "" {
		a.Origin = update.Origin
	}
	a.Metadata.Merge(update.Metadata)
	a.Tags.Merge(update.Tags)
}

func (a *BaseAsset) Visit(o Assetlike) {
	other := o.GetBase()
	a.Visited = other.Visited
	if a.Source == SelfSource && a.IsStatus(Pending) && other.IsStatus(Active) {
		a.Status = other.Status
	}
	if a.IsStatus(Active) && a.TTL != 0 {
		a.TTL = other.TTL
	}
	if IsPermanentSource(other.Source) {
		a.TTL = 0
	}
	if other.TTL == 0 {
		a.TTL = 0
	}
	if a.Origin == "" {
		a.Origin = other.Origin
	}

	a.Secret = other.Secret
	a.Metadata.Visit(other.Metadata)
	a.Tags.Visit(other.Tags)
}

func (a *BaseAsset) System() bool {
	return !IsPermanentSource(a.Source)
}

func (a *BaseAsset) State() string {
	return string(a.Status[0])
}

func (a *BaseAsset) Substate() string {
	if len(a.Status) > 1 {
		return a.Status[1:]
	}
	return ""
}

func (a *BaseAsset) SetStatusFromLastSeen(lastSeenStr string, layout string) {
	a.Status = Pending
	if lastSeen, err := time.Parse(layout, lastSeenStr); err == nil {
		if time.Since(lastSeen) < 24*time.Hour {
			a.Status = Active
		}
	}
}

func (a *BaseAsset) SetUsername(username string) {
	a.Username = username
}

func (a *BaseAsset) GetAgent() string {
	return a.MLProperties.Agent
}

// GetSecret returns the secret reference for this asset
func (a *BaseAsset) GetSecret() string {
	if a.Secret != nil {
		return *a.Secret
	}
	return ""
}

// GetPartitionKey returns a partition key for load distribution.
// Uses the Identifier field for natural partitioning by asset identity.
// This ensures jobs for the same asset are grouped together while
// distributing load across different assets.
func (a *BaseAsset) GetPartitionKey() string {
	return a.Identifier
}

func (a *BaseAsset) Defaulted() {
	a.Status = Active
	a.Source = SelfSource
	a.Created = Now()
	a.Visited = Now()
	a.TTL = Future(30 * 24)
}

func NewBaseAsset(identifier, group string) BaseAsset {
	a := BaseAsset{Identifier: identifier, Group: group}
	// don't call Defaulted or CallHooks here; leave that to the parent/caller
	return a
}

// Metadata is a collection of 1:1 fields that can be present on an asset, as well as 1:many
// fields where relationships and string searching (enums, essentially) are not relevant.
// This is partially an experiment. I'd like to see how this grows/changes over time.
// The goal is to replace all 1:1 relationships and enums currently represented by attributes with a property
// of the asset itself.
type Metadata struct {
	registry.BaseModel
	ASNumber string `neo4j:"asnumber,omitempty" json:"asnumber,omitempty" desc:"Autonomous System number." example:"AS15169"`
	ASName   string `neo4j:"asname,omitempty" json:"asname,omitempty" desc:"Autonomous System name." example:"GOOGLE"`
	ASRange  string `neo4j:"asrange,omitempty" json:"asrange,omitempty" desc:"Autonomous System IP range." example:"172.217.0.0/16"`

	Country  string `neo4j:"country,omitempty" json:"country,omitempty" desc:"Country associated with the asset." example:"US"`
	Province string `neo4j:"province,omitempty" json:"province,omitempty" desc:"Province or state associated with the asset." example:"California"`
	City     string `neo4j:"city,omitempty" json:"city,omitempty" desc:"City associated with the asset." example:"Mountain View"`

	Purchased  string `neo4j:"purchased,omitempty" json:"purchased,omitempty" desc:"Date the asset (e.g., domain) was purchased (RFC3339)." example:"2002-09-15T00:00:00Z"`
	Updated    string `neo4j:"updated,omitempty" json:"updated,omitempty" desc:"Date the asset registration was last updated (RFC3339)." example:"2023-09-15T10:00:00Z"`
	Expiration string `neo4j:"expiration,omitempty" json:"expiration,omitempty" desc:"Date the asset registration expires (RFC3339)." example:"2024-09-15T00:00:00Z"`

	Registrant string `neo4j:"registrant,omitempty" json:"registrant,omitempty" desc:"Registered owner of the asset (e.g., domain)." example:"Google LLC"`
	Registrar  string `neo4j:"registrar,omitempty" json:"registrar,omitempty" desc:"Registrar managing the asset (e.g., domain)." example:"MarkMonitor Inc."`
	Email      string `neo4j:"email,omitempty" json:"email,omitempty" desc:"Optional contact email associated with the seed." example:"contact@example.com"`

	CloudService string `neo4j:"cloudService,omitempty" json:"cloudService,omitempty" desc:"Name of the cloud service provider (e.g., AWS, GCP, Azure)." example:"GCP"`
	CloudId      string `neo4j:"cloudId,omitempty" json:"cloudId,omitempty" desc:"Unique identifier within the cloud provider." example:"project-id-12345"`
	CloudRoot    string `neo4j:"cloudRoot,omitempty" json:"cloudRoot,omitempty" desc:"Root identifier for the cloud environment (e.g., organization ID)." example:"organizations/1234567890"`
	CloudAccount string `neo4j:"cloudAccount,omitempty" json:"cloudAccount,omitempty" desc:"Specific account identifier within the cloud provider." example:"billing-account-id"`
	OriginationData
}

func (m *Metadata) Merge(other Metadata) {
	m.updateFields(other)
	m.OriginationData.Merge(other.OriginationData)
}

func (m *Metadata) Visit(other Metadata) {
	m.updateFields(other)
	m.OriginationData.Visit(other.OriginationData)
}

// updateFields will copy over any non-empty fields from the other metadata into this metadata.
// Uses reflection here to maintain type-safety elsewhere in the codebase
func (m *Metadata) updateFields(other Metadata) {
	v := reflect.ValueOf(m).Elem()
	otherV := reflect.ValueOf(other)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		otherField := otherV.Field(i)

		if field.Kind() == reflect.String && otherField.String() != "" {
			field.SetString(otherField.String())
		}
	}
}

// GetDescription returns a description for the Metadata model.
func (m *Metadata) GetDescription() string {
	return "Contains metadata about an asset, including discovery information, relationships, and attributes."
}

// GetDescription returns a description for the Asset model.
func (a *BaseAsset) GetDescription() string {
	return "Base logic for an asset."
}
