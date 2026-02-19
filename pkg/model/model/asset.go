package model

import (
	"fmt"
	"golang.org/x/net/idna"
	"net"
	"regexp"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
	"golang.org/x/net/publicsuffix"
)

type Asset struct {
	BaseAsset
	LabelSettableEmbed
	// Attributes
	DNS     string `neo4j:"dns" json:"dns" desc:"The DNS name, or group identifier associated with this asset." example:"example.com" capmodel:"Asset,IP=ip,Domain=domain"`
	Name    string `neo4j:"name" json:"name" desc:"Name of the asset, or the same value as DNS if this asset represents the group." example:"169.254.169.254" capmodel:"Asset,IP=ip,Domain=domain"`
	Private bool   `neo4j:"private" json:"private" desc:"Flag indicating if the asset is considered private (e.g., internal IP)." example:"false"`
}

func init() {
	registry.Registry.MustRegisterModel(&Asset{})
}

var (
	aws      = regexp.MustCompile(`^arn:aws:`)
	s3       = regexp.MustCompile(`^(s3://)([^/]+)/?$`)
	domain   = regexp.MustCompile(`^(https?://)?((xn--[a-zA-Z0-9-]+|[a-zA-Z0-9-]+)\.)+([a-zA-Z]{2,})$`)
	assetKey = regexp.MustCompile(`^#asset(#[^#]+){2,}$`)
)

const AssetLabel = "Asset"

func (a *Asset) GetLabels() []string {
	labels := []string{AssetLabel, TTLLabel}
	if a.Source == SeedSource {
		labels = append(labels, SeedLabel)
	}
	return labels
}

func (a *Asset) GetClass() string {
	name := []func(string) (string, bool){
		func(s string) (string, bool) {
			return "aws", aws.MatchString(s)
		},
		func(s string) (string, bool) {
			return "s3", s3.FindStringSubmatch(s) != nil
		},
		func(s string) (string, bool) {
			ip := net.ParseIP(s)
			if ip == nil {
				return "", false
			}
			if ip4 := ip.To4(); ip4 != nil {
				return "ipv4", true
			}
			return "ipv6", true
		},
	}

	dns := []func(string) (string, bool){
		func(s string) (string, bool) {
			_, _, err := net.ParseCIDR(s)
			return "cidr", err == nil
		},
		func(s string) (string, bool) {
			if !domain.MatchString(s) {
				return "", false
			}

			tld, icann := publicsuffix.PublicSuffix(s)
			parts := strings.Split(strings.TrimSuffix(s, tld), ".")
			return "tld", len(parts) == 2 && icann && !strings.Contains(s, "/")
		},
		func(s string) (string, bool) {
			return "domain", domain.MatchString(s)
		},
	}

	if a.Source == AccountSource {
		return a.DNS
	}

	for _, class := range name {
		if c, ok := class(a.Name); ok {
			return c
		}
	}

	for _, class := range dns {
		if c, ok := class(a.DNS); ok {
			return c
		}
	}

	return ""
}

func (a *Asset) IsPrivate() bool {
	if a.IsClass("ipv") {
		return net.ParseIP(a.Name).IsPrivate()
	}

	if a.IsClass("cidr") {
		_, cidr, err := net.ParseCIDR(a.Name)
		if err != nil {
			return false
		}
		return cidr.IP.IsPrivate()
	}
	return false
}

func (a *Asset) Valid() bool {
	return assetKey.MatchString(a.Key)
}

func (a *Asset) GetPartitionKey() string {
	return a.Name
}

func (a *Asset) Merge(o Assetlike) {
	other, ok := o.(*Asset)
	if !ok {
		return
	}
	MergeWithPromotionCheck(&a.BaseAsset, &a.LabelSettableEmbed, other)
}

func (a *Asset) Visit(o Assetlike) {
	other, ok := o.(*Asset)
	if !ok {
		return
	}
	if IsSeedPromotion(&a.BaseAsset, &other.BaseAsset) {
		ApplySeedLabels(&a.BaseAsset, &a.LabelSettableEmbed)
	}
	a.BaseAsset.Visit(other)
	// allow asset enrichments to control asset privateness
	a.Private = other.Private
}

func (a *Asset) Spawn(dns, name string) Asset {
	asset := NewAsset(dns, name)
	asset.Status = a.Status
	return asset
}

func (a *Asset) SeedModels() []Seedable {
	copy := *a
	return []Seedable{&copy}
}

func (a *Asset) DomainVerificationJob(parentJob *Job, config ...string) Job {
	isDomain := a.Class == "domain" || a.Class == "tld"
	isSeed := a.Source == SeedSource
	if !isDomain || !isSeed {
		return Job{}
	}

	copy := *a
	job := Job{
		Target:  TargetWrapper{Model: &copy},
		Config:  make(map[string]string),
		Created: Now(),
		Updated: Now(),
		TTL:     Future(12),
		Queue:   Standard,
		Parent:  parentJob.Target,
		Full:    true,
	}
	job.SetStatus(Queued)
	job.SetCapability("whois")
	registry.CallHooks(&job)

	if job.Target.Model != nil {
		template := fmt.Sprintf("#job#%%s#%s#%s", job.Target.Model.Identifier(), job.GetCapability())
		if len(template) <= 1024 {
			shortenedDNS := job.Target.Model.Group()[:min(1024-len(template), len(job.Target.Model.Group()))]
			job.DNS = shortenedDNS
			job.Key = fmt.Sprintf(template, shortenedDNS)
		}
	}

	job.Config["source"] = parentJob.GetCapability()
	for i := 0; i < len(config); i += 2 {
		job.Config[config[i]] = config[i+1]
	}
	return job
}

func (a *Asset) WithStatus(status string) Target {
	ret := *a
	ret.Status = status
	return &ret
}

func (a *Asset) Group() string {
	return a.DNS
}

func (a *Asset) Identifier() string {
	return a.Name
}

func (a *Asset) SetSource(source string) {
	a.BaseAsset.SetSource(source)
	a.Class = a.GetClass()
}

func (a *Asset) Attribute(name, value string) Attribute {
	return NewAttribute(name, value, a)
}

func (a *Asset) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Description: "normalize unicode characters with punycode",
			Call: func() error {
				var err error
				a.DNS, err = idna.ToASCII(a.DNS)
				if err != nil {
					return err
				}

				a.Name, err = idna.ToASCII(a.Name)
				if err != nil {
					return err
				}

				return nil
			},
		},
		useGroupAndIdentifier(a, &a.DNS, &a.Name),
		{
			Call: func() error {
				a.Key = strings.ToLower(fmt.Sprintf("#asset#%s#%s", a.DNS, a.Name))
				a.Class = a.GetClass()
				a.Private = a.IsPrivate()
				if a.Private && (a.IsClass("ip") || a.IsClass("cidr")) {
					a.ASName = "Non-Routable"
					a.ASNumber = "0"
				}
				return nil
			},
		},
		setGroupAndIdentifier(a, &a.DNS, &a.Name),
	}
}

func NewAsset(dns, name string) Asset {
	a := Asset{
		DNS:  dns,
		Name: name,
	}

	a.Defaulted()
	registry.CallHooks(&a)

	return a
}

func NewAssetSeed(name string) Asset {
	a := NewAsset(name, name)
	a.Source = SeedSource
	a.Status = Pending
	a.TTL = 0
	return a
}

// GetDescription returns a description for the Asset model.
func (a *Asset) GetDescription() string {
	return "Represents a discoverable entity within an infrastructure, such as a host, service, or application."
}
