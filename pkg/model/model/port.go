package model

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

type PortProtocol string

const (
	PortProtocolTCP PortProtocol = "tcp"
	PortProtocolUDP PortProtocol = "udp"
)

type Port struct {
	registry.BaseModel
	Username   string            `neo4j:"username" json:"username" desc:"Chariot username associated with the port." example:"user@example.com"`
	Key        string            `neo4j:"key" json:"key" desc:"Unique key identifying the port." example:"#port#tcp#80#asset#example.com#example.com"`
	Source     string            `neo4j:"source" json:"source" desc:"Key of the parent asset this port belongs to." example:"#asset#example.com#example.com"`
	Protocol   string            `neo4j:"protocol" json:"protocol" desc:"The protocol of this port." example:"tcp"`
	PortNumber int               `neo4j:"port" json:"port" desc:"The port number of this port." example:"80"`
	Service    string            `neo4j:"service" json:"service" desc:"The name of the service identified on this port." example:"https"`
	Status     string            `neo4j:"status" json:"status" desc:"Status of the port." example:"A"`
	Created    string            `neo4j:"created" json:"created" desc:"Timestamp when the port was created (RFC3339)." example:"2023-10-27T10:00:00Z"`
	Visited    string            `neo4j:"visited" json:"visited" desc:"Timestamp when the port was last visited or confirmed (RFC3339)." example:"2023-10-27T11:00:00Z"`
	TTL        int64             `neo4j:"ttl" json:"ttl" desc:"Time-to-live for the port record (Unix timestamp)." example:"1706353200"`
	Parent     GraphModelWrapper `neo4j:"-" json:"parent" desc:"Port parent asset."`
}

const PortLabel = "Port"

func init() {
	registry.Registry.MustRegisterModel(&Port{})
}

func (p *Port) GetKey() string {
	return p.Key
}

func (p *Port) GetLabels() []string {
	return []string{PortLabel, TTLLabel}
}

func (p *Port) Target() string {
	asset := p.Asset()
	if p.Service != "" {
		return fmt.Sprintf("%s://%s:%d", p.Service, asset.DNS, p.PortNumber)
	}
	return fmt.Sprintf("%s:%d", asset.Name, p.PortNumber)
}

func (p *Port) Asset() Asset {
	parts := strings.Split(p.Source, "#")
	if len(parts) != 4 {
		return Asset{}
	}
	return NewAsset(parts[2], parts[3])
}

func (p *Port) Valid() bool {
	return p.Key != "" && p.PortNumber > 0 && p.PortNumber <= 65535
}

func (p *Port) Visit(other Port) {
	p.Visited = other.Visited
	if other.Status != Pending {
		p.Status = other.Status
	}
	if other.TTL != 0 {
		p.TTL = other.TTL
	}
	if other.Service != "" {
		p.Service = other.Service
	}
	p.Parent = other.Parent
}

func (p *Port) IsClass(value string) bool {
	return strings.HasPrefix(p.Service, value) || fmt.Sprintf("%v", p.PortNumber) == value
}

func (p *Port) GetStatus() string {
	return p.Status
}

func (p *Port) WithStatus(status string) Target {
	p.Status = status
	return p
}

func (p *Port) IsStatus(value string) bool {
	return strings.HasPrefix(p.Status, value)
}

func (p *Port) Group() string {
	return p.Asset().DNS
}

func (p *Port) Identifier() string {
	return p.Target()
}

func (p *Port) IsPrivate() bool {
	parent := p.Asset()
	return parent.IsPrivate()
}

func (p *Port) Defaulted() {
	p.Status = Active
	p.Visited = Now()
	p.Created = Now()
	p.TTL = Future(14 * 24)
}

func (p *Port) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				if p.Parent.Model == nil {
					return fmt.Errorf("parent is required")
				}
				p.Key = fmt.Sprintf("#port#%s#%d%s", p.Protocol, p.PortNumber, p.Parent.Model.GetKey())
				p.Source = p.Parent.Model.GetKey()
				return nil
			},
		},
	}
}

func NewPort(protocol string, portNumber int, parent GraphModel) Port {
	p := Port{
		Protocol:   protocol,
		PortNumber: portNumber,
		Parent:     NewGraphModelWrapper(parent),
	}
	p.Defaulted()
	registry.CallHooks(&p)
	return p
}

// GetDescription returns a description for the Port model.
func (p *Port) GetDescription() string {
	return "Represents an open port on an asset with protocol, port number, and optional service information."
}

func PortConditions(port Port) []Condition {
	portStr := strconv.Itoa(port.PortNumber)
	return []Condition{
		NewCondition("port", ""),
		NewCondition("port", portStr),
	}
}
