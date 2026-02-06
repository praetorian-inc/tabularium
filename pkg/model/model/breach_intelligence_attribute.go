package model

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// BreachIntelligenceAttribute represents breach intelligence data associated with an asset
type BreachIntelligenceAttribute struct {
	registry.BaseModel
	Username string `neo4j:"username" json:"username" desc:"Chariot username associated with the attribute." example:"user@example.com"`
	Key      string `neo4j:"key" json:"key" desc:"Unique key identifying the breach intelligence attribute." example:"#breach_intelligence#asset#example.com#example.com"`

	// Core breach intelligence fields
	AssetID       string   `neo4j:"asset_id" json:"asset_id" desc:"Asset identifier this breach intelligence belongs to." example:"#asset#example.com#example.com"`
	BreachStatus  string   `neo4j:"breach_status" json:"breach_status" desc:"Status of breach check (BREACHED, CLEAN, UNKNOWN, NOT_CHECKED)." example:"BREACHED"`
	RiskLevel     string   `neo4j:"risk_level" json:"risk_level" desc:"Risk level based on breach data (CRITICAL, HIGH, MEDIUM, LOW, NONE)." example:"HIGH"`
	BreachCount   int      `neo4j:"breach_count" json:"breach_count,omitempty" desc:"Number of breaches detected." example:"3"`
	CheckedAt     string   `neo4j:"checked_at" json:"checked_at" desc:"Timestamp when breach check was performed (RFC3339)." example:"2024-02-04T10:00:00Z"`
	ExpiresAt     string   `neo4j:"expires_at" json:"expires_at,omitempty" desc:"Timestamp when this breach intelligence expires (RFC3339)." example:"2024-03-04T10:00:00Z"`

	// Detailed breach information
	MostRecentBreach string   `neo4j:"most_recent_breach" json:"most_recent_breach,omitempty" desc:"Timestamp of most recent breach (RFC3339)." example:"2023-12-15T00:00:00Z"`
	PasswordExposed  bool     `neo4j:"password_exposed" json:"password_exposed,omitempty" desc:"Whether passwords were exposed in breaches." example:"true"`
	DataClasses      []string `neo4j:"data_classes" json:"data_classes,omitempty" desc:"Types of data exposed in breaches." example:"[\"email\", \"password\", \"username\"]"`
	BreachSources    []string `neo4j:"breach_sources" json:"breach_sources,omitempty" desc:"Sources of breach data." example:"[\"HIBP\", \"DeHashed\"]"`
	RelevanceScore   float64  `neo4j:"relevance_score" json:"relevance_score,omitempty" desc:"Relevance score for the breach (0.0-1.0)." example:"0.85"`

	// Metadata
	RawFindings map[string]interface{} `neo4j:"raw_findings" json:"raw_findings,omitempty" desc:"Raw JSON findings from breach check." example:"{\"breaches\": [{\"name\": \"LinkedIn\"}]}"`
	Status      string                 `neo4j:"status" json:"status" desc:"Status of the attribute record." example:"A"`
	Created     string                 `neo4j:"created" json:"created" desc:"Timestamp when the attribute was created (RFC3339)." example:"2024-02-04T10:00:00Z"`
	Visited     string                 `neo4j:"visited" json:"visited" desc:"Timestamp when the attribute was last visited (RFC3339)." example:"2024-02-04T11:00:00Z"`
	TTL         int64                  `neo4j:"ttl" json:"ttl" desc:"Time-to-live for the attribute record (Unix timestamp)." example:"1706353200"`
	Parent      GraphModelWrapper      `neo4j:"-" json:"parent" desc:"Attribute parent asset."`
}

const BreachIntelligenceAttributeLabel = "BreachIntelligenceAttribute"

func init() {
	registry.Registry.MustRegisterModel(&BreachIntelligenceAttribute{})
}

func (b *BreachIntelligenceAttribute) GetKey() string {
	return b.Key
}

func (b *BreachIntelligenceAttribute) GetLabels() []string {
	return []string{BreachIntelligenceAttributeLabel, AttributeLabel, TTLLabel}
}

func (b *BreachIntelligenceAttribute) Valid() bool {
	return b.Key != "" && b.AssetID != "" && b.BreachStatus != "" && b.RiskLevel != "" && b.CheckedAt != ""
}

func (b *BreachIntelligenceAttribute) SetSource(source string) {
	// Parent source tracking if needed
}

func (b *BreachIntelligenceAttribute) GetSource() string {
	return b.AssetID
}

func (b *BreachIntelligenceAttribute) Defaulted() {
	b.Status = Active
	b.Visited = Now()
	b.Created = Now()
	b.TTL = Future(30 * 24) // 30 days default TTL for breach intelligence
	if b.BreachStatus == "" {
		b.BreachStatus = "NOT_CHECKED"
	}
	if b.RiskLevel == "" {
		b.RiskLevel = "NONE"
	}
}

func (b *BreachIntelligenceAttribute) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				if b.Parent.Model == nil && b.AssetID == "" {
					return fmt.Errorf("parent asset or asset_id is required")
				}
				if b.Parent.Model != nil {
					b.AssetID = b.Parent.Model.GetKey()
				}
				b.Key = fmt.Sprintf("#breach_intelligence#%s", b.AssetID)
				return nil
			},
		},
	}
}

// NewBreachIntelligenceAttribute creates a new breach intelligence attribute for an asset
func NewBreachIntelligenceAttribute(assetID string, breachStatus string, riskLevel string, parent GraphModel) BreachIntelligenceAttribute {
	b := BreachIntelligenceAttribute{
		AssetID:      assetID,
		BreachStatus: breachStatus,
		RiskLevel:    riskLevel,
		CheckedAt:    Now(),
	}
	if parent != nil {
		b.Parent = NewGraphModelWrapper(parent)
	}
	b.Defaulted()
	registry.CallHooks(&b)
	return b
}

// GetDescription returns a description for the BreachIntelligenceAttribute model.
func (b *BreachIntelligenceAttribute) GetDescription() string {
	return "Represents breach intelligence data for an asset, including breach status, risk level, and detailed findings from HIBP/DeHashed."
}
