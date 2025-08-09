package model

import (
	"fmt"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/model/filters"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

type Preseed struct {
	registry.BaseModel
	Username   string `neo4j:"username" json:"username" desc:"Chariot username associated with the preseed record." example:"user@example.com"`
	Key        string `neo4j:"key" json:"key" desc:"Unique key identifying the preseed record." example:"#preseed#whois#registrant_email#test@example.com"`
	Type       string `neo4j:"type" json:"type" desc:"Type of the preseed data." example:"whois"`
	Title      string `neo4j:"title" json:"title" desc:"Title or category within the preseed type." example:"registrant_email"`
	Value      string `neo4j:"value" json:"value" desc:"The actual preseed value." example:"test@example.com"`
	Display    string `neo4j:"display" json:"display" desc:"Hint for UI display type (e.g., text, image, base64)." example:"text"`
	Status     string `neo4j:"status" json:"status" desc:"Status of the preseed record." example:"A"`
	Created    string `neo4j:"created" json:"created" desc:"Timestamp when the preseed record was created (RFC3339)." example:"2023-10-27T10:00:00Z"`
	Visited    string `neo4j:"visited" json:"visited" desc:"Timestamp when the preseed record was last visited or processed (RFC3339)." example:"2023-10-27T11:00:00Z"`
	Capability string `neo4j:"capability" json:"capability,omitempty" desc:"Capability associated with processing this preseed record." example:"whois-lookup"`
	TTL        int64  `neo4j:"ttl" json:"ttl" desc:"Time-to-live for the preseed record (Unix timestamp)." example:"1706353200"`
}

func init() {
	registry.Registry.MustRegisterModel(&Preseed{})
}

var PreseedLabel = NewLabel("Preseed")

func (p *Preseed) IsPrivate() bool {
	return false
}

func (p *Preseed) GetLabels() []string {
	return []string{PreseedLabel, TTLLabel}
}

func (p *Preseed) Class() string {
	return strings.Split(p.Type, "+")[0]
}

func (p *Preseed) Visit(other Preseed) {
	if other.Status != Pending {
		p.Status = other.Status
	}
	if p.TTL != 0 {
		p.TTL = other.TTL
	}
	p.Visited = other.Visited
}

func (p *Preseed) Valid() bool {
	return p.Key != ""
}

func (p *Preseed) ToAttribute(source GraphModel, metadata ...string) Attribute {
	attr := NewAttribute("preseed", p.Key, source)
	for i := 0; i < len(metadata); i += 2 {
		attr.Metadata[metadata[i]] = metadata[i+1]
	}
	return attr
}

func (p *Preseed) Synonymous() []filters.Filter {
	switch p.Class() {
	case "whois":
		// match other whois attributes with the same values
		return []filters.Filter{
			filters.NewFilter("type", filters.OperatorStartsWith, "whois"),
			filters.NewFilter("title", filters.OperatorEqual, p.Title),
			filters.NewFilter("value", filters.OperatorEqual, p.Value),
		}
	default:
		// otherwise, exact match
		return []filters.Filter{
			filters.NewFilter("key", filters.OperatorEqual, p.Key),
		}
	}
}

func (p *Preseed) GetKey() string {
	return p.Key
}

func (p *Preseed) GetStatus() string {
	return p.Status
}

func (p *Preseed) WithStatus(status string) Target {
	p.Status = status
	return p
}

func (p *Preseed) Group() string {
	return p.Type
}

func (p *Preseed) Identifier() string {
	return p.Value
}

func (p *Preseed) IsStatus(value string) bool {
	return strings.HasPrefix(p.Status, value)
}

func (p *Preseed) IsClass(value string) bool {
	return strings.HasPrefix(p.Type, value)
}

func (p *Preseed) Defaulted() {
	p.Status = Pending
	p.Visited = Now()
	p.Created = Now()
	p.TTL = Future(14 * 24)
	p.Display = generatePreseedDisplay(p.Type)
}

func (p *Preseed) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				keyTemplate := fmt.Sprintf("#preseed#%s#%s#%%s", p.Type, p.Title)
				shortenedValue := p.Value[:min(1024-len(keyTemplate), len(p.Value))]
				p.Key = fmt.Sprintf(keyTemplate, shortenedValue)
				return nil
			},
		},
	}
}

func NewPreseed(preseedType, title, value string) Preseed {
	p := Preseed{
		Type:  preseedType,
		Title: title,
		Value: value,
	}
	p.Defaulted()
	registry.CallHooks(&p)
	return p
}

func generatePreseedDisplay(preseedType string) string {
	if preseedType == "csp" {
		return "base64"
	} else if preseedType == "favicon" {
		return "image"
	} else if preseedType == "tlscert" {
		return "tlscert"
	}
	return "text"
}

// GetDescription returns a description for the Preseed model.
func (p *Preseed) GetDescription() string {
	return "Represents pre-seeded information about an asset or entity, often used to bootstrap discovery."
}
