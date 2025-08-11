package model

import (
	"fmt"
	"github.com/praetorian-inc/tabularium/pkg/model/label"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

type Attribute struct {
	registry.BaseModel
	Username string `neo4j:"username" json:"username" desc:"Chariot username associated with the attribute." example:"user@example.com"`
	Key      string `neo4j:"key" json:"key" desc:"Unique key identifying the attribute." example:"#attribute#open_port#80#asset#example.com#example.com"`
	// Attributes
	Source     string            `neo4j:"source" json:"source" desc:"Key of the parent model this attribute belongs to." example:"#asset#example.com#example.com"`
	Name       string            `neo4j:"name" json:"name" desc:"Name of the attribute." example:"https"`
	Value      string            `neo4j:"value" json:"value" desc:"Value of the attribute." example:"443"`
	Status     string            `neo4j:"status" json:"status" desc:"Status of the attribute." example:"A"`
	Created    string            `neo4j:"created" json:"created" desc:"Timestamp when the attribute was created (RFC3339)." example:"2023-10-27T10:00:00Z"`
	Visited    string            `neo4j:"visited" json:"visited" desc:"Timestamp when the attribute was last visited or confirmed (RFC3339)." example:"2023-10-27T11:00:00Z"`
	Capability string            `neo4j:"capability" json:"capability,omitempty" desc:"Capability that discovered this attribute." example:"portscan"`
	TTL        int64             `neo4j:"ttl" json:"ttl" desc:"Time-to-live for the attribute record (Unix timestamp)." example:"1706353200"`
	Metadata   map[string]string `neo4j:"metadata" json:"metadata,omitempty" desc:"Additional metadata associated with the attribute." example:"{\"tool\": \"masscan\"}"`
	Parent     GraphModelWrapper `neo4j:"-" json:"parent" desc:"Attribute parent."`
}

var AttributeLabel = label.New("Attribute")

func init() {
	registry.Registry.MustRegisterModel(&Attribute{})
}

func (a *Attribute) GetKey() string {
	return a.Key
}

func (a *Attribute) GetLabels() []string {
	return []string{AttributeLabel, TTLLabel}
}

func (a *Attribute) Target() string {
	parts := strings.Split(a.Source, "#")
	if len(parts) < 3 {
		return ""
	}
	sourceDNS := parts[2]

	target := fmt.Sprintf("%s://%s:%s", a.Name, sourceDNS, a.Value)

	switch a.Name {
	case "port":
		target = strings.TrimPrefix(target, "port://")
	case "protocol":
		target = strings.Replace(target, "protocol://", a.Value+"://", 1)
		target = strings.TrimSuffix(target, ":"+a.Value)
	}

	return target
}

func (a *Attribute) Asset() Asset {
	parts := strings.Split(a.Source, "#")
	if len(parts) != 4 {
		return Asset{}
	}
	return NewAsset(parts[2], parts[3])
}

func (a *Attribute) Preseed() Preseed {
	parts := strings.Split(a.Value, "#")
	if len(parts) != 5 || !strings.HasPrefix(a.Value, "#preseed") {
		return Preseed{}
	}

	return NewPreseed(parts[2], parts[3], parts[4])
}

func (a *Attribute) Valid() bool {
	return a.Key != ""
}

func (a *Attribute) Visit(attr Attribute) {
	a.Visited = attr.Visited
	if attr.Status != Pending {
		a.Status = attr.Status
	}
	if a.TTL != 0 {
		a.TTL = attr.TTL
	}
	a.Capability = attr.Capability
	if len(attr.Metadata) > 0 {
		a.Metadata = attr.Metadata
	}
	a.Parent = attr.Parent
}

func (a *Attribute) IsStatus(value string) bool {
	return strings.HasPrefix(a.Status, value)
}

func (a *Attribute) IsClass(value string) bool {
	return strings.HasPrefix(a.Name, value)
}

func (a *Attribute) GetStatus() string {
	return a.Status
}

func (a *Attribute) WithStatus(status string) Target {
	a.Status = status
	return a
}

func (a *Attribute) Group() string {
	return a.Asset().DNS
}

func (a *Attribute) Identifier() string {
	return a.Target()
}

func (a *Attribute) IsPrivate() bool {
	parent := a.Asset()
	return parent.IsPrivate()
}

func (a *Attribute) Defaulted() {
	a.Status = Active
	a.Metadata = map[string]string{}
	a.Visited = Now()
	a.Created = Now()
	a.TTL = Future(14 * 24)
}

func (a *Attribute) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				if a.Parent.Model == nil {
					return fmt.Errorf("parent is required")
				}
				template := fmt.Sprintf("#attribute#%s#%%s%s", a.Name, a.Parent.Model.GetKey())
				shortenedValue := a.Value[:min(1024-len(template), len(a.Value))]
				a.Key = fmt.Sprintf(template, shortenedValue)
				a.Source = a.Parent.Model.GetKey()
				return nil
			},
		},
	}
}

func NewAttribute(name, value string, parent GraphModel) Attribute {
	a := Attribute{
		Name:   name,
		Value:  value,
		Parent: NewGraphModelWrapper(parent),
	}
	a.Defaulted()
	registry.CallHooks(&a)
	return a
}

// GetDescription returns a description for the Attribute model.
func (a *Attribute) GetDescription() string {
	return "Represents a key-value pair attribute associated with an entity, often used for tagging or additional properties."
}
