package model

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// BreachIntelligenceAttribute represents breach intelligence data associated with an asset
type BreachIntelligenceAttribute struct {
	registry.BaseModel
	Username string `neo4j:"username" json:"username" desc:"Chariot username associated with the attribute." example:"user@example.com"`
	Key      string `neo4j:"key" json:"key" desc:"Unique key identifying the breach intelligence attribute." example:"#breach_intelligence#person#john.doe@example.com#John Doe#LinkedIn"`

	// Core breach intelligence fields
	AssetID       string  `neo4j:"asset_id" json:"asset_id" desc:"Asset identifier this breach intelligence belongs to." example:"#asset#example.com#example.com"`
	PersonKey     string  `neo4j:"person_key" json:"person_key,omitempty" desc:"Person key this breach intelligence belongs to." example:"#person#john.doe@example.com#John Doe"`
	EntryID       *string `neo4j:"entry_id,omitempty" json:"entry_id,omitempty" desc:"Unique identifier for this breach entry from DeHashed." example:"12345678"`
	BreachStatus  string  `neo4j:"breach_status" json:"breach_status" desc:"Status of breach check (BREACHED, CLEAN, UNKNOWN, NOT_CHECKED)." example:"BREACHED"`
	RiskLevel     string  `neo4j:"risk_level" json:"risk_level" desc:"Risk level based on breach data (CRITICAL, HIGH, MEDIUM, LOW, NONE)." example:"HIGH"`
	BreachCount   *int    `neo4j:"breach_count,omitempty" json:"breach_count,omitempty" desc:"Number of breaches detected." example:"3"`
	CheckedAt     string  `neo4j:"checked_at" json:"checked_at" desc:"Timestamp when breach check was performed (RFC3339)." example:"2024-02-04T10:00:00Z"`
	ExpiresAt     string  `neo4j:"expires_at" json:"expires_at,omitempty" desc:"Timestamp when this breach intelligence expires (RFC3339)." example:"2024-03-04T10:00:00Z"`

	// Detailed breach information
	MostRecentBreach string   `neo4j:"most_recent_breach" json:"most_recent_breach,omitempty" desc:"Timestamp of most recent breach (RFC3339)." example:"2023-12-15T00:00:00Z"`
	PasswordExposed  *bool    `neo4j:"password_exposed,omitempty" json:"password_exposed,omitempty" desc:"Whether passwords were exposed in breaches." example:"true"`
	DataClasses      []string `neo4j:"data_classes" json:"data_classes,omitempty" desc:"Types of data exposed in breaches." example:"[\"email\", \"password\", \"username\"]"`
	BreachSources    []string `neo4j:"breach_sources" json:"breach_sources,omitempty" desc:"Sources of breach data." example:"[\"HIBP\", \"DeHashed\"]"`
	// DeHashed per-entry fields
	DatabaseName    *string `neo4j:"database_name,omitempty" json:"database_name,omitempty" desc:"Name of the breached database." example:"LinkedIn"`
	ObtainedDate    *string `neo4j:"obtained_date,omitempty" json:"obtained_date,omitempty" desc:"Date the breach data was obtained (YYYY-MM-DD)." example:"2021-06-22"`
	HashedPassword  *string `neo4j:"hashed_password,omitempty" json:"hashed_password,omitempty" desc:"Hashed password found in the breach." example:"a9f8dfn49sj2g1jrgjsng43rgeo0w2"`
	PlaintextFound  *bool   `neo4j:"plaintext_found,omitempty" json:"plaintext_found,omitempty" desc:"Whether a plaintext password was found in the breach." example:"false"`
	BreachUsername  *string `neo4j:"breach_username,omitempty" json:"breach_username,omitempty" desc:"Username found in the breach entry." example:"johndoe"`
	BreachIPAddress *string `neo4j:"breach_ip_address,omitempty" json:"breach_ip_address,omitempty" desc:"IP address found in the breach entry." example:"192.168.1.1"`
	BreachName      *string `neo4j:"breach_name,omitempty" json:"breach_name,omitempty" desc:"Full name found in the breach entry." example:"John Doe"`
	BreachDOB       *string `neo4j:"breach_dob,omitempty" json:"breach_dob,omitempty" desc:"Date of birth found in the breach entry." example:"1990-01-15"`
	BreachAddress   *string `neo4j:"breach_address,omitempty" json:"breach_address,omitempty" desc:"Physical address found in the breach entry." example:"123 Main St"`
	BreachPhone     *string `neo4j:"breach_phone,omitempty" json:"breach_phone,omitempty" desc:"Phone number found in the breach entry." example:"+1-555-123-4567"`

	// HIBP enrichment fields (from free tier cross-reference by database name)
	PwnCount        *int       `neo4j:"pwn_count,omitempty" json:"pwn_count,omitempty" desc:"Total accounts affected in this breach (from HIBP free metadata)." example:"700000000"`
	HIBPDataClasses *[]string  `neo4j:"hibp_data_classes,omitempty" json:"hibp_data_classes,omitempty" desc:"Types of data exposed as reported by HIBP." example:"[\"email\", \"password\"]"`
	IsVerified      *bool      `neo4j:"is_verified,omitempty" json:"is_verified,omitempty" desc:"Whether this breach is verified by HIBP." example:"true"`

	RelevanceScore *float64 `neo4j:"relevance_score,omitempty" json:"relevance_score,omitempty" desc:"Relevance score for the breach (0.0-1.0)." example:"0.85"`

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
	return b.Key != "" && (b.AssetID != "" || b.PersonKey != "") && b.EntryID != nil && *b.EntryID != "" && b.BreachStatus != "" && b.RiskLevel != "" && b.CheckedAt != ""
}

