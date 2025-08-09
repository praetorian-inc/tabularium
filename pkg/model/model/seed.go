package model

import (
	"fmt"
	"net"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
	"golang.org/x/net/publicsuffix"
)

func init() {
	registry.Registry.MustRegisterModel(&Seed{})
}

// GetDescription returns a description for the Seed model.
func (s *Seed) GetDescription() string {
	return "Represents a seed input for discovery, such as a domain, IP range, or ASN."
}

type Seed struct {
	registry.BaseModel
	Username  string  `neo4j:"username" json:"username" desc:"Chariot username associated with the seed." example:"user@example.com"`
	Key       string  `neo4j:"key" json:"key" desc:"Unique key identifying the seed." example:"#seed#domain#example.com"`
	DNS       string  `neo4j:"dns" json:"dns" desc:"The DNS name or IP address of the seed." example:"example.com"`
	Status    string  `neo4j:"status" json:"status" desc:"Composite of the current status of the seed and the seed type." example:"domain#P"`
	Source    string  `neo4j:"source" json:"source,omitempty" desc:"Source from which the seed was obtained." example:"ns1"`
	Name      *string `neo4j:"name,omitempty" json:"name,omitempty" desc:"Optional name associated with the seed (e.g., company name)." example:"Example Corp"`
	Location  *string `neo4j:"location,omitempty" json:"location,omitempty" desc:"Optional location associated with the seed." example:"Headquarters"`
	Email     *string `neo4j:"email,omitempty" json:"email,omitempty" desc:"Optional contact email associated with the seed." example:"contact@example.com"`
	Registrar *string `neo4j:"registrar,omitempty" json:"registrar,omitempty" desc:"Optional registrar information for the seed (if domain)." example:"MarkMonitor Inc."`
	Created   string  `neo4j:"created" json:"created" desc:"Timestamp when the seed was created (RFC3339)." example:"2023-10-27T09:00:00Z"`
	Visited   string  `neo4j:"visited" json:"visited" desc:"Timestamp when the seed was last processed or visited (RFC3339)." example:"2023-10-27T09:30:00Z"`
	Comment   string  `neo4j:"-" json:"comment,omitempty" desc:"User-provided comment about the seed." example:"Initial customer seed"`
	Class     string  `neo4j:"class" json:"class" desc:"Classification of the seed type (e.g., domain, tld, ip, cidr)." example:"domain"`
	Type      string  `neo4j:"type" json:"type" desc:"Broader type category (e.g., domain, ip)." example:"domain"`
	History
}

var SeedLabel = NewLabel("Seed")

func (s *Seed) GetLabels() []string {
	return []string{SeedLabel}
}

func (s *Seed) GetKey() string {
	return s.Key
}

func (s *Seed) GetClass() string {
	ip := net.ParseIP(s.DNS)
	if ip != nil {
		return "ip"
	}

	cidr, net, err := net.ParseCIDR(s.DNS)
	if cidr != nil && net != nil && err == nil {
		return "cidr"
	}

	tld, err := publicsuffix.EffectiveTLDPlusOne(s.DNS)
	if tld == s.DNS && err == nil {
		return "tld"
	}
	return "domain"
}

func (s *Seed) GetType() string {
	switch s.GetClass() {
	case "ip":
		fallthrough
	case "cidr":
		return "ip"
	case "tld":
		fallthrough
	case "domain":
		fallthrough
	default:
		return "domain"
	}
}

func (s *Seed) Merge(update Seed) {
	if s.History.Update(s.GetStatus(), update.GetStatus(), update.Source, update.Comment, update.History) {
		s.SetStatus(update.GetStatus())
	}
}

func (s *Seed) Visit(update Seed) {
	if update.Name != nil {
		s.Name = update.Name
	}
	if update.Location != nil {
		s.Location = update.Location
	}
	if update.Email != nil {
		s.Email = update.Email
	}
	if update.Registrar != nil {
		s.Registrar = update.Registrar
	}
	s.Visited = Now()
}

func (s *Seed) Asset() Asset {
	a := NewAsset(s.DNS, s.DNS)
	a.SetSource(SeedSource)
	a.SetStatus(s.GetStatus())
	a.TTL = 0

	a.Class = a.GetClass()
	a.Private = a.IsPrivate()
	return a
}

func (s *Seed) SetStatus(status string) {
	s.Status = fmt.Sprintf("%s#%s", s.GetType(), status)
}

func (s *Seed) GetStatus() string {
	return strings.TrimPrefix(s.Status, fmt.Sprintf("%s#", s.GetType()))
}

func (s *Seed) DomainVerificationJob(parentJob *Job, config ...string) Job {
	a := s.Asset()
	job := Job{
		Source:  "whois",
		Target:  TargetWrapper{Model: &a},
		Status:  fmt.Sprintf("%s#%s", Queued, "whois"),
		Config:  make(map[string]string),
		Created: Now(),
		Updated: Now(),
		TTL:     Future(12),
		Queue:   Standard,
		Parent:  parentJob.Target,
		Full:    true,
	}

	if job.Target.Model != nil {
		template := fmt.Sprintf("#job#%%s#%s#%s", job.Target.Model.Identifier(), job.Source)
		if len(template) <= 1024 {
			shortenedDNS := job.Target.Model.Group()[:min(1024-len(template), len(job.Target.Model.Group()))]
			job.DNS = shortenedDNS
			job.Key = fmt.Sprintf(template, shortenedDNS)
		}
	}

	job.Config["source"] = parentJob.Source
	for i := 0; i < len(config); i += 2 {
		job.Config[config[i]] = config[i+1]
	}
	return job
}

func (s *Seed) Attribute(name, value string) Attribute {
	return NewAttribute(name, value, s)
}

func (s *Seed) Is(status string) bool {
	return strings.HasPrefix(s.GetStatus(), status)
}

func (s *Seed) Valid() bool {
	if s.GetClass() == "domain" {
		return domain.MatchString(s.DNS)
	}
	return true
}

func (s *Seed) Defaulted() {
	s.Created = Now()
	s.Visited = Now()
}

func (s *Seed) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				s.Class = s.GetClass()
				s.Type = s.GetType()
				s.SetStatus(Pending)
				s.Key = fmt.Sprintf("#seed#%s#%s", s.GetType(), s.DNS)
				return nil
			},
		},
	}
}

func NewSeed(dns string) Seed {
	s := Seed{
		DNS: strings.ToLower(dns),
	}
	s.Defaulted()
	registry.CallHooks(&s)
	return s
}
