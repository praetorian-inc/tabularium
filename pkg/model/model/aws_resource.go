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
			ResourceType: CloudResourceType(rtype),
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
}

func (a *AWSResource) GetHooks() []registry.Hook {
	hooks := []registry.Hook{
		{
			Call: func() error {
				a.CloudResource.Key = fmt.Sprintf("#awsresource#%s#%s", a.AccountRef, a.Name)
				a.CloudResource.Labels = []string{AWSResourceLabel}
				a.CloudResource.IPs = a.GetIPs()
				a.CloudResource.URLs = a.GetURLs()
				return nil
			},
		},
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

func (a *AWSResource) Group() string { return "awsresource" }

func (a *AWSResource) Merge(otherModel any) {
	other, ok := otherModel.(*AWSResource)
	if !ok {
		return
	}
	a.CloudResource.Merge(&other.CloudResource)
}

func (a *AWSResource) Visit(otherModel any) error {
	other, ok := otherModel.(*AWSResource)
	if !ok {
		return fmt.Errorf("expected *AWSResource, got %T", otherModel)
	}
	a.CloudResource.Visit(&other.CloudResource)
	return nil
}

// Return an Asset that matches the legacy integration
func (a *AWSResource) NewAsset() []Asset {
	assets := make([]Asset, 0)
	dns := a.GetDNS()
	ips := a.GetIPs()
	urls := a.GetURLs()

	// Extract service name from ARN (same logic as Amazon capability)
	service := a.extractService()

	// Create assets from URLs - NewAsset(url, arn)
	for _, url := range urls {
		asset := NewAsset(url, a.Name)
		asset.CloudId = a.Name
		asset.CloudService = service
		asset.CloudAccount = a.AccountRef
		assets = append(assets, asset)
	}

	// Create assets from IPs
	for _, ip := range ips {
		var asset Asset
		if dns != "" {
			// NewAsset(dns, dns) when both DNS and IP exist - DNS takes precedence as identifier
			asset = NewAsset(dns, dns)
		} else {
			// NewAsset(ip, ip) when only IP exists
			asset = NewAsset(ip, ip)
		}
		asset.CloudId = a.Name
		asset.CloudService = service
		asset.CloudAccount = a.AccountRef
		assets = append(assets, asset)
	}

	// If no URLs or IPs, create a fallback asset using the ARN
	if len(assets) == 0 {
		identifier := a.Name // Use ARN as fallback
		if dns != "" {
			identifier = dns
		}
		asset := NewAsset(identifier, identifier)
		asset.CloudId = a.Name
		asset.CloudService = service
		asset.CloudAccount = a.AccountRef
		assets = append(assets, asset)
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
