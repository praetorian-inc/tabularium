package model

import (
	"fmt"
	"maps"
	"net"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

type AWSResource struct {
	CloudResource
}

// NewAWSResource creates an AWSResource from the given ARN, resource type, account reference, and properties.
func NewAWSResource(name, accountRef string, rtype CloudResourceType, properties map[string]any) (AWSResource, error) {
	parsedARN, err := arn.Parse(name)
	if err != nil {
		return AWSResource{}, fmt.Errorf("invalid ARN: %s", name)
	}

	key := fmt.Sprintf("#awsresource#%s#%s", accountRef, name)

	r := AWSResource{
		CloudResource: CloudResource{
			Key:          key,
			Name:         name,
			DisplayName:  parsedARN.Resource,
			Provider:     "aws",
			Properties:   properties,
			ResourceType: CloudResourceType(rtype),
			Region:       parsedARN.Region,
			AccountRef:   accountRef,
			Labels:       []string{"AWSResource"},
		},
	}

	r.Defaulted()
	registry.CallHooks(&r)
	return r, nil
}

func (a *AWSResource) GetHooks() []registry.Hook {
	return a.CloudResource.GetHooks()
}

// WithStatus method for AWSResource to prevent type erasure
// Overrides the embedded CloudResource.WithStatus() to return proper type
func (a *AWSResource) WithStatus(status string) Target {
	ret := *a // Copy the full AWSResource, not just CloudResource
	ret.Status = status
	return &ret
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

func (a *AWSResource) GetURL() string {
	return ""
}

func (a *AWSResource) GetDNS() string {
	if dns, ok := a.Properties["PublicDnsName"].(string); ok {
		return dns
	}
	return ""
}

func (a *AWSResource) Group() string { return "awsresource" }

// Insertable interface methods
func (a *AWSResource) Merge(otherModel any) {
	other, ok := otherModel.(*AWSResource)
	if !ok {
		return
	}
	a.Status = other.Status
	a.Visited = other.Visited

	// Safely copy properties with nil checks
	if a.Properties == nil {
		a.Properties = make(map[string]any)
	}
	if other.Properties != nil {
		maps.Copy(a.Properties, other.Properties)
	}
}

func (a *AWSResource) Visit(otherModel any) error {
	other, ok := otherModel.(*AWSResource)
	if !ok {
		return fmt.Errorf("expected *AWSResource, got %T", otherModel)
	}
	a.Visited = other.Visited
	a.Status = other.Status

	// Safely copy properties with nil checks
	if a.Properties == nil {
		a.Properties = make(map[string]any)
	}
	if other.Properties != nil {
		maps.Copy(a.Properties, other.Properties)
	}

	// Fix TTL update logic: update if other has a valid TTL
	if other.TTL != 0 {
		a.TTL = other.TTL
	}
	return nil
}

// Return an Asset that matches the legacy integration
func (a *AWSResource) NewAsset() Asset {
	dns := a.GetDNS()
	var ip string
	if ips := a.GetIPs(); len(ips) > 0 {
		ip = ips[0] // Use first IP for legacy Asset compatibility
	}

	// Determine the identifier for the Asset
	// Priority: DNS > IP > ARN (fallback for resources without DNS/IP)
	identifier := ""
	if dns != "" {
		identifier = dns
	} else if ip != "" {
		identifier = ip
	} else {
		// Fallback to ARN for resources without DNS/IP (S3, Lambda, etc.)
		identifier = a.Name
	}

	// Correctly extract AWS service from ARN
	// ARN format: arn:partition:service:region:account:resource
	service := "Unknown Service"

	parts := strings.Split(a.Name, ":")
	if len(parts) > 2 {
		service = parts[2] // Index 2 is the service (e.g., "lambda", "s3", "ec2")
	}

	// Create Asset with proper DNS and Name to ensure valid key
	// When DNS is empty, use identifier for both to avoid "#asset##name" pattern
	assetDNS := dns
	assetName := identifier
	if dns == "" {
		// Use identifier for DNS as well to create valid key "#asset#identifier#identifier"
		assetDNS = identifier
	}

	asset := NewAsset(assetDNS, assetName)
	asset.CloudId = a.Name
	asset.CloudService = service
	asset.CloudAccount = a.AccountRef

	return asset
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
	if url := a.GetURL(); url != "" {
		return false // Has public URL = not private
	}

	// No public IPs or URL = assume private
	return true
}