func (b *BreachIntelligenceAttribute) SetSource(source string) {
	if len(source) > 7 && source[:8] == "#person#" {
		b.PersonKey = source
	} else {
		b.AssetID = source
	}
}

func (b *BreachIntelligenceAttribute) GetSource() string {
	if b.PersonKey != "" {
		return b.PersonKey
	}
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
				// Support both Person and Asset as parent
				if b.Parent.Model == nil && b.AssetID == "" && b.PersonKey == "" {
					return fmt.Errorf("parent (asset or person) or asset_id/person_key is required")
				}
				if b.Parent.Model != nil {
					parentKey := b.Parent.Model.GetKey()
					// Detect parent type from key prefix
					if len(parentKey) > 7 && parentKey[:8] == "#person#" {
						b.PersonKey = parentKey
					} else {
						b.AssetID = parentKey
					}
				}
				// Key includes database name and entry ID for per-entry uniqueness
				dbName := "unknown"
				if b.DatabaseName != nil && *b.DatabaseName != "" {
					dbName = *b.DatabaseName
				}
				entryID := ""
				if b.EntryID != nil && *b.EntryID != "" {
					entryID = *b.EntryID
				}
				if b.PersonKey != "" {
					b.Key = fmt.Sprintf("#breach_intelligence#%s#%s#%s", b.PersonKey, dbName, entryID)
				} else {
					b.Key = fmt.Sprintf("#breach_intelligence#%s#%s#%s", b.AssetID, dbName, entryID)
				}
				return nil
			},
		},
	}
}

// NewBreachIntelligenceAttribute creates a new breach intelligence attribute for an asset or person
func NewBreachIntelligenceAttribute(parentKey string, breachStatus string, riskLevel string, parent GraphModel) BreachIntelligenceAttribute {
	b := BreachIntelligenceAttribute{
		BreachStatus: breachStatus,
		RiskLevel:    riskLevel,
		CheckedAt:    Now(),
	}
	// Detect parent type from key prefix
	if len(parentKey) > 7 && parentKey[:8] == "#person#" {
		b.PersonKey = parentKey
	} else {
		b.AssetID = parentKey
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
	return "Represents a breach intelligence record for a person or asset, including breach details from DeHashed enriched with HIBP metadata."
}
