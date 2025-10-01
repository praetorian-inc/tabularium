package model

import (
	"fmt"
	"net"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

const (
	AWSProvider      = "aws"
	AWSResourceLabel = "AWSResource"
)

type AWSResource struct {
	CloudResource
	OrgPolicy []byte `neo4j:"-" json:"orgPolicy"`
}

func init() {
	MustRegisterLabel(AWSResourceLabel)
	registry.Registry.MustRegisterModel(&AWSResource{})
}

// NewAWSResource creates an AWSResource from the given ARN, resource type, account reference, and properties.
func NewAWSResource(name, accountRef string, rtype CloudResourceType, properties map[string]any) (AWSResource, error) {
	parsedARN, err := arn.Parse(name)
	if err != nil {
		return AWSResource{}, fmt.Errorf("invalid ARN: %s", name)
	}

	r := AWSResource{
		CloudResource: CloudResource{
			Name:         name,
			DisplayName:  parsedARN.Resource,
			Provider:     AWSProvider,
			Properties:   properties,
			ResourceType: rtype,
			Region:       parsedARN.Region,
			AccountRef:   accountRef,
		},
	}

	r.Defaulted()
	registry.CallHooks(&r)
	return r, nil
}

func (a *AWSResource) Defaulted() {
	a.Origins = []string{"amazon"}
	a.AttackSurface = []string{"cloud"}
	a.CloudResource.Defaulted()
	a.BaseAsset.Defaulted()
}

func (a *AWSResource) GetHooks() []registry.Hook {
	hooks := []registry.Hook{
		useGroupAndIdentifier(a, &a.AccountRef, &a.Name),
		{
			Call: func() error {
				a.CloudResource.Key = fmt.Sprintf("#awsresource#%s#%s", a.AccountRef, a.Name)
				a.CloudResource.Labels = []string{AWSResourceLabel}
				a.CloudResource.IPs = a.GetIPs()
				a.CloudResource.URLs = a.GetURLs()
				return nil
			},
		},
		setGroupAndIdentifier(a, &a.AccountRef, &a.Name),
	}

	hooks = append(hooks, a.CloudResource.GetHooks()...)
	return hooks
}

// WithStatus method for AWSResource to prevent type erasure
// Overrides the embedded CloudResource.WithStatus() to return proper type
func (a *AWSResource) WithStatus(status string) Target {
	ret := *a // Copy the full AWSResource, not just CloudResource
	ret.Status = status
	return &ret
}

func (c *AWSResource) HydratableFilepath() string {
	if c.OrgPolicy == nil {
		return ""
	}
	return c.GetOrgPolicyFilename()
}

func (a *AWSResource) GetOrgPolicyFilename() string {
	return fmt.Sprintf("awsresource/%s/%s/org-policies.json", a.AccountRef, RemoveReservedCharacters(a.Identifier()))
}

func (c *AWSResource) Hydrate(data []byte) error {
	c.OrgPolicy = data
	return nil
}

func (c *AWSResource) HydratedFile() File {
	filepath := c.HydratableFilepath()
	if filepath == "" {
		return File{}
	}

	file := NewFile(filepath)
	file.Bytes = c.OrgPolicy
	return file
}

func (c *AWSResource) Dehydrate() Hydratable {
	dehydrated := *c
	dehydrated.OrgPolicy = nil
	return &dehydrated
}

func (a *AWSResource) Visit(other Assetlike) {
	otherResource, ok := other.(*AWSResource)
	if !ok {
		return
	}

	a.CloudResource.Visit(&otherResource.CloudResource)
	a.BaseAsset.Visit(otherResource)
}

func (a *AWSResource) Merge(other Assetlike) {
	otherResource, ok := other.(*AWSResource)
	if !ok {
		return
	}

	a.CloudResource.Merge(&otherResource.CloudResource)
	a.BaseAsset.Merge(otherResource)
}

func (a *AWSResource) GetIPs() []string {
	ips := make([]string, 0) // Initialize with empty slice instead of nil
	switch a.ResourceType {
	case AWSEC2Instance:
		if ip, ok := a.Properties["PublicIp"].(string); ok && ip != "" {
			ips = append(ips, ip)
		}
		// Also check for PrivateIp in case it's needed for validation
		if ip, ok := a.Properties["PrivateIp"].(string); ok && ip != "" {
			ips = append(ips, ip)
		}
	}
	return ips
}

func (a *AWSResource) GetURLs() []string {
	return []string{}
}

func (a *AWSResource) GetDNS() string {
	if dns, ok := a.Properties["PublicDnsName"].(string); ok {
		return dns
	}
	return ""
}

func (a *AWSResource) Group() string {
	return a.AccountRef
}

func (a *AWSResource) Identifier() string {
	return a.Name
}

// Return an Asset that matches the legacy integration
func (a *AWSResource) NewAssets() []Asset {
	assets := make([]Asset, 0)
	dns := a.GetDNS()
	ips := a.GetIPs()
	urls := a.GetURLs()

	// Extract service name from ARN (same logic as Amazon capability)
	service := a.extractService()

	record := func(asset Asset) {
		asset.CloudId = a.Name
		asset.CloudService = service
		asset.CloudAccount = a.AccountRef
		assets = append(assets, asset)
	}

	// Create assets from URLs - NewAsset(url, arn)
	for _, url := range urls {
		record(NewAsset(url, a.Name))
	}

	for _, ip := range ips {
		if dns != "" {
			record(NewAsset(dns, ip))
		}
		record(NewAsset(ip, ip))
	}

	if len(assets) == 0 {
		identifier := a.Name // Use ARN as fallback
		if dns != "" {
			identifier = dns
		}
		record(NewAsset(identifier, identifier))
	}

	return assets
}

// extractService extracts the service name from the ARN (same logic as Amazon capability)
func (a *AWSResource) extractService() string {
	parts := strings.Split(a.Name, ":")
	if len(parts) > 2 {
		return parts[2]
	}
	return "Unknown Service"
}

// IsPrivate determines if this AWS resource is private based on its IP/URL
func (a *AWSResource) IsPrivate() bool {
	// Check if resource has any public IP addresses
	if ips := a.GetIPs(); len(ips) > 0 {
		for _, ip := range ips {
			if ip != "" {
				parsedIP := net.ParseIP(ip)
				if parsedIP != nil && !parsedIP.IsPrivate() {
					return false // Has at least one public IP = not private
				}
			}
		}
	}

	// Check if resource has a public URL/endpoint
	if urls := a.GetURLs(); len(urls) > 0 {
		return false // Has public URL = not private
	}

	// No public IPs or URL = assume private
	return true
}
