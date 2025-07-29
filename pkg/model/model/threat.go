package model

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

const ThreatLabel = "Threat"

func init() {
	registry.Registry.MustRegisterModel(&Threat{})
}

// GetDescription returns a description for the Threat model.
func (t *Threat) GetDescription() string {
	return "Represents a known threat, such as a malware family or campaign."
}

type Threat struct {
	registry.BaseModel
	baseTableModel
	Username string `dynamodbav:"username" json:"username" desc:"Chariot username associated with the threat data." example:"system"`
	Key      string `dynamodbav:"key" json:"key" desc:"Unique key identifying the threat record." example:"#threat#vulncheck#CVE-2023-12345"`
	// Attributes
	Source  string `dynamodbav:"source" json:"source" desc:"Identifier for the source of the threat data (e.g., CVE ID)." example:"CVE-2023-12345"`
	Data    any    `dynamodbav:"data" json:"data" desc:"The actual threat intelligence data." example:"{\"cvss\": 9.8, \"exploits\": [\"metasploit\"]}"`
	Created string `dynamodbav:"created" json:"created" desc:"Timestamp associated with the threat data creation (often includes feed name)." example:"#threat#vulncheck#2023-11-01T00:00:00Z"`
	Updated string `dynamodbav:"updated" json:"updated" desc:"Timestamp when the threat record was last updated (RFC3339)." example:"2023-11-10T12:00:00Z"`
	Feed    string `dynamodbav:"-" json:"feed" desc:"The feed this threat was found in"`
}

func (t *Threat) Defaulted() {
	t.Updated = Now()
}

func (t *Threat) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				t.Created = fmt.Sprintf("#threat#%s#%s", t.Feed, t.Created)
				k := fmt.Sprintf("#threat#%s#%s", t.Feed, t.Source)
				t.Key = k[:min(1024, len(k))]
				return nil
			},
		},
	}
}

func NewThreat(feed string, cve string, created string, data any) Threat {
	t := Threat{
		Source:  cve,
		Data:    data,
		Feed:    feed,
		Created: created,
	}
	t.Defaulted()
	registry.CallHooks(&t)
	return t
}
