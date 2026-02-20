package model

import (
	"fmt"
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

	OrgPolicyFilename   string `neo4j:"orgPolicyFilename" json:"orgPolicyFilename"`
	OrgPolicy           []byte `neo4j:"-" json:"orgPolicy"`
	IsManagementAccount bool   `neo4j:"isManagementAccount" json:"isManagementAccount"`
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

				a.IsManagementAccount = a.isManagementAccount()
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

func (c *AWSResource) SetOrgPolicy(policy []byte) {
	c.OrgPolicy = policy
	c.OrgPolicyFilename = c.BuildOrgPolicyFilename()
}

func (c *AWSResource) GetOrgPolicy() []byte {
	return c.OrgPolicy
}

func (c *AWSResource) HydratableFilepath() string {
	return c.OrgPolicyFilename
}

func (c *AWSResource) Hydrate(data []byte) error {
	c.SetOrgPolicy(data)
	return nil
}

func (c *AWSResource) HydratedFile() File {
	if c.OrgPolicy == nil {
		return File{}
	}

	file := NewFile(c.BuildOrgPolicyFilename())
	file.Bytes = c.OrgPolicy

	c.OrgPolicyFilename = file.Name
	return file
}

func (c *AWSResource) Dehydrate() Hydratable {
	dehydrated := *c

	if dehydrated.OrgPolicy != nil {
		dehydrated.OrgPolicyFilename = c.BuildOrgPolicyFilename()
	}

	dehydrated.OrgPolicy = nil
	return &dehydrated
}

func (a *AWSResource) BuildOrgPolicyFilename() string {
	return fmt.Sprintf("awsresource/%s/%s/org-policies.json", a.AccountRef, RemoveReservedCharacters(a.Identifier()))
}

func (a *AWSResource) Visit(other Assetlike) {
	otherResource, ok := other.(*AWSResource)
	if !ok {
		return
	}

	if otherResource.OrgPolicyFilename != "" {
		a.OrgPolicyFilename = otherResource.OrgPolicyFilename
	}

	a.CloudResource.Visit(&otherResource.CloudResource)
	a.BaseAsset.Visit(otherResource)
}

func (a *AWSResource) Merge(other Assetlike) {
	otherResource, ok := other.(*AWSResource)
	if !ok {
		return
	}

	if otherResource.OrgPolicyFilename != "" {
		a.OrgPolicyFilename = otherResource.OrgPolicyFilename
	}

	a.CloudResource.Merge(&otherResource.CloudResource)
	a.BaseAsset.Merge(otherResource)
}

func (a *AWSResource) isManagementAccount() bool {
	if a.ResourceType != AWSOrganization {
		return false
	}

	// management account IDs are listed in the middle of Organization ARNs, and the actual account's ID is listed at the end
	// If they match, that means the actual account IS the management account
	// e.g., management account ARN: arn:aws:organizations::123456789012:account/o-b5qlad4a9o/123456789012
	//   non-management account ARN: arn:aws:organizations::123456789012:account/o-b5qlad4a9o/098765432109
	parts := strings.Split(a.Identifier(), ":")
	if len(parts) < 5 {
		return false
	}

	managementAccountID := parts[4]
	isManagementAccount := strings.HasSuffix(a.Identifier(), "/"+managementAccountID)

	return isManagementAccount
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

func (a *AWSResource) IsPrivate() bool {
	return false
}
