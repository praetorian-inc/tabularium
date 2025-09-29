package model

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func init() {
	registry.Registry.MustRegisterModel(&Technology{})
}

var technologyKeyRegex = regexp.MustCompile(`^#technology#cpe:2\.3(:[^:]+){11}$`)

// This is so we can have CPE as a field as well
type AliasCPE = CPE
type Technology struct {
	registry.BaseModel
	Username string `neo4j:"username" json:"username" desc:"Chariot username associated with the technology record." example:"user@example.com"`
	Key      string `neo4j:"key" json:"key" desc:"Unique key identifying the technology (based on CPE string)." example:"#technology#cpe:2.3:a:apache:http_server:2.4.50:*:*:*:*:*:*:*"`
	// Attributes
	Source string `neo4j:"-" json:"source,omitempty" desc:"Source that identified this technology." example:"wappalyzer"`
	AliasCPE
	CPE     string `neo4j:"cpe" json:"cpe" desc:"The full CPE string representation." example:"cpe:2.3:a:apache:http_server:2.4.50:*:*:*:*:*:*:*"`
	Name    string `neo4j:"name" json:"name,omitempty" desc:"Optional common name for the technology." example:"Apache httpd"`
	TTL     int64  `neo4j:"ttl" json:"ttl" desc:"Time-to-live for the technology record (Unix timestamp)." example:"1706353200"`
	Comment string `neo4j:"-" json:"comment,omitempty" desc:"User-provided comment about the technology record." example:"Found on main web server"`
	Created string `neo4j:"created" json:"created" desc:"Timestamp when the technology record was created (RFC3339)." example:"2023-10-27T10:00:00Z"`
	Visited string `neo4j:"visited" json:"visited" desc:"Timestamp when the technology was last visited or confirmed (RFC3339)." example:"2023-10-27T11:00:00Z"`
	History
	Tags
}

const TechnologyLabel = "Technology"

func (t *Technology) GetKey() string {
	return t.Key
}

func (t *Technology) GetLabels() []string {
	return []string{TechnologyLabel, TTLLabel}
}

func (t *Technology) Valid() bool {
	return technologyKeyRegex.MatchString(t.Key)
}

// No CPE fields are upgraded on a premise that visit is only
// used in exact searches
func (t *Technology) Visit(other Technology) {
	if t.TTL != 0 {
		t.TTL = other.TTL
	}
	if t.Name == "" {
		t.Name = other.Name
	}
	t.Visited = other.Visited
}

func (t *Technology) Merge(update Technology) {
	t.History.Update("", "", update.Source, update.Comment, update.History)
	t.Tags.Merge(update.Tags)
}

func (t *Technology) Proof(bits []byte, asset *Asset, transportProtocol, port string) File {
	file := NewFile(fmt.Sprintf("proofs/%s/%s/%s/%s/%s", t.CPE, asset.DNS, asset.Name, transportProtocol, port))
	file.Bytes = bits
	return file
}

func (t *Technology) Attribute(name, value string) Attribute {
	attr := NewAttribute(name, value, t)
	return attr
}

func NewTechnology(cpe string) (Technology, error) {
	parsed, err := NewCPE(strings.TrimSpace(cpe))
	if err == nil {
		return NewTechnologyWithCPE(parsed), nil
	}

	parsed, err = NewCPEFromURI(strings.TrimSpace(cpe))
	if err == nil {
		return NewTechnologyWithCPE(parsed), nil
	}

	return Technology{}, fmt.Errorf("could not parse CPE: %s", cpe)
}

func (t *Technology) Defaulted() {
	t.TTL = Future(14 * 24)
	t.Created = Now()
	t.Visited = Now()
}

func (t *Technology) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				t.Key = fmt.Sprintf("#technology#%s", t.CPE)
				return nil
			},
		},
	}
}

func NewTechnologyWithCPE(cpe CPE) Technology {
	t := Technology{
		AliasCPE: cpe,
		CPE:      cpe.String(),
	}
	t.Defaulted()
	registry.CallHooks(&t)
	return t
}

// GetDescription returns a description for the Technology model.
func (t *Technology) GetDescription() string {
	return "Represents a specific technology (e.g., software, library, framework) identified on an asset."
}
